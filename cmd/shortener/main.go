package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/t1mon-ggg/go_shortner.git/internal/app"
)

// {"url":"<some_url>"}
// {"result":"<shorten_url>"}
type SURL struct {
	ShortURL string `json:"result"`
}

type LURL struct {
	LongURL string `json:"url"`
}

type tmpDB map[string]string

func (db *tmpDB) Router(r chi.Router) {
	r.Get("/", DefaultGetHandler)
	r.Get("/{^[a-zA-Z]}", db.GetHandler)
	r.Post("/", db.PostHandler)
	r.Post("/api/shorten", db.PostApiHandler)
	r.MethodNotAllowed(OtherHandler)
}

func OtherHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Bad request", http.StatusBadRequest)
}

func (db *tmpDB) PostHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	blongURL, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	slongURL := string(blongURL)
	surl := app.RandStringRunes(8)
	(*db)[surl] = slongURL
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("http://%s/%s", r.Host, surl)))
}

func (db *tmpDB) PostApiHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctype := r.Header.Get("Content-Type")
	if ctype != "application/json" {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
	longURL := LURL{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Body Error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	err = json.Unmarshal(body, &longURL)
	if err != nil {
		log.Println("JSON Unmarshal error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	short := app.RandStringRunes(8)
	(*db)[short] = longURL.LongURL
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	jbody := SURL{ShortURL: fmt.Sprintf("http://%s/%s", r.Host, short)}
	abody, err := json.Marshal(jbody)
	if err != nil {
		log.Println("JSON Marshal error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	w.Write(abody)
}

func (db tmpDB) GetHandler(w http.ResponseWriter, r *http.Request) {

	if len(db) == 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
	p := r.RequestURI
	p = p[1:]
	if _, ok := db[p]; !ok {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
	if len(p) != 8 {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
	lurl := db[p]
	w.Header().Set("Location", lurl)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte{})
}

func DefaultGetHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Empty request", http.StatusBadRequest)
}

func main() {
	db := make(tmpDB)
	db["ABCDabcd"] = "https://yandex.ru"

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", db.Router)

	http.ListenAndServe(":8080", r)

}
