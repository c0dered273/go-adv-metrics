package client

import (
	"context"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestMetricClient_SendContinuously(t *testing.T) {
	type want struct {
		url string
	}
	tests := []struct {
		name  string
		stats metric.Stats
		want  want
	}{
		{
			name: "successfully return gauge metric",
			stats: metric.Stats{
				Gauges: []metric.Gauge{
					{
						Name:  "Alloc",
						Value: 31773.001,
					},
				},
			},
			want: want{
				url: "/update/gauge/Alloc/31773.001",
			},
		},
		{
			name: "successfully return counter metric",
			stats: metric.Stats{
				Counters: []metric.Counter{
					{
						Name:  "PollCounter",
						Value: 12345,
					},
				},
			},
			want: want{
				url: "/update/counter/PollCounter/12345",
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

			container := metric.NewContainer([]metric.Source{&tt.stats})
			metricClient := NewMetricClient(ctx, &wg, Settings{ServerAddr: srv.URL})
			metricClient.SendUpdateContinuously(&container)

			time.Sleep(20 * time.Millisecond)
			cancel()
		})
	}
}
