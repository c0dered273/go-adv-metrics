package client

import (
	"context"
	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/go-resty/resty/v2"
	"sync"
	"time"
)

const (
	updateEndpoint = "/update"
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second

	retryCount       = 3
	retryWaitTime    = 5 * time.Second
	retryMaxWaitTime = 15 * time.Second
)

type Settings struct {
	ServerAddr string
}

type MetricClient struct {
	Ctx      context.Context
	Wg       *sync.WaitGroup
	Settings Settings
}

func NewMetricClient(ctx context.Context, wg *sync.WaitGroup, settings Settings) MetricClient {
	if len(settings.ServerAddr) == 0 {
		settings.ServerAddr = "http://localhost:8080"
	}
	return MetricClient{
		ctx,
		wg,
		settings,
	}
}

func (c *MetricClient) update(container *metric.Container, metricUpdate *metricUpdate) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		metricUpdate.set(container.UpdateAndGet())
		select {
		case <-ticker.C:
			continue
		case <-c.Ctx.Done():
			c.Wg.Done()
			return
		}
	}
}

func (c *MetricClient) send(client *resty.Client, metricUpdate *metricUpdate) {
	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()
	for {
		metrics := metricUpdate.get()
		if len(metrics) != 0 {
			for _, m := range metrics {
				err := c.sendMetric(client, m)
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

func (c *MetricClient) sendMetric(client *resty.Client, metric metric.Metric) error {
	pathParams := map[string]string{
		"type":  metric.Type.String(),
		"name":  metric.Name,
		"value": metric.Value,
	}
	response, err := client.R().
		SetContext(c.Ctx).
		SetHeader("Content-Type", "text/plain").
		SetPathParams(pathParams).
		Post(c.Settings.ServerAddr + updateEndpoint + "/{type}/{name}/{value}")
	if err != nil {
		return err
	}
	log.Info.Printf("Metric update success %v %v %v", response.StatusCode(), response.Request.Method, response.Request.URL)
	return nil
}

type metricUpdate struct {
	mu    sync.RWMutex
	value []metric.Metric
}

func (m *metricUpdate) set(newValue []metric.Metric) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.value = newValue
}

func (m *metricUpdate) get() []metric.Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.value
}

func (c *MetricClient) SendUpdateContinuously(container *metric.Container) {
	client := resty.New()
	client.
		SetRetryCount(retryCount).
		SetRetryWaitTime(retryWaitTime).
		SetRetryMaxWaitTime(retryMaxWaitTime)

	var metricUpdate metricUpdate

	go c.update(container, &metricUpdate)
	time.AfterFunc(10*time.Millisecond, func() {
		c.send(client, &metricUpdate)
	})
}

func (c *MetricClient) SendAllMetricsContinuously() {
	container := metric.NewContainer([]metric.Source{
		metric.NewMemStats(),
	})
	c.SendUpdateContinuously(&container)
}
