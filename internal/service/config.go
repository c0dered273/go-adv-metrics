package service

import (
	"strings"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
	"github.com/caarlos0/env/v6"
)

func NewServerConfig() config.Server {
	var cfg config.Server
	if err := env.Parse(&cfg); err != nil {
		log.Error.Fatal(err)
	}

	if hasSchema(cfg.Address) {
		split := strings.Split(cfg.Address, "//")
		cfg.Address = split[1]
	}

	cfg.Repo = storage.GetMemStorageInstance()

	return cfg
}

func NewAgentConfig() config.Agent {
	var cfg config.Agent
	if err := env.Parse(&cfg); err != nil {
		log.Error.Fatal(err)
	}

	if !hasSchema(cfg.Address) {
		cfg.Address = "http://" + cfg.Address
	}

	return cfg
}

func hasSchema(addr string) bool {
	return strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://")
}
