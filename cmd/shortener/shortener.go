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

var cfg *config.OsVars
var AppData *webhandlers.DB

func init() {
	cfg := config.NewConfig()
	err := cfg.Read()
	if err != nil {
		log.Fatal(err)
	}
	AppData = webhandlers.NewApp()
	AppData.Config = *cfg
	AppData.Storage = *storage.NewFileDB(AppData.Config.FileStoragePath)
	AppData.Data, err = AppData.Storage.Read()
	if err != nil {
		log.Fatal(err)
	}

}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", AppData.Router)

	http.ListenAndServe(AppData.Config.ServerAddress, r)

}
