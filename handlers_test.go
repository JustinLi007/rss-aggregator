package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type createFeedPayload struct {
	Feed       Feed             `json:"feed"`
	FeedFollow UsersFeedsFollow `json:"feed_follow"`
}

var testUserName = "Sample User"
var testFeedName = "Sample Feed"
var testFeedURL = "www.url.com"
var testUser = User{}
var testFeedCreatePayload = createFeedPayload{}

func executeRequest(c *http.Client, r *http.Request) ([]byte, error) {
	resp, err := c.Do(r)
	if err != nil {
		return make([]byte, 0), errors.New(fmt.Sprintf("Failed to complete request: %v", err))
	}
	defer resp.Body.Close()

	if status := resp.StatusCode; status != http.StatusOK {
		return make([]byte, 0), errors.New(fmt.Sprintf("Expected status code %v, got %v", http.StatusOK, status))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return make([]byte, 0), errors.New(fmt.Sprintf("Failed to read resp body: %v", err))
	}

	return data, nil
}

func userIsSet(actual User, expectedUserName string, t *testing.T) {
	if err := uuid.Validate(actual.ID.String()); err != nil {
		t.Errorf("Malformed UUID: %v", err)
	}
	if actual.CreatedAt.IsZero() {
		t.Errorf("Created_at unset")
	}
	if actual.UpdatedAt.IsZero() {
		t.Errorf("Updated_at unset")
	}
	if actual.Name != expectedUserName {
		t.Errorf("Expected name %v, got %v", expectedUserName, actual.Name)
	}
	if actual.ApiKey == "" {
		t.Errorf("ApiKey unset")
	}
}

func feedIsSet(actual Feed, user User, expectedFeedName, expectedURL string, t *testing.T) {
	if err := uuid.Validate(actual.ID.String()); err != nil {
		t.Errorf("Malformed UUID: %v", err)
	}
	if actual.CreatedAt.IsZero() {
		t.Errorf("Created_at unset")
	}
	if actual.UpdatedAt.IsZero() {
		t.Errorf("Updated_at unset")
	}
	if actual.Name != expectedFeedName {
		t.Errorf("Expected name %v, got %v", expectedFeedName, actual.Name)
	}
	if actual.Url != expectedURL {
		t.Errorf("Expected url %v, got %v", expectedURL, actual.Url)
	}
	if actual.UserID.String() != user.ID.String() {
		t.Errorf("Expected user_id %v, got %v", user.ID, actual.UserID)
	}
}

func userFeedFollowIsSet(actual UsersFeedsFollow, feed Feed, user User, t *testing.T) {
	if err := uuid.Validate(actual.ID.String()); err != nil {
		t.Errorf("Malformed UUID: %v", err)
	}
	if actual.CreatedAt.IsZero() {
		t.Errorf("Created_at unset")
	}
	if actual.UpdatedAt.IsZero() {
		t.Errorf("Updated_at unset")
	}
	if actual.UserID.String() != user.ID.String() {
		t.Errorf("Expected user_id %v, got %v", user.ID, actual.UserID)
	}
	if actual.FeedID.String() != feed.ID.String() {
		t.Errorf("Expected feed_id %v, got %v", feed.ID, actual.UserID)
	}
}

func compareUser(expected User, actual User, t *testing.T) {
	if actual.ID.String() != expected.ID.String() {
		t.Errorf("Expected UUID %v, got %v", expected.ID, actual.ID)
	}
	if actual.CreatedAt.Compare(expected.CreatedAt) != 0 {
		t.Errorf("Expected created_at %v, got %v", expected.CreatedAt, actual.CreatedAt)
	}
	if actual.UpdatedAt.Compare(expected.UpdatedAt) != 0 {
		t.Errorf("Expected updated_at %v, got %v", expected.UpdatedAt, actual.UpdatedAt)
	}
	if actual.Name != expected.Name {
		t.Errorf("Expected name %v, got %v", expected.Name, actual.Name)
	}
	if actual.ApiKey != expected.ApiKey {
		t.Errorf("Expected ApiKey %v, got %v", expected.ApiKey, actual.ApiKey)
	}
}

