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
	nextFeeds []Feed
	limit     int32
}

type feedXMLStruct struct {
	XMLName xml.Name
	Attrs   []xml.Attr      `xml:",any,attr"`
	Nodes   []feedXMLStruct `xml:",any"`
	Content string          `xml:",chardata"`
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
		DB:        database.New(db),
		nextFeeds: make([]Feed, 0),
		limit:     1,
	}

	go apiCfg.initFetchFeedWorker(time.Second * 5)

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

func (cfg *apiConfig) initFetchFeedWorker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cfg.nextFeeds = cfg.getNextFeeds(cfg.limit)
			fetchFeedWorker(cfg.nextFeeds)
		}
	}
}

func fetchFeedWorker(feeds []Feed) {
	var wg sync.WaitGroup
	n := len(feeds)
	resultChan := make(chan string, n)

	log.Printf("Start fetchFeedWorker")
	for i := 1; i <= n; i++ {
		wg.Add(1)
		go fetchFeed(feeds[i-1].Url, i, n, &wg, resultChan)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for rs := range resultChan {
		fmt.Println(rs)
	}

	log.Printf("End fetchFeedWorker")
}

func fetchFeed(url string, i, n int, wg *sync.WaitGroup, resultChan chan<- string) {
	defer wg.Done()

	log.Printf("Fetching %v of %v feeds: %v", i, n, url)

	resp, err := http.Get(url)
	if err != nil {
		if resp != nil {
			log.Printf("(%v/%v) Status code: %v. Failed to fetch feed at %v: %v", i, n, resp.StatusCode, url, err)
		} else {
			log.Printf("(%v/%v) Failed to fetch feed at %v: %v", i, n, url, err)
		}
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("(%v/%v) Failed to read resp body: %v", i, n, err)
		return
	}

	result := feedXMLStruct{}
	if err := xml.Unmarshal(data, &result); err != nil {
		log.Printf("(%v/%v) Failed to unmarshal resp data: %v", i, n, err)
		return
	}

	resultChan <- fmt.Sprintf("(%v/%v) Not yet implemented", i, n)
	//resultChan <- result
}

func (cfg *apiConfig) getNextFeeds(limit int32) []Feed {
	feeds, err := cfg.DB.GetNextFeedsToFetch(context.TODO(), limit)
	if err != nil {
		log.Printf("Failed to fetch feeds: %v", err.Error())
		return make([]Feed, 0)
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

	result := databaseFeedsToFeeds(feeds)
	for i := range result {
		result[i].Url = "https://wagslane.dev/index.xml"
	}
	return result
	//return databaseFeedsToFeeds(feeds)
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
