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
	"github.com/spf13/viper"
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
	envVarsServer = []string{
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

type ServerConfig struct {
	Address            string        `mapstructure:"address"`
	DatabaseDsn        string        `mapstructure:"database_dsn"`
	StoreInterval      time.Duration `mapstructure:"store_interval"`
	StoreFile          string        `mapstructure:"store_file"`
	Restore            bool          `mapstructure:"restore"`
	Key                string        `mapstructure:"key"`
	PrivateKeyFileName string        `mapstructure:"crypto_key"`
	ConfigFileName     string        `mapstructure:"config"`
	Logger             zerolog.Logger
	PrivateKey         *rsa.PrivateKey
	Repo               storage.Repository
}

func serverSetDefaults() {
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.RegisterAlias("filename", "store_file")
	viper.RegisterAlias("crypto-key", "crypto_key")

	viper.SetDefault("address", Address)
	viper.SetDefault("store_interval", StoreInterval)
	viper.SetDefault("filename", StoreFile)
	viper.SetDefault("restore", Restore)
}

func serverGetPFlags() {
	pflag.StringP("address", "a", viper.GetString("address"), "Server address:port")
	pflag.StringP("databaseDsn", "d", "", "Database url")
	pflag.DurationP("store_interval", "i", viper.GetDuration("store_interval"), "Writing metrics to disk interval")
	pflag.StringP("filename", "f", viper.GetString("filename"), "Storage filename")
	pflag.StringP("key", "k", "", "Metric sign hash key")
	pflag.BoolP("restore", "r", viper.GetBool("restore"), "Is restore metrics from disk")
	pflag.String("crypto-key", "", "Private RSA key")
	pflag.StringP("config", "c", "", "Config file name")

	f := pflag.CommandLine
	f.SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		switch name {
		case "databaseDsn":
			name = "database_dsn"
		}
		return pflag.NormalizedName(name)
	})

	pflag.Parse()
}

func newServerConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{}
	err := viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
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
	serverSetDefaults()

	serverGetPFlags()

	err := bindConfigFile("config")
	if err != nil {
		return nil, err
	}

	err = bindPFlags()
	if err != nil {
		return nil, err
	}

	err = bindEnvVars(envVarsServer)
	if err != nil {
		return nil, err
	}

	srvCfg, err := newServerConfig()
	if err != nil {
		return nil, err
	}

	srvCfg.Logger = logger

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
			logger.Fatal().Err(err).Send()
		}
		srvCfg.PrivateKey = prvKey
	}

	return srvCfg, nil
}
