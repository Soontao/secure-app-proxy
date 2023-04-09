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
)

func main() {
	upstream := os.Getenv("UPSTREAM")
	if len(upstream) == 0 {
		log.Fatal("must provide upstream!")
	}
	u, err := url.Parse(upstream)
	if err != nil {
		log.Fatalf("parse upstream url failed %s", err)
	}
	log.Printf("upstream endpoint %s", upstream)

	jwtSecret := os.Getenv("JWT_SECRET")
	enableJwtVerification := len(jwtSecret) > 0
	if enableJwtVerification {
		log.Println("JWT verification enabled")
	}

	rp := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) { pr.SetURL(u) },
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if enableJwtVerification {
			tokenText := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			_, err := jwt.Parse(tokenText, func(t *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(&map[string]string{
					"error":  "jwt validate failed",
					"detail": err.Error(),
				})
				return
			}
		}

		rp.ServeHTTP(w, r)
	})

	log.Println("Listening on port: 8080")
	http.ListenAndServe(":8080", nil)
}
