package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
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

	c.Body = cleanProfanity(c.Body)
	respondWithJSON(w, http.StatusOK, returnVals{
		CleanedBody: c.Body,
	})
}

func cleanProfanity(s string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(s, " ")
	for i := range words {
		for _, badWord := range badWords {
			if strings.ToLower(words[i]) == badWord {
				words[i] = "****"
			}
		}
	}
	return strings.Join(words, " ")
}
