package metric

import (
	"fmt"
	"strconv"
)

type Type int

const (
	GaugeType Type = iota
	CounterType
)

var metricTypes = [...]string{
	"gauge",
	"counter",
}

func ParseMetricType(s string) (m Type, err error) {
	index := -1
	for i, mt := range metricTypes {
		if mt == s {
			index = i
		}
	}
	if index == -1 {
		return m, fmt.Errorf("cannot parse: [%s] as MetricType", s)
	}
	return Type(index), nil
}

func (m Type) String() string {
	return metricTypes[m]
}

type Gauge struct {
	Name  string
	Value float64
}

type Counter struct {
	Name  string
	Value int64
}

type Metric struct {
	Name  string
	Type  Type
	Value string
}

func (m *Metric) String() string {
	return fmt.Sprintf("/%v/%v/%v", m.Type.String(), m.Name, m.Value)
}

func NewGaugeMt(g Gauge) Metric {
	return Metric{
		Name:  g.Name,
		Value: strconv.FormatFloat(g.Value, 'f', 3, 64),
		Type:  GaugeType,
	}
}

func NewCounterMt(c Counter) Metric {
	return Metric{
		Name:  c.Name,
		Value: strconv.FormatInt(c.Value, 10),
		Type:  CounterType,
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

type Stats struct {
	Gauges   []Gauge
	Counters []Counter
}

func (m *Stats) toMetrics() []Metric {
	gaugesMt := gaugeSliceToMt(m.Gauges)
	counterMt := counterSliceToMt(m.Counters)
	return concatSlices(gaugesMt, counterMt)
}

type Source interface {
	toMetrics() []Metric
}

type Container struct {
	sources []Source
	metrics []Metric
}

func NewContainer(sources []Source) Container {
	return Container{
		sources: sources,
	}
}

func (m *Container) UpdateAndGet() []Metric {
	for _, s := range m.sources {
		m.metrics = append(m.metrics, s.toMetrics()...)
	}
	return m.metrics
}
