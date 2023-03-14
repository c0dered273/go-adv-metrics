package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemStats(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{
			name: "should return slice of updatable metrics",
			want: 29,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, len(NewMemStats()), "NewMemStats()")
		})
	}
}
