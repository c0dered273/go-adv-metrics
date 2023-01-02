package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/metric"
	middleware2 "github.com/c0dered273/go-adv-metrics/internal/middleware"
	"github.com/c0dered273/go-adv-metrics/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type IndexData struct {
	Title   string
	Metrics []string
}

func rootHandler(c *service.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		indexTemplate, err := template.ParseFiles("templates/index.html")
		if err != nil {
			c.Logger.Fatal().Err(err).Msg("handler: failed parse template file")
		}

		allMetrics, _ := c.Repo.FindAll(r.Context())
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
			c.Logger.Error().Err(err).Msg("handler: can`t execute templates")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func connectionPingHandler(c *service.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := c.Repo.Ping(); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func metricStore(c *service.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newMetric, appError := metric.NewMetric(
			chi.URLParam(r, "name"), chi.URLParam(r, "type"), chi.URLParam(r, "value"), "")
		if appError.Error != nil && appError.TypeError {
			c.Logger.Error().Err(appError.Error).Msg("handler: unknown metric type")
			http.Error(w, "Unknown metric type", http.StatusNotImplemented)
			return
		}
		if appError.Error != nil && appError.ValueError {
			c.Logger.Error().Err(appError.Error).Msg("handler: wrong metric value")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		err := c.Repo.Save(r.Context(), newMetric)
		if err != nil {
			c.Logger.Error().Err(err).Msg("handler: failed to save metric")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func metricJSONStore(c *service.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newMetric metric.Metric

		if err := json.NewDecoder(r.Body).Decode(&newMetric); err != nil {
			c.Logger.Error().Err(err).Msg("handler: failed to unmarshall request")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if !metric.IsValid(newMetric) {
			c.Logger.Error().Msgf("handler: invalid metric: %v", newMetric)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		ok, err := newMetric.CheckHash(c.Key)
		if err != nil {
			c.Logger.Error().Err(err).Msg("handler: failed to check metric hash")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		if !ok {
			c.Logger.Error().Msg("handler: invalid metric hash")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		err = c.Repo.Save(r.Context(), newMetric)
		if err != nil {
			c.Logger.Error().Err(err).Msg("handler: failed to save metric")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func metricStoreAll(c *service.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		buf := make([]metric.Metric, 0)
		if err := json.NewDecoder(r.Body).Decode(&buf); err != nil {
			c.Logger.Error().Err(err).Msg("handler: failed to unmarshall request")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		newMetrics := metric.Metrics{Metrics: buf}

		if ok, m := newMetrics.IsValid(); !ok {
			c.Logger.Error().Msgf("handler: invalid metric: %v", m)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		ok, err := newMetrics.CheckHash(c.Key)
		if err != nil {
			c.Logger.Error().Err(err).Msg("handler: failed to check metric hash")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		if !ok {
			c.Logger.Error().Msg("handler: invalid metric hash")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		err = c.Repo.SaveAll(r.Context(), newMetrics.Metrics)
		if err != nil {
			c.Logger.Error().Err(err).Msg("handler: failed to save metric")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func metricJSONLoad(c *service.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var keyMetric metric.Metric

		if err := json.NewDecoder(r.Body).Decode(&keyMetric); err != nil {
			c.Logger.Error().Err(err).Msg("handler: failed to unmarshall request")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		resultMetric, err := c.Repo.FindByID(r.Context(), keyMetric)
		if err != nil {
			c.Logger.
				Error().
				Msgf("handler: metric not found with id: %v, type: %v", keyMetric.ID, keyMetric.MType.String())
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}

		resultMetric.SetHash(c.Key)

		resultBody, err := json.Marshal(resultMetric)
		if err != nil {
			c.Logger.
				Error().
				Err(err).
				Str("metric", resultMetric.String()).
				Msg("handler: failed to marshall metric")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(resultBody)
		if err != nil {
			c.Logger.Error().Err(err).Msg("handler: failed to write response body")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func metricLoad(c *service.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mName := chi.URLParam(r, "name")
		mType := chi.URLParam(r, "type")
		keyMetric, createErr := metric.NewMetric(mName, mType, "0", "")
		if createErr.Error != nil {
			c.Logger.
				Error().
				Err(createErr.Error).
				Msg("handler: failed to create metric")
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}
		tmpMetric, err := c.Repo.FindByID(r.Context(), keyMetric)
		if err != nil {
			c.Logger.
				Error().
				Msgf("handler: metric not found with id: %v, type: %v", keyMetric.ID, keyMetric.MType.String())
			http.Error(w, "Metric not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		_, err = w.Write([]byte(tmpMetric.GetStringValue()))
		if err != nil {
			c.Logger.Error().Err(err).Msg("handler: failed to write response body")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

func Service(config *service.ServerConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware2.GzipResponseEncoder)
	r.Use(middleware2.GzipRequestDecoder)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/", rootHandler(config))
	r.Get("/ping", connectionPingHandler(config))
	r.Post("/value/", metricJSONLoad(config))
	r.Get("/value/{type}/{name}", metricLoad(config))
	r.Post("/update/", metricJSONStore(config))
	r.Post("/updates/", metricStoreAll(config))
	r.Post("/update/{type}/{name}/{value}", metricStore(config))

	return r
}
