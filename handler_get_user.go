package main

import (
	"net/http"

	"github.com/JustinLi007/rss-aggregator/internal/database"
)

func (cfg *apiConfig) handlerGetUserAuthed(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJSON(w, http.StatusOK, databaseUserToUser(user))
}
