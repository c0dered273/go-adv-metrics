package handler

import (
	"html/template"
	"net/http"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/c0dered273/go-adv-metrics/internal/service"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ServerConfig struct {
	Repo *storage.MemStorage
}

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

func metricStore(persist service.PersistMetric) http.HandlerFunc {
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

		err := persist.SaveMetric(newMetric)
		if err != nil {
			log.Error.Println("Can`t save metric ", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func metricLoad(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mName := chi.URLParam(r, "name")
		mType := chi.URLParam(r, "type")
		keyMetric, appErr := metric.NewMetric(mName, mType, "0")
		if appErr.Error != nil {
			log.Error.Println(appErr.Error)
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		tmpMetric, findErr := repository.FindByID(keyMetric)
		if findErr != nil {
			log.Error.Printf("Metric not found with name: %v, type: %v", mName, mType)
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		_, wrErr := w.Write([]byte(tmpMetric.GetStringValue()))
		if wrErr != nil {
			log.Error.Println(wrErr.Error())
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func Service(config ServerConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/", rootHandler(config.Repo))
	r.Post("/update/{type}/{name}/{value}", metricStore(service.PersistMetric{Repo: config.Repo}))
	r.Get("/value/{type}/{name}", metricLoad(config.Repo))

	return r
}
