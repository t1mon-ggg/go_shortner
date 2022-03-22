package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/t1mon-ggg/go_shortner.git/internal/app"
)

const (
	addr = "127.0.0.1:8080"
)

type tmpDB map[string]string

// func (db tmpDB) Database(post, get, answer chan string) {
// 	select {
// 	case <-post:
// 		data := <-post

// 	case <-get:

// 	}
// }

func (db tmpDB) Router(r chi.Router) {
	r.Get("/", DefaultGetHandler)
	r.Get("/{^[a-zA-Z]}", db.GetHandler)
	r.Post("/", db.PostHandler)
	r.MethodNotAllowed(OtherHandler)
}

func OtherHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Bad request", http.StatusBadRequest)
	return
}

func (db tmpDB) PostHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	blongURL, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	slongURL := string(blongURL)
	surl := app.RandStringRunes(8)
	db[surl] = slongURL
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("http://%s/%s", r.Host, surl)))
	return
}

func (db tmpDB) GetHandler(w http.ResponseWriter, r *http.Request) {

	if len(db) == 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	p := r.RequestURI
	p = p[1:]
	if _, ok := db[p]; !ok {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if len(p) != 8 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	lurl := db[p]
	w.Header().Set("Location", lurl)
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Write([]byte{})
	return
}

func DefaultGetHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Empty request", http.StatusBadRequest)
	return
}

func main() {
	db := make(tmpDB)
	db["ABCDabcd"] = "https://yandex.ru"

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// r.Use(mymiddleware.TimerTrace)

	r.Route("/", db.Router)

	http.ListenAndServe(":8080", r)

}
