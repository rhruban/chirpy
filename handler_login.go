package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rhruban/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	type response struct {
		User
	}

	const (
		MAX_EXPIRE = 60 * 60
	)

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	expiry := params.ExpiresInSeconds
	if params.ExpiresInSeconds <= 0 || params.ExpiresInSeconds > MAX_EXPIRE {
		expiry = MAX_EXPIRE
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Duration(expiry)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not make token", err)
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Token:     token,
		},
	})
}
