package main

import (
	"log"
	"net/http"
	"os"

	"github.com/samber/lo"
)

func main() {

	middlewares := []Middleware{
		NewOdicMiddleware(),
		NewJwtMiddleware(),
		NewRateLimiterMiddleware(),
	}

	addr := os.Getenv("LISTEN_ADDR")

	if len(addr) == 0 {
		addr = ":8080"
	}

	// major handler
	handler := createProxyHandler()

	// apply middlewares
	for _, middleware := range lo.Reverse(middlewares) {
		if middleware.Enabled() {
			log.Printf("middleware %s is enabled", middleware.Name())
			handler = middleware.Handler(handler)
		}
	}

	log.Println("Listening on", addr)

	http.ListenAndServe(addr, handler)
}
