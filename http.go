package main

import (
	"encoding/json"
	"net/http"
)

func flushJsonErrorResponse(w http.ResponseWriter, errMessage string, code string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(&ErrorMessage{
		ErrorMessage: errMessage,
		Code:         code,
	})
}

func flushHttpResponseError(w http.ResponseWriter, errMessage string, code string) {
	flushJsonErrorResponse(w, errMessage, code, http.StatusUnauthorized)
}
