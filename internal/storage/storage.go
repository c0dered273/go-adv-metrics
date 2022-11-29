package storage

import "github.com/c0dered273/go-adv-metrics/internal/metric"

type Repository interface {
	Save(metric.Metric) (saved metric.Metric, err error)
	FindAll() (metrics []metric.Metric, err error)
}
