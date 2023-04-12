package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	t.Setenv("RATE_LIMIT", "60-S")
	t.Setenv("LISTEN_ADDR", "127.0.0.1:43533")
	t.Setenv("UPSTREAM", "https://httpbin.org/")
	log.SetOutput(ioutil.Discard)
	// Call the main function, which will set up the server and handle the request.
	go main()

	// Wait for the server to start listening on port 8080.
	// This is necessary because the main function runs in a separate goroutine.
	for {
		_, err := http.Get("http://localhost:43533/get")
		time.Sleep(1 * time.Second)
		if err == nil {
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
