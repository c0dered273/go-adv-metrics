package config

import (
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

type AgentCmd struct {
	Address        string
	ReportInterval time.Duration
	PollInterval   time.Duration
	Key            string
}

func GetAgentConfig() AgentCmd {
	agentFlag := AgentCmd{}
	pflag.StringVarP(&agentFlag.Address, "address", "a", Address, "Server address:port")
	pflag.DurationVarP(&agentFlag.ReportInterval, "report_interval", "r", ReportInterval, "Send metrics to server interval")
	pflag.DurationVarP(&agentFlag.PollInterval, "poll_interval", "p", PollInterval, "Collect metrics interval")
	pflag.StringVarP(&agentFlag.Key, "key", "k", "", "Metric sign hash key")
	pflag.Parse()

	return AgentCmd{
		Address:        lookupEnvOrString("ADDRESS", agentFlag.Address),
		ReportInterval: lookupEnvOrDuration("REPORT_INTERVAL", agentFlag.ReportInterval),
		PollInterval:   lookupEnvOrDuration("POLL_INTERVAL", agentFlag.PollInterval),
		Key:            lookupEnvOrString("KEY", agentFlag.Key),
	}
}

type AgentConfig struct {
	AgentCmd
	Logger zerolog.Logger
}

func NewAgentConfig(logger zerolog.Logger) *AgentConfig {
	agentCfg := AgentConfig{
		AgentCmd: GetAgentConfig(),
		Logger:   logger,
	}

	if !hasSchema(agentCfg.Address) {
		agentCfg.Address = "http://" + agentCfg.Address
	}

	return &agentCfg
}
