package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/rhruban/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerChirpyRedPayment(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Could not get apiKey", err)
		return
	}
	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Invalid apiKey", err)
	}
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode in coming body", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = cfg.db.UpdateRed(req.Context(), params.Data.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Could not find user", err)
		}
		respondWithError(w, http.StatusInternalServerError, "Could not update user", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
