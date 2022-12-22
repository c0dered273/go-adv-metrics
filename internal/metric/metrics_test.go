package metric

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUpdatable(t *testing.T) {
	tests := []struct {
		name      string
		sourceOne []Metric
		sourceTwo []Metric
		want      []string
	}{
		{
			name: "successfully return updatable slice of metrics",
			sourceOne: []Metric{
				NewGaugeMetric("FirstGauge", 31337.1),
				NewCounterMetric("FirstCounter", 12345),
			},
			sourceTwo: []Metric{
				NewGaugeMetric("SecondGauge", float64(42)),
				NewCounterMetric("SecondCounter", 321),
			},
			want: []string{
				"/gauge/FirstGauge/31337.1",
				"/counter/FirstCounter/12345",
				"/gauge/SecondGauge/42",
				"/counter/SecondCounter/321",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatable := GetUpdatable(func() []Metric { return tt.sourceOne }, func() []Metric { return tt.sourceTwo })
			var actual []string
			for _, m := range updatable() {
				actual = append(actual, m.String())
			}

			assert.Equal(t, tt.want, actual)
		})
	}
}

func TestNewMetric(t *testing.T) {
	tests := []struct {
		name   string
		mName  string
		mType  string
		mValue string
		wantM  Metric
	}{
		{
			name:   "successfully parse gauge metric",
			mName:  "GaugeOne",
			mType:  "gauge",
			mValue: "31337",
			wantM: Metric{
				ID:    "GaugeOne",
				MType: Gauge,
				Value: func(v float64) *float64 { return &v }(31337),
			},
		},
		{
			name:   "successfully parse counter metric",
			mName:  "CounterOne",
			mType:  "counter",
			mValue: "31337",
			wantM: Metric{
				ID:    "CounterOne",
				MType: Counter,
				Delta: func(v int64) *int64 { return &v }(31337),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotM, _ := NewMetric(tt.mName, tt.mType, tt.mValue)
			assert.Equal(t, true, tt.wantM.Equal(&gotM))
		})
	}
}

func TestNewMetricErrors(t *testing.T) {
	tests := []struct {
		name    string
		mName   string
		mType   string
		mValue  string
		want    NewMetricError
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "error parse unknown metric type",
			mName:  "ErrorOne",
			mType:  "unknown",
			mValue: "31337",
			want: NewMetricError{
				TypeError:  true,
				ValueError: false,
			},
			wantErr: assert.Error,
		},
		{
			name:   "error parse metric value",
			mName:  "ErrorTwo",
			mType:  "counter",
			mValue: "fake_value",
			want: NewMetricError{
				TypeError:  false,
				ValueError: true,
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, appError := NewMetric(tt.mName, tt.mType, tt.mValue)
			if !tt.wantErr(t, appError.Error) {
				panic(appError.Error)
			}
			tt.want.Error = appError.Error
			assert.Equal(t, tt.want, appError)
		})
	}
}

func TestMetric_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		metric  Metric
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "should successfully return json from gauge",
			metric: NewGaugeMetric("Alloc", 31337.999),
			want: []byte(`{
						"id": "Alloc",
						"type": "gauge",
						"value": 31337.999
						}`),
			wantErr: assert.NoError,
		},
		{
			name:   "should successfully return json from counter",
			metric: NewCounterMetric("Poll", 31337999),
			want: []byte(`{
						"id": "Poll",
						"type": "counter",
						"delta": 31337999
						}`),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.metric)
			if !tt.wantErr(t, err) {
				panic(err)
			}
			expect := new(bytes.Buffer)
			err = json.Compact(expect, tt.want)
			if err != nil {
				panic(err)
			}
			assert.JSONEq(t, expect.String(), string(got))
		})
	}
}

func TestMetric_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    []byte
		want    Metric
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "should successfully return gauge from json",
			json: []byte(`{
							"id": "Alloc",
							"type": "gauge",
							"value": 31337.991
						}`),
			want:    NewGaugeMetric("Alloc", 31337.991),
			wantErr: assert.NoError,
		},
		{
			name: "should successfully return counter from json",
			json: []byte(`{
							"id": "Poll",
							"type": "counter",
							"delta": 123456
						}`),
			want:    NewCounterMetric("Poll", 123456),
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := new(bytes.Buffer)
			resp := Metric{}
			err := json.Compact(req, tt.json)
			if err != nil {
				panic(err)
			}

			err = json.Unmarshal(req.Bytes(), &resp)
			if !tt.wantErr(t, err) {
				panic(err)
			}
			assert.Equal(t, tt.want, resp)
		})
	}
}

func TestIsValid(t *testing.T) {
	type args struct {
		m Metric
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should return true when valid gauge struct",
			args: args{m: NewGaugeMetric("Alloc", 123.456)},
			want: true,
		},
		{
			name: "should return true when valid counter struct",
			args: args{m: NewCounterMetric("Poll", 123)},
			want: true,
		},
		{
			name: "should return false when invalid struct",
			args: args{m: Metric{ID: "Invalid", MType: Gauge, Delta: func(i int64) *int64 { return &i }(123)}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsValid(tt.args.m), "IsValid(%v)", tt.args.m)
		})
	}
}
