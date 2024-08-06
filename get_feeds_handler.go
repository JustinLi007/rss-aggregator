package main

import "net/http"

func (cfg *apiConfig) getFeedsHandler(w http.ResponseWriter, r *http.Request) {
	feeds, err := cfg.DB.GetFeeds(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve feeds")
		return
	}

	allFeeds := make([]Feed, 0)
	for _, f := range feeds {
		allFeeds = append(allFeeds, databaseFeedToFeed(f))
	}

	respondWithJSON(w, http.StatusOK, allFeeds)
}
