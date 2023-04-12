package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestNewJwtMiddleware_ValidToken(t *testing.T) {
	// Generate a valid JWT token with a UUID secret
	secret := uuid.New().String()
	claims := jwt.RegisteredClaims{
		Subject: "123",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Set the JWT_SECRET environment variable to the UUID secret
	t.Setenv("JWT_SECRET", secret)

	// Create a new JwtMiddleware
	middleware := NewJwtMiddleware()

	// Create a test request with the valid JWT token in the Authorization header
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Call the Handler method of the middleware with the test request and response recorder
	middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the X-User-Subject header is set to the subject of the JWT token
		subject := r.Context().Value("X-User-Subject").(string)
		if subject != claims.Subject {
			t.Errorf("X-User-Subject header is incorrect: got %v, want %v", subject, claims.Subject)
		}
	})).ServeHTTP(rr, req)

	// Check that the response status code is 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, want %v", status, http.StatusOK)
	}

	// Check that the response body is empty
	if body := rr.Body.String(); body != "" {
		t.Errorf("Handler returned non-empty body: %v", body)
	}
}

func TestJwtMiddleware_Handler_InvalidToken(t *testing.T) {
	// Create a mock HTTP handler
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, world!"))
	})

	// Create an instance of the JwtMiddleware struct
	jwtMiddleware := NewJwtMiddleware()

	// Call the Handler method with the mock handler
	handler := jwtMiddleware.Handler(mockHandler)

	// Create a request with an invalid JWT token
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Make the request to the handler
	handler.ServeHTTP(rr, req)

	// Check that the response has a 401 status code and a message indicating that the JWT validation failed
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d but got %d", http.StatusUnauthorized, rr.Code)
	}
	expectedBody := `{"Code":"JWT_VALIDATE_FAILED","ErrorMessage":"token is malformed: token contains an invalid number of segments"}`
	if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(expectedBody) {
		t.Errorf("expected body %q but got %q", expectedBody, rr.Body.String())
	}
}
