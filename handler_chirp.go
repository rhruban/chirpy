package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rhruban/chirpy/internal/auth"
	"github.com/rhruban/chirpy/internal/database"
)

type Chrip struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirpsGetAll(w http.ResponseWriter, req *http.Request) {
	s := req.URL.Query().Get("author_id")
	order := req.URL.Query().Get("sort")
	var allChirps []database.Chirp
	var err error
	if s != "" {
		u, err := uuid.Parse(s)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "could not parse user id", err)
			return
		}
		allChirps, err = cfg.db.GetChirpsByUser(req.Context(), u)
	} else {
		allChirps, err = cfg.db.GetChirps(req.Context())
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrieve chirps", err)
		return
	}

	if order == "desc" {
		sort.Slice(allChirps, func(i, j int) bool { return allChirps[i].CreatedAt.After(allChirps[j].CreatedAt) })
	}

	returnVal := []Chrip{}
	for _, entry := range allChirps {
		returnVal = append(returnVal, Chrip{
			ID:        entry.ID,
			CreatedAt: entry.CreatedAt,
			UpdatedAt: entry.UpdatedAt,
			Body:      entry.Body,
			UserID:    entry.UserID,
		})
	}
	respondWithJSON(w, http.StatusOK, returnVal)
}

func (cfg *apiConfig) handlerChirpsGetSingle(w http.ResponseWriter, req *http.Request) {
	chirpID := req.PathValue("chirpID")
	u, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not parse chirp ID", err)
		return
	}
	chirp, err := cfg.db.GetChirpByID(req.Context(), u)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chrip{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	validUser, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Could not validate JWT", err)
		return
	}

	decoder := json.NewDecoder(req.Body)
	c := parameters{}
	err = decoder.Decode(&c)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode body", err)
		return
	}

	cleaned, err := validateChrip(c.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	c_db, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: validUser,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chrip{
		ID:        c_db.ID,
		CreatedAt: c_db.CreatedAt,
		UpdatedAt: c_db.UpdatedAt,
		Body:      c_db.Body,
		UserID:    c_db.UserID,
	})
}

func validateChrip(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(body, badWords)
	return cleaned, nil
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

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	validUser, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Could not validate JWT", err)
		return
	}

	chirpID := req.PathValue("chirpID")
	u, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not parse chirp ID", err)
		return
	}
	chirp, err := cfg.db.GetChirpByID(req.Context(), u)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	if validUser != chirp.UserID {
		respondWithError(w, http.StatusForbidden, "Not your chirp", err)
		return
	}

	err = cfg.db.DeleteChrip(req.Context(), u)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
