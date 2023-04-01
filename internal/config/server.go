package config

import (
	"context"
	"strings"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/storage"
	"github.com/rs/zerolog"
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

// GetServerConfig получает конфигурацией сервера из командной строки или переменных окружения.
// Параметры из переменных окружения имеют приоритет.
// ADDRESS - адрес и порт на котором необходимо поднять сервер
// DATABASE_DSN - строка подключения к БД
// STORE_INTERVAL - интервал сброса метрик на диск (необязательно)
// STORE_FILE - имя файла для хранения метрик (необязательно)
// RESTORE - сохранять ли метрики с предыдущего сеанса (по умолчанию нет)
// KEY - ключ для подписи метрик должен быть одинаковым на сервере и агенте
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

type ServerConfig struct {
	ServerCmd
	Logger zerolog.Logger
	Repo   storage.Repository
}

// NewServerConfig возвращает структуру с необходимыми настройками сервера
func NewServerConfig(ctx context.Context, logger zerolog.Logger, srvCmd ServerCmd) *ServerConfig {
	srvCfg := ServerConfig{
		ServerCmd: srvCmd,
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
			storage.NewFileStorage(ctx, srvCfg.StoreFile, srvCfg.StoreInterval, srvCfg.Restore, logger),
		)
	}

	return &srvCfg
}
