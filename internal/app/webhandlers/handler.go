package webhandlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/t1mon-ggg/go_shortner/internal/app/config"
	"github.com/t1mon-ggg/go_shortner/internal/app/helpers"
	"github.com/t1mon-ggg/go_shortner/internal/app/models"
	"github.com/t1mon-ggg/go_shortner/internal/app/mymiddlewares"
	"github.com/t1mon-ggg/go_shortner/internal/app/storage"
)

type app struct {
	Storage storage.Database
	Config  *config.Config
	DelBuf  chan models.DelWorker
}

//NewApp - функция для создания новой структуры для работы приложения
func NewApp() *app {
	s := app{}
	s.Config = config.NewConfig()
	s.DelBuf = make(chan models.DelWorker)
	return &s
}

func (application *app) Router(r chi.Router) {
	go application.Storage.Cleaner(application.DelBuf, 10)
	r.Get("/", defaultGetHandler)
	r.Get("/ping", application.ConnectionTest)
	r.Get("/{^[a-zA-Z]}", application.getHandler)
	r.Get("/api/user/urls", application.userURLs)
	r.Post("/", application.postHandler)
	r.Post("/api/shorten", application.postAPIHandler)
	r.Post("/api/shorten/batch", application.postAPIBatch)
	r.Delete("/api/user/urls", application.deleteTags)
	r.MethodNotAllowed(otherHandler)
}

func defaultGetHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Empty request", http.StatusBadRequest)
}

func otherHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Bad request", http.StatusBadRequest)
}

