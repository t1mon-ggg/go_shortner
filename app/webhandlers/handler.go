package webhandlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/t1mon-ggg/go_shortner/api"
	"github.com/t1mon-ggg/go_shortner/app/config"
	"github.com/t1mon-ggg/go_shortner/app/helpers"
	"github.com/t1mon-ggg/go_shortner/app/models"
	"github.com/t1mon-ggg/go_shortner/app/mymiddlewares"
	"github.com/t1mon-ggg/go_shortner/app/storage"
)

// @Title URL SHORTNER API
// @Description Сервис сокращения ссылок
// @Version 1.0

// @Contact.email t1mon.ggg@yandex.ru

// @BasePath /
// @Host 127.0.0.1:8080

// App - application struct
type App struct {
	Storage storage.Storage
	Config  *config.Config
	DelBuf  chan models.DelWorker
	wg      *sync.WaitGroup
	done    chan os.Signal
}

type answer struct {
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}

type input struct {
	Correlation string `json:"correlation_id"`
	Long        string `json:"original_url"`
}

type output struct {
	Correlation string `json:"correlation_id"`
	Short       string `json:"short_url"`
}

type sURL struct {
	ShortURL string `json:"result"`
}

type lURL struct {
	LongURL string `json:"url"`
}

// NewApp - функция для создания новой структуры для работы приложения
func NewApp() *App {
	s := App{}
	s.Config = config.New()
	s.DelBuf = make(chan models.DelWorker)
	return &s
}

// Wait - return application sync.WaitGroup pointer
func (application *App) Wait() *sync.WaitGroup {
	return application.wg
}

// Signal - return os.Signal channel
func (application *App) Signal() chan os.Signal {
	return application.done
}

func (application *App) NewStorage() error {
	var err error
	application.Storage, err = application.Config.NewStorage()
	if err != nil {
		return err
	}
	return nil
}

// NewWebProcessor - создание новго роутера для обработки веб запросов
//  workers int - количество потоков для удаления сокращенных ссылок
func (application *App) NewWebProcessor(workers int) *chi.Mux {
	application.done = make(chan os.Signal, 1)
	application.wg = &sync.WaitGroup{}
	signal.Notify(application.done, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go application.Storage.Cleaner(application.done, application.wg, application.DelBuf, workers)
	r := chi.NewRouter()
	application.middlewares(r)
	r.Route("/", application.router)
	return r
}

// Router - creates chi router and cleaner
func (application *App) router(r chi.Router) {
	r.Get("/", defaultGetHandler)
	r.Get("/ping", application.connectionTest)
	r.Get("/{^[a-zA-Z]}", application.getHandler)
	r.Get("/api/user/urls", application.userURLs)
	r.Get("/api/internal/stats", application.getStats)
	r.Post("/", application.postHandler)
	r.Post("/api/shorten", application.postAPIHandler)
	r.Post("/api/shorten/batch", application.postAPIBatch)
	r.Delete("/api/user/urls", application.deleteTags)
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(application.Config.BaseURL+"/swagger/doc.json")))
	r.MethodNotAllowed(otherHandler)
	r.NotFound(otherHandler)
}

// defaultGetHandler - handler for "/"
func defaultGetHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Empty request", http.StatusBadRequest)
}

// otherHandler - handler for unknown request and unknow paths
func otherHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Bad request", http.StatusBadRequest)
}

