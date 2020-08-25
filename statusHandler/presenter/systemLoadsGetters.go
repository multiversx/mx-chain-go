package presenter

import (
	"github.com/ElrondNetwork/elrond-go/core"
)

// GetCpuLoadPercent wil return cpu load
func (psh *PresenterStatusHandler) GetCpuLoadPercent() uint64 {
	return psh.getFromCacheAsUint64(core.MetricCpuLoadPercent)
}

// GetMemLoadPercent will return mem percent usage of node
func (psh *PresenterStatusHandler) GetMemLoadPercent() uint64 {
	return psh.getFromCacheAsUint64(core.MetricMemLoadPercent)
}

// GetTotalMem will return how much memory does the computer have
func (psh *PresenterStatusHandler) GetTotalMem() uint64 {
	return psh.getFromCacheAsUint64(core.MetricMemTotal)
}

// GetMemUsedByNode will return how many memory node uses
func (psh *PresenterStatusHandler) GetMemUsedByNode() uint64 {
	return psh.getFromCacheAsUint64(core.MetricMemUsedGolang)
}

// GetNetworkRecvPercent will return network receive percent
func (psh *PresenterStatusHandler) GetNetworkRecvPercent() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNetworkRecvPercent)
}

// GetNetworkRecvBps will return network received bytes per second
func (psh *PresenterStatusHandler) GetNetworkRecvBps() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNetworkRecvBps)
}

// GetNetworkRecvBpsPeak will return received bytes per seconds peak
func (psh *PresenterStatusHandler) GetNetworkRecvBpsPeak() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNetworkRecvBpsPeak)
}

// GetNetworkSentPercent will return network sent percent
func (psh *PresenterStatusHandler) GetNetworkSentPercent() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNetworkSentPercent)
}

// GetNetworkSentBps will return network sent bytes per second
func (psh *PresenterStatusHandler) GetNetworkSentBps() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNetworkSentBps)
}

// GetNetworkSentBpsPeak will return sent bytes per seconds peak
func (psh *PresenterStatusHandler) GetNetworkSentBpsPeak() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNetworkSentBpsPeak)
}

// GetNetworkSentBytesInEpoch will return the number of bytes sent in current epoch
func (psh *PresenterStatusHandler) GetNetworkSentBytesInEpoch() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNetworkSendBytesInCurrentEpochPerHost)
}

// GetNetworkReceivedBytesInEpoch will return the number of bytes received in current epoch
func (psh *PresenterStatusHandler) GetNetworkReceivedBytesInEpoch() uint64 {
	return psh.getFromCacheAsUint64(core.MetricNetworkRecvBytesInCurrentEpochPerHost)
}
