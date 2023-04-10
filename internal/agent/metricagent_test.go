package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/stretchr/testify/assert"
)

func TestMetricClient_SendUpdateContinuously(t *testing.T) {
	type want struct {
		url  string
		body []byte
	}
	tests := []struct {
		name   string
		metric metric.UpdatableMetric
		want   want
	}{
		{
			name:   "successfully return gauge metric",
			metric: metric.NewUpdatableGauge("FirstGauge", func() float64 { return 31337.1 }),
			want: want{
				url: "/updates/",
				body: []byte(`[{
								"id": "FirstGauge",
								"type": "gauge",
								"value": 31337.1
							}]`),
			},
		},
		{
			name:   "successfully return counter metric",
			metric: metric.NewUpdatableCounter("FirstCounter", func() int64 { return 12345 }),
			want: want{
				url: "/updates/",
				body: []byte(`[{
								"id": "FirstCounter",
								"type": "counter",
								"delta": 12345
							}]`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			var wg sync.WaitGroup
			wg.Add(2)

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var expectBody bytes.Buffer
				_ = json.Compact(&expectBody, tt.want.body)

				actualBody, err := io.ReadAll(r.Body)
				if err != nil {
					panic(err)
				}
				assert.Equal(t, tt.want.url, r.URL.Path)
				assert.JSONEq(t, expectBody.String(), string(actualBody))
			}))
			defer srv.Close()

			cfg := &config.AgentConfig{
				AgentInParams: &config.AgentInParams{
					Address:        srv.URL,
					ReportInterval: 10 * time.Second,
					PollInterval:   2 * time.Second,
				},
			}

			upd := metric.ConcatSources([]metric.UpdatableMetric{tt.metric})
			metricClient := NewMetricAgent(ctx, &wg, cfg)
			metricClient.SendAllMetricsContinuously(upd)

			time.Sleep(20 * time.Millisecond)
			cancel()
		})
	}
}
