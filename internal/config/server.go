package config

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"strings"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/storage"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

var (
	// Параметры из переменных окружения имеют приоритет.
	// ADDRESS - адрес и порт на котором необходимо поднять сервер
	// DATABASE_DSN - строка подключения к БД
	// STORE_INTERVAL - интервал сброса метрик на диск (необязательно)
	// STORE_FILE - имя файла для хранения метрик (необязательно)
	// RESTORE - сохранять ли метрики с предыдущего сеанса (по умолчанию нет)
	// KEY - ключ для подписи метрик должен быть одинаковым на сервере и агенте
	// CRYPTO_KEY - имя файла с приватным RSA ключом, должен соответствовать публичному ключу клиента
	// CONFIG - имя файла конфигурации в формате json
	serverEnvVars = []string{
		"ADDRESS",
		"DATABASE_DSN",
		"STORE_INTERVAL",
		"STORE_FILE",
		"RESTORE",
		"KEY",
		"CRYPTO_KEY",
		"CONFIG",
	}
)

type ServerConfigFileParams struct {
	Address            string        `json:"address"`
	DatabaseDsn        string        `json:"database_dsn"`
	StoreInterval      time.Duration `json:"store_interval"`
	StoreFile          string        `json:"store_file"`
	Restore            bool          `json:"restore"`
	PrivateKeyFileName string        `json:"crypto_key"`
}

type ServerInParams struct {
	Address            string        `mapstructure:"address"`
	DatabaseDsn        string        `mapstructure:"database_dsn"`
	StoreInterval      time.Duration `mapstructure:"store_interval"`
	StoreFile          string        `mapstructure:"store_file"`
	Restore            bool          `mapstructure:"restore"`
	Key                string        `mapstructure:"key"`
	PrivateKeyFileName string        `mapstructure:"crypto_key"`
}

// getServerPFlag получает конфигурацией сервера из командной строки.
func getServerPFlag() Params {
	pflag.StringP("address", "a", Address, "Server address:port")
	pflag.StringP("databaseDsn", "d", "", "Database url")
	pflag.DurationP("store_interval", "i", StoreInterval, "Writing metrics to disk interval")
	pflag.StringP("filename", "f", StoreFile, "Storage filename")
	pflag.StringP("key", "k", "", "Metric sign hash key")
	pflag.BoolP("restore", "r", Restore, "Is restore metrics from disk")
	pflag.String("crypto-key", "", "Private RSA key")
	pflag.StringP("config", "c", "", "Имя файла конфигурации")
	pflag.Parse()

	params := make(map[string]any)
	pflag.CommandLine.VisitAll(func(flag *pflag.Flag) {
		if len(flag.Value.String()) > 0 {
			name := strings.ReplaceAll(flag.Name, "-", "_")

			if name == "databaseDsn" {
				name = "database_dsn"
			}

			if name == "filename" {
				name = "store_file"
			}

			params[name] = flag.Value.String()
		}
	})
	return params
}

type ServerConfig struct {
	*ServerInParams
	Logger     zerolog.Logger
	PrivateKey *rsa.PrivateKey
	Repo       storage.Repository
}

func getRSAPrivateKey(fileName string) (*rsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	prv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return prv, nil
}

// NewServerConfig возвращает структуру с необходимыми настройками сервера
func NewServerConfig(ctx context.Context, logger zerolog.Logger) (*ServerConfig, error) {
	pFlagCfg := getServerPFlag()
	envCfg := getEnvCfg(serverEnvVars)

	mergedCfg := merge(pFlagCfg, envCfg)
	fileCfg, err := getFileCfg(mergedCfg)
	if err != nil {
		return nil, err
	}

	mergedCfg = merge(fileCfg, pFlagCfg, envCfg)
	serverParams := &ServerInParams{}
	err = bindParams(mergedCfg, serverParams)
	if err != nil {
		return nil, err
	}

	srvCfg := ServerConfig{
		ServerInParams: serverParams,
		Logger:         logger,
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

	if len(srvCfg.PrivateKeyFileName) > 0 {
		prvKey, err := getRSAPrivateKey(srvCfg.PrivateKeyFileName)
		if err != nil {
			return nil, err
		}
		srvCfg.PrivateKey = prvKey
	}

	return &srvCfg, nil
}
