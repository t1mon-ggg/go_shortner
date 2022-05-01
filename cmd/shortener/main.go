package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"

	"github.com/t1mon-ggg/go_shortner/internal/app/storage"
	"github.com/t1mon-ggg/go_shortner/internal/app/webhandlers"
)

func main() {
	application := webhandlers.NewApp()

	err := application.Config.Init()
	if err != nil {
		log.Fatal(err)
	}
	application.Storage, err = storage.SetStorage(application.Config)
	if err != nil {
		log.Fatalln("Coud not set storagre", err)
	}

	r := chi.NewRouter()

	application.Middlewares(r)

	r.Route("/", application.Router)

	http.ListenAndServe(application.Config.ServerAddress, r)

}
