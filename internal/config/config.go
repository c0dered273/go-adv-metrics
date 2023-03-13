package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Настройки агента по умолчанию.
// Если не передана строка соединения с БД, метрики хранятся локально в json файле.
const (
	// Address Адрес сервера
	Address = "localhost:8080"
	// StoreInterval Интервал сброса метрик на диск
	StoreInterval = 300 * time.Second
	// StoreFile Путь к файлу хранения метрик
	StoreFile = "/tmp/devops-metrics-db.json"
	// Restore Флаг показывает сохранять ли метрики с прошлого сеанса или очистить БД
	Restore = true

	// ReportInterval Интервал отправки обновлений на сервер
	ReportInterval = 10 * time.Second
	// PollInterval Интервал обновления метрик
	PollInterval = 2 * time.Second
)

func lookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func lookupEnvOrDuration(key string, defaultVal time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		durationVal, err := time.ParseDuration(val)
		if err != nil {
			log.Fatal().Msgf("config: failed to parse duration: %v", val)
		}
		return durationVal
	}
	return defaultVal
}

func lookupEnvOrBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		parseBool, err := strconv.ParseBool(val)
		if err != nil {
			log.Fatal().Msgf("config: failed to parse bool: %v", val)
		}
		return parseBool
	}
	return defaultVal
}

func hasSchema(addr string) bool {
	return strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://")
}
