package handler

import (
	"encoding/json"
	"net/http"
)

// Ping - публичный эндпоинт
func Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "pong"}); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}

// SecurePing - защищенный эндпоинт
func SecurePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "pong-secure"}); err != nil {
		http.Error(w, "failed to write response", http.StatusInternalServerError)
	}
}
