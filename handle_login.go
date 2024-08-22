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
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	refreshToken, err := auth.RefreshToken()
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't create refresh token")
		return
	}

	tokens, err := cfg.DB.CreateRefreshToken(token, refreshToken)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't generate refreshToken")
		return
	}

	respondWithJson(writer, http.StatusOK, response{
		User: User{
			ID:    user.ID,
			Email: user.Email},
		Token:        tokens.Token,
		RefreshToken: tokens.RefreshToken,
	})
}

func (cfg *apiConfig) handleRefresh(writer http.ResponseWriter, request *http.Request) {

	type response struct {
		Token string `json:"token"`
	}

	auth, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Invalid header")
		return
	}

	token, err := cfg.DB.GetToken(auth)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "No token")
		return
	}

	if token.ExpiresAt < time.Duration(time.Now().Second()) {
		respondWithError(writer, http.StatusUnauthorized, "Timed out")
		return
	}

	respondWithJson(writer, http.StatusOK, response{
		Token: token.Token,
	})
}

func (cfg *apiConfig) handleRevoke(writer http.ResponseWriter, request *http.Request) {

	auth, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Invalid header")
		return
	}

	err = cfg.DB.DeleteToken(auth)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "No token found")
		return
	}

	respondWithJson(writer, http.StatusNoContent, "")
}
