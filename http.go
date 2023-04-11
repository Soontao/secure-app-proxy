package main

import (
	"encoding/json"
	"net/http"
)

func flushHttpResponseError(w http.ResponseWriter, errMessage string, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(&ErrorMessage{
		ErrorMessage: errMessage,
		Code:         code,
	})
}
