package machine

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
)

// machineStatistics aggregates all available local machine statistics like CPU and memory usage and
// network stats
type machineStatistics struct {
	appStatusHandler      core.AppStatusHandler
	cpu                   *cpuStatistics
	mem                   *memStatistics
	net                   *netStatistics
	metricsUpdateInterval time.Duration
}

// NewMachineStatistics starts the update processes for cpu, mem, net and create a new instance of machineStatistics
func NewMachineStatistics(
	appStatusHandler core.AppStatusHandler,
	metricsUpdateInterval time.Duration,
) (*machineStatistics, error) {

	ms := &machineStatistics{
		cpu:                   &cpuStatistics{},
		mem:                   &memStatistics{},
		net:                   &netStatistics{},
		appStatusHandler:      appStatusHandler,
		metricsUpdateInterval: metricsUpdateInterval,
	}

	ms.startGettingCpuStatistics()
	ms.startGettingMemStatistics()
	ms.startGettingNetStatistics()
	ms.updateStatistics()

	return ms, nil
}

func (ms *machineStatistics) startGettingCpuStatistics() {
	go func() {
		for {
			ms.cpu.getCpuStatistics()
		}
	}()
}

func (ms *machineStatistics) startGettingMemStatistics() {
	go func() {
		for {
			ms.mem.getMemStatistics()
		}
	}()
}

func (ms *machineStatistics) startGettingNetStatistics() {
	go func() {
		for {
			ms.net.getNetStatistics()
		}
	}()
}

func (ms *machineStatistics) updateStatistics() {
	go func() {
		for {
			time.Sleep(ms.metricsUpdateInterval)

			ms.appStatusHandler.SetUInt64Value(core.MetricCpuLoadPercent, ms.cpu.CpuPercentUsage())
			ms.appStatusHandler.SetUInt64Value(core.MetricMemLoadPercent, ms.mem.MemPercentUsage())
			ms.appStatusHandler.SetUInt64Value(core.MetricTotalMem, ms.mem.TotalMemory())

			ms.appStatusHandler.SetUInt64Value(core.MetricNetworkRecvBps, ms.net.BpsRecv())
			ms.appStatusHandler.SetUInt64Value(core.MetricNetworkRecvBpsPeak, ms.net.BpsRecvPeak())
			ms.appStatusHandler.SetUInt64Value(core.MetricNetworkRecvPercent, ms.net.PercentRecv())
			ms.appStatusHandler.SetUInt64Value(core.MetricNetworkSentBps, ms.net.BpsSent())
			ms.appStatusHandler.SetUInt64Value(core.MetricNetworkSentBpsPeak, ms.net.BpsSentPeak())
			ms.appStatusHandler.SetUInt64Value(core.MetricNetworkSentPercent, ms.net.PercentSent())
		}
	}()
}
