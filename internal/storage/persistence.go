package storage

import (
	"context"

	"github.com/c0dered273/go-adv-metrics/internal/metric"
)

type PersistenceRepo struct {
	Repository
}

func (p *PersistenceRepo) Save(ctx context.Context, newMetric metric.Metric) error {
	switch newMetric.GetType() {
	case metric.Gauge:
		{
			err := p.Repository.Save(ctx, newMetric)
			if err != nil {
				return err
			}
		}
	case metric.Counter:
		{
			existMetric, fndErr := p.FindByID(ctx, newMetric)
			if fndErr != nil {
				err := p.Repository.Save(ctx, newMetric)
				if err != nil {
					return err
				}
				return nil
			}
			newValue := existMetric.GetCounterValue() + newMetric.GetCounterValue()
			err := p.Repository.Save(ctx, metric.NewCounterMetric(existMetric.GetName(), newValue))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func NewPersistenceRepo(r Repository) Repository {
	return &PersistenceRepo{Repository: r}
}
