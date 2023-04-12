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
)

var (
	// Параметры из переменных окружения имеют приоритет.
	// ADDRESS - адрес сервера метрик
	// REPORT_INTERVAL - интервал отправки обновлений на сервер
	// POLL_INTERVAL - интервал обновления метрик
	// KEY - ключ для подписи метрик должен быть одинаковым на сервере и агенте
	// CRYPTO_KEY - имя файла с публичным RSA ключом, должен соответствовать приватному ключу сервера
	// CONFIG - имя файла конфигурации в формате json
	agentEnvVars = []string{
		"ADDRESS",
		"REPORT_INTERVAL",
		"POLL_INTERVAL",
		"KEY",
		"CRYPTO_KEY",
		"CONFIG",
	}
)

type AgentConfigFileParams struct {
	Address           string        `json:"address"`
	ReportInterval    time.Duration `json:"report_interval"`
	PollInterval      time.Duration `json:"poll_interval"`
	PublicKeyFileName string        `json:"crypto_key"`
}

type AgentInParams struct {
	Address           string        `mapstructure:"address"`
	ReportInterval    time.Duration `mapstructure:"report_interval"`
	PollInterval      time.Duration `mapstructure:"poll_interval"`
	Key               string        `mapstructure:"key"`
	PublicKeyFileName string        `mapstructure:"crypto_key"`
	ConfigFileName    string        `mapstructure:"config"`
}

// getAgentPFlag получает конфигурацию агента из командной строки.
func getAgentPFlag() Params {
	pflag.StringP("address", "a", Address, "Server address:port")
	pflag.DurationP("report_interval", "r", ReportInterval, "Send metrics to server interval")
	pflag.DurationP("poll_interval", "p", PollInterval, "Collect metrics interval")
	pflag.StringP("key", "k", "", "Metric sign hash key")
	pflag.String("crypto-key", "", "Public RSA key")
	pflag.StringP("config", "c", "", "Имя файла конфигурации")
	pflag.Parse()

	params := make(map[string]any)
	pflag.CommandLine.VisitAll(func(flag *pflag.Flag) {
		if len(flag.Value.String()) > 0 {
			name := strings.ReplaceAll(flag.Name, "-", "_")

			params[name] = flag.Value.String()
		}
	})
	return params
}

type AgentConfig struct {
	*AgentInParams
	PublicKey *rsa.PublicKey
	Logger    zerolog.Logger
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
	pFlagCfg := getAgentPFlag()
	envCfg := getEnvCfg(agentEnvVars)

	mergedCfg := merge(pFlagCfg, envCfg)
	fileCfg, err := getFileCfg(mergedCfg)
	if err != nil {
		return nil, err
	}

	mergedCfg = merge(fileCfg, pFlagCfg, envCfg)
	agentParams := &AgentInParams{}
	err = bindParams(mergedCfg, agentParams)
	if err != nil {
		return nil, err
	}

	agentCfg := AgentConfig{
		AgentInParams: agentParams,
		Logger:        logger,
	}

	if !hasSchema(agentCfg.Address) {
		agentCfg.Address = "http://" + agentCfg.Address
	}

	if len(agentCfg.PublicKeyFileName) > 0 {
		pubKey, err := getRSAPublicKey(agentCfg.PublicKeyFileName)
		if err != nil {
			return nil, err
		}
		agentCfg.PublicKey = pubKey
	}

	return &agentCfg, nil
}
