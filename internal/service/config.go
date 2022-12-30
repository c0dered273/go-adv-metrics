package service

import (
	"context"
	"strings"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
)

func NewServerConfig(ctx context.Context) *config.ServerConfig {
	srvCfg := config.GetServerConfig()

	if hasSchema(srvCfg.Address) {
		split := strings.Split(srvCfg.Address, "//")
		srvCfg.Address = split[1]
	}

	srvCfg.Repo = storage.NewFileStorage(srvCfg.StoreFile, srvCfg.StoreInterval, srvCfg.Restore, ctx)

	// TODO("Заменить файловое хранилище на базу данных")
	if srvCfg.DatabaseDsn != "" {
		srvCfg.DBRepo = storage.NewDBStorage(srvCfg.DatabaseDsn, ctx)
	}

	return srvCfg
}

func NewAgentConfig() *config.AgentConfig {
	agentCfg := config.GetAgentConfig()

	if !hasSchema(agentCfg.Address) {
		agentCfg.Address = "http://" + agentCfg.Address
	}

	return agentCfg
}

func hasSchema(addr string) bool {
	return strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://")
}
