package config

import (
	"time"

	"github.com/spf13/pflag"
)

type ServerCmd struct {
	Address       string
	DatabaseDsn   string
	StoreInterval time.Duration
	StoreFile     string
	Restore       bool
	Key           string
}

func GetServerConfig() ServerCmd {
	srvFlag := ServerCmd{}
	pflag.StringVarP(&srvFlag.Address, "address", "a", Address, "Server address:port")
	pflag.StringVarP(&srvFlag.DatabaseDsn, "databaseDsn", "d", "", "Database url")
	pflag.DurationVarP(&srvFlag.StoreInterval, "store_interval", "i", StoreInterval, "Writing metrics to disk interval")
	pflag.StringVarP(&srvFlag.StoreFile, "filename", "f", StoreFile, "Storage filename")
	pflag.StringVarP(&srvFlag.Key, "key", "k", "", "Metric sign hash key")
	pflag.BoolVarP(&srvFlag.Restore, "restore", "r", Restore, "Is restore metrics from disk")
	pflag.Parse()

	return ServerCmd{
		Address:       lookupEnvOrString("ADDRESS", srvFlag.Address),
		DatabaseDsn:   lookupEnvOrString("DATABASE_DSN", srvFlag.DatabaseDsn),
		StoreInterval: lookupEnvOrDuration("STORE_INTERVAL", srvFlag.StoreInterval),
		StoreFile:     lookupEnvOrString("STORE_FILE", srvFlag.StoreFile),
		Restore:       lookupEnvOrBool("RESTORE", srvFlag.Restore),
		Key:           lookupEnvOrString("KEY", srvFlag.Key),
	}
}
