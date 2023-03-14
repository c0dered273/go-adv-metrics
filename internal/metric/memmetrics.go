package metric

import (
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Ограничение по частоте обновления метрик
const cacheTimeMem = 1 * time.Second

var (
	m           runtime.MemStats
	pollCounter int64

	lastUpdateMem = ConcurrentTime{
		time: time.Now(),
		mu:   new(sync.RWMutex),
	}
)

func updateMemStats() {
	if time.Since(lastUpdateMem.get()) > cacheTimeMem {
		runtime.ReadMemStats(&m)
		lastUpdateMem.set(time.Now())
	}
}

// NewMemStats возвращает обновляемые метрики.
// Значение метрик кэшируется, метрики обновляются не чаще одного раза в секунду.
// Отдает метрики приложения из пакета runtime
func NewMemStats() []UpdatableMetric {
	runtime.ReadMemStats(&m)

	seed := rand.NewSource(time.Now().UnixNano())

	return []UpdatableMetric{
		NewUpdatableGauge("Alloc", func() float64 {
			updateMemStats()
			return float64(m.Alloc)
		}),
		NewUpdatableGauge("BuckHashSys", func() float64 {
			updateMemStats()
			return float64(m.BuckHashSys)
		}),
		NewUpdatableGauge("Frees", func() float64 {
			updateMemStats()
			return float64(m.Frees)
		}),
		NewUpdatableGauge("GCCPUFraction", func() float64 {
			updateMemStats()
			return m.GCCPUFraction
		}),
		NewUpdatableGauge("GCSys", func() float64 {
			updateMemStats()
			return float64(m.GCSys)
		}),
		NewUpdatableGauge("HeapAlloc", func() float64 {
			updateMemStats()
			return float64(m.HeapAlloc)
		}),
		NewUpdatableGauge("HeapIdle", func() float64 {
			updateMemStats()
			return float64(m.HeapIdle)
		}),
		NewUpdatableGauge("HeapInuse", func() float64 {
			updateMemStats()
			return float64(m.HeapInuse)
		}),
		NewUpdatableGauge("HeapObjects", func() float64 {
			updateMemStats()
			return float64(m.HeapObjects)
		}),
		NewUpdatableGauge("HeapReleased", func() float64 {
			updateMemStats()
			return float64(m.HeapReleased)
		}),
		NewUpdatableGauge("HeapSys", func() float64 {
			updateMemStats()
			return float64(m.HeapSys)
		}),
		NewUpdatableGauge("LastGC", func() float64 {
			updateMemStats()
			return float64(m.LastGC)
		}),
		NewUpdatableGauge("Lookups", func() float64 {
			updateMemStats()
			return float64(m.Lookups)
		}),
		NewUpdatableGauge("MCacheInuse", func() float64 {
			updateMemStats()
			return float64(m.MCacheInuse)
		}),
		NewUpdatableGauge("MCacheSys", func() float64 {
			updateMemStats()
			return float64(m.MCacheSys)
		}),
		NewUpdatableGauge("MSpanInuse", func() float64 {
			updateMemStats()
			return float64(m.MSpanInuse)
		}),
		NewUpdatableGauge("MSpanSys", func() float64 {
			updateMemStats()
			return float64(m.MSpanSys)
		}),
		NewUpdatableGauge("Mallocs", func() float64 {
			updateMemStats()
			return float64(m.Mallocs)
		}),
		NewUpdatableGauge("NextGC", func() float64 {
			updateMemStats()
			return float64(m.NextGC)
		}),
		NewUpdatableGauge("NumForcedGC", func() float64 {
			updateMemStats()
			return float64(m.NumForcedGC)
		}),
		NewUpdatableGauge("NumGC", func() float64 {
			updateMemStats()
			return float64(m.NumGC)
		}),
		NewUpdatableGauge("OtherSys", func() float64 {
			updateMemStats()
			return float64(m.OtherSys)
		}),
		NewUpdatableGauge("PauseTotalNs", func() float64 {
			runtime.ReadMemStats(&m)
			return float64(m.PauseTotalNs)
		}),
		NewUpdatableGauge("StackInuse", func() float64 {
			updateMemStats()
			return float64(m.StackInuse)
		}),
		NewUpdatableGauge("StackSys", func() float64 {
			updateMemStats()
			return float64(m.StackSys)
		}),
		NewUpdatableGauge("Sys", func() float64 {
			updateMemStats()
			return float64(m.Sys)
		}),
		NewUpdatableGauge("TotalAlloc", func() float64 {
			updateMemStats()
			return float64(m.TotalAlloc)
		}),
		NewUpdatableGauge("RandomValue", func() float64 {
			r := rand.New(seed)
			return r.Float64() * 1000000
		}),
		NewUpdatableCounter("PollCount", func() int64 {
			atomic.AddInt64(&pollCounter, 1)
			return pollCounter
		}),
	}
}
