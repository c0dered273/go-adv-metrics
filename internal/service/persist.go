package service

import (
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
)

type PersistMetric struct {
	Repo storage.Repository
}

func (p *PersistMetric) SaveMetric(newMetric metric.Metric) error {
	switch newMetric.GetType() {
	case metric.Gauge:
		{
			err := p.Repo.Save(newMetric)
			if err != nil {
				return err
			}
		}
	case metric.Counter:
		{
			existMetric, fndErr := p.Repo.FindById(newMetric)
			if fndErr != nil {
				err := p.Repo.Save(newMetric)
				if err != nil {
					return err
				}
				return nil
			}
			newValue := existMetric.GetCounterValue() + newMetric.GetCounterValue()
			err := p.Repo.Save(metric.NewCounterMetric(existMetric.GetName(), newValue))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
