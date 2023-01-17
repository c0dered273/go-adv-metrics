package storage

import (
	"context"

	"github.com/c0dered273/go-adv-metrics/internal/metric"
)

type Repository interface {
	Save(context.Context, metric.Metric) error
	SaveAll(context.Context, []metric.Metric) error
	FindByID(context.Context, metric.Metric) (metric.Metric, error)
	FindAll(context.Context) ([]metric.Metric, error)
	Ping() error
	Close() error
}
