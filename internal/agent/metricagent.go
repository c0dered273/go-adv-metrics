// Package agent модуль периодически обновляет метрики и корзинами отправляет на сервер
package agent

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"sync"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/go-resty/resty/v2"
)

// Настройки отправки обновлений от агента
const (
	updateEndpoint   = "/updates/"
	retryCount       = 3
	retryWaitTime    = 5 * time.Second
	retryMaxWaitTime = 15 * time.Second
	BufferLen        = 3
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

// MetricAgent предоставляет методы для обновления и отправки метрик на сервер
type MetricAgent struct {
	Ctx    context.Context
	Wg     *sync.WaitGroup
	Config *config.AgentConfig
	client *resty.Client
	buffer []metric.UpdatableMetric
}

// NewMetricAgent возвращает настроенного агента
func NewMetricAgent(ctx context.Context, wg *sync.WaitGroup, config *config.AgentConfig) MetricAgent {
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
	body, err := encryptBody(metrics, ma.Config.PublicKey)
	if err != nil {
		return err
	}

	response, err := ma.client.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
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

func encryptBody(metrics []metric.UpdatableMetric, key *rsa.PublicKey) ([]byte, error) {
	m, err := json.Marshal(metrics)
	if err != nil {
		return nil, err
	}
	if key != nil {
		return rsa.EncryptOAEP(sha256.New(), rand.Reader, key, m, nil)
	}

	return m, nil
}

// SendAllMetricsContinuously метод инкапсулирует периодическое обновление и отправку метрик
// На вход получает слайс с сырыми метриками и последовательно обновляет каждую
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
