package metrics

import (
	"fmt"
	"strconv"
)

type MetricType int

const (
	GaugeType MetricType = iota
	CounterType
)

func (m MetricType) String() string {
	switch m {
	case GaugeType:
		return "gauge"
	case CounterType:
		return "counter"
	}
	return "unknown"
}

type Gauge struct {
	name  string
	value float64
}

type Counter struct {
	name  string
	value int64
}

type Metric struct {
	name   string
	stType MetricType
	value  string
}

func (m *Metric) String() string {
	return fmt.Sprintf("/%v/%v/%v", m.stType.String(), m.name, m.value)
}

func NewGaugeMt(g Gauge) Metric {
	return Metric{
		name:   g.name,
		value:  strconv.FormatFloat(g.value, 'f', 3, 64),
		stType: GaugeType,
	}
}

func NewCounterMt(c Counter) Metric {
	return Metric{
		name:   c.name,
		value:  strconv.FormatInt(c.value, 10),
		stType: CounterType,
	}
}

func gaugeSliceToMt(g []Gauge) []Metric {
	result := make([]Metric, len(g))
	for i := 0; i < len(g); i++ {
		result[i] = NewGaugeMt(g[i])
	}
	return result
}

func counterSliceToMt(c []Counter) []Metric {
	result := make([]Metric, len(c))
	for i := 0; i < len(c); i++ {
		result[i] = NewCounterMt(c[i])
	}
	return result
}

func concatSlices(slices ...[]Metric) []Metric {
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}

	result := make([]Metric, totalLen)
	var i int
	for _, s := range slices {
		i += copy(result[i:], s)
	}
	return result
}

func GetAllMetrics() []Metric {
	memStats := gaugeSliceToMt(GetMemStats())
	cusCounter := counterSliceToMt(GetCustomCounter())
	cusGauge := gaugeSliceToMt(GetCustomGauge())

	return concatSlices(memStats, cusCounter, cusGauge)
}
