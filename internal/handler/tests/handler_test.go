package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/c0dered273/go-adv-metrics/internal/handler"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

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
			url:    "http://localhost:8080/update",
			body: func() []byte {
				var req bytes.Buffer
				r := []byte(`{
						"id": "Alloc",
						"type": "gauge",
						"value": 31337.9
					}`)
				err := json.Compact(&req, r)
				if err != nil {
					return nil
				}
				return req.Bytes()
			}(),
			want: want{
				code: 200,
			},
		},
		{
			name:   "should response 200 when valid counter",
			method: "POST",
			url:    "http://localhost:8080/update",
			body: func() []byte {
				var req bytes.Buffer
				r := []byte(`{
						"id": "Poll",
						"type": "counter",
						"value": 313379
					}`)
				err := json.Compact(&req, r)
				if err != nil {
					return nil
				}
				return req.Bytes()
			}(),
			want: want{
				code: 200,
			},
		},
	}

	cfg := handler.ServerConfig{
		Repo: storage.GetMemStorageInstance(),
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

	cfg := handler.ServerConfig{
		Repo: storage.GetMemStorageInstance(),
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
