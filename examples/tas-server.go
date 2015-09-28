package main

import (
	"log"
	"runtime"
)

import (
	"github.com/chango/tas/tas"
)

func main() {
	runtime.GOMAXPROCS(10)
	tasConfig := tas.NewDefaultTASConfig()
	svr, err := tas.NewTASServer(tasConfig)
	if err != nil {
		log.Println("Failed to start TAS: %s", err)
		return
	}
	svr.Run()
}
