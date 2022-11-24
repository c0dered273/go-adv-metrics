package handler

import (
	"github.com/c0dered273/go-adv-metrics/internal/log"
	"net/http"
)

func DefaultNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `Not found`, http.StatusNotFound)
	log.Error.Printf("Path not found %v", r)
}

func RequestLoggerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Allowed only POST method", http.StatusBadRequest)
		log.Error.Printf("Not allowed method %v", r)
		return
	}
	if contentType := r.Header.Get("Content-Type"); contentType != "text/plain" {
		http.Error(w, "Allowed only Content-Type: text/plain", http.StatusBadRequest)
	}
	log.Info.Printf("Incoming request %v %v", r.Method, r.URL)
}
