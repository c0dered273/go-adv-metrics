package client

import (
	"context"
	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"net/http"
	"sync"
	"time"
)

type MetricUpdate struct {
	mu    sync.RWMutex
	value []metric.Metric
}

func (m *MetricUpdate) Set(newValue []metric.Metric) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.value = newValue
}

func (m *MetricUpdate) Get() []metric.Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.value
}

type Settings struct {
	UpdateEndpoint string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

type MetricClient struct {
	Client   *http.Client
	Ctx      context.Context
	Wg       *sync.WaitGroup
	Settings Settings
}

func NewMetricClient(
	client *http.Client,
	ctx context.Context,
	wg *sync.WaitGroup,
	updateEndpoint string,
	pollInterval time.Duration,
	ReportInterval time.Duration) MetricClient {
	return MetricClient{
		client,
		ctx,
		wg,
		Settings{
			updateEndpoint,
			pollInterval,
			ReportInterval,
		},
	}
}

func sendMetric(client *http.Client, endpoint string) error {
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

func (c *MetricClient) updateMetrics(container *metric.Container, metricUpdate *MetricUpdate) {
	ticker := time.NewTicker(c.Settings.PollInterval)
	defer ticker.Stop()
	for {
		metricUpdate.Set(container.UpdateAndGet())
		select {
		case <-ticker.C:
			continue
		case <-c.Ctx.Done():
			c.Wg.Done()
			return
		}
	}
}

func (c *MetricClient) sendAllMetrics(metricUpdate *MetricUpdate) {
	ticker := time.NewTicker(c.Settings.ReportInterval)
	defer ticker.Stop()
	for {
		metricsSlice := metricUpdate.Get()
		if len(metricsSlice) != 0 {
			for _, m := range metricsSlice {
				endpoint := c.Settings.UpdateEndpoint + m.String()
				err := sendMetric(c.Client, endpoint)
				if err != nil {
					log.Error.Println("Unable to send update request ", err)
				}
			}
		}
		select {
		case <-ticker.C:
			continue
		case <-c.Ctx.Done():
			c.Wg.Done()
			return
		}
	}
}

func (c *MetricClient) SendContinuously(m *metric.Container) {
	var metricUpdate MetricUpdate
	go c.updateMetrics(m, &metricUpdate)
	time.AfterFunc(50*time.Millisecond, func() {
		c.sendAllMetrics(&metricUpdate)
	})
}
