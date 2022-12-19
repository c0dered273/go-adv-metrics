package config

import (
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
)

type AgentEnv struct {
	Address        string        `env:"ADDRESS" envDefault:"http://localhost:8080"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
}
type AgentConfig struct {
	Address        string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

func GetAgentConfig() *AgentConfig {
	agentCfg := &AgentConfig{}
	agentEnv := &AgentEnv{}
	if err := env.Parse(agentEnv); err != nil {
		log.Error.Fatal(err)
	}

	pflag.StringVar(&agentCfg.Address, "a", agentEnv.Address, "Server address:port")
	pflag.DurationVar(&agentCfg.ReportInterval, "r", agentEnv.ReportInterval, "Send metrics to server interval")
	pflag.DurationVar(&agentCfg.PollInterval, "p", agentEnv.PollInterval, "Collect metrics interval")
	pflag.Parse()

	return &AgentConfig{
		Address:        agentCfg.Address,
		ReportInterval: agentCfg.ReportInterval,
		PollInterval:   agentCfg.PollInterval,
	}
}
