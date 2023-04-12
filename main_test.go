package main

import (
	"net/http"
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	os.Setenv("LISTEN_ADDR", "127.0.0.1:43533")
	os.Setenv("UPSTREAM", "https://httpbin.org/")

	// Call the main function, which will set up the server and handle the request.
	go main()

	// Wait for the server to start listening on port 8080.
	// This is necessary because the main function runs in a separate goroutine.
	for {
		resp, err := http.Get("http://localhost:43533/get")
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
	}

	// Send the request to the server.
	resp, err := http.Get("http://localhost:43533/get")
	if err != nil {
		t.Fatal(err)
	}

	// Check the status code of the response.
	if status := resp.StatusCode; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
