package metric

import (
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

const cacheTimeCpu = 1 * time.Second

var (
	virtMem       *mem.VirtualMemoryStat
	cpuStats      []float64
	lastUpdateCpu = ConcurrentTime{
		time: time.Now(),
		mu:   new(sync.RWMutex),
	}
)

func updatePSStats() {
	if time.Since(lastUpdateCpu.get()) > cacheTimeCpu {
		cpuStats, _ = cpu.Percent(0, true)
		virtMem, _ = mem.VirtualMemory()
		lastUpdateCpu.set(time.Now())
	}
}
func NewPsUtilStats() []UpdatableMetric {
	result := make([]UpdatableMetric, 0)
	virtMem, _ = mem.VirtualMemory()
	result = append(result, NewUpdatableGauge("TotalMemory", func() float64 {
		updatePSStats()
		return float64(virtMem.Total)
	}))
	result = append(result, NewUpdatableGauge("FreeMemory", func() float64 {
		updatePSStats()
		return float64(virtMem.Free)
	}))

	cpuStats, _ = cpu.Percent(0, true)
	for i := range cpuStats {
		func(cpuNo int) {
			result = append(result, NewUpdatableGauge(fmt.Sprintf("CPUutilization%d", cpuNo), func() float64 {
				updatePSStats()
				return cpuStats[cpuNo]
			}))
		}(i)
	}

	return result
}
