package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

func createRewriter() func(pr *httputil.ProxyRequest) {
	upstream := os.Getenv("UPSTREAM")
	if len(upstream) == 0 {
		log.Fatal("must provide upstream!")
	}
	u, err := url.Parse(upstream)
	if err != nil {
		log.Fatalf("parse upstream url failed %s", err)
	}
	log.Printf("upstream endpoint %s", upstream)

	rewriteSteps := []func(pr *httputil.ProxyRequest){}

	rewriteSteps = append(rewriteSteps, func(pr *httputil.ProxyRequest) {
		pr.SetURL(u)
	})
	// TODO: only jwt/auth enabled ?
	rewriteSteps = append(rewriteSteps, func(pr *httputil.ProxyRequest) {
		userSubject := pr.In.Context().Value("X-User-Subject")
		if userSubject != nil {
			pr.Out.Header.Set("X-User-Subject", userSubject.(string))
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
		if strings.HasPrefix(key, "DELETE_REQ_HEADERS") {
			delReqHeaders = append(delReqHeaders, strings.TrimPrefix(key, "DELETE_REQ_HEADERS_"))
		}
		if strings.HasPrefix(key, "APPEND_REQ_HEADERS") {
			setReqHeaders[strings.TrimPrefix(key, "APPEND_REQ_HEADERS_")] = value
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

func createModifier() func(r *http.Response) error {
	modifierSteps := []func(r *http.Response) error{}

	delResHeaders := []string{}
	setResHeaders := map[string]string{}

	for _, v := range os.Environ() {
		parts := strings.SplitN(v, "=", 2)
		key := parts[0]
		value := parts[1]
		if strings.HasPrefix(key, "DELETE_RES_HEADERS") {
			delResHeaders = append(delResHeaders, strings.TrimPrefix(key, "DELETE_RES_HEADERS_"))
		}
		if strings.HasPrefix(key, "APPEND_RES_HEADERS") {
			setResHeaders[strings.TrimPrefix(key, "APPEND_RES_HEADERS_")] = value
		}
	}

	if len(delResHeaders) > 0 {
		modifierSteps = append(modifierSteps, func(r *http.Response) error {
			for _, delHeader := range delResHeaders {
				r.Header.Del(delHeader)
			}
			return nil
		})
	}

	if len(setResHeaders) > 0 {
		modifierSteps = append(modifierSteps, func(r *http.Response) error {
			for setHeader, value := range setResHeaders {
				r.Header.Set(setHeader, value)
			}
			return nil
		})
	}

	return func(r *http.Response) error {
		for _, step := range modifierSteps {
			if err := step(r); err != nil {
				return err
			}
		}
		return nil
	}
}

func createProxyHandler() http.Handler {
	rp := &httputil.ReverseProxy{
		Rewrite:        createRewriter(),
		ModifyResponse: createModifier(),
	}
	return http.HandlerFunc(rp.ServeHTTP)
}
