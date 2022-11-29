package handler

import (
	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"html/template"
	"net/http"
	"time"
)

type IndexData struct {
	Title   string
	Metrics []string
}

func rootHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		indexTemplate, err := template.ParseFiles("templates/index.html")
		if err != nil {
			log.Error.Fatal(err)
		}

		allMetrics, _ := repository.FindAll()
		mtr := make([]string, len(allMetrics))
		for i := 0; i < len(allMetrics); i++ {
			mtr[i] = allMetrics[i].String()
		}
		indexData := IndexData{
			Title:   "All metrics",
			Metrics: mtr,
		}
		err = indexTemplate.Execute(w, indexData)
		if err != nil {
			log.Error.Println("Can`t execute templates", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func metricSetHandler(repository storage.Repository) http.HandlerFunc {
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
			return
		}

		savedMetric, err := repository.Save(newMetric)
		if err != nil {
			log.Error.Printf("Can`t save metric: %v", savedMetric)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func metricGetHandler(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mName := chi.URLParam(r, "name")
		mType := chi.URLParam(r, "type")
		allMetrics, _ := repository.FindAll()
		var value string
		for _, m := range allMetrics {
			if mName == m.GetName() && mType == m.GetType().String() {
				value = m.GetStringValue()
			}
		}
		if len(value) == 0 {
			log.Error.Printf("Metric not found name: %v, type: %v", mName, mType)
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		_, err := w.Write([]byte(value))
		if err != nil {
			log.Error.Println(err.Error())
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

	r.Get("/", rootHandler(storage.GetMemStorage()))
	r.Post("/update/{type}/{name}/{value}", metricSetHandler(storage.GetMemStorage()))
	r.Get("/value/{type}/{name}", metricGetHandler(storage.GetMemStorage()))

	return r
}
