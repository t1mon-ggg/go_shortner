package main

import (
	"fmt"
	"log"

	"github.com/t1mon-ggg/go_shortner/app/webhandlers"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func init() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
}

func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n\n", buildVersion, buildDate, buildCommit)
	application := webhandlers.NewApp()
	err := application.NewStorage()
	if err != nil {
		log.Fatalln("Coud not set storage", err)
	}
	r := application.NewWebProcessor(10)
	application.Config.NewListner(r)
}
