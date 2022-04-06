package webhandlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/t1mon-ggg/go_shortner/internal/app/config"
	"github.com/t1mon-ggg/go_shortner/internal/app/rand"
	"github.com/t1mon-ggg/go_shortner/internal/app/storage"
)

type app struct {
	Storage storage.FileDB
	Config  config.OsVars
	Data    map[string]string
}

//type WebData map[string]string

func NewData() map[string]string {
	s := make(map[string]string)
	return s
}

func NewApp() *app {
	s := app{}
	s.Storage = storage.FileDB{}
	s.Config = config.OsVars{}
	s.Data = NewData()
	return &s
}

func (db *app) Router(r chi.Router) {
	r.Get("/", defaultGetHandler)
	r.Get("/{^[a-zA-Z]}", db.GetHandler)
	r.Post("/", db.postHandler)
	r.Post("/api/shorten", db.PostAPIHandler)
	r.MethodNotAllowed(otherHandler)
}

func defaultGetHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Empty request", http.StatusBadRequest)
}

func otherHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Bad request", http.StatusBadRequest)
}

func (db *app) postHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	blongURL, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	slongURL := string(blongURL)
	surl := rand.RandStringRunes(8)
	(*db).Data[surl] = slongURL
	rec := map[string]string{surl: slongURL}
	db.Storage.Write(rec)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", db.Config.BaseURL, surl)))
}

func (db *app) PostAPIHandler(w http.ResponseWriter, r *http.Request) {
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
	short := rand.RandStringRunes(8)
	(*db).Data[short] = longURL.LongURL
	rec := map[string]string{short: longURL.LongURL}
	db.Storage.Write(rec)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	jbody := sURL{ShortURL: fmt.Sprintf("%s/%s", (*db).Config.BaseURL, short)}
	abody, err := json.Marshal(jbody)
	if err != nil {
		log.Println("JSON Marshal error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Write(abody)
}

func (db app) GetHandler(w http.ResponseWriter, r *http.Request) {
	if len(db.Data) == 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	p := r.RequestURI
	p = p[1:]
	if _, ok := db.Data[p]; !ok {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if len(p) != 8 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	lurl := db.Data[p]
	w.Header().Set("Location", lurl)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte{})
}

func DecompressRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (strings.Contains(r.Header.Get("Content-Encoding"), "gzip")) || (strings.Contains(r.Header.Get("Content-Encoding"), "br")) || (strings.Contains(r.Header.Get("Content-Encoding"), "deflate")) {
			defer r.Body.Close()
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			defer gz.Close()
			body, err := io.ReadAll(gz)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			r.ContentLength = int64(len(body))
			r.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		next.ServeHTTP(w, r)
	})
}

func TimeTracer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tStart := time.Now()
		next.ServeHTTP(w, r)
		tEnd := time.Since(tStart)
		log.Printf("Duration for a request %s\r\n", tEnd)
	})
}
