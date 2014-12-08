package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func RouteHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write it back to the client.
		log.Println("[RouteHandler]", r.Method, r.URL, r.UserAgent())
		fmt.Fprintf(w, "howdy")
	})
}

func main() {
	port := 8080
	portstring := strconv.Itoa(port)

	http.Handle("/", RouteHandler())

	log.Println("Starting GO webserver on port", portstring)
	err := http.ListenAndServe(":"+portstring, nil)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
		os.Exit(1)
	}
}
