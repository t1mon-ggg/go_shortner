package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/t1mon-ggg/go_shortner/internal/app/config"
	"github.com/t1mon-ggg/go_shortner/internal/app/storage"
	"github.com/t1mon-ggg/go_shortner/internal/app/webhandlers"
)

func main() {
	AppData := webhandlers.NewApp()
	AppData.Config = *config.NewConfig()
	err := AppData.Config.ReadEnv()
	if err != nil {
		log.Fatal(err)
	}
	AppData.Config.ReadCli()
	AppData.Storage = *storage.NewFileDB(AppData.Config.FileStoragePath)
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

	r.Route("/", AppData.Router)

	http.ListenAndServe(AppData.Config.ServerAddress, r)

}
