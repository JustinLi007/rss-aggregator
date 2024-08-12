package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/JustinLi007/rss-aggregator/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB        *database.Queries
	nextFeeds nextFetch
}

type nextFetch struct {
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
		nextFeeds: nextFetch{
			limit: 2,
		},
	}

	//apiCfg.initFetchFeedWorker(time.Second * 60)

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

	log.Printf("Server listening on port: %v", port)
	log.Fatal(server.ListenAndServe())
}

type feedXMLStruct struct{}

func (cfg *apiConfig) initFetchFeedWorker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			feedsToFetch := cfg.getNextFeeds(cfg.nextFeeds.limit)
			cfg.nextFeeds = feedsToFetch
			fetchFeedWorker(feedsToFetch.nextFeeds)
		}
	}
}

func fetchFeedWorker(feeds []Feed) {
	var wg sync.WaitGroup

	for i := 1; i <= len(feeds); i++ {
		wg.Add(i)
		go fetchFeed(feeds[i-1].Url, &wg)
	}

	wg.Wait()
}

func fetchFeed(url string, wg *sync.WaitGroup) feedXMLStruct {
	defer wg.Done()
	log.Printf("Fetching %v", url)
	resp, err := http.Get(url)
	if err != nil {
		if resp != nil {
			log.Printf("Status code: %v", resp.StatusCode)
		}
		log.Printf("Failed to fetch feed at %v: %v", url, err)
		return feedXMLStruct{}
	}
	defer resp.Body.Close()

	log.Printf("Response status code: %v", resp.StatusCode)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read resp body: %v", err)
		return feedXMLStruct{}
	}

	result := feedXMLStruct{}
	if err := xml.Unmarshal(data, &result); err != nil {
		log.Printf("Failed to unmarshal resp data: %v", err)
		return feedXMLStruct{}
	}

	return result
}

func (cfg *apiConfig) getNextFeeds(limit int32) nextFetch {
	feeds, err := cfg.DB.GetNextFeedsToFetch(context.TODO(), limit)
	if err != nil {
		log.Printf("Failed to fetch feeds: %v", err.Error())
		return nextFetch{}
	}

	for _, v := range feeds {
		cfg.DB.MarkFeedFetched(context.TODO(), database.MarkFeedFetchedParams{
			LastFetchedAt: sql.NullTime{
				Time:  time.Now().UTC(),
				Valid: true,
			},
			UpdatedAt: time.Now().UTC(),
			ID:        v.ID,
		})
	}

	result := nextFetch{
		nextFeeds: databaseFeedsToFeeds(feeds),
		limit:     limit,
	}
	return result
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
