package storage

import (
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	metrics := []metric.Metric{
		{
			Name:  "Test1",
			Type:  metric.GaugeType,
			Value: "31337.000",
		},
		{
			Name:  "Test2",
			Type:  metric.CounterType,
			Value: "123456",
		},
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
