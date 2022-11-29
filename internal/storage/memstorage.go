package storage

import "github.com/c0dered273/go-adv-metrics/internal/metric"

var str []metric.Metric

type MemStorage struct {
}

func (m *MemStorage) Save(newMetric metric.Metric) (saved metric.Metric, err error) {
	if len(str) == 0 {
		str = append(str, newMetric)
	}
	for i := 0; i < len(str); i++ {
		if newMetric.GetName() == str[i].GetName() && newMetric.GetType() == str[i].GetType() {
			m, appError := metric.NewMetric(newMetric.GetName(), newMetric.GetType().String(), newMetric.GetStringValue())
			if err != nil {
				return saved, appError.Error
			}
			str[i] = m
			break
		} else {
			str = append(str, newMetric)
		}
	}
	return newMetric, nil
}

func (m *MemStorage) FindAll() (metrics []metric.Metric, err error) {
	return str, nil
}

func GetMemStorage() *MemStorage {
	return &MemStorage{}
}
