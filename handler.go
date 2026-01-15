package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type parameters struct {
	Body string `json:"body"`
}


type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}


func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cfg.fileserverHits.Add(1)
        next.ServeHTTP(w, r)
    })
}


func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	
	if err := cfg.db.ResetUsers(r.Context()); err != nil {
		log.Printf("ERROR resetting user table: %v", err)
	}
	w.Write([]byte("User table reset"))

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


func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email string `json:"email"`
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR decoding request: %v" , err)
		return
	}
	
	user, err := cfg.db.CreateUser(r.Context(), req.Email)
	if err != nil {
		log.Printf("ERROR creating user in db: %v", err)
		return
	}
	
	output := User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	data, err := json.Marshal(output)
	if err != nil {
		log.Printf("ERROR marshalling output: %v", err)
		return
	}
	w.Write(data)
}

/***** HELPER *****/
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