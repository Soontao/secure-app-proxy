package main

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"testing"
)

func TestCreateRewriter(t *testing.T) {
	// Set up environment variables
	os.Setenv("UPSTREAM", "http://example.com")
	os.Setenv("APPEND_FORWARD_HEADERS", "true")
	os.Setenv("DELETE_REQ_HEADERS_FOO", "true")
	os.Setenv("APPEND_REQ_HEADERS_BAR", "baz")

	// Call createRewriter function
	rewriter := createRewriter()

	// Create a mock ProxyRequest
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.RemoteAddr = "localhost:8080"
	ctx := context.WithValue(req.Context(), "X-User-Subject", "user123")
	req = req.WithContext(ctx)

	pr := &httputil.ProxyRequest{
		In:  req,
		Out: &http.Request{Header: http.Header{}, URL: &url.URL{}},
	}

	// Call rewriter function on the mock ProxyRequest
	rewriter(pr)

	// Check if URL was set correctly
	if pr.In.URL.String() != "http://example.com" {
		t.Errorf("Expected URL to be http://example.com, but got %s", pr.In.URL.String())
	}

	// Check if X-User-Subject header was set correctly
	if pr.Out.Header.Get("X-User-Subject") != "user123" {
		t.Errorf("Expected X-User-Subject header to be user123, but got %s", pr.Out.Header.Get("X-User-Subject"))
	}

	// Check if X-Forwarded-* headers were appended correctly
	if pr.Out.Header.Get("X-Forwarded-For") == "" || pr.Out.Header.Get("X-Forwarded-Proto") == "" {
		t.Error("Expected X-Forwarded-* headers to be appended, but they were not")
	}

	// Check if FOO header was deleted
	if pr.Out.Header.Get("FOO") != "" {
		t.Error("Expected FOO header to be deleted, but it was not")
	}

	// Check if BAR header was appended with value "baz"
	if pr.Out.Header.Get("BAR") != "baz" {
		t.Errorf("Expected BAR header to be appended with value baz, but got %s", pr.Out.Header.Get("BAR"))
	}
}

func TestCreateModifier(t *testing.T) {

	// Test case 2: DELETE_RES_HEADERS environment variable set
	os.Setenv("DELETE_RES_HEADERS_FOO", "true")
	modifier := createModifier()
	resp := &http.Response{Header: http.Header{"Foo": []string{"bar"}}}
	err := modifier(resp)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if resp.Header.Get("Foo") != "" {
		t.Errorf("Expected header Foo to be deleted, but it still exists")
	}

	// Test case 3: APPEND_RES_HEADERS environment variable set
	os.Setenv("APPEND_RES_HEADERS_FOO", "bar")
	modifier = createModifier()
	resp = &http.Response{Header: http.Header{}}
	err = modifier(resp)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if resp.Header.Get("Foo") != "bar" {
		t.Errorf("Expected header Foo to be set to bar, but got %s", resp.Header.Get("Foo"))
	}

	// Test case 4: Both DELETE_RES_HEADERS and APPEND_RES_HEADERS environment variables set
	os.Unsetenv("APPEND_RES_HEADERS_FOO")
	os.Setenv("DELETE_RES_HEADERS_FOO", "true")
	os.Setenv("APPEND_RES_HEADERS_BAR", "baz")
	modifier = createModifier()
	resp = &http.Response{Header: http.Header{"Foo": []string{"bar"}}}
	err = modifier(resp)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if resp.Header.Get("Foo") != "" {
		t.Errorf("Expected header Foo to be deleted, but it still exists")
	}
	if resp.Header.Get("Bar") != "baz" {
		t.Errorf("Expected header Bar to be set to baz, but got %s", resp.Header.Get("Bar"))
	}
}
