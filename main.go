package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"
	"github.com/OriElbaz/chirpy/internal/database"
	_ "github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)


type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
	platform string
}


func main(){
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Error opening postgres: %v", err)
	}
	dbQueries := database.New(db)
	
	fmt.Println("Server started!")
	mux := http.NewServeMux()
	apiCfg := &apiConfig{
		db: dbQueries,
		platform: os.Getenv("PLATFORM"),
	}

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

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerAllChirps)

	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	
	// ListenAndServe() blocks the main function until the server shuts down
	s.ListenAndServe()
}

