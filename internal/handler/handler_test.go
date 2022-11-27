package handler

import (
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestMetricsHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct {
		name   string
		method string
		url    string
		want   want
	}{
		{
			name:   "should response 200 when valid request #1",
			method: "POST",
			url:    "http://localhost:8080/update/gauge/Alloc/31337",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.url, nil)
			writer := httptest.NewRecorder()
			h := MetricsHandler{}
			h.ServeHTTP(writer, request)
			res := writer.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
		})
	}
}
