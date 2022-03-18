package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func MyHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if len(r.RequestURI) == 1 {
			http.Error(w, "Bad request", http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("some_long_url"))
		}
		return
	case http.MethodPost:
		defer r.Body.Close()
		blongURL, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
		}
		slongURL := string(blongURL)
		log.Println(slongURL)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("some_short_url"))
		return
	default:
		log.Println(r)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
}

func main() {
	addr := "127.0.0.1:8080"
	mux := http.NewServeMux()
	mux.HandleFunc("/", MyHandler)
	log.Println(fmt.Sprintf("Запуск сервера на %s", addr))
	err := http.ListenAndServe(addr, mux)
	if err != nil {
		panic(err)
	}
}
