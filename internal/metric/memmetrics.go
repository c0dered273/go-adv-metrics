package metric

import (
	"math/rand"
	"runtime"
	"sync/atomic"
	"time"
)

var pollCounter int64

func NewMemStats() []Metric {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	atomic.AddInt64(&pollCounter, 1)

	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)

	return []Metric{
		NewGaugeMetric("Alloc", float64(m.Alloc)),
		NewGaugeMetric("BuckHashSys", float64(m.BuckHashSys)),
		NewGaugeMetric("Frees", float64(m.Frees)),
		NewGaugeMetric("GCCPUFraction", m.GCCPUFraction),
		NewGaugeMetric("GCSys", float64(m.GCSys)),
		NewGaugeMetric("HeapAlloc", float64(m.HeapAlloc)),
		NewGaugeMetric("HeapIdle", float64(m.HeapIdle)),
		NewGaugeMetric("HeapInuse", float64(m.HeapInuse)),
		NewGaugeMetric("HeapObjects", float64(m.HeapObjects)),
		NewGaugeMetric("HeapReleased", float64(m.HeapReleased)),
		NewGaugeMetric("HeapSys", float64(m.HeapSys)),
		NewGaugeMetric("LastGC", float64(m.LastGC)),
		NewGaugeMetric("Lookups", float64(m.Lookups)),
		NewGaugeMetric("MCacheInuse", float64(m.MCacheInuse)),
		NewGaugeMetric("MCacheSys", float64(m.MCacheSys)),
		NewGaugeMetric("MSpanInuse", float64(m.MSpanInuse)),
		NewGaugeMetric("MSpanSys", float64(m.MSpanSys)),
		NewGaugeMetric("Mallocs", float64(m.Mallocs)),
		NewGaugeMetric("NextGC", float64(m.NextGC)),
		NewGaugeMetric("NumForcedGC", float64(m.NumForcedGC)),
		NewGaugeMetric("NumGC", float64(m.NumGC)),
		NewGaugeMetric("OtherSys", float64(m.OtherSys)),
		NewGaugeMetric("PauseTotalINs", float64(m.PauseTotalNs)),
		NewGaugeMetric("StackInuse", float64(m.StackInuse)),
		NewGaugeMetric("StackSys", float64(m.StackSys)),
		NewGaugeMetric("Sys", float64(m.Sys)),
		NewGaugeMetric("TotalAlloc", float64(m.TotalAlloc)),
		NewGaugeMetric("RandomValue", r.Float64()*1000000),
		NewCounterMetric("PollCounter", pollCounter),
	}
}
