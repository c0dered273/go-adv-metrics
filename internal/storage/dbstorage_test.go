package storage

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/docker/go-connections/nat"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbName   = "goadv"
	dbUser   = "postgres"
	dbPasswd = "postgres"
)

type postgresContainer struct {
	testcontainers.Container
}

type postgresContainerOption func(req *testcontainers.ContainerRequest)

func WithWaitStrategy(strategies ...wait.Strategy) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.WaitingFor = wait.ForAll(strategies...).WithDeadline(1 * time.Minute)
	}
}

func WithPort(port string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.ExposedPorts = append(req.ExposedPorts, port)
	}
}

func WithInitialDatabase(user string, password string, dbName string) func(req *testcontainers.ContainerRequest) {
	return func(req *testcontainers.ContainerRequest) {
		req.Env["POSTGRES_USER"] = user
		req.Env["POSTGRES_PASSWORD"] = password
		req.Env["POSTGRES_DB"] = dbName
	}
}

func startContainer(ctx context.Context, opts ...postgresContainerOption) (*postgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image: "postgres:14-alpine",
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "goadv",
		},
		ExposedPorts: []string{},
		Cmd:          []string{"postgres", "-c", "fsync=off"},
	}

	for _, opt := range opts {
		opt(&req)
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &postgresContainer{Container: container}, nil
}

func GetTestDB(ctx context.Context, t *testing.T) *DBStorage {
	port, err := nat.NewPort("tcp", "5432")
	require.NoError(t, err)

	postgresC, err := startContainer(ctx,
		WithPort(port.Port()),
		WithInitialDatabase(dbUser, dbPasswd, dbName),
		WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}

	endpoint, err := postgresC.Endpoint(ctx, "")
	assert.NoError(t, err)

	dsn := fmt.Sprintf("postgres://postgres:postgres@%v/goadv", endpoint)
	db := NewDBStorage(
		dsn,
		false,
		log.Logger,
		ctx,
	)

	t.Cleanup(func() {
		err = db.Close()
		if err != nil {
			t.Fatal(err)
		}

		if err := postgresC.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	return db
}

func TestDBStorage_Save(t *testing.T) {
	ctx := context.Background()
	db := GetTestDB(ctx, t)

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
			err := db.Save(ctx, tt.store)
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
	db := GetTestDB(ctx, t)

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
			err := db.Save(ctx, tt.store)
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
	db := GetTestDB(ctx, t)

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
			err := db.SaveAll(ctx, tt.want)
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

func BenchmarkDBStorage_SaveAll(b *testing.B) {
	ctx := context.Background()
	port, err := nat.NewPort("tcp", "5432")
	if err != nil {
		log.Fatal().Err(err).Send()
		return
	}

	postgresC, err := startContainer(ctx,
		WithPort(port.Port()),
		WithInitialDatabase(dbUser, dbPasswd, dbName),
		WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		log.Fatal().Err(err).Send()
		return
	}

	endpoint, err := postgresC.Endpoint(ctx, "")
	if err != nil {
		log.Fatal().Err(err).Send()
		return
	}

	dsn := fmt.Sprintf("postgres://postgres:postgres@%v/goadv", endpoint)
	db := NewDBStorage(
		dsn,
		false,
		log.Logger,
		ctx,
	)

	defer func() {
		err = db.Close()
		if err != nil {
			log.Fatal().Err(err).Send()
		}

		if err := postgresC.Terminate(ctx); err != nil {
			if err != nil {
				log.Fatal().Err(err).Send()
			}
		}
	}()

	n := 20
	metrics := make([]metric.Metric, n)
	for i := 0; i < n; i++ {
		metrics[i] = metric.NewGaugeMetric("TestMetric"+strconv.Itoa(i), float64(100+i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := db.SaveAll(ctx, metrics)
		if err != nil {
			panic(err)
		}
	}
}
