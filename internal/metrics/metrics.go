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

var metricTypes = [...]string{
	"gauge",
	"counter",
}

func ParseMetricType(s string) (m MetricType, err error) {
	index := -1
	for i, mt := range metricTypes {
		if mt == s {
			index = i
		}
	}
	if index == -1 {
		return m, fmt.Errorf(`cannot parse:[%s] as MetricType`, s)
	}
	return MetricType(index), nil
}

func (m MetricType) String() string {
	return metricTypes[m]
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
	Name  string
	MType MetricType
	Value string
}

func (m *Metric) String() string {
	return fmt.Sprintf("/%v/%v/%v", m.MType.String(), m.Name, m.Value)
}

func NewGaugeMt(g Gauge) Metric {
	return Metric{
		Name:  g.name,
		Value: strconv.FormatFloat(g.value, 'f', 3, 64),
		MType: GaugeType,
	}
}

func NewCounterMt(c Counter) Metric {
	return Metric{
		Name:  c.name,
		Value: strconv.FormatInt(c.value, 10),
		MType: CounterType,
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
