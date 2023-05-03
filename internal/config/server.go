package config

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net"
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
	// CA_CERT_FILE - файл с корневым сертификатом
	// SERVER_CERT_FILE - файл с серверным сертификатом
	// SERVER_KEY_FILE - файл с серверным ключом
	serverEnvVars = []string{
		"ADDRESS",
		"GRPC_ADDRESS",
		"DATABASE_DSN",
		"STORE_INTERVAL",
		"STORE_FILE",
		"RESTORE",
		"KEY",
		"CRYPTO_KEY",
		"CONFIG",
		"TRUSTED_SUBNET",
		"CA_CERT_FILE",
		"SERVER_CERT_FILE",
		"SERVER_KEY_FILE",
	}
)

type ServerConfigFileParams struct {
	Address            string        `json:"address"`
	GRPCAddress        string        `json:"grpc_address"`
	DatabaseDsn        string        `json:"database_dsn"`
	StoreInterval      time.Duration `json:"store_interval"`
	StoreFile          string        `json:"store_file"`
	Restore            bool          `json:"restore"`
	PrivateKeyFileName string        `json:"crypto_key"`
	TrustedSubnet      string        `json:"trusted_subnet"`
	CACertFile         string        `json:"ca_cert_file"`
	ServerCertFile     string        `json:"server_cert_file"`
	ServerKeyFile      string        `json:"server_key_file"`
}

type ServerInParams struct {
	Address            string        `mapstructure:"address"`
	GRPCAddress        string        `mapstructure:"grpc_address"`
	DatabaseDsn        string        `mapstructure:"database_dsn"`
	StoreInterval      time.Duration `mapstructure:"store_interval"`
	StoreFile          string        `mapstructure:"store_file"`
	Restore            bool          `mapstructure:"restore"`
	Key                string        `mapstructure:"key"`
	PrivateKeyFileName string        `mapstructure:"crypto_key"`
	TrustedSubnet      *net.IPNet    `mapstructure:"trusted_subnet"`
	CACertFile         string        `mapstructure:"ca_cert_file"`
	ServerCertFile     string        `mapstructure:"server_cert_file"`
	ServerKeyFile      string        `mapstructure:"server_key_file"`
}

// getServerPFlag получает конфигурацией сервера из командной строки.
func getServerPFlag() Params {
	pflag.StringP("address", "a", "", "Server address:port")
	pflag.StringP("grpc_address", "g", "", "gRPC Server address:port")
	pflag.StringP("databaseDsn", "d", "", "Database url")
	pflag.StringP("store_interval", "i", "", "Writing metrics to disk interval")
	pflag.StringP("filename", "f", "", "Storage filename")
	pflag.StringP("key", "k", "", "Metric sign hash key")
	pflag.StringP("restore", "r", "", "Is restore metrics from disk")
	pflag.String("crypto-key", "", "Private RSA key")
	pflag.StringP("config", "c", "", "Config file name")
	pflag.StringP("trusted_subnet", "t", "", "Trusted subnet")

	pflag.String("ca_cert_file", "", "CA certificate")
	pflag.String("server_cert_file", "", "Server certificate")
	pflag.String("server_key_file", "", "Server certificate key")

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

func getSrvDefaults() Params {
	return map[string]any{
		"address":        Address,
		"grpc_address":   GRPCAddress,
		"store_interval": StoreInterval,
		"restore":        Restore,
		"store_file":     StoreFile,
	}
}

type ServerConfig struct {
	*ServerInParams
	Logger       zerolog.Logger
	PrivateKey   *rsa.PrivateKey
	IsTLSEnabled bool
	Repo         storage.Repository
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
	defaults := getSrvDefaults()
	pFlagCfg := getServerPFlag()
	envCfg := getEnvCfg(serverEnvVars)

	mergedCfg := merge(defaults, pFlagCfg, envCfg)
	fileCfg, err := getFileCfg(mergedCfg)
	if err != nil {
		return nil, err
	}

	mergedCfg = merge(defaults, fileCfg, pFlagCfg, envCfg)
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

	if srvCfg.CACertFile != "" && srvCfg.ServerCertFile != "" && srvCfg.ServerKeyFile != "" {
		srvCfg.IsTLSEnabled = true
	}

	return &srvCfg, nil
}
