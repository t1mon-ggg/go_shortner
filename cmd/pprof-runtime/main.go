package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/t1mon-ggg/go_shortner/app/webhandlers"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
	memprofile = flag.String("memprofile", "", "write memory profile to `file`")
)

func shortner() {
	application := webhandlers.NewApp()
	err := application.NewStorage()
	if err != nil {
		log.Fatalln("Coud not set storage", err)
	}
	r := application.NewWebProcessor(10)
	http.ListenAndServe(application.Config.ServerAddress, r)
}

// cpuProfile - cpu profiling
func cpuProfile() {
	f, err := os.Create(*cpuprofile)
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer f.Close() // error handling omitted for example
	defer pprof.StopCPUProfile()
	time.Sleep(240 * time.Second)

}

// memProfile - mem profiling
func memProfile() {
	f, err := os.Create(*memprofile)
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
	defer f.Close() // error handling omitted for example
}
func main() {
	flag.Parse()
	if *cpuprofile != "" {
		go cpuProfile()
	}

	go shortner()
	time.Sleep(120 * time.Second)

	if *memprofile != "" {
		memProfile()
	}
}
