package main

import (
	"net/http"

	"github.com/JustinLi007/rss-aggregator/internal/database"
)

func (cfg *apiConfig) getFeedFollowsAuthedHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	feedFollows, err := cfg.DB.GetFeedFollows(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve feed follows")
		return
	}

	respondWithJSON(w, http.StatusOK, databaseFeedFollowsToFeedFollows(feedFollows))
}
