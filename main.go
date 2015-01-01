package main

import (
	"github.com/dr4ke616/GoProxy/server"
	"log"
	"os"
)

func main() {

	server.DEVIL = false

	var err error
	proxy := server.Proxy{}

	// Load the proxy from a config file
	c := "/etc/goproxy/config.json"
	if server.DEVIL {
		c = "config.json"
	}
	err = server.LoadFromConfig(&proxy, c)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Start the Proxy server
	err = server.StartProxy(&proxy)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
