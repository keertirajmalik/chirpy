package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/keertirajmalik/chirpy/internal/auth"
)

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

func (cfg *apiConfig) handleChirpCreate(writer http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}

	subject, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}

	userID, err := strconv.Atoi(subject)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't parse user ID")
		return
	}

	decoder := json.NewDecoder(request.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(writer, http.StatusBadRequest, err.Error())
		return
	}

	chirp, err := cfg.DB.CreateChirp(cleaned, userID)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	respondWithJson(writer, http.StatusCreated, Chirp{
		ID:       chirp.ID,
		Body:     chirp.Body,
		AuthorID: chirp.AuthorID,
	})
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
		ID:       dbChirps.ID,
		Body:     dbChirps.Body,
		AuthorID: dbChirps.AuthorID,
	})
}

func (cfg *apiConfig) handleChirpDelete(writer http.ResponseWriter, request *http.Request) {
	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}

	subject, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Couldn't validate JWT")
		return
	}

	userID, err := strconv.Atoi(subject)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't parse user ID")
		return
	}

	chirpID, err := strconv.Atoi(request.PathValue("chirpID"))
	if err != nil {
		respondWithError(writer, http.StatusNotFound, "Invalid chrip ID")
		return
	}

	dbChirp, err:= cfg.DB.GetChirp(chirpID)
	if err != nil {
		 respondWithError(writer, http.StatusNotFound, "Couldn't get chirp")
         return
	}

	if dbChirp.AuthorID != userID {
		respondWithError(writer, http.StatusForbidden, "You can't delete this chirp")
        return
	}

	err = cfg.DB.DeleteChirp(chirpID, userID)
	if err != nil {
		respondWithError(writer, http.StatusForbidden, "Couldn't delete chirp")
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
