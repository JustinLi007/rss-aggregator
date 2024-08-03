package main

import (
	"net/http"

	"github.com/JustinLi007/rss-aggregator/internal/auth"
)

func (cfg *apiConfig) getUserHandler(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetApiKeyToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	user, err := cfg.DB.GetUser(r.Context(), apiKey)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, databaseUserToUser(user))
}
