package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rhruban/chirpy/internal/auth"
	"github.com/rhruban/chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Token     string    `json:"token"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, req *http.Request) {
	type in_val struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
	}

	decoder := json.NewDecoder(req.Body)
	in_data := in_val{}
	err := decoder.Decode(&in_data)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode in coming body", err)
		return
	}

	hashedPW, err := auth.HashPassword(in_data.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not hash password", err)
		return
	}

	user, err := cfg.db.CreateUser(req.Context(), database.CreateUserParams{
		Email:          in_data.Email,
		HashedPassword: hashedPW,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	})
}
