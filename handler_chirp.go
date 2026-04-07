package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rhruban/chirpy/internal/database"
)

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, req *http.Request) {
	type chirp struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	type returnVals struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdateAt  time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(req.Body)
	c := chirp{}
	err := decoder.Decode(&c)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode body", err)
		return
	}

	const maxChirpLength = 140
	if len(c.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(c.Body, badWords)

	c_db, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: c.UserId,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, returnVals{
		ID:        c_db.ID,
		CreatedAt: c_db.CreatedAt,
		UpdateAt:  c_db.UpdatedAt,
		Body:      c_db.Body,
		UserID:    c_db.UserID,
	})
}

func getCleanedBody(s string, badWords map[string]struct{}) string {
	words := strings.Split(s, " ")
	for i := range words {
		if _, ok := badWords[strings.ToLower(words[i])]; ok {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}
