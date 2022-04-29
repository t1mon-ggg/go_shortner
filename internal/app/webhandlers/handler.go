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
	"strings"

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
}

//NewData - создание пустого массива данных
func NewData() map[string]models.WebData {
	s := make(map[string]models.WebData)
	return s
}

//NewApp - функция для создания новой структуры для работы приложения
func NewApp() *app {
	s := app{}
	s.Config = config.NewConfig()
	return &s
}

func (application *app) Router(r chi.Router) {
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
	value := idCookieValue(w, r)
	data, err := application.Storage.ReadByCookie(value)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if len(data) == 0 {
		http.Error(w, "No Content", http.StatusNoContent)
		return
	}
	a := make([]answer, 0)
	for i, content := range data[value].Short {
		a = append(a, answer{Short: fmt.Sprintf("%s/%s", application.Config.BaseURL, i), Original: content})
	}
	d, err := json.Marshal(a)
	if err != nil {
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
	data := make(map[string]models.WebData)
	value := idCookieValue(w, r)
	entry := models.WebData{}
	entry.Key = ""
	entry.Short = make(map[string]string)
	defer r.Body.Close()
	blongURL, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	slongURL := string(blongURL)
	surl := helpers.RandStringRunes(8)
	entry.Short[surl] = slongURL
	data[value] = entry
	err = application.Storage.Write(data)
	if err != nil {
		if err.Error() == "not uniquie url" {
			s, err := application.Storage.TagByURL(slongURL)
			if err != nil {
				http.Error(w, "Storage error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(fmt.Sprintf("%s/%s", application.Config.BaseURL, s)))
			return
		}
		http.Error(w, "Storage error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", application.Config.BaseURL, surl)))
}

func (application *app) postAPIHandler(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]models.WebData)
	value := idCookieValue(w, r)
	data[value] = models.WebData{}
	newentry := data[value]
	newentry.Key = ""
	newentry.Short = make(map[string]string)
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
	err = json.Unmarshal(body, &longURL)
	if err != nil {
		log.Println("JSON Unmarshal error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	short := helpers.RandStringRunes(8)
	newentry.Short[short] = longURL.LongURL
	data[value] = newentry
	err = application.Storage.Write(data)
	if err != nil {
		if err.Error() == "not uniquie url" {
			s, err := application.Storage.TagByURL(longURL.LongURL)
			if err != nil {
				http.Error(w, "Storage error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			jbody := sURL{ShortURL: fmt.Sprintf("%s/%s", (*application).Config.BaseURL, s)}
			abody, err := json.Marshal(jbody)
			if err != nil {
				log.Println("JSON Marshal error", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			w.Write(abody)
			return
		}
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
	value := idCookieValue(w, r)
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
		data := make(map[string]models.WebData)
		entry := models.WebData{}
		entry.Key = ""
		entry.Short = make(map[string]string)
		short := helpers.RandStringRunes(8)
		entry.Short[short] = in[i].Long
		data[value] = entry
		err := application.Storage.Write(data)
		if err != nil {
			if err.Error() == "not uniquie url" {
				s, err := application.Storage.TagByURL(in[i].Long)
				if err != nil {
					http.Error(w, "Storage error", http.StatusInternalServerError)
					return
				}
				out = append(out, output{Correlation: in[i].Correlation, Short: fmt.Sprintf("%s/%s", (*application).Config.BaseURL, s)})
			} else {
				http.Error(w, "Storage error", http.StatusInternalServerError)
				return
			}
		} else {
			out = append(out, output{Correlation: in[i].Correlation, Short: fmt.Sprintf("%s/%s", (*application).Config.BaseURL, short)})
		}
	}
	batch, err := json.Marshal(out)
	if err != nil {
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
		http.Error(w, "URI process error", http.StatusInternalServerError)
		return
	}
	if !matched || len(p) > 8 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	data, err := application.Storage.ReadByTag(p)
	if err != nil {
		http.Error(w, "DB read error", http.StatusInternalServerError)
		return
	}
	var count int
	for i := range data {
		if i == p {
			count++
			w.Header().Set("Location", data[i])
			w.WriteHeader(http.StatusTemporaryRedirect)
			w.Write([]byte{})
		}
	}
	if count == 0 {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

}

func (application *app) deleteTags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte{})
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	re := regexp.MustCompile(`\w+`)
	tags := re.FindAllString(string(body), -1)
	log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!", tags)
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
		str := strings.Split(w.Header().Get("Set-Cookie"), "=")
		if len(str) == 2 {
			newcookieval = str[1]
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
	}
	data := make(map[string]models.WebData)
	entry := models.WebData{}
	entry.Key = key
	entry.Short = make(map[string]string)
	data[value] = entry
	err := application.Storage.Write(data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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
	h := hmac.New(sha256.New, []byte(checkdata[data].Key))
	h.Write([]byte(data))
	signed := h.Sum(nil)
	return hmac.Equal(sign, signed)
}
