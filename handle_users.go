package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/keertirajmalik/chirpy/internal/auth"
	"github.com/keertirajmalik/chirpy/internal/database"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

func (cfg *apiConfig) handleUsersCreate(writer http.ResponseWriter, request *http.Request) {

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(request.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't hash password")
		return
	}

	user, err := cfg.DB.CreateUser(params.Email, hashedPassword)
	if err != nil {
		if errors.Is(err, database.ErrAlreadyExists) {
			respondWithError(writer, http.StatusConflict, "User already exists")
			return
		}

		respondWithError(writer, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	respondWithJson(writer, http.StatusCreated, User{
		ID:    user.ID,
		Email: user.Email})
}

func (cfg *apiConfig) handleUsersUpdate(writer http.ResponseWriter, request *http.Request) {

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
	}

	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Couldn't find JWT")
		return
	}

	subject, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {

		//refreshToken, err := cfg.DB.GetTokenByRefreshToken(token)
		//subject, err = auth.ValidateJWT(refreshToken.Token, cfg.jwtSecret)
		//if err != nil {
		respondWithError(writer, http.StatusUnauthorized, "Couldn't validate JWT")
		return
		//}
	}

	decoder := json.NewDecoder(request.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't hash password")
		return
	}

	userIDInt, err := strconv.Atoi(subject)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't parse user ID")
		return
	}

	user, err := cfg.DB.UpdateUser(userIDInt, params.Email, hashedPassword)
	if err != nil {
		respondWithError(writer, http.StatusInternalServerError, "Couldn't create user")
		return
	}

	respondWithJson(writer, http.StatusOK, response{
		User: User{
			ID:    user.ID,
			Email: user.Email,
		},
	})
}
