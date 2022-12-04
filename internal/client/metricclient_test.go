package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/stretchr/testify/assert"
)

func TestMetricClient_SendUpdateContinuously(t *testing.T) {
	type want struct {
		url string
	}
	tests := []struct {
		name   string
		metric metric.Metric
		want   want
	}{
		{
			name:   "successfully return gauge metric",
			metric: metric.NewGaugeMetric("FirstGauge", 31337.1),
			want: want{
				url: "/update/gauge/FirstGauge/31337.1",
			},
		},
		{
			name:   "successfully return counter metric",
			metric: metric.NewCounterMetric("FirstCounter", 12345),
			want: want{
				url: "/update/counter/FirstCounter/12345",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			var wg sync.WaitGroup
			wg.Add(2)

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.want.url, r.URL.Path)
			}))
			defer srv.Close()

			upd := metric.GetUpdatable(func() []metric.Metric { return []metric.Metric{tt.metric} })
			metricClient := NewMetricClient(ctx, &wg, Settings{ServerAddr: srv.URL})
			metricClient.SendUpdateContinuously(upd)

			time.Sleep(20 * time.Millisecond)
			cancel()
		})
	}
}
