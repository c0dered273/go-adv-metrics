package handler

import (
	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"
)

func metricHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newMetric, appError := metric.NewMetric(chi.URLParam(r, "name"), chi.URLParam(r, "type"), chi.URLParam(r, "value"))
		if appError.Error != nil && appError.TypeError {
			log.Error.Printf(appError.Error.Error())
			http.Error(w, "Unknown metric type", http.StatusNotImplemented)
			return
		}
		if appError.Error != nil && appError.ValueError {
			log.Error.Printf(appError.Error.Error())
			http.Error(w, "Bad request", http.StatusBadRequest)
		}

		savedMetric, err := repository.Save(newMetric)
		if err != nil {
			log.Error.Printf("Can`t save metric: %v", savedMetric)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func Service() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Post("/update/{type}/{name}/{value}", metricHandler(storage.NewMemStorage()))

	return r
}
