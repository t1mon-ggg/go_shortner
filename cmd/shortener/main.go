package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
)

const (
	addr = "127.0.0.1:8080"
)

//Deleteme
var tmpDB map[string]string

func RandStringRunes(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func MyHandler(w http.ResponseWriter, r *http.Request) {
	if tmpDB == nil {
		tmpDB = make(map[string]string)
	}
	switch r.Method {
	case http.MethodGet:
		if len(r.RequestURI) == 1 {
			http.Error(w, "Bad request", http.StatusBadRequest)
		} else {
			lurl := tmpDB[r.RequestURI]
			w.Header().Set("Location", lurl)
			w.WriteHeader(http.StatusTemporaryRedirect)
			w.Write([]byte{})
		}
		return
	case http.MethodPost:
		defer r.Body.Close()
		blongURL, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
		}
		slongURL := string(blongURL)
		// log.Println(fmt.Sprintf("Got url %s", slongURL))
		surl := "/" + RandStringRunes(8)
		tmpDB[surl] = slongURL
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("http://%s%s", addr, surl)))
		return
	default:
		// log.Println(r)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
}

func main() {
	tmpDB = make(map[string]string)
	mux := http.NewServeMux()
	mux.HandleFunc("/", MyHandler)
	log.Println(fmt.Sprintf("Запуск сервера на %s", addr))
	err := http.ListenAndServe(addr, mux)
	if err != nil {
		panic(err)
	}
}
