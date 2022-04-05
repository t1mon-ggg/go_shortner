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

func main() {
	cfg := config.NewConfig()
	err := cfg.Read()
	if err != nil {
		log.Fatal(err)
	}
	AppData := webhandlers.NewApp()
	AppData.Config = *cfg
	if err != nil {
		log.Fatal(err)
	}

	AppData.Config.Cli()
	AppData.Storage = *storage.NewFileDB(AppData.Config.FileStoragePath)
	AppData.Data, err = AppData.Storage.Read()
	if err != nil {
		log.Fatal(err)
	}
	r := chi.NewRouter()

	r.Use(middleware.AllowContentEncoding("deflate", "gzip", "br"))
	r.Use(middleware.Compress(6))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", AppData.Router)

	http.ListenAndServe(AppData.Config.ServerAddress, r)

}
