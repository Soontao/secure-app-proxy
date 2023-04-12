package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiterMiddleware(t *testing.T) {
	os.Setenv("RATE_LIMIT", "10-M")

	// Create a new RateLimiterMiddleware instance
	rlm := NewRateLimiterMiddleware()

	// Create a new HTTP request to test the middleware
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new HTTP recorder to capture the response from the middleware
	rr := httptest.NewRecorder()

	// Create a new HTTP handler that just writes a 200 response
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Call the RateLimiterMiddleware.Handler method with our handler
	middleware := rlm.Handler(handler)

	// Call ServeHTTP on the middleware with our request and recorder
	middleware.ServeHTTP(rr, req)

	// Check that the response status code is 200 (the rate limit should not have been reached)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check that the headers are set correctly
	assert.Equal(t, "10", rr.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "9", rr.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, rr.Header().Get("X-RateLimit-Reset"))
}

func TestRateLimiterMiddleware_Handler_RateLimitReached(t *testing.T) {
	// Create a new RateLimiterMiddleware instance with a mock store and rate limit of 1 request per second
	os.Setenv("RATE_LIMIT", "1-S")
	middleware := NewRateLimiterMiddleware()

	// Create a new HTTP request with a mock handler
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// Call the Handler function twice to exceed the rate limit
	middleware.Handler(handler).ServeHTTP(httptest.NewRecorder(), req)
	rr := httptest.NewRecorder()
	middleware.Handler(handler).ServeHTTP(rr, req)

	// Check if the response status code is correct
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}

	// Check if the response body is correct
	expectedBody := `{"ErrorMessage":"Rate Limit Reached","Code":"RATE_LIMIT_REACH"}`
	if !jsonEqual(rr.Body.Bytes(), []byte(expectedBody)) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expectedBody)
	}
}

// jsonEqual checks if two JSON byte arrays are equal
func jsonEqual(a, b []byte) bool {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false
	}
	return reflect.DeepEqual(j, j2)
}
