package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/c0dered273/go-adv-metrics/internal/storage/mocks"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

type TestFileInfo struct {
}

func (F TestFileInfo) Name() string {
	return "TestFile"
}

func (F TestFileInfo) Size() int64 {
	return 1
}

func (F TestFileInfo) Mode() fs.FileMode {
	return 0
}

func (F TestFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (F TestFileInfo) IsDir() bool {
	return false
}

func (F TestFileInfo) Sys() any {
	return nil
}

func TestFileStorage_FindAll(t *testing.T) {
	type fields struct {
		rw      *mocks.MockFileReaderWriter
		storage *mocks.MockRepository
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		prepare func(f *fields)
		want    []metric.Metric
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should successful find all metric",
			args: args{
				ctx: context.Background(),
			},
			prepare: func(f *fields) {
				m := metric.Metrics{
					Metrics: []metric.Metric{
						metric.NewGaugeMetric("TestGauge", 123.456),
						metric.NewCounterMetric("TestCounter", 112233),
					},
				}

				f.storage.EXPECT().FindAll(gomock.Any()).Return(m.Metrics, nil)
				f.rw.EXPECT().Stat().Return(TestFileInfo{}, nil)
				f.rw.EXPECT().Read(gomock.Any()).Return(1, nil)
			},
			want: []metric.Metric{
				metric.NewGaugeMetric("TestGauge", 123.456),
				metric.NewCounterMetric("TestCounter", 112233),
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			f := fields{
				rw:      mocks.NewMockFileReaderWriter(ctrl),
				storage: mocks.NewMockRepository(ctrl),
			}

			if tt.prepare != nil {
				tt.prepare(&f)
			}

			s := FileStorage{
				ctx:         context.Background(),
				logger:      log.Logger,
				mx:          new(sync.Mutex),
				file:        f.rw,
				encoder:     json.NewEncoder(f.rw),
				decoder:     json.NewDecoder(f.rw),
				memCache:    f.storage,
				isSyncStore: true,
			}

			got, err := s.FindAll(tt.args.ctx)
			_ = s.ReadMetrics()
			tt.wantErr(t, err, fmt.Sprint("FindAll()"))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFileStorage_FindByID(t *testing.T) {
	type fields struct {
		rw      *mocks.MockFileReaderWriter
		storage *mocks.MockRepository
	}
	type args struct {
		ctx context.Context
		id  metric.Metric
	}
	tests := []struct {
		name    string
		args    args
		prepare func(f *fields)
		want    metric.Metric
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should successful find metric by id",
			args: args{
				ctx: context.Background(),
				id:  metric.NewCounterMetric("TestCounter", 112233),
			},
			prepare: func(f *fields) {
				m := metric.NewCounterMetric("TestCounter", 112233)

				f.storage.EXPECT().FindByID(gomock.Any(), m).Return(m, nil)
			},
			want:    metric.NewCounterMetric("TestCounter", 112233),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			f := fields{
				rw:      mocks.NewMockFileReaderWriter(ctrl),
				storage: mocks.NewMockRepository(ctrl),
			}

			if tt.prepare != nil {
				tt.prepare(&f)
			}

			s := FileStorage{
				ctx:         context.Background(),
				logger:      log.Logger,
				mx:          new(sync.Mutex),
				file:        f.rw,
				encoder:     json.NewEncoder(f.rw),
				decoder:     json.NewDecoder(f.rw),
				memCache:    f.storage,
				isSyncStore: true,
			}

			got, err := s.FindByID(tt.args.ctx, tt.args.id)
			tt.wantErr(t, err, fmt.Sprintf("FindById(%v)", tt.args.id))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFileStorage_Ping(t *testing.T) {
	type fields struct {
		ctx         context.Context
		logger      zerolog.Logger
		mx          *sync.Mutex
		file        *os.File
		encoder     *json.Encoder
		decoder     *json.Decoder
		memCache    Repository
		isSyncStore bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "should always return no error",
			fields:  fields{},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileStorage{
				ctx:         tt.fields.ctx,
				logger:      tt.fields.logger,
				mx:          tt.fields.mx,
				file:        tt.fields.file,
				encoder:     tt.fields.encoder,
				decoder:     tt.fields.decoder,
				memCache:    tt.fields.memCache,
				isSyncStore: tt.fields.isSyncStore,
			}
			tt.wantErr(t, f.Ping(), fmt.Sprintf("Ping()"))
		})
	}
}

func TestFileStorage_Save(t *testing.T) {
	type fields struct {
		rw *mocks.MockFileReaderWriter
	}
	type args struct {
		ctx       context.Context
		newMetric metric.Metric
	}
	tests := []struct {
		name    string
		args    args
		prepare func(f *fields)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should successful save metric",
			args: args{
				ctx:       context.Background(),
				newMetric: metric.NewGaugeMetric("TestGauge", 123.456),
			},
			prepare: func(f *fields) {
				m := metric.Metrics{
					Metrics: []metric.Metric{
						metric.NewGaugeMetric("TestGauge", 123.456),
					},
				}
				byteMetrics, _ := json.MarshalIndent(&m, "", "  ")
				byteMetrics = append(byteMetrics, 10)

				f.rw.EXPECT().Seek(int64(0), 0).Return(int64(0), nil)
				f.rw.EXPECT().Write(byteMetrics).Return(1, nil)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			f := fields{
				rw: mocks.NewMockFileReaderWriter(ctrl),
			}

			if tt.prepare != nil {
				tt.prepare(&f)
			}

			s := CreateFileStorage(
				context.Background(),
				f.rw,
				0*time.Second,
				false,
				log.Logger,
			)

			tt.wantErr(t, s.Save(tt.args.ctx, tt.args.newMetric), fmt.Sprintf("Save(%v, %v)", tt.args.ctx, tt.args.newMetric))
		})
	}
}

func TestFileStorage_SaveAll(t *testing.T) {
	type fields struct {
		rw *mocks.MockFileReaderWriter
	}
	type args struct {
		ctx        context.Context
		newMetrics []metric.Metric
	}
	tests := []struct {
		name    string
		args    args
		prepare func(f *fields)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should successful save all metric",
			args: args{
				ctx: context.Background(),
				newMetrics: []metric.Metric{
					metric.NewGaugeMetric("TestGauge", 123.456),
					metric.NewCounterMetric("TestCounter", 112233),
				},
			},
			prepare: func(f *fields) {
				m := metric.Metrics{
					Metrics: []metric.Metric{
						metric.NewGaugeMetric("TestGauge", 123.456),
						metric.NewCounterMetric("TestCounter", 112233),
					},
				}
				byteMetrics, _ := json.MarshalIndent(&m, "", "  ")
				byteMetrics = append(byteMetrics, 10)

				f.rw.EXPECT().Seek(int64(0), 0).Return(int64(0), nil)
				f.rw.EXPECT().Write(byteMetrics).Return(1, nil)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			f := fields{
				rw: mocks.NewMockFileReaderWriter(ctrl),
			}

			if tt.prepare != nil {
				tt.prepare(&f)
			}

			s := CreateFileStorage(
				context.Background(),
				f.rw,
				0*time.Second,
				false,
				log.Logger,
			)

			tt.wantErr(t, s.SaveAll(tt.args.ctx, tt.args.newMetrics), fmt.Sprintf("SaveAll(%v, %v)", tt.args.ctx, tt.args.newMetrics))
		})
	}
}

func TestFileStorage_ReadMetrics(t *testing.T) {
	type fields struct {
		rw      *mocks.MockFileReaderWriter
		storage *mocks.MockRepository
	}
	tests := []struct {
		name    string
		prepare func(f *fields)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should successful call Read method",
			prepare: func(f *fields) {
				f.rw.EXPECT().Stat().Return(TestFileInfo{}, nil)
				f.rw.EXPECT().Read(gomock.Any()).Return(1, nil)
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			f := fields{
				rw:      mocks.NewMockFileReaderWriter(ctrl),
				storage: mocks.NewMockRepository(ctrl),
			}

			if tt.prepare != nil {
				tt.prepare(&f)
			}

			s := FileStorage{
				ctx:         context.Background(),
				logger:      log.Logger,
				mx:          new(sync.Mutex),
				file:        f.rw,
				encoder:     json.NewEncoder(f.rw),
				decoder:     json.NewDecoder(f.rw),
				memCache:    f.storage,
				isSyncStore: true,
			}

			tt.wantErr(t, s.ReadMetrics(), fmt.Sprint("ReadMetrics()"))
		})
	}
}

func TestFileStorage_WriteMetrics(t *testing.T) {
	type fields struct {
		rw      *mocks.MockFileReaderWriter
		storage *mocks.MockRepository
	}
	tests := []struct {
		name    string
		prepare func(f *fields)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should successful write to file",
			prepare: func(f *fields) {
				m := metric.Metrics{
					Metrics: []metric.Metric{
						metric.NewGaugeMetric("TestGauge", 123.456),
						metric.NewCounterMetric("TestCounter", 112233),
					},
				}
				byteMetrics, _ := json.Marshal(&m)
				byteMetrics = append(byteMetrics, 10)

				f.storage.EXPECT().FindAll(gomock.Any()).Return(m.Metrics, nil)
				f.rw.EXPECT().Seek(int64(0), 0).Return(int64(0), nil)
				f.rw.EXPECT().Write(byteMetrics).Return(1, nil)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			f := fields{
				rw:      mocks.NewMockFileReaderWriter(ctrl),
				storage: mocks.NewMockRepository(ctrl),
			}

			if tt.prepare != nil {
				tt.prepare(&f)
			}

			s := FileStorage{
				ctx:         context.Background(),
				logger:      log.Logger,
				mx:          new(sync.Mutex),
				file:        f.rw,
				encoder:     json.NewEncoder(f.rw),
				decoder:     json.NewDecoder(f.rw),
				memCache:    f.storage,
				isSyncStore: true,
			}

			tt.wantErr(t, s.WriteMetrics(), fmt.Sprint("WriteMetrics()"))
		})
	}
}
