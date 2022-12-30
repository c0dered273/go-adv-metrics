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

func (ds *DBStorage) Save(ctx context.Context, m metric.Metric) error {
	statement := `INSERT INTO metrics VALUES ($1, $2, $3, $4, $5)
						ON CONFLICT (metric_name, metric_type)
							DO UPDATE SET metric_delta = $3, metric_value = $4, hash = $5`
	_, err := ds.db.ExecContext(ctx, statement, m.ID, m.MType.String(), m.Delta, m.Val, m.Hash)
	if err != nil {
		return err
	}
	return nil
}

func (ds *DBStorage) SaveAll(ctx context.Context, metrics []metric.Metric) error {
	for _, m := range metrics {
		err := ds.Save(ctx, m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ds *DBStorage) FindByID(ctx context.Context, keyMetric metric.Metric) (metric.Metric, error) {
	m := metric.Metric{}
	statement := "SELECT * FROM metrics WHERE metric_name = $1 AND metric_type = $2"
	row := ds.db.QueryRowContext(ctx, statement, keyMetric.ID, keyMetric.MType.String())
	err := row.Scan(&m.ID, &m.MType, &m.Delta, &m.Val, &m.Hash)
	if err != nil {
		return metric.Metric{}, err
	}
	return m, nil
}

func (ds *DBStorage) FindAll(ctx context.Context) ([]metric.Metric, error) {
	result := make([]metric.Metric, 0, 0)
	statement := "SELECT * FROM metrics"
	rows, err := ds.db.QueryContext(ctx, statement)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		m := metric.Metric{}
		err := rows.Scan(&m.ID, &m.MType, &m.Delta, &m.Val, &m.Hash)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}

	return result, nil
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

func (ds *DBStorage) initDB(isRestore bool) error {
	ctx, cancel := context.WithTimeout(ds.ctx, ds.QueryTimeout)
	defer cancel()
	go func() {
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			log.Error.Fatalln("dbStorage: db schema init timeout", ctx.Err())
		}
	}()

	if !isRestore {
		_, dErr := ds.db.ExecContext(ctx, "DROP TABLE IF EXISTS metrics")
		if dErr != nil {
			log.Error.Println("dbStorage: can't drop table 'metrics'")
			return dErr
		}
	}

	statement := `CREATE TABLE IF NOT EXISTS metrics 
				(
					metric_name varchar(32),
					metric_type varchar(32),
					metric_delta bigint,
					metric_value double precision,
					hash varchar(64),
    				CONSTRAINT metric_pk PRIMARY KEY(metric_name, metric_type)
				)`
	_, cErr := ds.db.ExecContext(ctx, statement)
	if cErr != nil {
		log.Error.Println("dbStorage: can't create table 'metrics'")
		return cErr
	}

	return nil
}

func NewDBStorage(databaseDsn string, isRestore bool, ctx context.Context) *DBStorage {
	connConfig, cErr := pgx.ParseConnectionString(databaseDsn)
	if cErr != nil {
		log.Error.Fatal("dbStorage: can`t parse connection config", cErr)
	}
	pool, pErr := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     connConfig,
		MaxConnections: 10,
	})
	if pErr != nil {
		log.Error.Fatalln("dbStorage: can`t create connection pool", pErr)
	}

	db := stdlib.OpenDBFromPool(pool)

	ds := &DBStorage{
		db:           db,
		ctx:          ctx,
		QueryTimeout: DefaultTimeout,
	}

	err := ds.initDB(isRestore)
	if err != nil {
		log.Error.Fatalln("dbStorage: can't init db")
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