// Ping godoc
// @Tags Ping
// @Summary Запрос состояния хранилища
// @Accept text/plain
// @Produce text/plain
// @Success 200   "Проверка подключения к хранилищу успешно завершено"
// @Failure 500   "Хранилище недоступно"
// @Router / [get]
// ConnectionTest - handler for "/ping"
func (application *App) connectionTest(w http.ResponseWriter, r *http.Request) {
	err := application.Storage.Ping()
	if err != nil {
		log.Println(err)
		http.Error(w, "Storage connection failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte{})
}

// Create godoc
// @Tags ListAll
// @Summary Запрос на получение всех сокращенных ссылок пользователя
// @Accept text/plain
// @Produce application/json
// @Param Client_ID header string true "Идентификационный cookie Client_ID"
// @Success 202   "Сокращения в базе данных отсутствуют"
// @Success 200 {array} answer "Список сокращений пользователя"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /api/user/urls [get]
// userURLs - handler for "/api/user/urls" GET Method
func (application *App) userURLs(w http.ResponseWriter, r *http.Request) {
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

// Create godoc
// @Tags GetStats
// @Summary Запрос статистики
// @Accept text/plain
// @Produce application/json
// @Param Client_ID header string true "Идентификационный cookie Client_ID"
// @Success 200 {array} answer "Статистика"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router api/internal/stats [get]
// userURLs - handler for "api/internal/stats" GET Method
func (application *App) getStats(w http.ResponseWriter, r *http.Request) {
	if application.Config.TrustedSubnet == "" {
		log.Println("endpoint locked")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	_, tSubnet, err := net.ParseCIDR(application.Config.TrustedSubnet)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	realIP := r.Header.Get("X-Real-IP")
	if realIP == "" {
		log.Println("Header X-Real-IP not provided")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if !tSubnet.Contains(net.ParseIP(realIP)) {
		log.Println("Header X-Real-IP provides wrong ip")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	stats, err := application.Storage.GetStats()
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	sts, err := json.Marshal(stats)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(sts)
}

// Create godoc
// @Tags Create
// @Summary Запрос на сокращение ссылки
// @Accept text/plain
// @Produce text/plain
// @Param Input body string true "Сокращаемый URL"
// @Param Client_ID header string false "Идентификационный cookie Client_ID"
// @Success 201 {string} string "Создана новая сокращенная ссылка"
// @Success 409 {string} string "Запрашиваемый URL уже существует"
// @Failure 500   "Внутренняя ошибка сервера"
// @Router / [post]
// postHandler - handler for "/" POST Method
func (application *App) postHandler(w http.ResponseWriter, r *http.Request) {
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
			s, errTag := application.Storage.TagByURL(slongURL, cookie)
			if errTag != nil {
				log.Println(errTag)
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

// APICreate godoc
// @Tags APICreate
// @Summary Запрос на сокращение ссылки
// @Accept application/json
// @Produce application/json
// @Param Client_ID header string false "Идентификационный cookie Client_ID"
// @Param Input body lURL true "Сокращаемый URL"
// @Success 201 {object} sURL "Создана новая сокращенная ссылка"
// @Success 409 {object} sURL "Запрашиваемый URL уже существует"
// @Failure 400   "Неверный запрос"
// @Failure 500   "Внутренняя ошибка сервера"
// @Router /api/shorten [post]
// postAPIHandler - handler for "/api/shorten" POST Method
func (application *App) postAPIHandler(w http.ResponseWriter, r *http.Request) {
	cookie := idCookieValue(w, r)
	entry := models.ClientData{}
	entry.Cookie = cookie
	entry.Key = ""
	entry.Short = make([]models.ShortData, 0)
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
			s, errTag := application.Storage.TagByURL(longURL.LongURL, cookie)
			if errTag != nil {
				log.Println(errTag)
				http.Error(w, "Storage error", http.StatusInternalServerError)
				return
			}
			jbody := sURL{ShortURL: fmt.Sprintf("%s/%s", application.Config.BaseURL, s)}
			abody, errJSON := json.Marshal(jbody)
			if errJSON != nil {
				log.Println("JSON Marshal error", errJSON)
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
	jbody := sURL{ShortURL: fmt.Sprintf("%s/%s", application.Config.BaseURL, short)}
	abody, err := json.Marshal(jbody)
	if err != nil {
		log.Println("JSON Marshal error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Write(abody)
}

// APICreateBatch godoc
// @Tags APICreate
// @Summary Запрос на сокращение ссылок списком
// @Accept application/json
// @Produce application/json
// @Param Client_ID header string false "Идентификационный cookie Client_ID"
// @Param Input body input true "Список сокращаемых URLs"
// @Success 201 {object} output "Список успешно обработан"
// @Failure 400   "Неверный запрос"
// @Failure 500   "Внутренняя ошибка сервера"
// @Router /api/shorten/batch [post]
// postAPIBatch - handler for "/api/shorten/batch" POST Method
func (application *App) postAPIBatch(w http.ResponseWriter, r *http.Request) {
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
		errWrite := application.Storage.Write(entry)
		if errWrite != nil {
			if errWrite.Error() == "not unique url" {
				s, errTag := application.Storage.TagByURL(in[i].Long, cookie)
				if errTag != nil {
					http.Error(w, "Storage error", http.StatusInternalServerError)
					return
				}
				out = append(out, output{Correlation: in[i].Correlation, Short: fmt.Sprintf("%s/%s", application.Config.BaseURL, s)})
			} else {
				log.Println(err)
				http.Error(w, "Storage error", http.StatusInternalServerError)
				return
			}
		} else {
			out = append(out, output{Correlation: in[i].Correlation, Short: fmt.Sprintf("%s/%s", application.Config.BaseURL, short)})
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

// getHandler - handler for "/{short_tag}" GET Method
// cjover short url to original url
func (application *App) getHandler(w http.ResponseWriter, r *http.Request) {
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

// APIDeleteShort godoc
// @Tags APIDelete
// @Summary Запрос на удаление короткой ссылки
// @Accept application/json
// @Produce application/json
// @Param Input body string true "Список удаляемых коротких идентификаторов"
// @Param Client_ID header string true "Идентификационный cookie Client_ID"
// @Success 202   "Запрос принят в обработку"
// @Failure 500   "Внутренняя ошибка сервера"
// @Router /api/user/urls [delete]
// deleteTags - handler for "/api/user/urls" DELETE Method
func (application *App) deleteTags(w http.ResponseWriter, r *http.Request) {
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

// middlewares - middleware definition
func (application *App) middlewares(r *chi.Mux) {
	r.Use(middleware.Compress(5))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(mymiddlewares.DecompressRequestAndTimeTracer)
	r.Use(application.cookieProcessor)
}

// idCookieValue - get cookie value fron request
func idCookieValue(w http.ResponseWriter, r *http.Request) string {
	if len(r.Cookies()) == 0 {
		re := regexp.MustCompile(`\w{96}`)
		cid := re.FindString(w.Header().Get("Set-Cookie"))
		if len(cid) != 0 {
			if len(cid) == 96 {
				value := cid[:32]
				return value
			}
		}
		return ""
	} else {
		for _, cookie := range r.Cookies() {
			if cookie.Name == "Client_ID" {
				if len(cookie.Value) == 96 {
					value := cookie.Value[:32]
					return value
				}
			}
		}
	}
	return ""
}

// cookieProcessor - cookie processor
func (application *App) cookieProcessor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.Cookies()) != 0 {
			found := false
			for _, cookie := range r.Cookies() {
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

// addCookie - add cookie to response
func (application *App) addCookie(w http.ResponseWriter, name, value string, key string) {
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

// checkCookie - cookie validation
func (application *App) checkCookie(cookie *http.Cookie) bool {
	data := cookie.Value[:32]
	signstring := cookie.Value[32:]
	sign, err := hex.DecodeString(signstring)
	if err != nil {
		log.Println(err)
		return false
	}
	checkdata, _ := application.Storage.ReadByCookie(data)
	h := hmac.New(sha256.New, []byte(checkdata.Key))
	h.Write([]byte(data))
	signed := h.Sum(nil)
	return hmac.Equal(sign, signed)
}
