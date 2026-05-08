package main

import (
	"net/http"
	"time"

	"github.com/rhruban/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find refresh token", err)
		return
	}

	refreshToken, err := cfg.db.GetRefreshByRefresh(r.Context(), token)
	if err != nil {

		return
	}

	if time.Now() > refreshToken.expires_at {
		cfg.db.RevokeRefresh(r.Context(), token)
		return
	}

	expiry := time.Hour

	accessToken, err := auth.MakeJWT(refreshToken.UserID, cfg.jwtSecret, expiry)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not make jwt access token", err)
		return
	}

	respondWithJSON()
}
