package main

import (
	"net/http"
	"net/http/httptest"
	"os"
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
