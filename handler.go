package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/OriElbaz/chirpy/internal/database"
	"github.com/google/uuid"
)

type parameters struct {
	Body string `json:"body"`
	UserID uuid.NullUUID `json:"user_id"`
}


type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}


type Chirp struct {
	ID        	uuid.UUID `json:"id"`
	CreatedAt 	time.Time `json:"created_at"`
	UpdatedAt 	time.Time `json:"updated_at"`
	Body     	string    `json:"body"`
	UserID      uuid.NullUUID `json:"user_id"`
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


func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
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
	resParams := database.CreateChirpParams{
		Body: outputBody,
		UserID: params.UserID,
	}
	c, err := cfg.db.CreateChirp(r.Context(), resParams)
	if err != nil {
		log.Printf("ERROR creating chirp: %v", err)
	}

	chirp := Chirp{
		ID: c.ID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Body: c.Body,
		UserID: c.UserID,

	}


	data, _ := json.Marshal(chirp)
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}


func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	c, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("ERROR getting all chirps: %v", err)
		return
	}

	var chirps []Chirp
	for _, chirp := range c {
		new := Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		}
		chirps = append(chirps, new)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	data, err := json.Marshal(chirps)
	if err != nil {
		log.Printf("ERROR marshalling chirps: %v", err)
		return
	}
	w.Write(data)

}


func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("ERROR parsing parameter to uuid: %v", err)
		return
	}
	chirp, err := cfg.db.GetChirpWithID(r.Context(), chirpID)
	if err != nil {
		log.Printf("ERROR getting chirp with id: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	
	w.WriteHeader(http.StatusOK)
	output := Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	}
	data, err := json.Marshal(output)
	if err != nil {
		log.Printf("ERROR marshalling chirp: %v", err)
		return
	}
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