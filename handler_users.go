package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, req *http.Request) {
	type in_val struct {
		Email string `json:"email"`
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

	user, err := cfg.db.CreateUser(req.Context(), in_data.Email)
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
