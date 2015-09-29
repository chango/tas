package main

import (
	"log"
)

import (
	"github.com/chango/tas/tas"
)

func main() {
	tasConfig := tas.NewDefaultTASConfig()
	svr, err := tas.NewTASServer(tasConfig)
	if err != nil {
		log.Println("Failed to start TAS:", err)
		return
	}
	svr.Run()
}
