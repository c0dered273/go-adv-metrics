// Package metric содержит методы для работы с метриками
package metric

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"
)

// Type Тип метрики
type Type int

const (
	Gauge Type = iota
	Counter
)

var types = [...]string{
	"gauge",
	"counter",
}

// Value Метод необходим для корректного преобразования типа метрики в значение, хранимое в БД
func (t Type) Value() (driver.Value, error) {
	return t.String(), nil
}

// Scan Метод преобразует тип метрики из БД в корректный вид
func (t *Type) Scan(src any) error {
	sType, err := NewType(src.(string))
	if err != nil {
		return err
	}
	*t = sType
	return nil
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

type Metrics struct {
	Metrics []Metric `json:"metrics"`
}

func (ms *Metrics) SetHash(hashKey string) {
	if hashKey != "" {
		for i := range ms.Metrics {
			ms.Metrics[i].SetHash(hashKey)
		}
	}
}

func (ms *Metrics) CheckHash(hashKey string) (bool, error) {
	for _, m := range ms.Metrics {
		ok, err := m.CheckHash(hashKey)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

func (ms *Metrics) IsValid() (bool, Metric) {
	for _, m := range ms.Metrics {
		if !IsValid(m) {
			return false, m
		}
	}
	return true, Metric{}
}

type Metric struct {
	ID    string   `json:"id"`
	MType Type     `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Val   *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
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
	return *m.Val
}

func (m *Metric) setGaugeValue(v float64) {
	m.MType = Gauge
	m.Val = &v
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
		return strconv.FormatFloat(*m.Val, 'f', -1, 64)
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
		fmt.Printf("*** %v ***", math.Abs(*m.Val-*other.Val))
		return math.Abs(*m.Val-*other.Val) <= math.SmallestNonzeroFloat64
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

func (m *Metric) getHashSrc() []byte {
	switch m.MType {
	case Gauge:
		return []byte(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Val))
	case Counter:
		return []byte(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta))
	}
	return []byte{}
}

func (m *Metric) generateHash(hashKey string) []byte {
	h := hmac.New(sha256.New, []byte(hashKey))
	h.Write(m.getHashSrc())
	return h.Sum(nil)
}

func (m *Metric) SetHash(hashKey string) {
	if hashKey != "" {
		m.Hash = hex.EncodeToString(m.generateHash(hashKey))
	}
}

func (m *Metric) CheckHash(hashKey string) (bool, error) {
	if hashKey != "" {
		hashActual, hexErr := hex.DecodeString(m.Hash)
		if hexErr != nil {
			return false, hexErr
		}
		hashExpected := m.generateHash(hashKey)
		return hmac.Equal(hashActual, hashExpected), nil
	}
	return true, nil
}

// NewGaugeMetric основная функция для создания метрики типа gauge
func NewGaugeMetric(ID string, value float64) Metric {
	var m Metric
	m.setName(ID)
	m.setGaugeValue(value)
	return m
}

// NewCounterMetric основная функция для создания метрики типа counter
func NewCounterMetric(ID string, value int64) Metric {
	var m Metric
	m.setName(ID)
	m.setCounterValue(value)
	return m
}

func IsValid(m Metric) bool {
	switch m.MType {
	case Gauge:
		return m.Val != nil
	case Counter:
		return m.Delta != nil
	default:
		return false
	}
}

type UpdatableMetric struct {
	Metric
	gaugeSource   func() float64
	counterSource func() int64
}

func (um *UpdatableMetric) Update() {
	switch um.MType {
	case Gauge:
		val := um.gaugeSource()
		um.Val = &val
	case Counter:
		val := um.counterSource()
		um.Delta = &val
	}
}

func NewUpdatableGauge(ID string, source func() float64) UpdatableMetric {
	return UpdatableMetric{
		NewGaugeMetric(ID, source()),
		source,
		nil,
	}
}

func NewUpdatableCounter(ID string, source func() int64) UpdatableMetric {
	return UpdatableMetric{
		NewCounterMetric(ID, source()),
		nil,
		source,
	}
}

type NewMetricError struct {
	Error      error
	TypeError  bool
	ValueError bool
}

func NewMetric(ID string, typeName string, value string, hashKey string) (m Metric, err NewMetricError) {
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
			m = NewGaugeMetric(ID, v)
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
			m = NewCounterMetric(ID, v)
		}
	}

	m.SetHash(hashKey)

	return m, NewMetricError{}
}

// ConcatSources функция эффективно склеивает массивы с обновляемыми метриками,
// позволяет удобно использовать несколько источников метрик
func ConcatSources(sources ...[]UpdatableMetric) []UpdatableMetric {
	var updates [][]UpdatableMetric
	var totalLen int
	for _, s := range sources {
		updates = append(updates, s)
		totalLen += len(s)
	}

	result := make([]UpdatableMetric, totalLen)
	var i int
	for _, s := range updates {
		i += copy(result[i:], s)
	}
	return result
}

type ConcurrentTime struct {
	time time.Time
	mu   *sync.RWMutex
}

func (t *ConcurrentTime) set(time time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.time = time
}

func (t *ConcurrentTime) get() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.time
}
