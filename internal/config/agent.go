package config

import (
	"time"

	"github.com/spf13/pflag"
)

type AgentConfig struct {
	Address        string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

func GetAgentConfig() *AgentConfig {
	agentFlag := AgentConfig{}
	pflag.StringVarP(&agentFlag.Address, "address", "a", Address, "Server address:port")
	pflag.DurationVarP(&agentFlag.ReportInterval, "report_interval", "r", ReportInterval, "Send metrics to server interval")
	pflag.DurationVarP(&agentFlag.PollInterval, "poll_interval", "p", PollInterval, "Collect metrics interval")
	pflag.Parse()

	return &AgentConfig{
		Address:        lookupEnvOrString("ADDRESS", agentFlag.Address),
		ReportInterval: lookupEnvOrDuration("REPORT_INTERVAL", agentFlag.ReportInterval),
		PollInterval:   lookupEnvOrDuration("POLL_INTERVAL", agentFlag.PollInterval),
	}
}
