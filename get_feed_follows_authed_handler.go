package main

import (
	"net/http"

	"github.com/JustinLi007/rss-aggregator/internal/database"
)

func (cfg *apiConfig) getFeedFollowsAuthedHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	allFeedFollows, err := cfg.DB.GetFeedFollows(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve feed follows")
		return
	}

	userFeedFollows := make([]UsersFeedsFollow, 0)
	for _, value := range allFeedFollows {
		if value.UserID.String() == user.ID.String() {
			userFeedFollows = append(userFeedFollows, databaseFeedFollowToFeedFollow(value))
		}
	}

	respondWithJSON(w, http.StatusOK, userFeedFollows)
}
