package main

import (
	"github.com/JustinLi007/rss-aggregator/internal/database"
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	ApiKey    string    `json:"api_key"`
}

type Feed struct {
	ID          uuid.UUID  `json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Name        string     `json:"name"`
	Url         string     `json:"url"`
	UserID      uuid.UUID  `json:"user_id"`
	LastFetchAt *time.Time `json:"last_fetched_at"`
}

type UsersFeedsFollow struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	FeedID    uuid.UUID `json:"feed_id"`
}

func databaseUserToUser(user database.User) User {
	return User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Name:      user.Name,
		ApiKey:    user.ApiKey,
	}
}

func databaseFeedToFeed(feed database.Feed) Feed {
	result := Feed{}
	result.ID = feed.ID
	result.CreatedAt = feed.CreatedAt
	result.UpdatedAt = feed.CreatedAt
	result.Name = feed.Name
	result.Url = feed.Url
	result.UserID = feed.UserID
	result.LastFetchAt = nil
	if feed.LastFetchedAt.Valid {
		result.LastFetchAt = &feed.LastFetchedAt.Time
	}

	return result
}

func databaseFeedsToFeeds(feeds []database.Feed) []Feed {
	result := make([]Feed, len(feeds))
	for i, v := range feeds {
		result[i] = databaseFeedToFeed(v)
	}
	return result
}

func databaseFeedFollowToFeedFollow(feedFollow database.UsersFeedsFollow) UsersFeedsFollow {
	return UsersFeedsFollow{
		ID:        feedFollow.ID,
		CreatedAt: feedFollow.CreatedAt,
		UpdatedAt: feedFollow.UpdatedAt,
		UserID:    feedFollow.UserID,
		FeedID:    feedFollow.FeedID,
	}
}

func databaseFeedFollowsToFeedFollows(feedFollows []database.UsersFeedsFollow) []UsersFeedsFollow {
	result := make([]UsersFeedsFollow, len(feedFollows))
	for i, v := range feedFollows {
		result[i] = databaseFeedFollowToFeedFollow(v)
	}
	return result
}
