package agent

import (
	"log"
	"os"

	"github.com/rs/zerolog"
)

// NewAgentLogger настраивает логгер для клиента. Лог пишется в консоль и в файл agent.log
func NewAgentLogger() zerolog.Logger {
	fileWriter, err := os.OpenFile("agent.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		log.Panicf("logger: failed to open log file: %v", err)
	}

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	multi := zerolog.MultiLevelWriter(consoleWriter, fileWriter)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	return zerolog.New(multi).With().Timestamp().Caller().Logger()
}
