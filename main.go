package main

import (
	"fmt"
	"net/http"
	"time"
)




func main(){
	fmt.Println("Server started!")
	mux := http.NewServeMux()
	apiCfg := &apiConfig{}

	fileServerHandler := http.FileServer(http.Dir("."))

	s := &http.Server{
		Addr:           ":8080",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	wrappedHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServerHandler))

	mux.Handle("/app/", wrappedHandler)

	mux.Handle("/app/assets/", wrappedHandler)

	mux.HandleFunc("GET /api/healthz", apiCfg.handlerHealthz)

	mux.HandleFunc("GET /api/metrics", apiCfg.handlerMetrics)

	mux.HandleFunc("POST /api/reset", apiCfg.handlerReset)
	
	// ListenAndServe() blocks the main function until the server shuts down
	s.ListenAndServe()
}