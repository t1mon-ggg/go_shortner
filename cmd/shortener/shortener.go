package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/t1mon-ggg/go_shortner/internal/app/storage"
	"github.com/t1mon-ggg/go_shortner/internal/app/webhandlers"
)

func main() {
	AppData := webhandlers.NewApp()
	err := AppData.Config.ReadEnv()
	if err != nil {
		log.Println(err)
	}
	AppData.Config.ReadCli()
	if AppData.Config.Database != "" {
		AppData.Storage, err = storage.NewDB(AppData.Config.Database)
		if err != nil {
			log.Println(err)
		}
	} else {
		AppData.Storage = storage.NewFileDB(AppData.Config.FileStoragePath)
	}
	AppData.Data, err = AppData.Storage.Read()
	if err != nil {
		log.Fatal(err)
	}
	r := chi.NewRouter()

	r.Use(middleware.Compress(5))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(webhandlers.DecompressRequest)
	r.Use(AppData.Cookies)

	r.Route("/", AppData.Router)

	log.Println("Current config", *AppData.Config)

	http.ListenAndServe(AppData.Config.ServerAddress, r)

}
