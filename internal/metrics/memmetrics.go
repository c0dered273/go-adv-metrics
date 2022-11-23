package metrics

import (
	"runtime"
)

func GetMemStats() []Gauge {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return []Gauge{
		{
			name:  "Alloc",
			value: float64(m.Alloc),
		},
		{
			name:  "BuckHashSys",
			value: float64(m.BuckHashSys),
		},
		{
			name:  "Frees",
			value: float64(m.Frees),
		},
		{
			name:  "GCCPUFraction",
			value: float64(m.GCCPUFraction),
		},
		{
			name:  "GCSys",
			value: float64(m.GCSys),
		},
		{
			name:  "HeapAlloc",
			value: float64(m.HeapAlloc),
		},
		{
			name:  "HeapIdle",
			value: float64(m.HeapIdle),
		},
		{
			name:  "HeapInuse",
			value: float64(m.HeapInuse),
		},
		{
			name:  "HeapObjects",
			value: float64(m.HeapObjects),
		},
		{
			name:  "HeapReleased",
			value: float64(m.HeapReleased),
		},
		{
			name:  "HeapSys",
			value: float64(m.HeapSys),
		},
		{
			name:  "LastGC",
			value: float64(m.LastGC),
		},
		{
			name:  "Lookups",
			value: float64(m.Lookups),
		},
		{
			name:  "MCacheInuse",
			value: float64(m.MCacheInuse),
		},
		{
			name:  "MCacheSys",
			value: float64(m.MCacheSys),
		},
		{
			name:  "MSpanInuse",
			value: float64(m.MSpanInuse),
		},
		{
			name:  "MSpanSys",
			value: float64(m.MSpanSys),
		},
		{
			name:  "Mallocs",
			value: float64(m.Mallocs),
		},
		{
			name:  "NextGC",
			value: float64(m.NextGC),
		},
		{
			name:  "NumForcedGC",
			value: float64(m.NumForcedGC),
		},
		{
			name:  "NumGC",
			value: float64(m.NumGC),
		},
		{
			name:  "OtherSys",
			value: float64(m.OtherSys),
		},
		{
			name:  "PauseTotalINs",
			value: float64(m.PauseTotalNs),
		},
		{
			name:  "StackInuse",
			value: float64(m.StackInuse),
		},
		{
			name:  "StackSys",
			value: float64(m.StackSys),
		},
		{
			name:  "Sys",
			value: float64(m.Sys),
		},
		{
			name:  "TotalAlloc",
			value: float64(m.TotalAlloc),
		},
	}
}
