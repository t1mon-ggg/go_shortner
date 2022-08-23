package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/t1mon-ggg/go_shortner/app/grpc"
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

func start(application *webhandlers.App) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		application.Start()
		wg.Done()
	}()
	go func() {
		grpc := grpc.New(application)
		grpc.Start()
		wg.Done()
	}()
	wg.Wait()
}

func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n\n", buildVersion, buildDate, buildCommit)
	application := webhandlers.NewApp()
	err := application.NewStorage()
	if err != nil {
		log.Fatalln("Coud not set storage", err)
	}
	start(application)
}
