package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/keertirajmalik/chirpy/internal/auth"
)

func (cfg *apiConfig) handleLogin(writer http.ResponseWriter, request *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	type response struct {
		User
		Token string `json:"token"`
	}

	decoder := json.NewDecoder(request.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	user, err := cfg.DB.GetUserByEmail(params.Email)

	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't get user")
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)

	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Invalid password")
		return
	}

	defaultExpiration := 60 * 60 * 24
	if params.ExpiresInSeconds == 0 {
		params.ExpiresInSeconds = defaultExpiration
	} else if params.ExpiresInSeconds > defaultExpiration {
		params.ExpiresInSeconds = defaultExpiration
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Duration(params.ExpiresInSeconds)*time.Second)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't create JWT")
		return
	}

	respondWithJson(writer, http.StatusOK, response{
		User: User{
			ID:    user.ID,
			Email: user.Email},
		Token: token,
	})
}
