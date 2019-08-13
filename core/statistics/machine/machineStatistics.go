package machine

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
)

type MachineStatistics struct {
	appStatusHandler      core.AppStatusHandler
	cpu                   *cpuStatistics
	mem                   *memStatistics
	net                   *netStatistics
	metricsUpdateInterval time.Duration
}

func NewMachineStatistics(
	appStatusHandler core.AppStatusHandler,
	metricsUpdateInterval time.Duration,
) (*MachineStatistics, error) {

	ms := &MachineStatistics{
		cpu:                   &cpuStatistics{},
		mem:                   &memStatistics{},
		net:                   &netStatistics{},
		appStatusHandler:      appStatusHandler,
		metricsUpdateInterval: metricsUpdateInterval,
	}

	go ms.cpu.getCpuStatistics()
	go ms.mem.getMemStatistics()
	go ms.net.getNetStatistics()
	go ms.updateStatistics()

	return ms, nil
}

func (ms *MachineStatistics) updateStatistics() {
	for {
		time.Sleep(ms.metricsUpdateInterval)

		ms.appStatusHandler.SetUInt64Value(core.MetricCpuLoadPercent, ms.cpu.CpuPercentUsage())
		ms.appStatusHandler.SetUInt64Value(core.MetricMemLoadPercent, ms.mem.MemPercentUsage())
		networkLoadPercent := (ms.net.percentRecv + ms.net.percentSent) / 2
		ms.appStatusHandler.SetUInt64Value(core.MetricNetworkLoadPercent, networkLoadPercent)
	}
}
