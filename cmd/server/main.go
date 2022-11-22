package main

import (
	"github.com/c0dered273/go-adv-metrics/internal/handler"
	"net/http"
	syslog "log"
)

func main() {
	http.HandleFunc("/", handler.DefaultNotFoundHandler)
	http.HandleFunc("/update/",handler.RequestLoggerHandler)
	syslog.Fatal(http.ListenAndServe(":8080", nil))
}
