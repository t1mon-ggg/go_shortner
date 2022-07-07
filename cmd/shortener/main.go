package main

import (
	"log"
	"net/http"

	"github.com/t1mon-ggg/go_shortner/app/webhandlers"
)

func main() {
	application := webhandlers.NewApp()
	err := application.NewStorage()
	if err != nil {
		log.Fatalln("Coud not set storage", err)
	}
	r := application.NewWebProcessor(10)
	http.ListenAndServe(application.Config.ServerAddress, r)
}
