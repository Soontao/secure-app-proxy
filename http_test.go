package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFlushHttpResponseError(t *testing.T) {
	// Create a new HTTP recorder to capture the response
	rr := httptest.NewRecorder()

	// Call the flushHttpResponseError function with our recorder
	flushHttpResponseError(rr, "Unauthorized", "UNAUTHORIZED")

	// Check that the response status code is 401 (Unauthorized)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("unexpected status code: got %d, want %d", rr.Code, http.StatusUnauthorized)
	}

	// Check that the Content-Type header is set to "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("unexpected Content-Type header: got %s, want application/json", ct)
	}

	// Decode the response body into an ErrorMessage struct
	var errMsg ErrorMessage
	if err := json.NewDecoder(rr.Body).Decode(&errMsg); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	// Check that the ErrorMessage and Code fields are set correctly
	if errMsg.ErrorMessage != "Unauthorized" {
		t.Errorf("unexpected ErrorMessage field: got %s, want Unauthorized", errMsg.ErrorMessage)
	}
	if errMsg.Code != "UNAUTHORIZED" {
		t.Errorf("unexpected Code field: got %s, want UNAUTHORIZED", errMsg.Code)
	}
}
