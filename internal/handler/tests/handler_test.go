package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/handler"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/c0dered273/go-adv-metrics/internal/service"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

func JSONtoByte(s string) []byte {
	var req bytes.Buffer
	if err := json.Compact(&req, []byte(s)); err != nil {
		return nil
	}
	return req.Bytes()
}

func TestService(t *testing.T) {
	type want struct {
		code  int
		value string
	}
	tests := []struct {
		name   string
		method string
		url    string
		body   []byte
		want   want
	}{
		{
			name:   "should response 200 when valid request #1",
			method: "POST",
			url:    "http://localhost:8080/update/gauge/Alloc/31337.1",
			want: want{
				code: 200,
			},
		},
		{
			name:   "should response 200 when valid request #2",
			method: "POST",
			url:    "http://localhost:8080/update/counter/PollCounter/123",
			want: want{
				code: 200,
			},
		},
		{
			name:   "should return value when valid request after POST",
			method: "GET",
			url:    "http://localhost:8080/value/gauge/Alloc",
			want: want{
				code:  200,
				value: "31337.1",
			},
		},
		{
			name:   "should return html with all metrics",
			method: "GET",
			url:    "http://localhost:8080/",
			want: want{
				code: 200,
			},
		},
		{
			name:   "should response 405 when invalid method",
			method: "GET",
			url:    "http://localhost:8080/update/gauge/Alloc/31337",
			want: want{
				code: 405,
			},
		},
		{
			name:   "should response 400 when not a number at metric value",
			method: "POST",
			url:    "http://localhost:8080/update/gauge/Alloc/invalid",
			want: want{
				code: 400,
			},
		},
		{
			name:   "should response 501 when unknown metric type",
			method: "POST",
			url:    "http://localhost:8080/update/unknown/Alloc/31337",
			want: want{
				code: 501,
			},
		},
		{
			name:   "should response 404 when invalid path #1",
			method: "POST",
			url:    "http://localhost:8080/update/gauge/Alloc",
			want: want{
				code: 404,
			},
		},
		{
			name:   "should response 404 when invalid path #2",
			method: "POST",
			url:    "http://localhost:8080/update/gauge",
			want: want{
				code: 404,
			},
		},
		{
			name:   "should response 404 when invalid path #3",
			method: "POST",
			url:    "http://localhost:8080/update",
			want: want{
				code: 404,
			},
		},
		{
			name:   "should response 404 when unknown value",
			method: "GET",
			url:    "http://localhost:8080/value/unknown/metric",
			want: want{
				code: 404,
			},
		},
		{
			name:   "should response 200 when valid gauge",
			method: "POST",
			url:    "http://localhost:8080/update/",
			body: JSONtoByte(`{
									"id": "Alloc",
									"type": "gauge",
									"value": 31337.9
								}`),
			want: want{
				code: 200,
			},
		},
		{
			name:   "should response 200 when valid counter",
			method: "POST",
			url:    "http://localhost:8080/update/",
			body: JSONtoByte(`{
									"id": "Poll",
									"type": "counter",
									"delta": 313379
								}`),
			want: want{
				code: 200,
			},
		},
	}

	cfg := &service.ServerConfig{
		ServerCmd: config.ServerCmd{
			Address: "localhost:8080",
		},
		Repo: storage.NewPersistenceRepo(storage.NewMemStorage()),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var request *http.Request
			var bodyReader *bytes.Reader

			if len(tt.body) != 0 {
				bodyReader = bytes.NewReader(tt.body)
				request = httptest.NewRequest(tt.method, tt.url, bodyReader)
			} else {
				request = httptest.NewRequest(tt.method, tt.url, nil)
			}

			writer := httptest.NewRecorder()
			h := handler.Service(cfg)
			h.ServeHTTP(writer, request)
			res := writer.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.code, res.StatusCode)
			if tt.want.value != "" {
				actual, _ := io.ReadAll(res.Body)
				assert.Equal(t, tt.want.value, string(actual))
			}
		})
	}
}

