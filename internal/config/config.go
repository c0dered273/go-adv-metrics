package config

import (
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Настройки агента по умолчанию.
// Если не передана строка соединения с БД, метрики хранятся локально в json файле.
const (
	// Address Адрес сервера
	Address = "localhost:8080"
	// StoreInterval Интервал сброса метрик на диск
	StoreInterval = 300 * time.Second
	// StoreFile Путь к файлу хранения метрик
	StoreFile = "/tmp/devops-metrics-db.json"
	// Restore Флаг показывает сохранять ли метрики с прошлого сеанса или очистить БД
	Restore = true

	// ReportInterval Интервал отправки обновлений на сервер
	ReportInterval = 10 * time.Second
	// PollInterval Интервал обновления метрик
	PollInterval = 2 * time.Second
)

var (
	// Тип файла конфигурации
	configFileType = "json"
	configFilePath = []string{
		".",
	}
)

func bindPFlags() error {
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return err
	}
	return nil
}

func bindConfigFile(fileNameFlag string) error {
	cfgFileFlag := pflag.Lookup(fileNameFlag)
	if cfgFileFlag != nil {
		cfgFileName := cfgFileFlag.Value.String()
		if len(cfgFileName) > 0 {
			err := readConfigFile(cfgFileName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func readConfigFile(filename string) error {
	viper.SetConfigName(filename)
	viper.SetConfigType(configFileType)
	for _, path := range configFilePath {
		viper.AddConfigPath(path)
	}
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return err
		} else {
			return err
		}
	}
	return nil
}

func bindEnvVars(envVars []string) error {
	for _, env := range envVars {
		err := viper.BindEnv(env)
		if err != nil {
			return err
		}
	}
	return nil
}

func hasSchema(addr string) bool {
	return strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://")
}
