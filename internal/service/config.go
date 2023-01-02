package service

import (
	"context"
	"strings"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
)

type ServerConfig struct {
	config.ServerCmd
	Repo storage.Repository
}

func NewServerConfig(ctx context.Context) *ServerConfig {
	srvCfg := ServerConfig{
		ServerCmd: config.GetServerConfig(),
	}

	if hasSchema(srvCfg.Address) {
		split := strings.Split(srvCfg.Address, "//")
		srvCfg.Address = split[1]
	}

	if srvCfg.DatabaseDsn != "" {
		srvCfg.Repo = storage.NewDBStorage(srvCfg.DatabaseDsn, srvCfg.Restore, ctx)
	} else {
		srvCfg.Repo = storage.NewPersistenceRepo(
			storage.NewFileStorage(srvCfg.StoreFile, srvCfg.StoreInterval, srvCfg.Restore, ctx),
		)
	}

	return &srvCfg
}

type AgentConfig struct {
	config.AgentCmd
}

func NewAgentConfig() *AgentConfig {
	agentCfg := AgentConfig{
		AgentCmd: config.GetAgentConfig(),
	}

	if !hasSchema(agentCfg.Address) {
		agentCfg.Address = "http://" + agentCfg.Address
	}

	return &agentCfg
}

func hasSchema(addr string) bool {
	return strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://")
}
