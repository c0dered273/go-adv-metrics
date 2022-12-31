package agent

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/c0dered273/go-adv-metrics/internal/service"
	"github.com/go-resty/resty/v2"
)

const (
	updateEndpoint   = "/update/"
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

type MetricAgent struct {
	Ctx    context.Context
	Wg     *sync.WaitGroup
	Config *service.AgentConfig
	client *resty.Client
}

func NewMetricAgent(ctx context.Context, wg *sync.WaitGroup, config *service.AgentConfig) MetricAgent {
	client := resty.New()
	client.
		SetRetryCount(retryCount).
		SetRetryWaitTime(retryWaitTime).
		SetRetryMaxWaitTime(retryMaxWaitTime)
	return MetricAgent{
		Ctx:    ctx,
		Wg:     wg,
		Config: config,
		client: client,
	}
}

func (c *MetricAgent) update(mUpdate metric.Updatable, metricUpdate *metricUpdate) {
	ticker := time.NewTicker(c.Config.PollInterval)
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

func (c *MetricAgent) send(metricUpdate *metricUpdate) {
	ticker := time.NewTicker(c.Config.ReportInterval)
	defer ticker.Stop()
	for {
		metrics := metricUpdate.get()
		if len(metrics) != 0 {
			for _, m := range metrics {
				m.SetHash(c.Config.Key)
				err := c.postMetric(m)
				if err != nil {
					log.Error.Println("unable to send update request ", err)
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

func (c *MetricAgent) postMetric(newMetric metric.Metric) error {
	body, marshErr := json.Marshal(newMetric)
	if marshErr != nil {
		return marshErr
	}

	response, err := c.client.R().
		SetContext(c.Ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(c.Config.Address + updateEndpoint)
	if err != nil {
		return err
	}
	if response.IsSuccess() {
		log.Info.Printf("Metric update success %v %v %v: %v", response.StatusCode(), response.Request.Method, response.Request.URL, newMetric.String())
	} else {
		log.Info.Printf("Unable to update metric, status %d - %v %v: %v", response.StatusCode(), response.Request.Method, response.Request.URL, newMetric.String())
	}
	return nil
}

func (c *MetricAgent) SendUpdateContinuously(mUpdate metric.Updatable) {
	var metricUpdate metricUpdate

	go c.update(mUpdate, &metricUpdate)
	time.AfterFunc(10*time.Millisecond, func() {
		c.send(&metricUpdate)
	})
}

func (c *MetricAgent) SendAllMetricsContinuously() {
	allMetrics := metric.GetUpdatable(
		metric.NewMemStats,
	)
	c.SendUpdateContinuously(allMetrics)
}
