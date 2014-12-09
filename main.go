package main

import (
	"github.com/dr4ke616/GoProxy/server"
	"log"
	"os"
)

func main() {
	var err error
	proxy := server.Proxy{}

	err = server.LoadFromConfig(&proxy)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	err = server.StartProxy(&proxy)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
