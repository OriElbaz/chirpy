package main

import (
	"net/http"
	"time"
)


func main(){
	// serveMux is a HTTP request router
	serveMux := http.NewServeMux()

	// http.Server is a struct that defines a server configuration
	s := &http.Server{
		Addr:           ":8080",
		Handler:        serveMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	serveMux.Handle("/", http.FileServer(http.Dir(".")))
	
	// ListenAndServe() blocks the main function until the server shuts down
	s.ListenAndServe()
}