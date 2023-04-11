package main

import "net/http"

type ErrorMessage struct {
	Code         string
	ErrorMessage string
}

type Middleware interface {
	// Name the name of middleware
	Name() string
	Enabled() bool
	Handler(next http.Handler) http.Handler
}
