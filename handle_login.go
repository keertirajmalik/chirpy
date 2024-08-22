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

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't create JWT")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't create refresh token")
		return
	}

	err = cfg.DB.SaveRefreshToken(user.ID, refreshToken)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't save refresh token")
		return
	}

	respondWithJson(writer, http.StatusOK, response{
		User: User{
			ID:    user.ID,
			Email: user.Email},
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handleRefresh(writer http.ResponseWriter, request *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Couldn't find token")
		return
	}

	user, err := cfg.DB.UserForRefershToken(refreshToken)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "No token")
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Couldn't validate token")
		return
	}
	respondWithJson(writer, http.StatusOK, response{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handleRevoke(writer http.ResponseWriter, request *http.Request) {
	refreshToken, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Invalid header")
		return
	}

	err = cfg.DB.RevokeToken(refreshToken)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Couldn't revoke session")
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
