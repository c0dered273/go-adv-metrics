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
	// GRPC_CLIENT - использовать gRPC для передачи метрик
	// CA_CERT_FILE - файл с корневым сертификатом
	agentEnvVars = []string{
		"ADDRESS",
		"REPORT_INTERVAL",
		"POLL_INTERVAL",
		"KEY",
		"CRYPTO_KEY",
		"CONFIG",
		"GRPC_CLIENT",
		"CA_CERT_FILE",
	}
)

type AgentConfigFileParams struct {
	Address           string        `json:"address"`
	ReportInterval    time.Duration `json:"report_interval"`
	PollInterval      time.Duration `json:"poll_interval"`
	PublicKeyFileName string        `json:"crypto_key"`
	GRPCClient        bool          `json:"grpc_client"`
	CACertFile        string        `json:"ca_cert_file"`
}

type AgentInParams struct {
	Address           string        `mapstructure:"address"`
	ReportInterval    time.Duration `mapstructure:"report_interval"`
	PollInterval      time.Duration `mapstructure:"poll_interval"`
	Key               string        `mapstructure:"key"`
	PublicKeyFileName string        `mapstructure:"crypto_key"`
	ConfigFileName    string        `mapstructure:"config"`
	GRPCClient        bool          `mapstructure:"grpc_client"`
	CACertFile        string        `mapstructure:"ca_cert_file"`
}

// getAgentPFlag получает конфигурацию агента из командной строки.
func getAgentPFlag() Params {
	pflag.StringP("address", "a", "", "Server address:port")
	pflag.StringP("report_interval", "r", "", "Send metrics to server interval")
	pflag.StringP("poll_interval", "p", "", "Collect metrics interval")
	pflag.StringP("key", "k", "", "Metric sign hash key")
	pflag.String("crypto-key", "", "Public RSA key")
	pflag.StringP("config", "c", "", "Config file")

	pflag.StringP("grpc_client", "g", "", "Use gRPC client")
	pflag.String("ca_cert_file", "", "CA certificate")

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

func getAgentDefaults() Params {
	return map[string]any{
		"address":         Address,
		"report_interval": ReportInterval,
		"poll_interval":   PollInterval,
		"grpc_client":     "false",
	}
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
	defaults := getAgentDefaults()
	pFlagCfg := getAgentPFlag()
	envCfg := getEnvCfg(agentEnvVars)

	mergedCfg := merge(defaults, pFlagCfg, envCfg)
	fileCfg, err := getFileCfg(mergedCfg)
	if err != nil {
		return nil, err
	}

	mergedCfg = merge(defaults, fileCfg, pFlagCfg, envCfg)
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
