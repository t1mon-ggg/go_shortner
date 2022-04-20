package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/t1mon-ggg/go_shortner/internal/app/webhandlers"
)

func main() {
	AppData := webhandlers.NewApp()
	err := AppData.Config.ReadEnv()
	if err != nil {
		log.Println(err)
	}
	AppData.Config.ReadCli()
	AppData.Storage, err = AppData.Config.SetStorage()
	if err != nil {
		log.Fatal(err)
	}
	AppData.Data, err = AppData.Storage.Read()
	if err != nil {
		log.Fatal(err)
	}
	r := chi.NewRouter()

	AppData.MyMiddlewares(r)

	r.Route("/", AppData.Router)

	http.ListenAndServe(AppData.Config.ServerAddress, r)

}
