package metrics

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/core/constants"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/appStatusPolling"
	"github.com/ElrondNetwork/elrond-go/core/statistics/machine"
)

// StartMachineStatisticsPolling will start read information about current  running machini
func StartMachineStatisticsPolling(ash core.AppStatusHandler, pollingInterval int) error {
	if ash == nil {
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

	err = registeNetStatistics(appStatusPollingHandler)
	if err != nil {
		return err
	}

	appStatusPollingHandler.Poll()

	return nil
}

func registerMemStatistics(appStatusPollingHandler *appStatusPolling.AppStatusPolling) error {
	memStats := &machine.MemStatistics{}
	go func() {
		for {
			memStats.ComputeStatistics()
		}
	}()

	return appStatusPollingHandler.RegisterPollingFunc(func(appStatusHandler core.AppStatusHandler) {
		appStatusHandler.SetUInt64Value(constants.MetricMemLoadPercent, memStats.MemPercentUsage())
		appStatusHandler.SetUInt64Value(constants.MetricMemTotal, memStats.TotalMemory())
		appStatusHandler.SetUInt64Value(constants.MetricMemUsedGolang, memStats.MemoryUsedByGolang())
		appStatusHandler.SetUInt64Value(constants.MetricMemUsedSystem, memStats.MemoryUsedBySystem())
	})
}

func registeNetStatistics(appStatusPollingHandler *appStatusPolling.AppStatusPolling) error {
	netStats := &machine.NetStatistics{}
	go func() {
		for {
			netStats.ComputeStatistics()
		}
	}()

	return appStatusPollingHandler.RegisterPollingFunc(func(appStatusHandler core.AppStatusHandler) {
		appStatusHandler.SetUInt64Value(constants.MetricNetworkRecvBps, netStats.BpsRecv())
		appStatusHandler.SetUInt64Value(constants.MetricNetworkRecvBpsPeak, netStats.BpsRecvPeak())
		appStatusHandler.SetUInt64Value(constants.MetricNetworkRecvPercent, netStats.PercentRecv())

		appStatusHandler.SetUInt64Value(constants.MetricNetworkSentBps, netStats.BpsSent())
		appStatusHandler.SetUInt64Value(constants.MetricNetworkSentBpsPeak, netStats.BpsSentPeak())
		appStatusHandler.SetUInt64Value(constants.MetricNetworkSentPercent, netStats.PercentSent())
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
		appStatusHandler.SetUInt64Value(constants.MetricCpuLoadPercent, cpuStats.CpuPercentUsage())
	})
}
