package main

import (
	"net/http"

	"github.com/JustinLi007/rss-aggregator/internal/database"
)

func (cfg *apiConfig) getFeedFollowsAuthedHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	// TODO: implement
	respondWithError(w, http.StatusNotImplemented, "Not implemented")
}
