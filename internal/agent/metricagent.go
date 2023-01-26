package agent

import (
	"context"
	"sync"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/c0dered273/go-adv-metrics/internal/service"
	"github.com/go-resty/resty/v2"
)

const (
	updateEndpoint   = "/updates/"
	retryCount       = 3
	retryWaitTime    = 5 * time.Second
	retryMaxWaitTime = 15 * time.Second
	BufferLen        = 10
)

type metricUpdate struct {
	mu    *sync.RWMutex
	value []metric.UpdatableMetric
}

func (m *metricUpdate) set(newValue []metric.UpdatableMetric) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.value = newValue
}

func (m *metricUpdate) get() []metric.UpdatableMetric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.value
}

type MetricAgent struct {
	Ctx    context.Context
	Wg     *sync.WaitGroup
	Config *service.AgentConfig
	client *resty.Client
	buffer []metric.UpdatableMetric
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
		buffer: make([]metric.UpdatableMetric, 0, BufferLen),
	}
}

func (ma *MetricAgent) update(allMetrics []metric.UpdatableMetric, metricUpdate *metricUpdate) {
	ticker := time.NewTicker(ma.Config.PollInterval)
	defer ticker.Stop()
	for {
		for i := range allMetrics {
			allMetrics[i].Update()
		}

		metricUpdate.set(allMetrics)

		select {
		case <-ticker.C:
			continue
		case <-ma.Ctx.Done():
			ma.Wg.Done()
			return
		}
	}
}

func (ma *MetricAgent) send(metricUpdate *metricUpdate) {
	ticker := time.NewTicker(ma.Config.ReportInterval)
	defer ticker.Stop()
	for {
		updated := metricUpdate.get()
		for i := range updated {
			updated[i].SetHash(ma.Config.Key)
			ma.buffer = append(ma.buffer, updated[i])

			if cap(ma.buffer) == len(ma.buffer) {
				err := ma.postMetric(ma.buffer)
				if err != nil {
					ma.Config.Logger.Error().Err(err).Msg("agent: failed to send update request")
				}
				ma.buffer = ma.buffer[:0]
			}
		}

		select {
		case <-ticker.C:
			continue
		case <-ma.Ctx.Done():
			err := ma.postMetric(ma.buffer)
			if err != nil {
				ma.Config.Logger.Error().Err(err).Msg("agent: failed to send update request")
			}
			ma.Wg.Done()
			return
		}
	}
}

func (ma *MetricAgent) postMetric(metrics []metric.UpdatableMetric) error {
	response, err := ma.client.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetBody(metrics).
		Post(ma.Config.Address + updateEndpoint)
	if err != nil {
		return err
	}
	if response.IsSuccess() {
		ma.Config.Logger.
			Info().
			Int("status_code", response.StatusCode()).
			Str("method", response.Request.Method).
			Str("url", response.Request.URL).
			Msg("send update success")
	} else {
		ma.Config.Logger.
			Error().
			Int("status_code", response.StatusCode()).
			Str("method", response.Request.Method).
			Str("url", response.Request.URL).
			Msg("agent: metric update failed")
	}
	return nil
}

func (ma *MetricAgent) SendAllMetricsContinuously(allMetrics []metric.UpdatableMetric) {
	mUpdate := &metricUpdate{
		mu:    new(sync.RWMutex),
		value: make([]metric.UpdatableMetric, 0),
	}

	go ma.update(allMetrics, mUpdate)

	time.AfterFunc(10*time.Millisecond, func() {
		ma.send(mUpdate)
	})
}
