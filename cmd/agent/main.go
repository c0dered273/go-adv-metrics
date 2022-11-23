package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/metrics"
)

const (
	updateEndpoint    = "http://127.0.0.1:8080/update"
	pollInterval      = 2 * time.Second
	reportInterval    = 10 * time.Second
	connectionTimeout = 15 * time.Second
)

type MetricUpdate struct {
	mu    sync.RWMutex
	value []metrics.Metric
}

func (m *MetricUpdate) Set(newValue []metrics.Metric) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.value = newValue
}

func (m *MetricUpdate) Get() []metrics.Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.value
}

func sendUpdate(client *http.Client, endpoint string) {
	request, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		log.Error.Fatal(err)
	}
	request.Header.Set("Content-Type", "text/plain")

	response, err := client.Do(request)
	if err != nil {
		log.Error.Printf("Update request unsuccessful %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error.Fatal("Unable to close response body", err)
		}
	}(response.Body)
	log.Info.Printf("Metric update success %v %v", response.Request.Method, response.Request.URL)
}

func updateMetrics(metricUpdate *MetricUpdate, ctx context.Context) {
	metricUpdate.Set(metrics.GetAllMetrics())

	ticker := time.NewTicker(pollInterval)
	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
			metricUpdate.Set(metrics.GetAllMetrics())
		}
	}
}

func sendAllMetrics(client *http.Client, metricUpdate *MetricUpdate, ctx context.Context) {
	metricsSlice := metricUpdate.Get()
	if len(metricsSlice) != 0 {
		for _, m := range metricsSlice {
			endpoint := updateEndpoint + m.String()
			sendUpdate(client, endpoint)
		}
	}

	ticker := time.NewTicker(reportInterval)
	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
			metricsSlice := metricUpdate.Get()
			for _, m := range metricsSlice {
				endpoint := updateEndpoint + m.String()
				sendUpdate(client, endpoint)
			}
		}
	}
}

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	log.Info.Println("Metrics agent started")
	transport := &http.Transport{}
	transport.MaxIdleConns = 20
	client := http.Client{}
	client.Transport = transport
	client.Timeout = connectionTimeout

	var metricUpdate MetricUpdate
	go updateMetrics(&metricUpdate, ctx)
	time.AfterFunc(100*time.Millisecond, func() {
		sendAllMetrics(&client, &metricUpdate, ctx)
	})

	<-shutdown
	cancel()
	log.Info.Println("Metrics agent shutdown")
}
