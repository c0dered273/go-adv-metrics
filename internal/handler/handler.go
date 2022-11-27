package handler

import (
	"fmt"
	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/metrics"
	"net/http"
	"regexp"
)

func init() {
	for _, m := range metrics.GetAllMetrics() {
		routes = append(routes, newRoute("POST", fmt.Sprintf("/update/%v/%v/[+-]?([0-9]*[.])?[0-9]+", m.MType.String(), m.Name), updateMetricsLoggerHandler))
	}
}

var routes []route

type route struct {
	method  string
	regex   *regexp.Regexp
	handler http.HandlerFunc
}

func newRoute(method string, pattern string, handler http.HandlerFunc) route {
	return route{method, regexp.MustCompile("^" + pattern + "$"), handler}
}

func defaultNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `Not found`, http.StatusNotFound)
	log.Error.Printf("Path not found %v", r)
}

func updateMetricsLoggerHandler(w http.ResponseWriter, r *http.Request) {
	log.Info.Printf("Incoming request %v %v", r.Method, r.URL)
	w.WriteHeader(http.StatusOK)
}

type MetricsHandler struct{}

func (h MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var notAllowed []string
	for _, route := range routes {
		matches := route.regex.FindStringSubmatch(r.URL.Path)
		if len(matches) > 0 {
			if r.Method != route.method {
				notAllowed = append(notAllowed, r.Method)
				continue
			}
			route.handler(w, r)
			return
		}
	}
	if len(notAllowed) > 0 {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Error.Printf("Not allowed method %v", r)
		return
	}
	defaultNotFoundHandler(w, r)
}
