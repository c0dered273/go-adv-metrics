package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	middleware2 "github.com/c0dered273/go-adv-metrics/internal/middleware"
	"github.com/c0dered273/go-adv-metrics/internal/service"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

		allMetrics, _ := repository.FindAll(r.Context())
		mtr := make([]string, len(allMetrics))
		for i := 0; i < len(allMetrics); i++ {
			mtr[i] = allMetrics[i].String()
		}
		indexData := IndexData{
			Title:   "All metrics",
			Metrics: mtr,
		}
		w.Header().Set("Content-Type", "text/html")
		err = indexTemplate.Execute(w, indexData)
		if err != nil {
			log.Error.Println("can`t execute templates", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func connectionPingHandler(config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := config.Repo.Ping(); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func metricStore(persist service.PersistMetric) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newMetric, appError := metric.NewMetric(
			chi.URLParam(r, "name"), chi.URLParam(r, "type"), chi.URLParam(r, "value"), "")
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

		err := persist.SaveMetric(r.Context(), newMetric)
		if err != nil {
			log.Error.Println("can`t save metric ", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func metricJSONStore(persist service.PersistMetric, config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newMetric metric.Metric

		if decErr := json.NewDecoder(r.Body).Decode(&newMetric); decErr != nil {
			log.Error.Println("can`t unmarshall request ", decErr)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if !metric.IsValid(newMetric) {
			log.Error.Printf("invalid metric: %v", newMetric)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		ok, checkErr := newMetric.CheckHash(config.Key)
		if checkErr != nil {
			log.Error.Println("can`t check metric hash ", checkErr)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		if !ok {
			log.Error.Println("invalid metric hash")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		persistErr := persist.SaveMetric(r.Context(), newMetric)
		if persistErr != nil {
			log.Error.Println("can`t save metric ", persistErr)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func metricJSONLoad(config *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var keyMetric metric.Metric

		if decErr := json.NewDecoder(r.Body).Decode(&keyMetric); decErr != nil {
			log.Error.Println("can`t unmarshall request ", decErr)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		resultMetric, findErr := config.Repo.FindByID(r.Context(), keyMetric)
		if findErr != nil {
			log.Error.Printf("metric not found with id: %v, type: %v", keyMetric.ID, keyMetric.MType.String())
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}

		resultMetric.SetHash(config.Key)

		resultBody, marshErr := json.Marshal(resultMetric)
		if marshErr != nil {
			log.Error.Printf("can`t marshal struct: %v", resultMetric)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(resultBody)
		if err != nil {
			log.Error.Print(err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func metricLoad(repository storage.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mName := chi.URLParam(r, "name")
		mType := chi.URLParam(r, "type")
		keyMetric, appErr := metric.NewMetric(mName, mType, "0", "")
		if appErr.Error != nil {
			log.Error.Println(appErr.Error)
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		tmpMetric, findErr := repository.FindByID(r.Context(), keyMetric)
		if findErr != nil {
			log.Error.Printf("metric not found with id: %v, type: %v", mName, mType)
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

func Service(config *config.ServerConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware2.GzipResponseEncoder)
	r.Use(middleware2.GzipRequestDecoder)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/", rootHandler(config.Repo))
	r.Get("/ping", connectionPingHandler(config))
	r.Post("/value/", metricJSONLoad(config))
	r.Get("/value/{type}/{name}", metricLoad(config.Repo))
	r.Post("/update/", metricJSONStore(service.PersistMetric{Repo: config.Repo}, config))
	r.Post("/update/{type}/{name}/{value}", metricStore(service.PersistMetric{Repo: config.Repo}))

	return r
}
