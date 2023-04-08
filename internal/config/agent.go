package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	// Параметры из переменных окружения имеют приоритет.
	// ADDRESS - адрес сервера метрик
	// REPORT_INTERVAL - интервал отправки обновлений на сервер
	// POLL_INTERVAL - интервал обновления метрик
	// KEY - ключ для подписи метрик должен быть одинаковым на сервере и агенте
	// CRYPTO_KEY - имя файла с публичным RSA ключом, должен соответствовать приватному ключу сервера
	// CONFIG - имя файла конфигурации в формате json
	envVarsAgent = []string{
		"ADDRESS",
		"REPORT_INTERVAL",
		"POLL_INTERVAL",
		"KEY",
		"CRYPTO_KEY",
		"CONFIG",
	}
)

type AgentConfig struct {
	Address           string        `mapstructure:"address"`
	ReportInterval    time.Duration `mapstructure:"report_interval"`
	PollInterval      time.Duration `mapstructure:"poll_interval"`
	Key               string        `mapstructure:"key"`
	PublicKeyFileName string        `mapstructure:"crypto_key"`
	ConfigFileName    string        `mapstructure:"config"`
	PublicKey         *rsa.PublicKey
	Logger            zerolog.Logger
}

func agentSetDefaults() {
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.RegisterAlias("crypto-key", "crypto_key")

	viper.SetDefault("address", Address)
	viper.SetDefault("report_interval", ReportInterval)
	viper.SetDefault("poll_interval", PollInterval)
}

func agentGetPFlags() {
	pflag.StringP("address", "a", viper.GetString("address"), "Server address:port")
	pflag.DurationP("report_interval", "r", viper.GetDuration("report_interval"), "Send metrics to server interval")
	pflag.DurationP("poll_interval", "p", viper.GetDuration("poll_interval"), "Collect metrics interval")
	pflag.StringP("key", "k", "", "Metric sign hash key")
	pflag.String("crypto_key", "", "Public RSA key")
	pflag.StringP("config", "c", "", "Config file name")
	pflag.Parse()
}

func newAgentConfig() (*AgentConfig, error) {
	cfg := &AgentConfig{}
	err := viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func getRSAPublicKey(fileName string) (*rsa.PublicKey, error) {
	keyBytes, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pub, nil
}

// NewAgentConfig отдает готовую структуру с необходимыми настройками для агента
func NewAgentConfig(logger zerolog.Logger) (*AgentConfig, error) {
	agentSetDefaults()

	agentGetPFlags()

	err := bindConfigFile("config")
	if err != nil {
		return nil, err
	}

	err = bindPFlags()
	if err != nil {
		return nil, err
	}

	err = bindEnvVars(envVarsAgent)
	if err != nil {
		return nil, err
	}

	agentCfg, err := newAgentConfig()
	if err != nil {
		return nil, err
	}

	agentCfg.Logger = logger

	if !hasSchema(agentCfg.Address) {
		agentCfg.Address = "http://" + agentCfg.Address
	}

	if len(agentCfg.PublicKeyFileName) > 0 {
		pubKey, err := getRSAPublicKey(agentCfg.PublicKeyFileName)
		if err != nil {
			logger.Fatal().Err(err).Send()
		}
		agentCfg.PublicKey = pubKey
	}

	return agentCfg, nil
}
