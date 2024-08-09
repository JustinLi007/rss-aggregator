package main

import (
	"net/http"

	"github.com/JustinLi007/rss-aggregator/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) unfollowFeedAuthedHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	feedFollowID := r.PathValue("feedFollowID")
	if feedFollowID == "" {
		respondWithError(w, http.StatusNotFound, "No feed follow ID included")
		return
	}

	if err := uuid.Validate(feedFollowID); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid feed follow ID")
		return
	}

	feedFollowUUID, err := uuid.Parse(feedFollowID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to parse feed follow id")
		return
	}

	feedFollow, err := cfg.DB.GetFeedFollowByID(r.Context(), feedFollowUUID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to locate followed feed")
		return
	}

	if feedFollow.UserID.String() == user.ID.String() {
		respondWithError(w, http.StatusBadRequest, "Feed not followed by user")
		return
	}

	if err = cfg.DB.UnfollowFeed(r.Context(), feedFollowUUID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to unfollow")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{})
}
