package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/JustinLi007/rss-aggregator/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Language    string    `xml:"language"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func initScraping(db *database.Queries, concurrency int, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(context.Background(), int32(concurrency))
		if err != nil {
			log.Printf("Failed to fetch feeds: %v", err.Error())
			continue
		}
		log.Printf("Found %v feeds to fetch.", len(feeds))

		scrapeFeedWorker(databaseFeedsToFeeds(feeds), db)
	}
}

func scrapeFeedWorker(feeds []Feed, db *database.Queries) {
	wg := new(sync.WaitGroup)
	n := len(feeds)
	resultChan := make(chan RSSFeed, n)

	for _, v := range feeds {
		wg.Add(1)
		go scrapeFeed(v, db, wg, resultChan)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for rs := range resultChan {
		fmt.Println(rs.Channel.Title)
		for _, v := range rs.Channel.Item {
			fmt.Printf(" - %v\n", v.Title)
		}
		fmt.Println()
	}
}

func scrapeFeed(feed Feed, db *database.Queries, wg *sync.WaitGroup, resultChan chan<- RSSFeed) {
	defer wg.Done()

	if _, err := db.MarkFeedFetched(context.Background(), feed.ID); err != nil {
		log.Printf("Failed to mark feed %v as fetched: %v", feed.Name, err)
		return
	}

	rssFeed, err := fetchFeed(feed.Url)
	if err != nil {
		log.Printf("Failed to fetch feed %v: %v", feed.Name, err)
		return
	}

	resultChan <- *rssFeed
}

func fetchFeed(url string) (*RSSFeed, error) {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rssFeed := new(RSSFeed)
	if err := xml.Unmarshal(data, rssFeed); err != nil {
		return nil, err
	}

	return rssFeed, nil
}