func Test_metricStore(t *testing.T) {
	type want struct {
		code  int
		value string
	}
	tests := []struct {
		name   string
		method string
		url1   string
		url2   string
		url3   string
		want   want
	}{
		{
			name:   "should return 200 and last update value",
			method: "POST",
			url1:   "http://localhost:8080/update/gauge/Alloc/11",
			url2:   "http://localhost:8080/update/gauge/Alloc/22",
			url3:   "http://localhost:8080/value/gauge/Alloc",
			want: want{
				code:  200,
				value: "22",
			},
		},
		{
			name:   "should return 200 and sum of updates values",
			method: "POST",
			url1:   "http://localhost:8080/update/counter/poll/11",
			url2:   "http://localhost:8080/update/counter/poll/22",
			url3:   "http://localhost:8080/value/counter/poll",
			want: want{
				code:  200,
				value: "33",
			},
		},
	}

	cfg := &service.ServerConfig{
		ServerCmd: config.ServerCmd{
			Address: "localhost:8080",
		},
		Repo: storage.NewPersistenceRepo(storage.NewMemStorage()),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request1 := httptest.NewRequest(tt.method, tt.url1, nil)
			request2 := httptest.NewRequest(tt.method, tt.url2, nil)
			request3 := httptest.NewRequest("GET", tt.url3, nil)
			writer := httptest.NewRecorder()
			h := handler.Service(cfg)
			h.ServeHTTP(writer, request1)
			h.ServeHTTP(writer, request2)
			h.ServeHTTP(writer, request3)
			res := writer.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.code, res.StatusCode)
			if tt.want.value != "" {
				actual, _ := io.ReadAll(res.Body)
				assert.Equal(t, tt.want.value, string(actual))
			}
		})
	}
}

func Test_metricJSONLoad(t *testing.T) {
	type want struct {
		code   int
		metric metric.Metric
	}
	tests := []struct {
		name      string
		method    string
		storeURL  string
		storeBody []byte
		loadURL   string
		loadBody  []byte
		want      want
	}{
		{
			name:     "should return 200 and update gauge value",
			method:   "POST",
			storeURL: "http://localhost:8080/update/",
			storeBody: JSONtoByte(`{
										"id":"Alloc",
										"type": "gauge",
										"value": 555.99
									}`),
			loadURL: "http://localhost:8080/value/",
			loadBody: JSONtoByte(`{
										"id":"Alloc",
										"type":"gauge"
									}`),
			want: want{
				code:   200,
				metric: metric.NewGaugeMetric("Alloc", 555.99),
			},
		},
		{
			name:     "should return 200 and update counter value",
			method:   "POST",
			storeURL: "http://localhost:8080/update/",
			storeBody: JSONtoByte(`{
										"id":"PollCounter",
										"type": "counter",
										"delta": 123456
									}`),
			loadURL: "http://localhost:8080/value/",
			loadBody: JSONtoByte(`{
										"id":"PollCounter",
										"type":"counter"
									}`),
			want: want{
				code:   200,
				metric: metric.NewCounterMetric("PollCounter", 123456),
			},
		},
		{
			name:     "should return 400 when invalid metric",
			method:   "POST",
			storeURL: "http://localhost:8080/update/",
			storeBody: JSONtoByte(`{
										"id":"Allocr",
										"type": "gauge",
										"delta": 123456
									}`),
			loadURL: "http://localhost:8080/value/",
			want: want{
				code: 400,
			},
		},
	}

	cfg := &service.ServerConfig{
		ServerCmd: config.ServerCmd{
			Address: "localhost:8080",
		},
		Repo: storage.NewPersistenceRepo(storage.NewMemStorage()),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storeReq := httptest.NewRequest(tt.method, tt.storeURL, bytes.NewReader(tt.storeBody))
			loadReq := httptest.NewRequest(tt.method, tt.loadURL, bytes.NewReader(tt.loadBody))
			writer := httptest.NewRecorder()
			h := handler.Service(cfg)
			h.ServeHTTP(writer, storeReq)
			h.ServeHTTP(writer, loadReq)
			res := writer.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.code, res.StatusCode)
			if res.StatusCode != http.StatusOK {
				return
			}

			actual, _ := io.ReadAll(res.Body)
			var actualMetric metric.Metric
			err := json.Unmarshal(actual, &actualMetric)
			if err != nil {
				panic(err)
			}

			assert.Equal(t, true, tt.want.metric.Equal(&actualMetric))
		})
	}
}
