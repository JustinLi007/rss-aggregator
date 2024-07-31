package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthzHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/v1/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(healthzHandler)

	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("Expected status code %v, got %v\n", http.StatusOK, status)
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
		t.Errorf("Expected body %v, got %v\n", string(expected), recorder.Body.String())
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
		t.Errorf("Expected status code %v, got %v\n", http.StatusInternalServerError, status)
	}

	type payload struct {
		Error string `json:"error"`
	}

	expected, err := json.Marshal(payload{
		Error: "Internal Server Error",
	})

	if recorder.Body.String() != string(expected) {
		t.Errorf("Expected body %v, got %v\n", string(expected), recorder.Body.String())
	}
}
