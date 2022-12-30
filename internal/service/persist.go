package service

import (
	"context"

	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
)

type PersistMetric struct {
	Repo storage.Repository
}

func (p *PersistMetric) SaveMetric(ctx context.Context, newMetric metric.Metric) error {
	switch newMetric.GetType() {
	case metric.Gauge:
		{
			err := p.Repo.Save(ctx, newMetric)
			if err != nil {
				return err
			}
		}
	case metric.Counter:
		{
			existMetric, fndErr := p.Repo.FindByID(ctx, newMetric)
			if fndErr != nil {
				err := p.Repo.Save(ctx, newMetric)
				if err != nil {
					return err
				}
				return nil
			}
			newValue := existMetric.GetCounterValue() + newMetric.GetCounterValue()
			err := p.Repo.Save(ctx, metric.NewCounterMetric(existMetric.GetName(), newValue))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
