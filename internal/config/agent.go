package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

type AgentCmd struct {
	Address           string
	ReportInterval    time.Duration
	PollInterval      time.Duration
	Key               string
	PublicKeyFileName string
}

// getAgentConfig получает конфигурацией агента из командной строки или переменных окружения.
// Параметры из переменных окружения имеют приоритет.
// ADDRESS - адрес сервера метрик
// REPORT_INTERVAL - интервал отправки обновлений на сервер
// POLL_INTERVAL - интервал обновления метрик
// KEY - ключ для подписи метрик должен быть одинаковым на сервере и агенте
// CRYPTO_KEY - имя файла с публичным RSA ключом, должен соответствовать приватному ключу сервера
func getAgentConfig() AgentCmd {
	agentFlag := AgentCmd{}
	pflag.StringVarP(&agentFlag.Address, "address", "a", Address, "Server address:port")
	pflag.DurationVarP(&agentFlag.ReportInterval, "report_interval", "r", ReportInterval, "Send metrics to server interval")
	pflag.DurationVarP(&agentFlag.PollInterval, "poll_interval", "p", PollInterval, "Collect metrics interval")
	pflag.StringVarP(&agentFlag.Key, "key", "k", "", "Metric sign hash key")
	pflag.StringVar(&agentFlag.PublicKeyFileName, "crypto-key", "", "Public RSA key")
	pflag.Parse()

	return AgentCmd{
		Address:           lookupEnvOrString("ADDRESS", agentFlag.Address),
		ReportInterval:    lookupEnvOrDuration("REPORT_INTERVAL", agentFlag.ReportInterval),
		PollInterval:      lookupEnvOrDuration("POLL_INTERVAL", agentFlag.PollInterval),
		Key:               lookupEnvOrString("KEY", agentFlag.Key),
		PublicKeyFileName: lookupEnvOrString("CRYPTO_KEY", agentFlag.PublicKeyFileName),
	}
}

type AgentConfig struct {
	AgentCmd
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
func NewAgentConfig(logger zerolog.Logger) *AgentConfig {
	agentCfg := AgentConfig{
		AgentCmd: getAgentConfig(),
		Logger:   logger,
	}

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

	return &agentCfg
}
