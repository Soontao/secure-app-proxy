package main

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type JwtMiddleware struct {
	secret  string
	enabled bool
}

func NewJwtMiddleware() *JwtMiddleware {
	jwtSecret := os.Getenv("JWT_SECRET")
	return &JwtMiddleware{
		secret:  jwtSecret,
		enabled: len(jwtSecret) > 0,
	}
}
func (m *JwtMiddleware) Name() string {
	return "JwtMiddleware"
}

func (m *JwtMiddleware) Enabled() bool {
	return m.enabled
}

func (m *JwtMiddleware) parseToken(tokenText string) (*jwt.Token, error) {
	return jwt.Parse(tokenText, func(t *jwt.Token) (interface{}, error) {
		return []byte(m.secret), nil
	})
}

func (m *JwtMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenText := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		token, err := m.parseToken(tokenText)
		if err != nil {
			flushHttpResponseError(
				w,
				err.Error(),
				"JWT_VALIDATE_FAILED",
			)
			return
		}
		sub, _ := token.Claims.GetSubject()
		if len(sub) > 0 {
			r = r.WithContext(
				context.WithValue(r.Context(), "X-User-Subject", sub),
			)
		}
		next.ServeHTTP(w, r)
	})
}
