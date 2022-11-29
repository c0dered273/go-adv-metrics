package handler

import (
	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func init() {
	routes = append(routes,
		newRoute("POST", "/update/[^/]+/[^/]+/[^/]+", updateMetricsHandler))
}

type handleFunc func(repository storage.Repository) http.HandlerFunc

var routes []route

type route struct {
	method  string
	regex   *regexp.Regexp
	handler handleFunc
}

func newRoute(method string, pattern string, handler handleFunc) route {
	return route{method, regexp.MustCompile("^" + pattern + "$"), handler}
}

func defaultNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `Not found`, http.StatusNotFound)
	log.Error.Printf("Path not found %v", r)
}

func updateMetricsHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathVars := strings.Split(r.URL.Path, "/")
		metricType, err := metric.ParseMetricType(pathVars[2])
		metricName := pathVars[3]
		if err != nil {
			log.Error.Printf("Unknown metric type: %v", metricType)
			http.Error(w, "Unknown metric type", http.StatusNotImplemented)
			return
		}
		_, err = strconv.ParseFloat(pathVars[4], 64)
		if err != nil {
			log.Error.Printf("Can`t parse metric value: %v", pathVars[4])
			http.Error(w, "Bad request", http.StatusBadRequest)
		}

		newMetric, err := repository.Save(metric.Metric{Name: metricName, Type: metricType, Value: pathVars[4]})
		if err != nil {
			log.Error.Printf("Can`t save metric: %v", newMetric)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		log.Info.Printf("Incoming request %v %v", r.Method, r.URL)
		w.WriteHeader(http.StatusOK)
	}
}

type MetricsHandler struct {
	Repository storage.Repository
}

func NewMetricHandler(repository storage.Repository) MetricsHandler {
	return MetricsHandler{
		Repository: repository,
	}
}

func (h MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var notAllowed []string
	for _, route := range routes {
		matches := route.regex.FindStringSubmatch(r.URL.Path)
		if len(matches) > 0 {
			if r.Method != route.method {
				notAllowed = append(notAllowed, r.Method)
				continue
			}
			route.handler(h.Repository)(w, r)
			return
		}
	}
	if len(notAllowed) > 0 {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Error.Printf("Method not allowed %v", r)
		return
	}
	defaultNotFoundHandler(w, r)
}
