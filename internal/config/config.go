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
	Address       = "localhost:8080"              // Адрес сервера
	StoreInterval = 300 * time.Second             // Интервал сброса метрик на диск
	StoreFile     = "/tmp/devops-metrics-db.json" // Путь к файлу хранения метрик
	Restore       = true                          // Флаг показывает сохранять ли метрики с прошлого сеанса или очистить БД

	ReportInterval = 10 * time.Second // Интервал отправки обновлений на сервер
	PollInterval   = 2 * time.Second  // Интервал обновления метрик
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
