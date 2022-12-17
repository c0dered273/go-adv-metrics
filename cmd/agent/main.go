package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	clients "github.com/c0dered273/go-adv-metrics/internal/client"
	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/caarlos0/env/v6"
)

func main() {
	var cfg config.Agent
	if err := env.Parse(&cfg); err != nil {
		log.Error.Fatal(err)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(2)

	metricClient := clients.NewMetricClient(ctx, &wg, cfg)
	metricClient.SendAllMetricsContinuously()

	<-shutdown
	cancel()
	wg.Wait()
	log.Info.Println("Metrics agent shutdown")
}
