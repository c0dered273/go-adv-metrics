package service

import (
	"context"
	"strings"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
	"github.com/rs/zerolog"
)

type ServerConfig struct {
	config.ServerCmd
	Logger zerolog.Logger
	Repo   storage.Repository
}

func NewServerConfig(logger zerolog.Logger, ctx context.Context) *ServerConfig {
	srvCfg := ServerConfig{
		ServerCmd: config.GetServerConfig(),
		Logger:    logger,
	}

	if hasSchema(srvCfg.Address) {
		split := strings.Split(srvCfg.Address, "//")
		srvCfg.Address = split[1]
	}

	if srvCfg.DatabaseDsn != "" {
		srvCfg.Repo = storage.NewDBStorage(srvCfg.DatabaseDsn, srvCfg.Restore, srvCfg.Logger, ctx)
	} else {
		srvCfg.Repo = storage.NewPersistenceRepo(
			storage.NewFileStorage(srvCfg.StoreFile, srvCfg.StoreInterval, srvCfg.Restore, logger, ctx),
		)
	}

	return &srvCfg
}

type AgentConfig struct {
	config.AgentCmd
	Logger zerolog.Logger
}

func NewAgentConfig(logger zerolog.Logger) *AgentConfig {
	agentCfg := AgentConfig{
		AgentCmd: config.GetAgentConfig(),
		Logger:   logger,
	}

	if !hasSchema(agentCfg.Address) {
		agentCfg.Address = "http://" + agentCfg.Address
	}

	return &agentCfg
}

func hasSchema(addr string) bool {
	return strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://")
}
