package storage

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/rs/zerolog"
)

// FileStorage позволяет использовать в качестве хранения файл на жестком диске.
// Метрики хранятся в текстовом формате в виде json объектов.
// Данные сначала кэшируются в памяти и с заданным интервалом сбрасываются на диск.
// Хранилище возможно сделать синхронным задав storeInterval = 0
type FileStorage struct {
	ctx         context.Context
	logger      zerolog.Logger
	mx          *sync.Mutex
	file        FileReaderWriter
	encoder     *json.Encoder
	decoder     *json.Decoder
	memCache    Repository
	isSyncStore bool
}

type FileReaderWriter interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Seek(offset int64, whence int) (ret int64, err error)
	Stat() (os.FileInfo, error)
	Close() error
}

func (f *FileStorage) Save(ctx context.Context, newMetric metric.Metric) error {
	if sErr := f.memCache.Save(ctx, newMetric); sErr != nil {
		return sErr
	}
	if f.isSyncStore {
		if wErr := f.WriteMetrics(); wErr != nil {
			return wErr
		}
	}
	return nil
}

func (f *FileStorage) SaveAll(ctx context.Context, metrics []metric.Metric) error {
	if sErr := f.memCache.SaveAll(ctx, metrics); sErr != nil {
		return sErr
	}
	if f.isSyncStore {
		if err := f.WriteMetrics(); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileStorage) FindByID(ctx context.Context, keyMetric metric.Metric) (metric.Metric, error) {
	return f.memCache.FindByID(ctx, keyMetric)
}

func (f *FileStorage) FindAll(ctx context.Context) ([]metric.Metric, error) {
	return f.memCache.FindAll(ctx)
}

func (f *FileStorage) Ping() error {
	return nil
}

func (f *FileStorage) ReadMetrics() error {
	data := metric.Metrics{}

	fileInfo, infoErr := f.file.Stat()
	if infoErr != nil {
		return infoErr
	}
	if fileInfo.Size() == 0 {
		return nil
	}

	f.mx.Lock()
	decErr := f.decoder.Decode(&data)
	if decErr != nil {
		return decErr
	}
	f.mx.Unlock()

	if err := f.memCache.SaveAll(f.ctx, data.Metrics); err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) WriteMetrics() error {
	cached, faErr := f.memCache.FindAll(f.ctx)
	if faErr != nil {
		return faErr
	}
	if len(cached) == 0 {
		return nil
	}
	data := metric.Metrics{Metrics: cached}

	f.mx.Lock()
	if _, sErr := f.file.Seek(0, 0); sErr != nil {
		return sErr
	}
	encErr := f.encoder.Encode(&data)
	if encErr != nil {
		return encErr
	}
	f.mx.Unlock()

	return nil
}

func (f *FileStorage) Close() error {
	if err := f.WriteMetrics(); err != nil {
		return err
	}
	if cErr := f.file.Close(); cErr != nil {
		return cErr
	}
	return nil
}

func (f *FileStorage) asyncStore(storeInterval time.Duration) {
	go func() {
		ticker := time.NewTicker(storeInterval)
		defer ticker.Stop()
		for {
			if err := f.WriteMetrics(); err != nil {
				f.logger.Error().Err(err).Msg("fileStore: failed to write metrics at async store")
				return
			}
			select {
			case <-ticker.C:
				continue
			case <-f.ctx.Done():
				return
			}
		}
	}()
}

// NewFileStorage возвращает настроенное файловое хранилище
func NewFileStorage(
	ctx context.Context, fileName string, storeInterval time.Duration, isRestore bool, logger zerolog.Logger,
) *FileStorage {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0777)
	if err != nil {
		logger.Error().Err(err).Msgf("fileStore: failed to open file: %v", fileName)
		panic(err)
	}

	return CreateFileStorage(ctx, file, storeInterval, isRestore, logger)
}

func CreateFileStorage(
	ctx context.Context, file FileReaderWriter, storeInterval time.Duration, isRestore bool, logger zerolog.Logger,
) *FileStorage {
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	decoder := json.NewDecoder(file)

	isSyncStore := storeInterval == 0

	fs := &FileStorage{
		ctx:         ctx,
		logger:      logger,
		mx:          new(sync.Mutex),
		file:        file,
		encoder:     encoder,
		decoder:     decoder,
		memCache:    NewMemStorage(),
		isSyncStore: isSyncStore,
	}

	if isRestore {
		err := fs.ReadMetrics()
		if err != nil {
			logger.Fatal().Err(err).Msg("fileStorage: failed to read metrics from disk")
		}
	}

	if !fs.isSyncStore {
		fs.asyncStore(storeInterval)
	}

	go func() {
		<-fs.ctx.Done()
		err := fs.Close()
		if err != nil {
			logger.Error().Err(err).Msg("fileStorage: can`t close file storage")
		}
	}()

	return fs
}
