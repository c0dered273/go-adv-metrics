package storage

import "github.com/c0dered273/go-adv-metrics/internal/metric"

type Repository interface {
	Save(metric.Metric) error
	FindAll() ([]metric.Metric, error)
	FindByID(metric.Metric) (metric.Metric, error)
}
