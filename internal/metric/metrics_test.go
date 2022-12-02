package metric

import (
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
		name    string
		mName   string
		mType   string
		mValue  string
		wantM   Metric
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "successfully parse gauge metric",
			mName:  "GaugeOne",
			mType:  "gauge",
			mValue: "31337",
			wantM: Metric{
				name:  "GaugeOne",
				mType: Gauge,
				value: float64(31337),
			},
			wantErr: assert.NoError,
		},
		{
			name:   "successfully parse counter metric",
			mName:  "CounterOne",
			mType:  "counter",
			mValue: "31337",
			wantM: Metric{
				name:  "CounterOne",
				mType: Counter,
				value: int64(31337),
			},
			wantErr: assert.NoError,
		},
		{
			name:    "error parse unknown metric type",
			mName:   "ErrorOne",
			mType:   "unknown",
			mValue:  "31337",
			wantErr: assert.Error,
		},
		{
			name:    "error parse metric value",
			mName:   "ErrorTwo",
			mType:   "counter",
			mValue:  "fake_value",
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotM, appError := NewMetric(tt.mName, tt.mType, tt.mValue)
			if !tt.wantErr(t, appError.Error) {
				return
			}
			assert.Equal(t, tt.wantM, gotM)
		})
	}
}
