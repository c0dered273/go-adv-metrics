package config

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
)

// Настройки агента по умолчанию.
// Если не передана строка соединения с БД, метрики хранятся локально в json файле.
const (
	// Address Адрес сервера
	Address     = "localhost:8080"
	GRPCAddress = "localhost:8081"
	// StoreInterval Интервал сброса метрик на диск
	StoreInterval = 300 * time.Second
	// StoreFile Путь к файлу хранения метрик
	StoreFile = "/tmp/devops-metrics-db"
	// Restore Флаг показывает сохранять ли метрики с прошлого сеанса или очистить БД
	Restore = true

	// ReportInterval Интервал отправки обновлений на сервер
	ReportInterval = 10 * time.Second
	// PollInterval Интервал обновления метрик
	PollInterval = 2 * time.Second
)

type Params map[string]any

// getFileCfg получает конфигурацию из json файла
func getFileCfg(cfg Params) (Params, error) {
	fileCfg := make(map[string]any)
	var err error

	if cfgFileName, ok := cfg["config"]; ok {
		if fileName := cfgFileName.(string); len(fileName) > 0 {
			fileCfg, err = readFileCfg(fileName)
			if err != nil {
				return nil, err
			}
		}
	}
	return fileCfg, nil
}

func readFileCfg(fileName string) (Params, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(file)

	params := make(map[string]any)
	err = decoder.Decode(&params)
	if err != nil {
		return nil, err
	}

	return params, nil
}

// getEnvCfg получает конфигурацию из переменных окружений
func getEnvCfg(envVars []string) Params {
	params := make(map[string]any)
	for _, env := range envVars {
		if val, ok := os.LookupEnv(env); ok {
			params[strings.ToLower(env)] = val
		}
	}

	return params
}

func merge(params ...Params) Params {
	output := make(map[string]any)
	for _, p := range params {
		for k, v := range p {
			output[k] = v
		}
	}
	return output
}

func bindParams(params Params, output any) error {
	c := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToIPNetHookFunc(),
		),
		WeaklyTypedInput: true,
		Metadata:         nil,
		Result:           output,
	}
	decoder, err := mapstructure.NewDecoder(c)
	if err != nil {
		return err
	}

	err = decoder.Decode(params)
	if err != nil {
		return err
	}

	return nil
}

func hasSchema(addr string) bool {
	return strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://")
}
