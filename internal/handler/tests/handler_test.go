package tests

import (
	"github.com/c0dered273/go-adv-metrics/internal/handler"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http/httptest"
	"testing"
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
				code:  200,
				value: "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <title>All metrics</title>\n</head>\n<body>\n    <h2>All metrics</h2>\n    <ol>\n        \n            <li>/gauge/Alloc/31337.1</li>\n        \n            <li>/counter/PollCounter/123</li>\n        \n    </ol>\n</body>\n</html>",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.url, nil)
			writer := httptest.NewRecorder()
			h := handler.Service()
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
