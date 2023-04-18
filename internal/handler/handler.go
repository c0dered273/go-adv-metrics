package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	middleware2 "github.com/c0dered273/go-adv-metrics/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

//	@Title			Metrics collection API
//	@Description	Сервис сбора и хранения метрик.
//	@Version		1.0

type IndexData struct {
	Title   string
	Metrics []string
}

// RootHandler godoc
//
//	@Tags			Index
//	@Summary		Отдает html со всеми метриками
//	@Description	Генерирует html страницу со списком всех метрик переданных на сервер.
//	@ID				rootHandler
//	@Produce		html
//	@Success		200
//	@Failure		500	{string}	string	"Internal error"
//	@Router			/ [get]
func RootHandler(c *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		indexTemplate, err := template.ParseFiles("templates/index.html")
		if err != nil {
			c.Logger.Fatal().Err(err).Msg("handler: failed parse template file")
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
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

// ConnectionPingHandler godoc
//
//	@Tags			Ping
//	@Summary		Проверяет соединение с БД
//	@Description	Позволяет проверить соединение с базой данных.
//	@ID				connectionPing
//	@Success		200
//	@Failure		500	{string}	string	"Internal error"
//	@Router			/ping [get]
func ConnectionPingHandler(c *config.ServerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := c.Repo.Ping(); err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
	}
}

// StoreMetricFromURLRequestHandler godoc
//
//	@Tags			Store
//	@Summary		Сохраняет метрику из запроса
//	@Description	Сохраняет или обновляет одну метрику через url запрос.
//	@ID				storeFromURL
//	@Param			type	path	string	true	"Metric type"
//	@Param			name	path	string	true	"Metric name"
//	@Param			value	path	string	true	"Metric value"
//	@Success		200
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		500	{string}	string	"Internal error"
//	@Failure		501	{string}	string	"Unknown metric type"
//	@Router			/update/{type}/{name}/{value} [post]
func StoreMetricFromURLRequestHandler(c *config.ServerConfig) http.HandlerFunc {
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

// StoreMetricFromJSONHandler godoc
//
//	@Tags			Store
//	@Summary		Сохраняет метрику из json
//	@Description	Сохраняет или обновляет одну метрику из json объекта.
//	@ID				storeFromJSON
//	@Accept			json
//	@Param			metric_data	body	metric.Metric	true	"Metric data"
//	@Success		200
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		500	{string}	string	"Internal error"
//	@Router			/update/ [post]
func StoreMetricFromJSONHandler(c *config.ServerConfig) http.HandlerFunc {
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

// StoreAllMetricsFromJSONHandler godoc
//
//	@Tags			Store
//	@Summary		Сохраняет метрики из json
//	@Description	Сохраняет или обновляет метрики из массива json объектов.
//	@ID				storeAllFromJSON
//	@Accept			json
//	@Param			metric_data	body	metric.Metrics	true	"Metric data"
//	@Success		200
//	@Failure		400	{string}	string	"Bad request"
//	@Failure		500	{string}	string	"Internal error"
//	@Router			/updates/ [post]
func StoreAllMetricsFromJSONHandler(c *config.ServerConfig) http.HandlerFunc {
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

// LoadMetricByJSONHandler godoc
//
//	@Tags			Load
//	@Summary		Отдает метрику из json
//	@Description	Отдает одну метрику согласно имени и типа метрики из json запроса.
//	@ID				loadFromJSON
//	@Accept			json
//	@Produce		json
//	@Param			metric_data	body		metric.Metric	true	"Metric data"
//	@Success		200			{object}	metric.Metric
//	@Failure		400			{string}	string	"Bad request"
//	@Failure		404			{string}	string	"Metric not found"
//	@Failure		500			{string}	string	"Internal error"
//	@Router			/value/ [post]
func LoadMetricByJSONHandler(c *config.ServerConfig) http.HandlerFunc {
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

// LoadMetricByURLRequestHandler
//
//	@Tags			Load
//	@Summary		Отдает метрику из запроса
//	@Description	Отдает одну метрику согласно имени и типа из url запроса.
//	@ID				LoadFromURL
//	@Produce		plain
//	@Param			type	path		string	true	"Metric type"
//	@Param			name	path		string	true	"Metric name"
//	@Success		200		{string}	string
//	@Failure		400		{string}	string	"Bad request"
//	@Failure		404		{string}	string	"Metric not found"
//	@Failure		500		{string}	string	"Internal error"
//	@Router			/value/{type}/{name} [get]
func LoadMetricByURLRequestHandler(c *config.ServerConfig) http.HandlerFunc {
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

func Service(config *config.ServerConfig, logger zerolog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware2.TrustedSubnet(config, logger))
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware2.GzipRequestDecoder)
	r.Use(middleware.Compress(5))
	r.Use(middleware2.RSADecrypt(config.PrivateKey))

	r.Mount("/debug", middleware.Profiler())

	r.Get("/", RootHandler(config))
	r.Get("/ping", ConnectionPingHandler(config))
	r.Post("/value/", LoadMetricByJSONHandler(config))
	r.Get("/value/{type}/{name}", LoadMetricByURLRequestHandler(config))
	r.Post("/update/", StoreMetricFromJSONHandler(config))
	r.Post("/updates/", StoreAllMetricsFromJSONHandler(config))
	r.Post("/update/{type}/{name}/{value}", StoreMetricFromURLRequestHandler(config))

	return r
}
