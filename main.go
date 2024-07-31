package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load("secrets.env"); err != nil {
		log.Fatal(err)
	}

	debug := flag.Bool("debug", false, "Debug mode")
	flag.Parse()

	if debug != nil && *debug {
		fmt.Println("Debug mode enabled")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable not set")
	}

	serveMux := http.NewServeMux()
	serveMux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("."))))
	serveMux.HandleFunc("GET /v1/healthz", healthzHandler)
	serveMux.HandleFunc("GET /v1/err", errorHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serveMux,
	}

	log.Fatal(server.ListenAndServe())
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	type payload struct {
		Status string `json:"status"`
	}

	respondWithJSON(w, http.StatusOK, payload{
		Status: "ok",
	})
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
}
