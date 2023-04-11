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

// threadLocalSubject store the current request subject information - user
var threadLocalSubject = routine.NewInheritableThreadLocal()

func flushHttpResponseError(w http.ResponseWriter, errMessage string, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(&ErrorMessage{
		ErrorMessage: errMessage,
		Code:         code,
	})
}

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
			flushHttpResponseError(
				w,
				err.Error(),
				"JWT_VALIDATE_FAILED",
			)
			return
		}
		sub, _ := token.Claims.GetSubject()
		if len(sub) > 0 {
			threadLocalSubject.Set(sub)
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
	middleware := limiterMiddlewareHandler.NewMiddleware(
		limiter.New(
			store,
			rate,
			limiter.WithTrustForwardHeader(true),
		),
		limiterMiddlewareHandler.WithLimitReachedHandler(func(w http.ResponseWriter, r *http.Request) {
			flushHttpResponseError(
				w,
				"Rate Limit Reached",
				"RATE_LIMIT_REACH",
			)
		}),
	)
	return middleware.Handler(next)
}

func createRewriter(u *url.URL) func(pr *httputil.ProxyRequest) {
	rewriteSteps := []func(pr *httputil.ProxyRequest){}

	rewriteSteps = append(rewriteSteps, func(pr *httputil.ProxyRequest) {
		pr.SetURL(u)
	})
	// TODO: only jwt/auth enabled ?
	rewriteSteps = append(rewriteSteps, func(pr *httputil.ProxyRequest) {
		if threadLocalSubject.Get() != nil {
			pr.Out.Header.Set("X-Auth-Subject", threadLocalSubject.Get().(string))
		}
	})
	if os.Getenv("APPEND_FORWARD_HEADERS") != "false" {
		rewriteSteps = append(rewriteSteps, func(pr *httputil.ProxyRequest) {
			pr.SetXForwarded()
		})
	}

	// >> prepare header rewrite
	delReqHeaders := []string{}
	setReqHeaders := map[string]string{}

	for _, v := range os.Environ() {
		parts := strings.SplitN(v, "=", 2)
		key := parts[0]
		value := parts[1]
		if strings.HasPrefix(key, "DELETE_SOURCE_HEADERS") {
			delReqHeaders = append(delReqHeaders, strings.TrimPrefix(key, "DELETE_SOURCE_HEADERS_"))
		}
		if strings.HasPrefix(key, "APPEND_CUSTOM_HEADERS") {
			setReqHeaders[strings.TrimPrefix(key, "APPEND_CUSTOM_HEADERS_")] = value
		}
	}

	if len(delReqHeaders) > 0 {
		rewriteSteps = append(rewriteSteps, func(pr *httputil.ProxyRequest) {
			for _, delReqHeader := range delReqHeaders {
				pr.Out.Header.Del(delReqHeader)
			}
		})
	}

	if len(setReqHeaders) > 0 {
		rewriteSteps = append(rewriteSteps, func(pr *httputil.ProxyRequest) {
			for setReqHeader, value := range setReqHeaders {
				pr.Out.Header.Set(setReqHeader, value)
			}
		})
	}

	return func(pr *httputil.ProxyRequest) {
		for _, step := range rewriteSteps {
			step(pr)
		}
	}
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
		Rewrite: createRewriter(u),
	}
	return http.HandlerFunc(rp.ServeHTTP)
}

func main() {
	log.Println("Listening on port: 8080")
	http.ListenAndServe(":8080", jwtMiddleware(rateMiddleware(createProxyHandler())))
}
