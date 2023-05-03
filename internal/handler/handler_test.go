package handler

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/metric"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
	"github.com/go-resty/resty/v2"
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
	_, trustedSubnet, _ := net.ParseCIDR("10.0.0.0/24")

	cfg := &config.ServerConfig{
		ServerInParams: &config.ServerInParams{
			Address: "localhost:8080",
		},
		Repo: storage.NewPersistenceRepo(storage.NewMemStorage()),
	}

	cfgWithHash := &config.ServerConfig{
		ServerInParams: &config.ServerInParams{
			Address: "localhost:8080",
			Key:     "some_hash_key",
		},
		Repo: storage.NewPersistenceRepo(storage.NewMemStorage()),
	}

	cfgWithTrustedSubnet := &config.ServerConfig{
		ServerInParams: &config.ServerInParams{
			Address:       "localhost:8080",
			TrustedSubnet: trustedSubnet,
		},
		Repo: storage.NewPersistenceRepo(storage.NewMemStorage()),
	}

	type want struct {
		code  int
		value string
	}
	tests := []struct {
		name    string
		srvCfg  *config.ServerConfig
		method  string
		url     string
		headers map[string]string
		body    []byte
		want    want
	}{
		{
			name:   "should response 200 when valid request #1",
			srvCfg: cfg,
			method: "POST",
			url:    "http://localhost:8080/update/gauge/Alloc/31337.1",
			want: want{
				code: 200,
			},
		},
		{
			name:   "should response 200 when valid request #2",
			srvCfg: cfg,
			method: "POST",
			url:    "http://localhost:8080/update/counter/PollCounter/123",
			want: want{
				code: 200,
			},
		},
		{
			name:   "should return value when valid request after POST",
			srvCfg: cfg,
			method: "GET",
			url:    "http://localhost:8080/value/gauge/Alloc",
			want: want{
				code:  200,
				value: "31337.1",
			},
		},
		{
			name:   "should return html with all metrics",
			srvCfg: cfg,
			method: "GET",
			url:    "http://localhost:8080/",
			want: want{
				code: 200,
			},
		},
		{
			name:   "should response 405 when invalid method",
			srvCfg: cfg,
			method: "GET",
			url:    "http://localhost:8080/update/gauge/Alloc/31337",
			want: want{
				code: 405,
			},
		},
		{
			name:   "should response 400 when not a number at metric value",
			srvCfg: cfg,
			method: "POST",
			url:    "http://localhost:8080/update/gauge/Alloc/invalid",
			want: want{
				code: 400,
			},
		},
		{
			name:   "should response 501 when unknown metric type",
			srvCfg: cfg,
			method: "POST",
			url:    "http://localhost:8080/update/unknown/Alloc/31337",
			want: want{
				code: 501,
			},
		},
		{
			name:   "should response 404 when invalid path #1",
			srvCfg: cfg,
			method: "POST",
			url:    "http://localhost:8080/update/gauge/Alloc",
			want: want{
				code: 404,
			},
		},
		{
			name:   "should response 404 when invalid path #2",
			srvCfg: cfg,
			method: "POST",
			url:    "http://localhost:8080/update/gauge",
			want: want{
				code: 404,
			},
		},
		{
			name:   "should response 404 when invalid path #3",
			srvCfg: cfg,
			method: "POST",
			url:    "http://localhost:8080/update",
			want: want{
				code: 404,
			},
		},
		{
			name:   "should response 404 when unknown value",
			srvCfg: cfg,
			method: "GET",
			url:    "http://localhost:8080/value/unknown/metric",
			want: want{
				code: 404,
			},
		},
		{
			name:   "should response 200 when valid gauge",
			srvCfg: cfg,
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
			srvCfg: cfg,
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
		{
			name:   "should response 400 when invalid json",
			srvCfg: cfg,
			method: "POST",
			url:    "http://localhost:8080/update/",
			body:   []byte("invalid_json"),
			want: want{
				code: 400,
			},
		},
		{
			name:   "should response 200 when valid hash key",
			srvCfg: cfgWithHash,
			method: "POST",
			url:    "http://localhost:8080/update/",
			body: JSONtoByte(`{
									"id": "Alloc",
									"value": 1552512,
									"hash": "8b95cb2c3c115495793dfae27498b432bc6242f22d62b51fe5148e6224a075d6",
									"type": "gauge"
									}`),
			want: want{
				code: 200,
			},
		},
		{
			name:   "should response 400 when invalid hash key",
			srvCfg: cfgWithHash,
			method: "POST",
			url:    "http://localhost:8080/update/",
			body: JSONtoByte(`{
									"id": "Alloc",
									"value": 1552512,
									"hash": "152246535903a8764b8b96305f9b4c7357da8c6d78d60cd81b4930781b5f3525",
									"type": "gauge"
									}`),
			want: want{
				code: 400,
			},
		},
		{
			name:   "should response 500 when invalid hash key format",
			srvCfg: cfgWithHash,
			method: "POST",
			url:    "http://localhost:8080/update/",
			body: JSONtoByte(`{
									"id": "Alloc",
									"value": 1552512,
									"hash": "invalid_hash_key",
									"type": "gauge"
									}`),
			want: want{
				code: 500,
			},
		},
		{
			name:   "should response 400 when invalid json array",
			srvCfg: cfg,
			method: "POST",
			url:    "http://localhost:8080/updates/",
			body:   []byte("invalid_json"),
			want: want{
				code: 400,
			},
		},
		{
			name:   "should response 200 when save array with valid hash key",
			srvCfg: cfgWithHash,
			method: "POST",
			url:    "http://localhost:8080/updates/",
			body: JSONtoByte(`[{
									"id": "Alloc",
									"value": 1552512,
									"hash": "8b95cb2c3c115495793dfae27498b432bc6242f22d62b51fe5148e6224a075d6",
									"type": "gauge"
									}]`),
			want: want{
				code: 200,
			},
		},
		{
			name:   "should response 400 when save array with invalid hash key",
			srvCfg: cfgWithHash,
			method: "POST",
			url:    "http://localhost:8080/updates/",
			body: JSONtoByte(`[{
									"id": "Alloc",
									"value": 1552512,
									"hash": "152246535903a8764b8b96305f9b4c7357da8c6d78d60cd81b4930781b5f3525",
									"type": "gauge"
									}]`),
			want: want{
				code: 400,
			},
		},
		{
			name:   "should response 500 when save array with invalid hash key format",
			srvCfg: cfgWithHash,
			method: "POST",
			url:    "http://localhost:8080/updates/",
			body: JSONtoByte(`[{
									"id": "Alloc",
									"value": 1552512,
									"hash": "invalid_hash_format",
									"type": "gauge"
									}]`),
			want: want{
				code: 500,
			},
		},
		{
			name:    "should response 200 when valid ping request with trusted ip",
			srvCfg:  cfgWithTrustedSubnet,
			method:  "GET",
			url:     "http://localhost:8080/ping",
			headers: map[string]string{"X-Real-IP": "10.0.0.11"},
			want: want{
				code: 200,
			},
		},
		{
			name:    "should response 403 when valid ping request with untrusted ip",
			srvCfg:  cfgWithTrustedSubnet,
			method:  "GET",
			url:     "http://localhost:8080/ping",
			headers: map[string]string{"X-Real-IP": "10.0.22.11"},
			want: want{
				code: 403,
			},
		},
		{
			name:    "should response 400 when valid ping request with invalid ip",
			srvCfg:  cfgWithTrustedSubnet,
			method:  "GET",
			url:     "http://localhost:8080/ping",
			headers: map[string]string{"X-Real-IP": "FAKE_IP"},
			want: want{
				code: 400,
			},
		},
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

			if len(tt.headers) != 0 {
				for k, v := range tt.headers {
					request.Header.Set(k, v)
				}
			}

			writer := httptest.NewRecorder()
			h := Service(tt.srvCfg)
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

func TestEncryptedBody(t *testing.T) {
	privateKey := `
-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`
	storeReq := []byte{40, 117, 98, 194, 200, 247, 183, 231, 37, 68, 197, 164, 209, 129, 148, 197, 144, 62, 190, 241,
		118, 155, 251, 0, 139, 182, 126, 236, 140, 136, 74, 131, 7, 141, 228, 238, 186, 229, 195, 61, 138, 70, 243, 82,
		99, 93, 145, 91, 14, 227, 144, 248, 124, 182, 51, 37, 65, 214, 43, 195, 115, 83, 254, 41, 100, 42, 36, 184,
		221, 91, 191, 102, 20, 114, 101, 157, 208, 181, 106, 238, 109, 155, 207, 74, 116, 98, 9, 126, 153, 135, 34,
		124, 23, 62, 14, 161, 95, 187, 94, 127, 113, 186, 188, 9, 85, 156, 241, 54, 243, 106, 170, 178, 22, 236, 75,
		222, 79, 119, 224, 190, 11, 236, 151, 138, 156, 19, 186, 242, 142, 4, 72, 253, 134, 195, 20, 227, 88, 8, 42,
		53, 253, 151, 75, 74, 18, 163, 158, 174, 110, 194, 9, 161, 54, 119, 231, 188, 54, 16, 14, 32, 65, 164, 239,
		178, 42, 2, 124, 183, 40, 159, 28, 191, 250, 31, 51, 120, 248, 206, 65, 190, 40, 213, 145, 66, 39, 162, 93,
		231, 180, 46, 227, 66, 66, 122, 97, 120, 204, 107, 137, 251, 159, 13, 254, 53, 4, 103, 213, 127, 133, 2, 222,
		24, 241, 219, 194, 18, 235, 27, 220, 65, 1, 20, 55, 176, 156, 50, 45, 197, 234, 85, 99, 23, 67, 144, 180, 27,
		239, 37, 221, 118, 120, 81, 36, 98, 178, 173, 107, 123, 252, 226, 71, 109, 151, 192, 156, 253, 5, 43, 129, 162,
		242, 1, 142, 28, 236, 74, 189, 128, 201, 103, 175, 69, 33, 211, 69, 78, 160, 205, 141, 239, 147, 15, 52, 109,
		13, 139, 212, 74, 76, 199, 100, 16, 94, 33, 36, 206, 223, 135, 137, 254, 46, 24, 25, 110, 172, 8, 52, 159, 2,
		162, 238, 1, 41, 151, 71, 78, 169, 121, 95, 187, 115, 13, 253, 45, 51, 19, 148, 190, 23, 123, 184, 227, 141,
		200, 139, 148, 114, 51, 24, 193, 106, 208, 212, 102, 6, 181, 239, 19, 96, 163, 112, 163, 80, 1, 146, 4, 22,
		171, 87, 83, 140, 186, 65, 2, 16, 244, 38, 23, 93, 96, 229, 85, 230, 56, 217, 123, 13, 117, 142, 246, 90, 105,
		17, 176, 239, 228, 201, 29, 126, 29, 223, 39, 23, 24, 193, 68, 9, 123, 66, 188, 134, 28, 160, 16, 137, 238,
		145, 165, 59, 146, 243, 17, 125, 27, 46, 136, 209, 113, 206, 92, 72, 223, 185, 91, 188, 104, 243, 7, 135, 108,
		89, 2, 231, 130, 153, 20, 38, 77, 45, 1, 190, 240, 175, 60, 86, 93, 59, 254, 117, 226, 44, 109, 55, 176, 241,
		144, 229, 231, 13, 227, 178, 104, 176, 126, 25, 130, 202, 178, 14, 29, 212, 92, 98, 88, 188, 255, 255, 63, 83,
		5, 4, 23, 201, 255, 138, 104, 248, 95, 146, 218, 111, 197, 123, 255, 187, 59, 161, 66, 30, 153, 232, 234, 28,
		107, 188, 139, 156, 114, 105, 223, 63, 114}

	loadReq := []byte{93, 31, 182, 211, 34, 139, 111, 199, 103, 99, 177, 122, 170, 247, 216, 48, 89, 49, 121, 30,
		129, 147, 208, 135, 123, 46, 220, 199, 187, 7, 36, 93, 135, 85, 137, 206, 132, 234, 238, 235, 19, 158, 79, 142,
		227, 223, 175, 115, 3, 39, 203, 225, 252, 230, 168, 72, 120, 115, 75, 215, 29, 214, 21, 60, 233, 13, 201, 233,
		62, 235, 250, 136, 251, 80, 27, 49, 86, 56, 20, 66, 102, 185, 214, 85, 195, 18, 192, 178, 200, 203, 228, 217,
		247, 229, 36, 113, 201, 12, 213, 203, 195, 245, 56, 63, 174, 171, 172, 158, 116, 227, 182, 155, 49, 133, 181,
		245, 203, 178, 21, 101, 65, 158, 9, 133, 147, 85, 228, 9, 62, 150, 61, 83, 239, 220, 214, 176, 112, 42, 213,
		85, 233, 34, 50, 78, 44, 40, 213, 110, 64, 37, 211, 113, 237, 140, 211, 226, 113, 71, 77, 237, 98, 118, 174,
		191, 109, 231, 153, 94, 193, 159, 81, 168, 36, 133, 74, 238, 2, 24, 81, 252, 107, 226, 167, 134, 195, 221, 82,
		174, 35, 238, 79, 223, 108, 179, 208, 97, 72, 194, 31, 42, 8, 169, 70, 169, 243, 104, 15, 67, 224, 53, 236,
		168, 154, 38, 3, 172, 23, 99, 165, 180, 59, 178, 115, 188, 159, 4, 209, 184, 11, 38, 83, 143, 246, 66, 111,
		216, 178, 97, 121, 155, 142, 1, 159, 81, 97, 30, 37, 174, 254, 141, 101, 209, 255, 161, 104, 101, 19, 239, 25,
		101, 77, 154, 166, 90, 13, 54, 12, 78, 193, 230, 93, 162, 192, 66, 70, 32, 189, 141, 22, 43, 242, 178, 11, 211,
		83, 236, 84, 137, 106, 255, 60, 75, 59, 76, 2, 195, 60, 147, 38, 143, 136, 188, 68, 65, 237, 120, 141, 143, 236,
		246, 251, 254, 254, 199, 244, 192, 122, 206, 44, 61, 235, 228, 27, 28, 190, 213, 153, 176, 71, 243, 111, 123,
		68, 250, 173, 7, 20, 228, 142, 191, 139, 167, 53, 60, 137, 44, 202, 26, 97, 49, 65, 252, 254, 241, 49, 109, 60,
		25, 116, 35, 26, 130, 212, 79, 45, 28, 33, 11, 173, 224, 102, 203, 104, 147, 149, 189, 38, 80, 165, 235, 154,
		55, 233, 147, 245, 122, 165, 84, 56, 71, 30, 20, 254, 134, 108, 192, 47, 77, 208, 224, 238, 105, 192, 109, 158,
		212, 52, 1, 122, 87, 23, 112, 214, 63, 32, 120, 230, 253, 66, 225, 60, 124, 72, 63, 33, 155, 8, 240, 77, 108,
		201, 154, 11, 201, 169, 197, 86, 211, 9, 242, 221, 158, 139, 225, 28, 191, 84, 56, 70, 10, 31, 4, 165, 249, 129,
		72, 57, 78, 86, 155, 96, 64, 122, 119, 228, 90, 248, 189, 37, 143, 236, 224, 38, 103, 125, 193, 138, 155, 143,
		183, 222, 99, 22, 35, 200, 83, 104, 140, 234, 181, 145, 113, 136, 143, 161, 157, 228, 254, 180, 220, 134, 177,
		116, 169, 96, 224, 44, 219, 28, 204, 21, 60}

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
			name:      "should return 200 and return gauge value when encrypted body",
			method:    "POST",
			storeURL:  "http://localhost:8080/updates/",
			storeBody: storeReq,
			loadURL:   "http://localhost:8080/value/",
			loadBody:  loadReq,
			want: want{
				code:   200,
				metric: metric.NewGaugeMetric("Alloc", 335120),
			},
		},
	}

	block, _ := pem.Decode([]byte(privateKey))
	prv, _ := x509.ParsePKCS1PrivateKey(block.Bytes)

	cfg := &config.ServerConfig{
		ServerInParams: &config.ServerInParams{
			Address: "localhost:8080",
		},
		PrivateKey: prv,
		Repo:       storage.NewPersistenceRepo(storage.NewMemStorage()),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storeReq := httptest.NewRequest(tt.method, tt.storeURL, bytes.NewReader(tt.storeBody))
			loadReq := httptest.NewRequest(tt.method, tt.loadURL, bytes.NewReader(tt.loadBody))
			writer := httptest.NewRecorder()
			h := Service(cfg)
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

	cfg := &config.ServerConfig{
		ServerInParams: &config.ServerInParams{
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
			h := Service(cfg)
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

	cfg := &config.ServerConfig{
		ServerInParams: &config.ServerInParams{
			Address: "localhost:8080",
		},
		Repo: storage.NewPersistenceRepo(storage.NewMemStorage()),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storeReq := httptest.NewRequest(tt.method, tt.storeURL, bytes.NewReader(tt.storeBody))
			loadReq := httptest.NewRequest(tt.method, tt.loadURL, bytes.NewReader(tt.loadBody))
			writer := httptest.NewRecorder()
			h := Service(cfg)
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

func ExampleStoreMetricFromJSONHandler() {
	gaugeMetric := metric.NewGaugeMetric("TestGauge1", 123.456)

	client := resty.New()
	_, _ = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(gaugeMetric).
		Post("localhost:8080/update/")

	// Output:
	//
}

func ExampleLoadMetricByJSONHandler() {
	reqMetric := metric.NewGaugeMetric("MetricName", 0)

	client := resty.New()
	resp, _ := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(reqMetric).
		Post("localhost:8080/value/")

	var responseMetric metric.Metric
	_ = json.Unmarshal(resp.Body(), &responseMetric)
	fmt.Println(responseMetric)

	// Output:
	// { gauge <nil> <nil> }
}
