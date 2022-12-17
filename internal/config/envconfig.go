package config

import (
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/storage"
)

type Agent struct {
	Address        string        `env:"ADDRESS" envDefault:"http://localhost:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
}

type Server struct {
	Address string `env:"ADDRESS" envDefault:"localhost:8080"`
	Repo    *storage.MemStorage
}
