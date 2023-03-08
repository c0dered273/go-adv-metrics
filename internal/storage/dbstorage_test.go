package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func NewPostgresContainer() (testcontainers.Container, error) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Name:  "goadv_postgres_test",
		Image: "postgres:14",
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "goadv",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
	}

	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            true,
	})
	if err != nil {
		return nil, err
	}

	return postgresC, nil
}

func TestDBStorage_Save(t *testing.T) {
	ctx := context.Background()
	postgresC, err := NewPostgresContainer()
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = postgresC.Terminate(context.Background())
		if err != nil {
			t.Fatalf("failed to terminate container: %s", err.Error())
		}
	}()

	db := NewDBStorage(
		"postgres://postgres:postgres@localhost:5432/goadv",
		false,
		log.Logger,
		ctx,
	)
	defer func() {
		err = db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	tests := []struct {
		name    string
		store   metric.Metric
		want    metric.Metric
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "should successfully save and find gauge metric",
			store:   metric.NewGaugeMetric("TestGauge1", 123.456),
			want:    metric.NewGaugeMetric("TestGauge1", 123.456),
			wantErr: assert.NoError,
		},
		{
			name:    "should return error when metric not exists",
			store:   metric.NewGaugeMetric("TestGauge1", 123.456),
			want:    metric.NewGaugeMetric("Fake", 0),
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = db.Save(ctx, tt.store)
			if err != nil {
				tt.wantErr(t, err, fmt.Sprintf("failed to save metric: %v", tt.store))
				return
			}

			got, err := db.FindByID(ctx, tt.want)
			if err != nil {
				tt.wantErr(t, err, fmt.Sprintf("failed to get metric: %v", tt.want))
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDBStorage_SaveCounter(t *testing.T) {
	ctx := context.Background()
	postgresC, err := NewPostgresContainer()
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = postgresC.Terminate(context.Background())
		if err != nil {
			t.Fatalf("failed to terminate container: %s", err.Error())
		}
	}()

	db := NewDBStorage(
		"postgres://postgres:postgres@localhost:5432/goadv",
		false,
		log.Logger,
		ctx,
	)
	defer func() {
		err = db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	tests := []struct {
		name    string
		store   metric.Metric
		want    metric.Metric
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "should successfully save and find counter metric",
			store:   metric.NewCounterMetric("TestCounter1", 12),
			want:    metric.NewCounterMetric("TestCounter1", 24),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = db.Save(ctx, tt.store)
			if err != nil {
				tt.wantErr(t, err, fmt.Sprintf("failed to save metric: %v", tt.store))
				return
			}
			err = db.Save(ctx, tt.store)
			if err != nil {
				tt.wantErr(t, err, fmt.Sprintf("failed to save metric: %v", tt.store))
				return
			}

			got, err := db.FindByID(ctx, tt.want)
			if err != nil {
				tt.wantErr(t, err, fmt.Sprintf("failed to get metric: %v", tt.want))
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDBStorage_SaveAll(t *testing.T) {
	ctx := context.Background()
	postgresC, err := NewPostgresContainer()
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = postgresC.Terminate(context.Background())
		if err != nil {
			t.Fatalf("failed to terminate container: %s", err.Error())
		}
	}()

	db := NewDBStorage(
		"postgres://postgres:postgres@localhost:5432/goadv",
		false,
		log.Logger,
		ctx,
	)
	defer func() {
		err = db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	tests := []struct {
		name    string
		want    []metric.Metric
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should successfully save and find array of metrics",
			want: []metric.Metric{
				metric.NewGaugeMetric("TestGauge1", 123.456),
				metric.NewCounterMetric("TestCounter1", 789),
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = db.SaveAll(ctx, tt.want)
			if err != nil {
				tt.wantErr(t, err, fmt.Sprintf("failed to save metrics: %v", tt.want))
			}

			got, err := db.FindAll(ctx)
			if err != nil {
				tt.wantErr(t, err, fmt.Sprintf("failed to get metric: %v", tt.want))
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
