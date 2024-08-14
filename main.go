package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/JustinLi007/rss-aggregator/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB        *database.Queries
	nextFeeds []Feed
	limit     int32
}

func main() {
	debug := flag.Bool("debug", false, "Debug mode")
	flag.Parse()

	if debug != nil && *debug {
		fmt.Println("Debug mode enabled")
	}

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable not set")
	}

	dbURL := os.Getenv("PSQL_CONNECTION_URL")
	if port == "" {
		log.Fatal("PSQL_CONNECTION_URL environment variable not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := apiConfig{
		DB: database.New(db),
	}

	serveMux := http.NewServeMux()
	serveMux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("."))))

	serveMux.HandleFunc("GET /v1/healthz", healthzHandler)
	serveMux.HandleFunc("GET /v1/err", errorHandler)

	serveMux.HandleFunc("POST /v1/users", apiCfg.createUsersHandler)
	serveMux.HandleFunc("GET /v1/users", apiCfg.middlewareAuth(apiCfg.getUserAuthedHandler))

	serveMux.HandleFunc("POST /v1/feeds", apiCfg.middlewareAuth(apiCfg.createFeedsAuthedHandler))
	serveMux.HandleFunc("GET /v1/feeds", apiCfg.getFeedsHandler)

	serveMux.HandleFunc("POST /v1/feed_follows", apiCfg.middlewareAuth(apiCfg.followFeedAuthedHandler))
	serveMux.HandleFunc("GET /v1/feed_follows", apiCfg.middlewareAuth(apiCfg.getFeedFollowsAuthedHandler))
	serveMux.HandleFunc("DELETE /v1/feed_follows/{feedFollowID}", apiCfg.middlewareAuth(apiCfg.unfollowFeedAuthedHandler))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serveMux,
	}

	const collectionConcurrency = 10
	const collectionInterval = time.Minute
	go initScraping(apiCfg.DB, collectionConcurrency, collectionInterval)

	log.Printf("Server listening on port: %v", port)
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
