package metric

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetricContainer_UpdateAndGet(t *testing.T) {
	tests := []struct {
		name  string
		stats Stats
		want  []string
	}{
		{
			name: "successfully return slice of endpoints",
			stats: Stats{
				Gauges: []Gauge{
					{
						Name:  "Alloc",
						Value: 31773.001,
					},
				},
				Counters: []Counter{
					{
						Name:  "PollCounter",
						Value: 12345,
					},
				},
			},
			want: []string{
				"/gauge/Alloc/31773.001",
				"/counter/PollCounter/12345",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := NewContainer([]Source{&tt.stats})
			metrics := container.UpdateAndGet()
			var expected []string
			for _, m := range metrics {
				expected = append(expected, m.String())
			}
			assert.Equal(t, tt.want, expected)
		})
	}
}

func TestParseMetricType(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		wantM   Type
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "successfully parse Gauge metric type",
			args:    args{"gauge"},
			wantM:   GaugeType,
			wantErr: assert.NoError,
		},
		{
			name:    "successfully parse Counter metric type",
			args:    args{"counter"},
			wantM:   CounterType,
			wantErr: assert.NoError,
		},
		{
			name:    "error when parse unknown type",
			args:    args{"unknown"},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotM, err := ParseMetricType(tt.args.s)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.wantM, gotM)
		})
	}
}
