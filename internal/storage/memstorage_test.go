package storage

import (
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	metrics := []metric.Metric{
		metric.NewGaugeMetric("FirstGauge", 31337.1),
		metric.NewCounterMetric("FirstCounter", 12345),
	}

	storage := NewMemStorage()

	for _, m := range metrics {
		_, err := storage.Save(m)
		if err != nil {
			return
		}
	}

	result, _ := storage.FindAll()
	assert.Equal(t, metrics, result)
}
