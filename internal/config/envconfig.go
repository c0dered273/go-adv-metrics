package config

import (
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/storage"
)

type Agent struct {
	Address        string        `env:"ADDRESS" envDefault:"http://127.0.0.1:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
}

type Server struct {
	Address    string `env:"ADDRESS" envDefault:"http://127.0.0.1:8080"`
	Properties Properties
}

type Properties struct {
	Repo *storage.MemStorage
}
