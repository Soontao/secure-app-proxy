package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

func TestNewOdicMiddleware(t *testing.T) {
	// Test when ODIC_CLIENT_ID and ODIC_CLIENT_SECRET are not set
	t.Setenv("ODIC_CLIENT_ID", "")
	t.Setenv("ODIC_CLIENT_SECRET", "")
	m := NewOdicMiddleware()
	if m.enabled {
		t.Errorf("Expected enabled to be false, but got true")
	}

	// Test when ODIC_CLIENT_ID and ODIC_CLIENT_SECRET are set
	t.Setenv("ODIC_CLIENT_ID", "client_id")
	t.Setenv("ODIC_CLIENT_SECRET", "client_secret")
	m = NewOdicMiddleware()
	if !m.enabled {
		t.Errorf("Expected enabled to be true, but got false")
	}
}

func TestVerifyIDToken(t *testing.T) {
	// Create a mock oidc provider
	provider, err := oidc.NewProvider(context.Background(), "https://theosunz.eu.auth0.com/")
	if err != nil {
		t.Fatalf("Failed to create mock oidc provider: %v", err)
	}

	// Create a mock oauth2 config
	conf := &oauth2.Config{
		ClientID:     "client_id",
		ClientSecret: "client_secret",
		RedirectURL:  "http://localhost:8080/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}

	// Create a mock oidc middleware
	m := &OidcMiddleware{
		provider: provider,
		conf:     conf,
	}

	// Create a mock oauth2 token
	token := &oauth2.Token{
		AccessToken: "access_token",
		Expiry:      time.Now(),
	}

	// Test when id_token field is not present in oauth2 token
	_, err = m.VerifyIDToken(context.Background(), token)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	// Test when id_token field is present in oauth2 token
	token = &oauth2.Token{
		AccessToken: "access_token",
		Expiry:      time.Now(),
	}
	token = token.WithExtra(map[string]interface{}{
		"id_token": "",
	})
	_, err = m.VerifyIDToken(context.Background(), token)
	if err == nil {
		t.Errorf("Expected error")
	}
}

func TestName(t *testing.T) {
	m := &OidcMiddleware{}
	if m.Name() != "OidcMiddleware" {
		t.Errorf("Expected Name to be OidcMiddleware, but got %s", m.Name())
	}
}

func TestEnabled(t *testing.T) {
	// Test when ODIC_CLIENT_ID and ODIC_CLIENT_SECRET are not set
	t.Setenv("ODIC_CLIENT_ID", "")
	t.Setenv("ODIC_CLIENT_SECRET", "")
	m := &OidcMiddleware{}
	if m.Enabled() {
		t.Errorf("Expected Enabled to be false, but got true")
	}

	// Test when ODIC_CLIENT_ID and ODIC_CLIENT_SECRET are set
	t.Setenv("ODIC_CLIENT_ID", "client_id")
	t.Setenv("ODIC_CLIENT_SECRET", "client_secret")
	m = NewOdicMiddleware()
	if !m.Enabled() {
		t.Errorf("Expected Enabled to be true, but got false")
	}
}

func TestHandler(t *testing.T) {
	t.Setenv("ODIC_SESSION_SECRET", "aaaaaaa")
	// Create a mock oidc provider
	t.Setenv("ODIC_ISSUER", "https://theosunz.eu.auth0.com/")

	provider, err := oidc.NewProvider(context.Background(), "https://theosunz.eu.auth0.com/")

	if err != nil {
		t.Fatalf("Failed to create mock oidc provider: %v", err)
	}

	// Create a mock oauth2 config
	conf := &oauth2.Config{
		ClientID:     "client_id",
		ClientSecret: "client_secret",
		RedirectURL:  "http://localhost:8080/callback",
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}

	// Create a mock oidc middleware
	m := &OidcMiddleware{
		provider: provider,
		conf:     conf,
	}

	// Create a mock http handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test when session handling fails
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create mock http request: %v", err)
	}
	rr := httptest.NewRecorder()
	m.Handler(handler).ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status code %d, but got %d", http.StatusTemporaryRedirect, rr.Code)
	}

	// Test when user is not authenticated
	store := sessions.NewCookieStore([]byte("session_secret"))
	req, err = http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create mock http request: %v", err)
	}
	rr = httptest.NewRecorder()
	s, _ := store.Get(req, "user")
	m.Handler(handler).ServeHTTP(rr, req)
	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status code %d, but got %d", http.StatusTemporaryRedirect, rr.Code)
	}
	if s.Values["oidc_state"] == "" {
		t.Errorf("Expected oidc_state to be set, but got empty string")
	}

}

func TestOidcMiddleware_handleCallback(t *testing.T) {
	// create a new OidcMiddleware instance
	m := NewOdicMiddleware()

	// create a new session
	s := sessions.NewSession(nil, "user")

	// create a new request
	req, err := http.NewRequest("GET", "/_/oidc/callback", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// create a new response recorder
	rr := httptest.NewRecorder()

	// call handleCallback function
	m.handleCallback(s, req, rr)

	// check if the response status code is 400
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status code %d, but got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestOidcMiddleware_HandleCallback_AuthFailed(t *testing.T) {
	// Create a new OidcMiddleware instance
	m := NewOdicMiddleware()

	// Create a new HTTP request with a query parameter that will cause the auth to fail
	req, err := http.NewRequest("GET", "/_/oidc/callback?code=invalid_code&state=111", nil)
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %v", err)
	}

	// Create a new HTTP response recorder
	rr := httptest.NewRecorder()

	store := sessions.NewCookieStore()

	session := sessions.NewSession(store, "user")
	session.Values["oidc_state"] = "111"

	m.conf = &oauth2.Config{}
	// Call the handleCallback function with a new session and the invalid request
	m.handleCallback(session, req, rr)

	// Check that the response status code is 401 Unauthorized
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, but got %d", http.StatusUnauthorized, rr.Code)
	}

	// Check that the response body contains the expected error message and code
}
