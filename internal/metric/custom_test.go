package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPsUtilStats(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{
			name: "should return slice of updatable metrics",
			want: 14,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, len(NewPsUtilStats()), "NewPsUtilStats()")
		})
	}
}
