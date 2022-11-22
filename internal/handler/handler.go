package handler

import (
	"github.com/c0dered273/go-adv-metrics/internal/log"
	"net/http"
)

func DefaultNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, ``, http.StatusNotFound)
	log.Error.Printf("Path not found %v", r)

}

func RequestLoggerHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Allowed only POST method", http.StatusBadRequest)
		log.Error.Printf("Not alowed method %v", r)
		return
	}
	log.Info.Printf("Incomming request %v", r)
}