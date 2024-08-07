package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/keertirajmalik/chirpy/internal/database"
)

type apiConfig struct {
	fileServerHits int
	DB             *database.DB
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/html")
	writer.WriteHeader(http.StatusOK)

	html := `<html>

<body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
</body>

</html>`

	writer.Write([]byte(fmt.Sprintf(html, cfg.fileServerHits)))
}

func (cfg *apiConfig) handleReset(writer http.ResponseWriter, request *http.Request) {
	cfg.fileServerHits = 0
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("Hit reset to 0"))
}

func (cfg *apiConfig) handleChirpCreate(writer http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(request.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, err.Error())
		return
	}

	chirp, err := cfg.DB.CreateChirp(cleaned)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	respondWithJson(writer, http.StatusCreated, Chirp{
		ID:   chirp.ID,
		Body: chirp.Body})
}

func validateChirp(body string) (string, error) {
	if len(body) > 140 {
		return "", errors.New("Chirp is too long")
	}
	badWords := map[string]struct{}{"kerfuffle": {}, "sharbert": {}, "fornax": {}}

	cleaned := getCleanedBody(body, badWords)

	return cleaned, nil
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")

	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}

func (cfg *apiConfig) handleChirpGet(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	chirps := []Chirp{}
	for _, dbChdbChirp := range dbChirps {
		chirps = append(chirps, Chirp{ID: dbChdbChirp.ID, Body: dbChdbChirp.Body})
	}

	sort.Slice(chirps, func(i, j int) bool { return chirps[i].ID < chirps[j].ID })

	respondWithJson(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handleChirpGetSpecific(w http.ResponseWriter, r *http.Request) {
	chirpId, err := strconv.Atoi(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Invalid chirp ID")
		return
	}

	dbChirps, err := cfg.DB.GetChirp(chirpId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		return
	}

	respondWithJson(w, http.StatusOK, Chirp{
		ID:   dbChirps.ID,
		Body: dbChirps.Body})
}
