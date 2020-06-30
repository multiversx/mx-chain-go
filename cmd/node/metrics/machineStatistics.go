package metrics

import (
	"errors"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/appStatusPolling"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/statistics/machine"
)

// StartMachineStatisticsPolling will start read information about current  running machini
func StartMachineStatisticsPolling(ash core.AppStatusHandler, pollingInterval time.Duration) error {
	if check.IfNil(ash) {
		return errors.New("nil AppStatusHandler")
	}

	appStatusPollingHandler, err := appStatusPolling.NewAppStatusPolling(ash, pollingInterval)
	if err != nil {
		return errors.New("cannot init AppStatusPolling")
	}

	err = registerCpuStatistics(appStatusPollingHandler)
	if err != nil {
		return err
	}

	err = registerMemStatistics(appStatusPollingHandler)
	if err != nil {
		return err
	}

	err = registerNetStatistics(appStatusPollingHandler)
	if err != nil {
		return err
	}

	appStatusPollingHandler.Poll()

	return nil
}

func registerMemStatistics(appStatusPollingHandler *appStatusPolling.AppStatusPolling) error {
	return appStatusPollingHandler.RegisterPollingFunc(func(appStatusHandler core.AppStatusHandler) {
		mem := machine.AcquireMemStatistics()

		appStatusHandler.SetUInt64Value(core.MetricMemLoadPercent, mem.PercentUsed)
		appStatusHandler.SetUInt64Value(core.MetricMemTotal, mem.Total)
		appStatusHandler.SetUInt64Value(core.MetricMemUsedGolang, mem.UsedByGolang)
		appStatusHandler.SetUInt64Value(core.MetricMemUsedSystem, mem.UsedBySystem)
		appStatusHandler.SetUInt64Value(core.MetricMemHeapInUse, mem.HeapInUse)
		appStatusHandler.SetUInt64Value(core.MetricMemStackInUse, mem.StackInUse)
	})
}

func registerNetStatistics(appStatusPollingHandler *appStatusPolling.AppStatusPolling) error {
	netStats := &machine.NetStatistics{}
	go func() {
		for {
			netStats.ComputeStatistics()
		}
	}()

	return appStatusPollingHandler.RegisterPollingFunc(func(appStatusHandler core.AppStatusHandler) {
		appStatusHandler.SetUInt64Value(core.MetricNetworkRecvBps, netStats.BpsRecv())
		appStatusHandler.SetUInt64Value(core.MetricNetworkRecvBpsPeak, netStats.BpsRecvPeak())
		appStatusHandler.SetUInt64Value(core.MetricNetworkRecvPercent, netStats.PercentRecv())

		appStatusHandler.SetUInt64Value(core.MetricNetworkSentBps, netStats.BpsSent())
		appStatusHandler.SetUInt64Value(core.MetricNetworkSentBpsPeak, netStats.BpsSentPeak())
		appStatusHandler.SetUInt64Value(core.MetricNetworkSentPercent, netStats.PercentSent())
	})
}

func registerCpuStatistics(appStatusPollingHandler *appStatusPolling.AppStatusPolling) error {
	cpuStats, err := machine.NewCpuStatistics()
	if err != nil {
		return err
	}

	go func() {
		for {
			cpuStats.ComputeStatistics()
		}
	}()

	return appStatusPollingHandler.RegisterPollingFunc(func(appStatusHandler core.AppStatusHandler) {
		appStatusHandler.SetUInt64Value(core.MetricCpuLoadPercent, cpuStats.CpuPercentUsage())
	})
}
