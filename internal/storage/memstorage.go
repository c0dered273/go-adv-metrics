package storage

import "github.com/c0dered273/go-adv-metrics/internal/metric"

var str []metric.Metric

type MemStorage struct {
}

func (m *MemStorage) Save(metric metric.Metric) (saved metric.Metric, err error) {
	str = append(str, metric)
	return metric, nil
}

func (m *MemStorage) FindAll() (metrics []metric.Metric, err error) {
	return str, nil
}

func NewMemStorage() *MemStorage {
	return &MemStorage{}
}
