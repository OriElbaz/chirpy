package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)


type apiConfig struct {
	fileserverHits atomic.Int32
}


func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cfg.fileserverHits.Add(1)
        next.ServeHTTP(w, r)
    })
}


func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	w.Write([]byte("File server hits reset to 0"))

}


func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		val := cfg.fileserverHits.Load()
		output := fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>
		`, val)
		w.Write([]byte(output))
}


func (cfg *apiConfig) handlerHealthz(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	
type parameters struct {
	Body string `json:"body"`
}

func (cfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Printf("ERROR decoding requests: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")

	if len(params.Body) > 140 {
		responseBody := map[string]string{"error":"Chirp is too long"}
		data, _ := json.Marshal(responseBody)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(data)
		return
	}

	outputBody := badWordReplacement(params)
	responseBody := map[string]string{"cleaned_body":outputBody}
	data, _ := json.Marshal(responseBody)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}


func badWordReplacement(p parameters) string {
	split := strings.Fields(p.Body)
	for i, word := range split {
		word = strings.ToLower(word)
		if word == "kerfuffle" || word == "sharbert" || word == "fornax" {
			split[i] = "****"
		}
	}

	return strings.Join(split, " ")
}