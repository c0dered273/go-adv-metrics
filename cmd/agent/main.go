package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	clients "github.com/c0dered273/go-adv-metrics/internal/agent"
	"github.com/c0dered273/go-adv-metrics/internal/log/agent"
	"github.com/c0dered273/go-adv-metrics/internal/service"
	"github.com/rs/zerolog/log"
)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())

	logger := agent.NewAgentLogger()
	cfg := service.NewAgentConfig(logger)

	var wg sync.WaitGroup
	wg.Add(2)

	metricClient := clients.NewMetricAgent(ctx, &wg, cfg)
	metricClient.SendAllMetricsContinuously()

	<-shutdown
	cancel()
	wg.Wait()
	log.Info().Msg("Metrics agent shutdown")
}
