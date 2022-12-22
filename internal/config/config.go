package config

import (
	"os"
	"strconv"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/log"
)

const (
	Address       = "localhost:8080"
	StoreInterval = 300 * time.Second
	StoreFile     = "/tmp/devops-metrics-db.json"
	Restore       = true

	ReportInterval = 10 * time.Second
	PollInterval   = 2 * time.Second
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
			log.Error.Fatal("can`t parse duration: ", val)
		}
		return durationVal
	}
	return defaultVal
}

func lookupEnvOrBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		parseBool, err := strconv.ParseBool(val)
		if err != nil {
			log.Error.Fatal("can`t parse bool: ", val)
		}
		return parseBool
	}
	return defaultVal
}
