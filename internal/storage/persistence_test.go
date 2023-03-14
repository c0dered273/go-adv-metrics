package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/stretchr/testify/assert"
)

func TestPersistenceRepo_Save(t *testing.T) {
	type fields struct {
		Repository Repository
	}
	type args struct {
		ctx       context.Context
		newMetric metric.Metric
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    metric.Metric
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should successfully save gauge metric",
			fields: fields{
				Repository: NewMemStorage(),
			},
			args: args{
				ctx:       context.Background(),
				newMetric: metric.NewGaugeMetric("TestGauge1", 123.456),
			},
			want:    metric.NewGaugeMetric("TestGauge1", 123.456),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PersistenceRepo{
				Repository: tt.fields.Repository,
			}
			err := p.Save(tt.args.ctx, tt.args.newMetric)
			if err != nil {
				tt.wantErr(t, err, fmt.Sprintf("Save(%v, %v)", tt.args.ctx, tt.args.newMetric))
				return
			}

			got, err := tt.fields.Repository.FindByID(tt.args.ctx, tt.args.newMetric)
			if err != nil {
				tt.wantErr(t, err, fmt.Sprintf("FindByID(%v, %v)", tt.args.ctx, tt.args.newMetric))
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPersistenceRepo_SaveCounter(t *testing.T) {
	type fields struct {
		Repository Repository
	}
	type args struct {
		ctx       context.Context
		newMetric metric.Metric
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    metric.Metric
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should successfully save counter metric",
			fields: fields{
				Repository: NewMemStorage(),
			},
			args: args{
				ctx:       context.Background(),
				newMetric: metric.NewCounterMetric("TestCounter1", 12),
			},
			want:    metric.NewCounterMetric("TestCounter1", 24),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PersistenceRepo{
				Repository: tt.fields.Repository,
			}
			err := p.Save(tt.args.ctx, tt.args.newMetric)
			if err != nil {
				tt.wantErr(t, err, fmt.Sprintf("Save(%v, %v)", tt.args.ctx, tt.args.newMetric))
				return
			}
			err = p.Save(tt.args.ctx, tt.args.newMetric)
			if err != nil {
				tt.wantErr(t, err, fmt.Sprintf("Save(%v, %v)", tt.args.ctx, tt.args.newMetric))
				return
			}

			got, err := tt.fields.Repository.FindByID(tt.args.ctx, tt.args.newMetric)
			if err != nil {
				tt.wantErr(t, err, fmt.Sprintf("FindByID(%v, %v)", tt.args.ctx, tt.args.newMetric))
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