func (application *app) ConnectionTest(w http.ResponseWriter, r *http.Request) {
	err := application.Storage.Ping()
	if err != nil {
		log.Println(err)
		http.Error(w, "Storage connection failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})
}

func (application *app) userURLs(w http.ResponseWriter, r *http.Request) {
	type answer struct {
		Short    string `json:"short_url"`
		Original string `json:"original_url"`
	}
	cookie := idCookieValue(w, r)
	data, err := application.Storage.ReadByCookie(cookie)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if len(data.Short) == 0 {
		http.Error(w, "No Content", http.StatusNoContent)
		return
	}
	a := make([]answer, 0)
	for _, content := range data.Short {
		a = append(a, answer{Short: fmt.Sprintf("%s/%s", application.Config.BaseURL, content.Short), Original: content.Long})
	}
	d, err := json.Marshal(a)
	if err != nil {
		log.Println(err)
		http.Error(w, "Json Error", http.StatusInternalServerError)
		return
	}
	if len(a) == 0 {
		http.Error(w, "No Content", http.StatusNoContent)
		return
	}
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(d)
}

func (application *app) postHandler(w http.ResponseWriter, r *http.Request) {
	cookie := idCookieValue(w, r)
	entry := models.ClientData{}
	entry.Cookie = cookie
	entry.Key = ""
	entry.Short = make([]models.ShortData, 0)
	defer r.Body.Close()
	blongURL, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	slongURL := string(blongURL)
	log.Println("Request body:", slongURL)
	surl := helpers.RandStringRunes(8)
	entry.Short = append(entry.Short, models.ShortData{Short: surl, Long: slongURL})
	err = application.Storage.Write(entry)
	if err != nil {
		if err.Error() == "not unique url" {
			s, err := application.Storage.TagByURL(slongURL, cookie)
			if err != nil {
				log.Println(err)
				http.Error(w, "Storage error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(fmt.Sprintf("%s/%s", application.Config.BaseURL, s)))
			return
		}
		log.Println(err)
		http.Error(w, "Storage error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", application.Config.BaseURL, surl)))
}

func (application *app) postAPIHandler(w http.ResponseWriter, r *http.Request) {
	cookie := idCookieValue(w, r)
	entry := models.ClientData{}
	entry.Cookie = cookie
	entry.Key = ""
	entry.Short = make([]models.ShortData, 0)
	type sURL struct {
		ShortURL string `json:"result"`
	}
	type lURL struct {
		LongURL string `json:"url"`
	}
	defer r.Body.Close()
	ctype := r.Header.Get("Content-Type")
	if ctype != "application/json" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	longURL := lURL{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Body Error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Println("Request body:", string(body))
	err = json.Unmarshal(body, &longURL)
	if err != nil {
		log.Println("JSON Unmarshal error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	short := helpers.RandStringRunes(8)
	entry.Short = append(entry.Short, models.ShortData{Short: short, Long: longURL.LongURL})
	err = application.Storage.Write(entry)
	if err != nil {
		if err.Error() == "not unique url" {
			s, err := application.Storage.TagByURL(longURL.LongURL, cookie)
			if err != nil {
				log.Println(err)
				http.Error(w, "Storage error", http.StatusInternalServerError)
				return
			}
			jbody := sURL{ShortURL: fmt.Sprintf("%s/%s", (*application).Config.BaseURL, s)}
			abody, err := json.Marshal(jbody)
			if err != nil {
				log.Println("JSON Marshal error", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			w.Write(abody)
			return
		}
		log.Println(err)
		http.Error(w, "Storage error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	jbody := sURL{ShortURL: fmt.Sprintf("%s/%s", (*application).Config.BaseURL, short)}
	abody, err := json.Marshal(jbody)
	if err != nil {
		log.Println("JSON Marshal error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Write(abody)
}
func (application *app) postAPIBatch(w http.ResponseWriter, r *http.Request) {
	type input struct {
		Correlation string `json:"correlation_id"`
		Long        string `json:"original_url"`
	}
	type output struct {
		Correlation string `json:"correlation_id"`
		Short       string `json:"short_url"`
	}
	cookie := idCookieValue(w, r)
	out := make([]output, 0)
	defer r.Body.Close()
	ctype := r.Header.Get("Content-Type")
	if ctype != "application/json" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	in := []input{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Body Error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(body, &in)
	if err != nil {
		log.Println("JSON Unmarshal error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	for i := range in {
		entry := models.ClientData{}
		entry.Cookie = cookie
		entry.Key = ""
		entry.Short = make([]models.ShortData, 0)
		short := helpers.RandStringRunes(8)
		entry.Short = append(entry.Short, models.ShortData{Short: short, Long: in[i].Long})
		err := application.Storage.Write(entry)
		if err != nil {
			if err.Error() == "not unique url" {
				s, err := application.Storage.TagByURL(in[i].Long, cookie)
				if err != nil {
					http.Error(w, "Storage error", http.StatusInternalServerError)
					return
				}
				out = append(out, output{Correlation: in[i].Correlation, Short: fmt.Sprintf("%s/%s", (*application).Config.BaseURL, s)})
			} else {
				log.Println(err)
				http.Error(w, "Storage error", http.StatusInternalServerError)
				return
			}
		} else {
			out = append(out, output{Correlation: in[i].Correlation, Short: fmt.Sprintf("%s/%s", (*application).Config.BaseURL, short)})
		}
	}
	batch, err := json.Marshal(out)
	if err != nil {
		log.Println(err)
		http.Error(w, "Answer error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(batch)
}

func (application *app) getHandler(w http.ResponseWriter, r *http.Request) {
	p := r.RequestURI
	p = p[1:]
	matched, err := regexp.Match(`\w{8}`, []byte(p))
	if err != nil {
		log.Println(err)
		http.Error(w, "URI process error", http.StatusInternalServerError)
		return
	}
	if !matched || len(p) > 8 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	data, err := application.Storage.ReadByTag(p)
	if err != nil {
		log.Println(err)
		http.Error(w, "DB read error", http.StatusInternalServerError)
		return
	}
	nilShort := models.ShortData{}
	if data == nilShort {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	if data.Deleted {
		w.WriteHeader(http.StatusGone)
		w.Write([]byte{})
		return
	}
	w.Header().Set("Location", data.Long)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte{})

}

func (application *app) deleteTags(w http.ResponseWriter, r *http.Request) {
	cookie := idCookieValue(w, r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte{})
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	re := regexp.MustCompile(`\w+`)
	tags := re.FindAllString(string(body), -1)
	task := models.DelWorker{Cookie: cookie, Tags: tags}
	application.DelBuf <- task
}

func (application *app) Middlewares(r *chi.Mux) {
	r.Use(middleware.Compress(5))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(mymiddlewares.DecompressRequest)
	r.Use(mymiddlewares.TimeTracer)
	r.Use(application.Cookies)
}

func idCookieValue(w http.ResponseWriter, r *http.Request) string {
	var newcookieval string
	var value string
	cookies := r.Cookies()
	if len(cookies) == 0 {
		re := regexp.MustCompile(`\w{96}`)
		cid := re.FindString(w.Header().Get("Set-Cookie"))
		if len(cid) != 0 {
			newcookieval = cid
			if len(newcookieval) == 96 {
				value = newcookieval[:32]

				return value
			}
		}
		return ""
	} else {
		for _, cookie := range cookies {
			if cookie.Name == "Client_ID" {
				if len(cookie.Value) == 96 {
					value = cookie.Value[:32]
					return value
				}
			}
		}
	}
	return ""
}

//Cookies - cookie processor
func (application *app) Cookies(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookies := r.Cookies()
		if len(cookies) != 0 {
			found := false
			for _, cookie := range cookies {
				if cookie.Name == "Client_ID" {
					if !application.checkCookie(cookie) {
						value := helpers.RandStringRunes(32)
						key := helpers.RandStringRunes(64)
						application.addCookie(w, "Client_ID", value, key)
					} else {
						found = true
					}
				}
			}
			if !found {
				value := helpers.RandStringRunes(32)
				key := helpers.RandStringRunes(64)
				application.addCookie(w, "Client_ID", value, key)
			}
		} else {
			value := helpers.RandStringRunes(32)
			key := helpers.RandStringRunes(64)
			application.addCookie(w, "Client_ID", value, key)
		}
		next.ServeHTTP(w, r)
	})
}

func (application *app) addCookie(w http.ResponseWriter, name, value string, key string) {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(value))
	signed := h.Sum(nil)
	sign := hex.EncodeToString(signed)
	cookie := http.Cookie{
		Name:   name,
		Value:  value + sign,
		MaxAge: 0,
		Path:   "/",
	}
	entry := models.ClientData{}
	entry.Cookie = value
	entry.Key = key
	entry.Short = make([]models.ShortData, 0)
	err := application.Storage.Write(entry)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &cookie)
}
func (application *app) checkCookie(cookie *http.Cookie) bool {
	data := cookie.Value[:32]
	signstring := cookie.Value[32:]
	sign, err := hex.DecodeString(signstring)
	if err != nil {
		log.Println(err)
	}
	checkdata, _ := application.Storage.ReadByCookie(data)
	h := hmac.New(sha256.New, []byte(checkdata.Key))
	h.Write([]byte(data))
	signed := h.Sum(nil)
	return hmac.Equal(sign, signed)
}
