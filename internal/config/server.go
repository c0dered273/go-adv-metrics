package config

import (
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/storage"
	"github.com/spf13/pflag"
)

type ServerConfig struct {
	Address       string
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
	Repo          storage.Repository
}

func GetServerConfig() *ServerConfig {
	srvFlag := ServerConfig{}
	pflag.StringVarP(&srvFlag.Address, "address", "a", Address, "Server address:port")
	pflag.DurationVarP(&srvFlag.StoreInterval, "store_interval", "i", StoreInterval, "Writing metrics to disk interval")
	pflag.StringVarP(&srvFlag.StoreFile, "filename", "f", StoreFile, "Storage filename")
	pflag.BoolVarP(&srvFlag.Restore, "restore", "r", Restore, "Is restore metrics from disk")
	pflag.Parse()

	return &ServerConfig{
		Address:       lookupEnvOrString("ADDRESS", srvFlag.Address),
		StoreInterval: lookupEnvOrDuration("STORE_INTERVAL", srvFlag.StoreInterval),
		StoreFile:     lookupEnvOrString("STORE_FILE", srvFlag.StoreFile),
		Restore:       lookupEnvOrBool("RESTORE", srvFlag.Restore),
	}
}
