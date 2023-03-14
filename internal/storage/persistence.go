package storage

import (
	"context"

	"github.com/c0dered273/go-adv-metrics/internal/metric"
)

// PersistenceRepo Репозиторий применяется как промежуточный для подключения логики сохранения метрик,
// в зависимости от типа.
// Используется с MemStorage и FileStorage
type PersistenceRepo struct {
	Repository
}

// Save метод содержит логику сохранения и обновления для каждого типа метрики.
// Gauge - если метрика с таким типом и именем уже существует, значение метрики обновляется на новое
// Counter - если метрика с таким типом и именем уже существует, новое значение прибавляется к существующему
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
