package presenter

import (
	"github.com/ElrondNetwork/elrond-go/core/constants"
)

// GetCpuLoadPercent wil return cpu load
func (psh *PresenterStatusHandler) GetCpuLoadPercent() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricCpuLoadPercent)
}

// GetMemLoadPercent will return mem percent usage of node
func (psh *PresenterStatusHandler) GetMemLoadPercent() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricMemLoadPercent)
}

// GetTotalMem will return how much memory does the computer have
func (psh *PresenterStatusHandler) GetTotalMem() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricMemTotal)
}

// GetMemUsedByNode will return how many memory node uses
func (psh *PresenterStatusHandler) GetMemUsedByNode() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricMemUsedGolang)
}

// GetNetworkRecvPercent will return network receive percent
func (psh *PresenterStatusHandler) GetNetworkRecvPercent() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricNetworkRecvPercent)
}

// GetNetworkRecvBps will return network received bytes per second
func (psh *PresenterStatusHandler) GetNetworkRecvBps() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricNetworkRecvBps)
}

// GetNetworkRecvBpsPeak will return received bytes per seconds peak
func (psh *PresenterStatusHandler) GetNetworkRecvBpsPeak() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricNetworkRecvBpsPeak)
}

// GetNetworkSentPercent will return network sent percent
func (psh *PresenterStatusHandler) GetNetworkSentPercent() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricNetworkSentPercent)
}

// GetNetworkSentBps will return network sent bytes per second
func (psh *PresenterStatusHandler) GetNetworkSentBps() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricNetworkSentBps)
}

// GetNetworkSentBpsPeak will return sent bytes per seconds peak
func (psh *PresenterStatusHandler) GetNetworkSentBpsPeak() uint64 {
	return psh.getFromCacheAsUint64(constants.MetricNetworkSentBpsPeak)
}
