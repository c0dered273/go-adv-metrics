package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	clients "github.com/c0dered273/go-adv-metrics/internal/agent"
	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/log/agent"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/rs/zerolog/log"
)

//	@Title			Metrics collection agrnt
//	@Description	Агент для сбора и отправки метрик.
//	@Version		1.0
//  Для сборки сервера с заполнением соответствующих переменных необходимо использовать флаги линковщика
//    go build -ldflags "-X main.buildVersion=v0.0.1 -X 'main.buildDate=$(date +'%Y/%m/%d')' -X 'main.buildCommit=$(git rev-parse HEAD)'"

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())

	logger := agent.NewAgentLogger()
	cfg, err := config.NewAgentConfig(logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("agent: configuration error")
	}

	var wg sync.WaitGroup
	wg.Add(2)

	metricClient, err := clients.NewMetricAgent(ctx, &wg, cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("agent: init failed")
	}

	metricClient.SendAllMetricsContinuously(
		metric.ConcatSources(
			metric.NewMemStats(),
			metric.NewPsUtilStats(),
		))

	<-shutdown
	cancel()
	wg.Wait()
	log.Info().Msg("Metrics agent shutdown")
}
