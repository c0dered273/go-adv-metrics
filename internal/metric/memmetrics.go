package metric

import (
	"math/rand"
	"runtime"
	"time"
)

var pollCounter int64

func NewMemStats() *Stats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	pollCounter += 1

	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)

	return &Stats{
		[]Gauge{
			{
				Name:  "Alloc",
				Value: float64(m.Alloc),
			},
			{
				Name:  "BuckHashSys",
				Value: float64(m.BuckHashSys),
			},
			{
				Name:  "Frees",
				Value: float64(m.Frees),
			},
			{
				Name:  "GCCPUFraction",
				Value: float64(m.GCCPUFraction),
			},
			{
				Name:  "GCSys",
				Value: float64(m.GCSys),
			},
			{
				Name:  "HeapAlloc",
				Value: float64(m.HeapAlloc),
			},
			{
				Name:  "HeapIdle",
				Value: float64(m.HeapIdle),
			},
			{
				Name:  "HeapInuse",
				Value: float64(m.HeapInuse),
			},
			{
				Name:  "HeapObjects",
				Value: float64(m.HeapObjects),
			},
			{
				Name:  "HeapReleased",
				Value: float64(m.HeapReleased),
			},
			{
				Name:  "HeapSys",
				Value: float64(m.HeapSys),
			},
			{
				Name:  "LastGC",
				Value: float64(m.LastGC),
			},
			{
				Name:  "Lookups",
				Value: float64(m.Lookups),
			},
			{
				Name:  "MCacheInuse",
				Value: float64(m.MCacheInuse),
			},
			{
				Name:  "MCacheSys",
				Value: float64(m.MCacheSys),
			},
			{
				Name:  "MSpanInuse",
				Value: float64(m.MSpanInuse),
			},
			{
				Name:  "MSpanSys",
				Value: float64(m.MSpanSys),
			},
			{
				Name:  "Mallocs",
				Value: float64(m.Mallocs),
			},
			{
				Name:  "NextGC",
				Value: float64(m.NextGC),
			},
			{
				Name:  "NumForcedGC",
				Value: float64(m.NumForcedGC),
			},
			{
				Name:  "NumGC",
				Value: float64(m.NumGC),
			},
			{
				Name:  "OtherSys",
				Value: float64(m.OtherSys),
			},
			{
				Name:  "PauseTotalINs",
				Value: float64(m.PauseTotalNs),
			},
			{
				Name:  "StackInuse",
				Value: float64(m.StackInuse),
			},
			{
				Name:  "StackSys",
				Value: float64(m.StackSys),
			},
			{
				Name:  "Sys",
				Value: float64(m.Sys),
			},
			{
				Name:  "TotalAlloc",
				Value: float64(m.TotalAlloc),
			},
			{
				Name:  "RandomValue",
				Value: r.Float64() * 1000000,
			},
		},
		[]Counter{
			{
				Name:  "PollCounter",
				Value: pollCounter,
			},
		},
	}
}
