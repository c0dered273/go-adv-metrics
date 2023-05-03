// Package agent модуль периодически обновляет метрики и корзинами отправляет на сервер
package agent

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/c0dered273/go-adv-metrics/internal/model"
	"github.com/c0dered273/go-adv-metrics/internal/service"
	"github.com/go-resty/resty/v2"
	"google.golang.org/grpc/metadata"
)

// Настройки отправки обновлений от агента
const (
	updateEndpoint   = "/updates/"
	connTimeout      = 5 * time.Second
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
	client Client
	buffer []metric.UpdatableMetric
}

type HTTPClient struct {
	ctx    context.Context
	config *config.AgentConfig
	client *resty.Client
}

func (c *HTTPClient) PostMetric(metrics []metric.UpdatableMetric) error {
	body, err := c.encryptBody(metrics, c.config.PublicKey)
	if err != nil {
		return err
	}

	response, err := c.client.R().
		SetContext(c.ctx).
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Real-IP", getPreferredHostIP(c.config.Address)).
		SetBody(body).
		Post(c.config.Address + updateEndpoint)
	if err != nil {
		return err
	}
	if response.IsSuccess() {
		c.config.Logger.
			Info().
			Int("status_code", response.StatusCode()).
			Str("method", response.Request.Method).
			Str("url", response.Request.URL).
			Msg("send update success")
	} else {
		c.config.Logger.
			Error().
			Int("status_code", response.StatusCode()).
			Str("method", response.Request.Method).
			Str("url", response.Request.URL).
			Msg("agent: metric update failed")
	}
	return nil
}

func (c *HTTPClient) encryptBody(metrics []metric.UpdatableMetric, key *rsa.PublicKey) (any, error) {
	if key != nil {
		m, err := json.Marshal(metrics)
		if err != nil {
			return nil, err
		}

		return rsa.EncryptOAEP(sha256.New(), rand.Reader, key, m, nil)
	}

	return metrics, nil
}

type GRPCClient struct {
	ctx          context.Context
	cfg          *config.AgentConfig
	metricClient service.MetricsServiceClient
}

func (c *GRPCClient) PostMetric(metrics []metric.UpdatableMetric) error {
	pbMetrics := make([]*model.Metric, len(metrics))
	err := service.MapSliceWithSerialization(service.ToSliceOfPointers(metrics), pbMetrics)
	if err != nil {
		return err
	}

	md := metadata.New(map[string]string{
		"X-Real-IP": getPreferredHostIP(c.cfg.Address),
	})
	outCtx := metadata.NewOutgoingContext(c.ctx, md)

	_, err = c.metricClient.SaveAll(outCtx, &model.Metrics{Metrics: pbMetrics})
	if err != nil {
		return err
	}

	return nil
}

// NewMetricAgent возвращает настроенного агента
func NewMetricAgent(ctx context.Context, wg *sync.WaitGroup, config *config.AgentConfig) (Agent, error) {
	var client Client
	var err error

	if config.GRPCClient {
		client, err = NewGRPCClient(ctx, config)
		if err != nil {
			return nil, err
		}
	} else {
		client, err = NewHTTPClient(ctx, config)
		if err != nil {
			return nil, err
		}
	}

	return &MetricAgent{
		Ctx:    ctx,
		Wg:     wg,
		Config: config,
		client: client,
		buffer: make([]metric.UpdatableMetric, 0, BufferLen),
	}, nil
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
				err := ma.client.PostMetric(ma.buffer)
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
			err := ma.client.PostMetric(ma.buffer)
			if err != nil {
				ma.Config.Logger.Error().Err(err).Msg("agent: failed to send update request")
			}
			ma.Wg.Done()
			return
		}
	}
}

func getPreferredHostIP(target string) string {
	targetURL, err := url.Parse(target)
	if err != nil {
		return ""
	}

	targetIPs, err := net.LookupIP(targetURL.Hostname())
	if err != nil {
		return ""
	}

	var targetIP net.IP
	for _, i := range targetIPs {
		if ip4 := i.To4(); ip4 != nil {
			targetIP = ip4
			break
		}
	}

	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addresses {
		if ipNet, ok := addr.(*net.IPNet); ok {
			if ipNet.Contains(targetIP) {
				if ipNet.IP.To4() != nil {
					return ipNet.IP.String()
				}
			}
		}
	}

	return ""
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
