package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ulule/limiter/v3"
	limiterMiddlewareHandler "github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	memory "github.com/ulule/limiter/v3/drivers/store/memory"
)

type RateLimiterMiddleware struct {
	store   limiter.Store
	rate    limiter.Rate
	enabled bool
}

func NewRateLimiterMiddleware() *RateLimiterMiddleware {
	rateLimit := os.Getenv("RATE_LIMIT")
	enabled := len(rateLimit) > 0
	if !enabled {
		return &RateLimiterMiddleware{
			enabled: enabled,
		}
	}
	rate, error := limiter.NewRateFromFormatted(rateLimit)
	if error != nil {
		log.Fatalf("%s is not a valid rate limit expression", rateLimit)
	}
	store := memory.NewStore()
	return &RateLimiterMiddleware{
		store:   store,
		rate:    rate,
		enabled: enabled,
	}
}

func (m *RateLimiterMiddleware) Name() string {
	return "RateLimiterMiddleware"
}

func (m *RateLimiterMiddleware) Enabled() bool {
	return m.enabled
}

func (m *RateLimiterMiddleware) Handler(next http.Handler) http.Handler {
	middleware := limiterMiddlewareHandler.NewMiddleware(
		limiter.New(
			m.store,
			m.rate,
			limiter.WithTrustForwardHeader(true),
		),
		limiterMiddlewareHandler.WithLimitReachedHandler(
			func(w http.ResponseWriter, r *http.Request) {
				flushHttpResponseError(
					w,
					"Rate Limit Reached",
					"RATE_LIMIT_REACH",
				)
			},
		),
	)
	return middleware.Handler(next)
}
