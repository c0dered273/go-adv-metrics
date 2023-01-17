package storage

import (
	"context"
	"testing"

	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	metrics := []metric.Metric{
		metric.NewGaugeMetric("FirstGauge", 31337.1),
		metric.NewCounterMetric("FirstCounter", 12345),
	}

	storage := NewMemStorage()

	for _, m := range metrics {
		err := storage.Save(context.Background(), m)
		if err != nil {
			panic(err)
		}
	}

	result, _ := storage.FindAll(context.Background())
	assert.ElementsMatch(t, metrics, result)
}
