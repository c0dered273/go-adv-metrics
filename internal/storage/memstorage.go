package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/c0dered273/go-adv-metrics/internal/metric"
)

type MemStorage struct {
	mx  *sync.RWMutex
	str map[string]metric.Metric
}

func (m *MemStorage) get(key string) (metric.Metric, bool) {
	m.mx.RLock()
	defer m.mx.RUnlock()
	v, ok := m.str[key]
	return v, ok
}

func (m *MemStorage) put(key string, value metric.Metric) {
	m.mx.Lock()
	defer m.mx.Unlock()
	m.str[key] = value
}

func (m *MemStorage) iterateValues() <-chan metric.Metric {
	c := make(chan metric.Metric)
	go func() {
		m.mx.RLock()
		defer m.mx.RUnlock()
		for _, v := range m.str {
			c <- v
		}
		close(c)
	}()
	return c
}

func (m *MemStorage) Save(ctx context.Context, newMetric metric.Metric) error {
	m.put(getID(newMetric), newMetric)
	return nil
}

func (m *MemStorage) SaveAll(ctx context.Context, metrics []metric.Metric) error {
	for _, mtr := range metrics {
		m.put(getID(mtr), mtr)
	}
	return nil
}

func (m *MemStorage) FindByID(ctx context.Context, keyMetric metric.Metric) (metric metric.Metric, err error) {
	if result, ok := m.get(getID(keyMetric)); ok {
		return result, nil
	}
	return metric, fmt.Errorf("storage: not found: %v %v", keyMetric.GetName(), keyMetric.GetType())
}

func (m *MemStorage) FindAll(ctx context.Context) (metrics []metric.Metric, err error) {
	var result []metric.Metric
	for v := range m.iterateValues() {
		result = append(result, v)
	}
	return result, nil
}

func (m *MemStorage) Ping() error {
	return nil
}

func (m *MemStorage) Close() error {
	return nil
}

func getID(newMetric metric.Metric) string {
	return newMetric.GetName() + newMetric.GetType().String()
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		str: make(map[string]metric.Metric),
		mx:  new(sync.RWMutex),
	}
}
