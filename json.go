package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)

	w.Write(data)
}

func respondWithError(w http.ResponseWriter, statusCode int, errorMessage string) {
	if statusCode > 499 {
		log.Printf("Responding with 5XX error: %v", errorMessage)
	}

	type payload struct {
		Error string `json:"error"`
	}

	respondWithJSON(w, statusCode, payload{
		Error: errorMessage,
	})
}
