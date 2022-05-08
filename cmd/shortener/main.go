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
	var err error
	application.Storage, err = storage.GetStorage(application.Config)
	if err != nil {
		log.Fatalln("Coud not set storagre", err)
	}

	r := chi.NewRouter()

	application.Middlewares(r)

	r.Route("/", application.Router)

	http.ListenAndServe(application.Config.ServerAddress, r)

}
