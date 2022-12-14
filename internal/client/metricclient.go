package client

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/go-resty/resty/v2"
)

const (
	updateEndpoint = "/update/"
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second

	retryCount       = 3
	retryWaitTime    = 5 * time.Second
	retryMaxWaitTime = 15 * time.Second
)

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

type Settings struct {
	ServerAddr string
}

type MetricClient struct {
	Ctx      context.Context
	Wg       *sync.WaitGroup
	Settings Settings
	client   *resty.Client
}

func NewMetricClient(ctx context.Context, wg *sync.WaitGroup, settings Settings) MetricClient {
	if len(settings.ServerAddr) == 0 {
		settings.ServerAddr = "http://localhost:8080"
	}
	client := resty.New()
	client.
		SetRetryCount(retryCount).
		SetRetryWaitTime(retryWaitTime).
		SetRetryMaxWaitTime(retryMaxWaitTime)
	return MetricClient{
		Ctx:      ctx,
		Wg:       wg,
		Settings: settings,
		client:   client,
	}
}

func (c *MetricClient) update(mUpdate metric.Updatable, metricUpdate *metricUpdate) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		metricUpdate.set(mUpdate())
		select {
		case <-ticker.C:
			continue
		case <-c.Ctx.Done():
			c.Wg.Done()
			return
		}
	}
}

func (c *MetricClient) send(metricUpdate *metricUpdate) {
	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()
	for {
		metrics := metricUpdate.get()
		if len(metrics) != 0 {
			for _, m := range metrics {
				err := c.postMetric(m)
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

func (c *MetricClient) postMetric(metric metric.Metric) error {
	body, marshErr := json.Marshal(metric)
	if marshErr != nil {
		return marshErr
	}

	response, err := c.client.R().
		SetContext(c.Ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(c.Settings.ServerAddr + updateEndpoint)
	if err != nil {
		return err
	}
	log.Info.Printf("Metric update success %v %v %v", response.StatusCode(), response.Request.Method, response.Request.URL)
	return nil
}

func (c *MetricClient) SendUpdateContinuously(mUpdate metric.Updatable) {
	var metricUpdate metricUpdate

	go c.update(mUpdate, &metricUpdate)
	time.AfterFunc(10*time.Millisecond, func() {
		c.send(&metricUpdate)
	})
}

func (c *MetricClient) SendAllMetricsContinuously() {
	allMetrics := metric.GetUpdatable(
		metric.NewMemStats,
	)
	c.SendUpdateContinuously(allMetrics)
}
