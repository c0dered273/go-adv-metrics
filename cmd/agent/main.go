package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	clients "github.com/c0dered273/go-adv-metrics/internal/agent"
	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/service"
)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())

	cfg := service.NewAgentConfig()

	var wg sync.WaitGroup
	wg.Add(2)

	metricClient := clients.NewMetricAgent(ctx, &wg, cfg)
	metricClient.SendAllMetricsContinuously()

	<-shutdown
	cancel()
	wg.Wait()
	log.Info.Println("Metrics agent shutdown")
}
