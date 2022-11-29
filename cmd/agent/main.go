package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	clients "github.com/c0dered273/go-adv-metrics/internal/client"
	"github.com/c0dered273/go-adv-metrics/internal/log"
)

const (
	serverAddr = "http://127.0.0.1:8080"
)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(2)

	metricClient := clients.NewMetricClient(ctx, &wg, clients.Settings{ServerAddr: serverAddr})
	metricClient.SendAllMetricsContinuously()

	<-shutdown
	cancel()
	wg.Wait()
	log.Info.Println("Metrics agent shutdown")
}
