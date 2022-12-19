package config

import (
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
)

type ServerEnv struct {
	Address       string        `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreInterval time.Duration `env:"STORE_INTERVAL" envDefault:"300s"`
	StoreFile     string        `env:"STORE_FILE" envDefault:"/tmp/devops-metrics-db.json"`
	Restore       bool          `env:"RESTORE" envDefault:"true"`
}

type ServerConfig struct {
	Address       string
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
	Repo          storage.Repository
}

func GetServerConfig() *ServerConfig {
	srvCfg := &ServerConfig{}
	srvEnv := &ServerEnv{}
	if err := env.Parse(srvEnv); err != nil {
		log.Error.Fatal(err)
	}

	pflag.StringVar(&srvCfg.Address, "a", srvEnv.Address, "Server address:port")
	pflag.DurationVar(&srvCfg.StoreInterval, "i", srvEnv.StoreInterval, "Writing metrics to disk interval")
	pflag.StringVar(&srvCfg.StoreFile, "f", srvEnv.StoreFile, "Storage filename")
	pflag.BoolVar(&srvCfg.Restore, "r", srvEnv.Restore, "Is restore metrics from disk")
	pflag.Parse()

	return &ServerConfig{
		Address:       srvCfg.Address,
		StoreInterval: srvCfg.StoreInterval,
		StoreFile:     srvCfg.StoreFile,
		Restore:       srvCfg.Restore,
	}
}
