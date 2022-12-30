package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
)

const (
	DefaultTimeout = 15 * time.Second
)

type DBStorage struct {
	db           *sql.DB
	ctx          context.Context
	QueryTimeout time.Duration
}

func (ds *DBStorage) Save(ctx context.Context, metric metric.Metric) error {
	//TODO implement me
	panic("implement me")
}

func (ds *DBStorage) SaveAll(ctx context.Context, metrics []metric.Metric) error {
	//TODO implement me
	panic("implement me")
}

func (ds *DBStorage) FindByID(ctx context.Context, metric metric.Metric) (metric.Metric, error) {
	//TODO implement me
	panic("implement me")
}

func (ds *DBStorage) FindAll(ctx context.Context) ([]metric.Metric, error) {
	//TODO implement me
	panic("implement me")
}

func (ds *DBStorage) Ping() error {
	ctx, cancel := context.WithTimeout(ds.ctx, ds.QueryTimeout)
	defer cancel()
	if err := ds.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func (ds *DBStorage) Close() error {
	return ds.db.Close()
}

func NewDBStorage(databaseDsn string, ctx context.Context) *DBStorage {
	connConfig, cErr := pgx.ParseConnectionString(databaseDsn)
	if cErr != nil {
		log.Error.Fatal("dbStorage: can`t parse connection config", cErr)
	}
	pool, pErr := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     connConfig,
		MaxConnections: 10,
	})
	if pErr != nil {
		log.Error.Fatal("dbStorage: can`t create connection pool", pErr)
	}

	db := stdlib.OpenDBFromPool(pool)

	ds := &DBStorage{
		db:           db,
		ctx:          ctx,
		QueryTimeout: DefaultTimeout,
	}

	go func() {
		<-ds.ctx.Done()
		err := ds.Close()
		if err != nil {
			log.Error.Println("dbStorage: can't close db", err)
		}
	}()

	return ds
}
