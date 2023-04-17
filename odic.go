package main

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type OidcMiddleware struct {
	provider *oidc.Provider
	conf     *oauth2.Config
	enabled  bool
}

func NewOdicMiddleware() *OidcMiddleware {
	return &OidcMiddleware{
		enabled: len(os.Getenv("ODIC_CLIENT_ID")) > 0 && len(os.Getenv("ODIC_CLIENT_SECRET")) > 0,
	}
}

// VerifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
func (m *OidcMiddleware) VerifyIDToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: m.conf.ClientID,
	}

	return m.provider.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}

func (m *OidcMiddleware) Name() string {
	return "OidcMiddleware"
}

func (m *OidcMiddleware) Enabled() bool {
	return m.enabled
}

func (m *OidcMiddleware) Handler(next http.Handler) http.Handler {
	m.provider, _ = oidc.NewProvider(
		context.Background(),
		os.Getenv("ODIC_ISSUER"),
	)
	m.conf = &oauth2.Config{
		ClientID:     os.Getenv("ODIC_CLIENT_ID"),
		ClientSecret: os.Getenv("ODIC_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("ODIC_CALLBACK_URL"),
		Endpoint:     m.provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}
	store := sessions.NewCookieStore([]byte(os.Getenv("ODIC_SESSION_SECRET")))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, err := store.Get(r, "user")

		if err != nil {
			flushJsonErrorResponse(w, "session handling failed", "ERR_SESSION", http.StatusBadRequest)
			return
		}

		if r.URL.Path == "/_/oidc/callback" {
			m.handleCallback(s, r, w)
			return
		}

		if s.IsNew || s.Values["token"] == nil {
			m.handleUnauthorized(s, r, w)
			return
		}
		r = r.WithContext(
			context.WithValue(r.Context(), "X-User-Subject", s.Values["profile_name"]),
		)
		next.ServeHTTP(w, r)
	})
}

func (m *OidcMiddleware) handleUnauthorized(s *sessions.Session, r *http.Request, w http.ResponseWriter) {
	newUUID, _ := uuid.NewRandom()
	stateId := newUUID.String()
	s.Values["odic_restore_url"] = r.URL.Path
	s.Values["oidc_state"] = stateId
	s.Save(r, w)
	s.Flashes()
	http.Redirect(w, r, m.conf.AuthCodeURL(stateId), http.StatusTemporaryRedirect)
}

func (m *OidcMiddleware) handleCallback(s *sessions.Session, r *http.Request, w http.ResponseWriter) {
	if s.Values["oidc_state"] != r.URL.Query().Get("state") {
		flushJsonErrorResponse(
			w,
			"OIDC state mismatch, avoid security issue we rejected your request", "ERR_OIDC_STATE_MISMATCH",
			http.StatusBadRequest,
		)
		return
	}
	token, err := m.conf.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		flushJsonErrorResponse(
			w,
			err.Error(),
			"ERR_OIDC_AUTH_FAILED",
			http.StatusUnauthorized,
		)
		return
	}
	idToken, err := m.VerifyIDToken(r.Context(), token)
	if err != nil {
		flushJsonErrorResponse(
			w,
			err.Error(),
			"ERR_OIDC_AUTH_FAILED",
			http.StatusUnauthorized,
		)
		return
	}

	profile := map[string]interface{}{}

	if err := idToken.Claims(&profile); err != nil {
		flushJsonErrorResponse(
			w,
			err.Error(),
			"ERR_OIDC_AUTH_RETRIEVE_PROFILE_FAILED",
			http.StatusUnauthorized,
		)
		return
	}
	s.Values["profile_name"] = profile["name"]
	s.Values["profile_email"] = profile["email"]
	s.Values["token"] = token.AccessToken
	if err := s.Save(r, w); err != nil {
		flushHttpResponseError(w, err.Error(), "ERR_SAVE_SESSION_FAILED")
		return
	}
	http.Redirect(
		w,
		r,
		s.Values["odic_restore_url"].(string),
		http.StatusTemporaryRedirect,
	)
}
