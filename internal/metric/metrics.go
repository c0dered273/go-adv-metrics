package metric

import (
	"encoding/json"
	"fmt"
	"math"
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

func NewType(s string) (Type, error) {
	for i, t := range types {
		if s == t {
			return (Type)(i), nil
		}
	}
	return 0, fmt.Errorf("unknown metric type from [%v]", s)
}

type Metric struct {
	ID    string   `json:"id"`
	MType Type     `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (m *Metric) GetName() string {
	return m.ID
}

func (m *Metric) setName(n string) {
	m.ID = n
}

func (m *Metric) GetType() Type {
	return m.MType
}

func (m *Metric) GetGaugeValue() float64 {
	return *m.Value
}

func (m *Metric) setGaugeValue(v float64) {
	m.MType = Gauge
	m.Value = &v
}

func (m *Metric) GetCounterValue() int64 {
	return *m.Delta
}

func (m *Metric) setCounterValue(v int64) {
	m.MType = Counter
	m.Delta = &v
}

func (m *Metric) GetStringValue() string {
	switch m.MType {
	case Gauge:
		return strconv.FormatFloat(*m.Value, 'f', -1, 64)
	case Counter:
		return strconv.FormatInt(*m.Delta, 10)
	default:
		return ""
	}
}

func (m *Metric) String() string {
	return fmt.Sprintf("/%v/%v/%v", m.GetType().String(), m.GetName(), m.GetStringValue())
}

func (m *Metric) Equal(other *Metric) bool {
	switch m.MType {
	case Gauge:
		fmt.Printf("*** %v ***", math.Abs(*m.Value-*other.Value))
		return math.Abs(*m.Value-*other.Value) <= math.SmallestNonzeroFloat64
	case Counter:
		return *m.Delta == *other.Delta
	default:
		return false
	}
}

func (m Metric) MarshalJSON() ([]byte, error) {
	type MetricAlias Metric
	aliasValue := &struct {
		MetricAlias
		MType string `json:"type"`
	}{
		MetricAlias: MetricAlias(m),
		MType:       m.MType.String(),
	}
	result, err := json.Marshal(aliasValue)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (m *Metric) UnmarshalJSON(bytes []byte) error {
	type MetricAlias Metric
	var strTypeName string
	aliasValue := &struct {
		*MetricAlias
		MType json.RawMessage `json:"type"`
	}{
		MetricAlias: (*MetricAlias)(m),
	}

	if aliasErr := json.Unmarshal(bytes, aliasValue); aliasErr != nil {
		return aliasErr
	}

	if typeErr := json.Unmarshal(aliasValue.MType, &strTypeName); typeErr != nil {
		return typeErr
	}

	mType, tErr := NewType(strTypeName)
	if tErr != nil {
		return tErr
	}

	m.MType = mType

	return nil
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

func NewMetric(name string, typeName string, value string) (m Metric, err NewMetricError) {
	t, typErr := NewType(typeName)
	if typErr != nil {
		return m, NewMetricError{
			Error:     typErr,
			TypeError: true,
		}
	}

	switch t {
	case Gauge:
		{
			v, fltErr := strconv.ParseFloat(value, 64)
			if fltErr != nil {
				return m, NewMetricError{
					Error:      fltErr,
					ValueError: true,
				}
			}
			m = NewGaugeMetric(name, v)
		}
	case Counter:
		{
			v, intErr := strconv.ParseInt(value, 10, 64)
			if intErr != nil {
				return m, NewMetricError{
					Error:      intErr,
					ValueError: true,
				}
			}
			m = NewCounterMetric(name, v)
		}
	}
	return m, NewMetricError{}
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
