package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
)

type FileStorage struct {
	ctx         context.Context
	mx          *sync.Mutex
	file        *os.File
	memCache    *MemStorage
	isSyncStore bool
}

func (f *FileStorage) Save(newMetric metric.Metric) error {
	if sErr := f.memCache.Save(newMetric); sErr != nil {
		return sErr
	}
	if f.isSyncStore {
		if wErr := f.WriteMetrics(); wErr != nil {
			return wErr
		}
	}
	return nil
}

func (f *FileStorage) SaveAll(metrics []metric.Metric) error {
	if sErr := f.memCache.SaveAll(metrics); sErr != nil {
		return sErr
	}
	if f.isSyncStore {
		if err := f.WriteMetrics(); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileStorage) FindByID(keyMetric metric.Metric) (metric.Metric, error) {
	return f.memCache.FindByID(keyMetric)
}

func (f *FileStorage) FindAll() ([]metric.Metric, error) {
	return f.memCache.FindAll()
}

func (f *FileStorage) ReadMetrics() error {
	var rsl []metric.Metric

	f.mx.Lock()
	scanner := bufio.NewScanner(f.file)
	for scanner.Scan() {
		var m metric.Metric
		if err := json.Unmarshal(scanner.Bytes(), &m); err != nil {
			return err
		}
		rsl = append(rsl, m)
	}
	f.mx.Unlock()

	if err := f.memCache.SaveAll(rsl); err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) WriteMetrics() error {
	cached, faErr := f.memCache.FindAll()
	if faErr != nil {
		return faErr
	}
	data := make([]byte, 0)

	for _, m := range cached {
		byteMetric, mErr := json.Marshal(m)
		if mErr != nil {
			return mErr
		}
		data = append(data, byteMetric...)
		data = append(data, '\n')
	}

	f.mx.Lock()
	if _, sErr := f.file.Seek(0, 0); sErr != nil {
		return sErr
	}
	writer := bufio.NewWriter(f.file)
	if _, wErr := writer.Write(data); wErr != nil {
		return wErr
	}

	if fErr := writer.Flush(); fErr != nil {
		return fErr
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
				log.Error.Println("can`t write metrics at async store ", err)
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

func NewFileStorage(fileName string, storeInterval time.Duration, isAppend bool, ctx context.Context) *FileStorage {
	var flags int
	if isAppend {
		flags = os.O_RDWR | os.O_CREATE
	} else {
		flags = os.O_RDWR | os.O_CREATE | os.O_TRUNC
	}

	file, oErr := os.OpenFile(fileName, flags, 0777)
	if oErr != nil {
		log.Error.Panic("can`t open file: ", fileName)
	}

	fs := &FileStorage{
		ctx:         ctx,
		mx:          new(sync.Mutex),
		file:        file,
		memCache:    NewMemStorage(),
		isSyncStore: storeInterval == 0,
	}

	if isAppend {
		rdErr := fs.ReadMetrics()
		if rdErr != nil {
			log.Error.Fatalln("can't read metrics from disk ", rdErr)
		}
	}

	if !fs.isSyncStore {
		fs.asyncStore(storeInterval)
	}

	go func() {
		<-fs.ctx.Done()
		err := fs.Close()
		if err != nil {
			log.Error.Println("can`t close file storage", err)
		}
	}()

	return fs
}
