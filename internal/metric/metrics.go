package metric

import (
	"fmt"
	"strconv"
)

type Type int

const (
	Gauge Type = iota
	Counter
)

var types = [...]string{
	"gauge",
	"counter",
}

func (t Type) String() string {
	return types[t]
}

type Metric struct {
	name  string
	mType Type
	value interface{}
}

func (m *Metric) GetName() string {
	return m.name
}

func (m *Metric) setName(n string) {
	m.name = n
}

func (m *Metric) GetType() Type {
	return m.mType
}

func (m *Metric) GetGaugeValue() float64 {
	return m.value.(float64)
}

func (m *Metric) setGaugeValue(v float64) {
	m.mType = Gauge
	m.value = v
}

func (m *Metric) GetCounterValue() int64 {
	return m.value.(int64)
}

func (m *Metric) setCounterValue(v int64) {
	m.mType = Counter
	m.value = v
}

func (m *Metric) GetStringValue() string {
	switch m.mType {
	case Gauge:
		return strconv.FormatFloat(m.value.(float64), 'f', -1, 64)
	case Counter:
		return strconv.FormatInt(m.value.(int64), 10)
	default:
		return ""
	}
}

func NewGaugeMetric(name string, value float64) Metric {
	var m Metric
	m.setName(name)
	m.setGaugeValue(value)
	return m
}

func NewCounterMetric(name string, value int64) Metric {
	var m Metric
	m.setName(name)
	m.setCounterValue(value)
	return m
}

type NewMetricError struct {
	Error      error
	TypeError  bool
	ValueError bool
}

func NewMetric(name string, typeName string, value string) (m Metric, appError *NewMetricError) {
	switch typeName {
	case Gauge.String():
		{
			v, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return m, &NewMetricError{err, false, true}
			}
			m = NewGaugeMetric(name, v)
		}
	case Counter.String():
		{
			v, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return m, &NewMetricError{err, false, true}
			}
			m = NewCounterMetric(name, v)
		}
	default:
		return m, &NewMetricError{
			fmt.Errorf("cannot parse: [%s] as metric type", typeName),
			true,
			false,
		}
	}
	return m, &NewMetricError{}
}

func (m *Metric) String() string {
	return fmt.Sprintf("/%v/%v/%v", m.GetType().String(), m.GetName(), m.GetStringValue())
}

type Updatable func() []Metric

func GetUpdatable(sources ...Updatable) Updatable {
	return func() []Metric {
		var updates [][]Metric
		var totalLen int
		for _, s := range sources {
			newSlice := s()
			updates = append(updates, newSlice)
			totalLen += len(newSlice)
		}

		result := make([]Metric, totalLen)
		var i int
		for _, s := range updates {
			i += copy(result[i:], s)
		}
		return result
	}
}
