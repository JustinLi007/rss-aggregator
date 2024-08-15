package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/JustinLi007/rss-aggregator/internal/database"
)

func (cfg *apiConfig) handlerGetPostsByUser(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		Limit int `json:"limit"`
	}

	params := parameters{
		Limit: 10,
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Printf("Failed to decode parameters")
	}

	posts, err := cfg.DB.GetPostByUser(r.Context(), database.GetPostByUserParams{
		UserID: user.ID,
		Limit:  int32(params.Limit),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve posts")
		return
	}

	respondWithJSON(w, http.StatusOK, databasePostsToPosts(posts))
}