func compareFeed(expected Feed, actual Feed, t *testing.T) {
	if actual.ID.String() != expected.ID.String() {
		t.Errorf("Expected UUID %v, got %v", expected.ID, actual.ID)
	}
	if actual.CreatedAt.Compare(expected.CreatedAt) != 0 {
		t.Errorf("Expected created_at %v, got %v", expected.CreatedAt, actual.CreatedAt)
	}
	if actual.UpdatedAt.Compare(expected.UpdatedAt) != 0 {
		t.Errorf("Expected updated_at %v, got %v", expected.UpdatedAt, actual.UpdatedAt)
	}
	if actual.Name != expected.Name {
		t.Errorf("Expected name %v, got %v", expected.Name, actual.Name)
	}
	if actual.Url != expected.Url {
		t.Errorf("Expected url %v, got %v", expected.Url, actual.Url)
	}
	if actual.UserID.String() != expected.UserID.String() {
		t.Errorf("Expected user_id %v, got %v", expected.UserID, actual.UserID)
	}
}

func TestCreateUsersEndpoint(t *testing.T) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	url := "http://localhost:8080/v1/users"
	reqPayload := fmt.Sprintf(`{"name":"%v"}`, testUserName)
	req, err := http.NewRequest("POST", url, strings.NewReader(reqPayload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	data, err := executeRequest(client, req)
	if err != nil {
		t.Fatalf(err.Error())
	}

	actual := User{}
	if err := json.Unmarshal(data, &actual); err != nil {
		t.Fatalf("Failed to unmarshal resp data: %v", err)
	}
	testUser = actual

	userIsSet(actual, testUserName, t)
}

func TestGetUsersEndpoint(t *testing.T) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	url := "http://localhost:8080/v1/users"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "ApiKey "+testUser.ApiKey)

	data, err := executeRequest(client, req)
	if err != nil {
		t.Fatalf(err.Error())
	}

	actual := User{}
	if err := json.Unmarshal(data, &actual); err != nil {
		t.Fatalf("Failed to unmarshal resp data: %v", err)
	}

	compareUser(testUser, actual, t)
}

func TestCreateFeedsEndpoint(t *testing.T) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	url := "http://localhost:8080/v1/feeds"
	reqPayload := fmt.Sprintf(`{"name":"%v", "url":"%v"}`, testFeedName, testFeedURL)
	req, err := http.NewRequest("POST", url, strings.NewReader(reqPayload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("ApiKey %v", testUser.ApiKey))

	data, err := executeRequest(client, req)
	if err != nil {
		t.Fatalf(err.Error())
	}

	respPayload := createFeedPayload{}
	if err := json.Unmarshal(data, &respPayload); err != nil {
		t.Fatalf("Failed to unmarshal resp data: %v", err)
	}
	testFeedCreatePayload = respPayload

	feedIsSet(respPayload.Feed, testUser, testFeedName, testFeedURL, t)
	userFeedFollowIsSet(respPayload.FeedFollow, respPayload.Feed, testUser, t)
}

func TestGetFeedsEndpoint(t *testing.T) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	url := "http://localhost:8080/v1/feeds"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	data, err := executeRequest(client, req)
	if err != nil {
		t.Fatalf(err.Error())
	}

	actual := []Feed{}
	if err := json.Unmarshal(data, &actual); err != nil {
		t.Fatalf("Failed to unmarshal resp data: %v", err)
	}

	if len(actual) <= 0 {
		t.Fatalf("Expected non-empty list, got %v", actual)
	}

	compareFeed(testFeedCreatePayload.Feed, actual[0], t)
}

func TestHealthzHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8080/v1/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(healthzHandler)

	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, status)
	}

	type payload struct {
		Status string `json:"status"`
	}

	expected, err := json.Marshal(payload{
		Status: "ok",
	})
	if err != nil {
		t.Fatal(err)
	}

	if recorder.Body.String() != string(expected) {
		t.Errorf("Expected body %v, got %v", string(expected), recorder.Body.String())
	}
}

func TestErrorHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/v1/err", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(errorHandler)

	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code %v, got %v", http.StatusInternalServerError, status)
	}

	type payload struct {
		Error string `json:"error"`
	}

	expected, err := json.Marshal(payload{
		Error: "Internal Server Error",
	})

	if recorder.Body.String() != string(expected) {
		t.Errorf("Expected body %v, got %v", string(expected), recorder.Body.String())
	}
}
