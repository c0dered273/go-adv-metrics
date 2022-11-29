package storage

import (
	"fmt"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
)

var (
	str = make(map[string]metric.Metric)
)

type MemStorage struct {
}

func (m *MemStorage) Save(newMetric metric.Metric) error {
	str[getID(newMetric)] = newMetric
	return nil
}

func (m *MemStorage) FindAll() (metrics []metric.Metric, err error) {
	var result []metric.Metric
	for _, v := range str {
		result = append(result, v)
	}
	return result, nil
}

func (m *MemStorage) FindByID(newMetric metric.Metric) (metric metric.Metric, err error) {
	if result, ok := str[getID(newMetric)]; ok {
		return result, nil
	}
	return metric, fmt.Errorf("storage: not found: %v %v", newMetric.GetName(), newMetric.GetType())
}

func getID(newMetric metric.Metric) string {
	return newMetric.GetName() + newMetric.GetType().String()
}

func GetMemStorage() *MemStorage {
	return &MemStorage{}
}
