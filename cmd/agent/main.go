package main

import (
	"context"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	clients "github.com/c0dered273/go-adv-metrics/internal/client"
	"github.com/c0dered273/go-adv-metrics/internal/log"
)

const (
	updateEndpoint    = "http://127.0.0.1:8080/update"
	pollInterval      = 2 * time.Second
	reportInterval    = 10 * time.Second
	connectionTimeout = 15 * time.Second
)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	log.Info.Println("Metrics agent started")
	transport := &http.Transport{}
	transport.MaxIdleConns = 5
	client := http.Client{}
	client.Transport = transport
	client.Timeout = connectionTimeout

	var wg sync.WaitGroup
	wg.Add(2)

	metricClient := clients.NewMetricClient(&client, ctx, &wg, updateEndpoint, pollInterval, reportInterval)
	allMetrics := metric.NewContainer([]metric.Source{
		metric.NewMemStats(),
	})
	metricClient.SendContinuously(&allMetrics)

	<-shutdown
	cancel()
	wg.Wait()
	log.Info.Println("Metrics agent shutdown")
}
