package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

func main() {
	u, err := url.Parse(os.Getenv("UPSTREAM"))
	if err != nil {
		log.Fatalf("parse upstream url failed %s", err)
	}
	log.Printf("upstream endpoint %s", u)
	rp := httputil.NewSingleHostReverseProxy(u)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rp.ServeHTTP(w, r)
	})
	fmt.Println("Listening on port: 8080")
	http.ListenAndServe(":8080", nil)
}
