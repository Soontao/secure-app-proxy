package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/timandy/routine"
	limiter "github.com/ulule/limiter/v3"
	limiterMiddlewareHandler "github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	memory "github.com/ulule/limiter/v3/drivers/store/memory"
)

var globalSubject = routine.NewInheritableThreadLocal()

func jwtMiddleware(next http.Handler) http.Handler {
	jwtSecret := os.Getenv("JWT_SECRET")
	enableJwtVerification := len(jwtSecret) > 0
	if !enableJwtVerification {
		return next
	}
	log.Println("JWT verification enabled")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenText := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		token, err := jwt.Parse(tokenText, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(&ErrorMessage{
				Message: "JWT validate failed",
				Code:    "JWT_VALIDATE_FAILED",
				Error:   err.Error(),
			})
			return
		}
		sub, _ := token.Claims.GetSubject()
		if len(sub) > 0 {
			globalSubject.Set(sub)
		}
		next.ServeHTTP(w, r)
	})
}

func rateMiddleware(next http.Handler) http.Handler {
	rateLimit := os.Getenv("RATE_LIMIT")
	rateLimitEnabled := len(rateLimit) > 0
	if !rateLimitEnabled {
		return next
	}
	log.Println("rate limiter enabled", rateLimit)
	rate, _ := limiter.NewRateFromFormatted(rateLimit)
	store := memory.NewStore()
	middleware := limiterMiddlewareHandler.NewMiddleware(limiter.New(store, rate, limiter.WithTrustForwardHeader(true)))
	return middleware.Handler(next)
}

func createProxyHandler() http.Handler {
	upstream := os.Getenv("UPSTREAM")
	if len(upstream) == 0 {
		log.Fatal("must provide upstream!")
	}
	u, err := url.Parse(upstream)
	if err != nil {
		log.Fatalf("parse upstream url failed %s", err)
	}
	log.Printf("upstream endpoint %s", upstream)
	rp := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(u)
			if globalSubject.Get() != nil {
				pr.Out.Header.Set("X-Auth-Subject", globalSubject.Get().(string))
			}
		},
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		rp.ServeHTTP(w, r)
	}
	return http.HandlerFunc(handler)
}

func main() {
	log.Println("Listening on port: 8080")
	http.ListenAndServe(":8080", jwtMiddleware(rateMiddleware(createProxyHandler())))
}
