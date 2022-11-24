package main

import (
	"context"
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

func sendUpdate(client *http.Client, endpoint string) error {
	request, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "text/plain")

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	log.Info.Printf("Metric update success %v %v", response.Request.Method, response.Request.URL)
	return nil
}

func updateMetrics(metricUpdate *MetricUpdate, ctx context.Context, wg *sync.WaitGroup) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		metricUpdate.Set(metrics.GetAllMetrics())
		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			wg.Done()
			return
		}
	}
}

func sendAllMetrics(client *http.Client, metricUpdate *MetricUpdate, ctx context.Context, wg *sync.WaitGroup) {
	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()
	for {
		metricsSlice := metricUpdate.Get()
		if len(metricsSlice) != 0 {
			for _, m := range metricsSlice {
				endpoint := updateEndpoint + m.String()
				err := sendUpdate(client, endpoint)
				if err != nil {
					log.Error.Printf("Unable to send update request %v", err)
				}
			}
		}
		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			wg.Done()
			return
		}
	}
}

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

	var metricUpdate MetricUpdate
	go updateMetrics(&metricUpdate, ctx, &wg)
	time.AfterFunc(100*time.Millisecond, func() {
		sendAllMetrics(&client, &metricUpdate, ctx, &wg)
	})

	<-shutdown
	cancel()
	wg.Wait()
	log.Info.Println("Metrics agent shutdown")
}
