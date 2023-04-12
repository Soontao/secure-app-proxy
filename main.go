package main

import (
	"log"
	"net/http"

	"github.com/samber/lo"
)

func main() {
	middlewares := []Middleware{
		NewJwtMiddleware(),
		NewRateLimiterMiddleware(),
	}
	log.Println("Listening on port: 8080")

	// major handler
	handler := createProxyHandler()

	// apply middlewares
	for _, middleware := range lo.Reverse(middlewares) {
		if middleware.Enabled() {
			log.Printf("middleware %s enabled", middleware.Name())
			handler = middleware.Handler(handler)
		}
	}

	http.ListenAndServe(":8080", handler)
}
