package metric

import (
	"fmt"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func NewPsUtilStats() []Metric {
	result := make([]Metric, 0)

	v, _ := mem.VirtualMemory()
	result = append(result, NewGaugeMetric("TotalMemory", float64(v.Total)))
	result = append(result, NewGaugeMetric("FreeMemory", float64(v.Free)))

	cpuUsed, _ := cpu.Percent(0, true)
	for i, used := range cpuUsed {
		result = append(result, NewGaugeMetric(fmt.Sprintf("CPUutilization%d", i), used))
	}

	return result
}
