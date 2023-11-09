package presenter

import (
	"github.com/multiversx/mx-chain-go/common"
)

// GetCpuLoadPercent wil return cpu load
func (psh *PresenterStatusHandler) GetCpuLoadPercent() uint64 {
	return psh.getFromCacheAsUint64(common.MetricCpuLoadPercent)
}

// GetMemLoadPercent will return mem percent usage of node
func (psh *PresenterStatusHandler) GetMemLoadPercent() uint64 {
	return psh.getFromCacheAsUint64(common.MetricMemLoadPercent)
}

// GetTotalMem will return how much memory does the computer have
func (psh *PresenterStatusHandler) GetTotalMem() uint64 {
	return psh.getFromCacheAsUint64(common.MetricMemTotal)
}

// GetMemUsedByNode will return how many memory node uses
func (psh *PresenterStatusHandler) GetMemUsedByNode() uint64 {
	return psh.getFromCacheAsUint64(common.MetricMemUsedGolang)
}

// GetNetworkRecvPercent will return network receive percent
func (psh *PresenterStatusHandler) GetNetworkRecvPercent() uint64 {
	return psh.getFromCacheAsUint64(common.MetricNetworkRecvPercent)
}

// GetNetworkRecvBps will return network received bytes per second
func (psh *PresenterStatusHandler) GetNetworkRecvBps() uint64 {
	return psh.getFromCacheAsUint64(common.MetricNetworkRecvBps)
}

// GetNetworkRecvBpsPeak will return received bytes per seconds peak
func (psh *PresenterStatusHandler) GetNetworkRecvBpsPeak() uint64 {
	return psh.getFromCacheAsUint64(common.MetricNetworkRecvBpsPeak)
}

// GetNetworkSentPercent will return network sent percent
func (psh *PresenterStatusHandler) GetNetworkSentPercent() uint64 {
	return psh.getFromCacheAsUint64(common.MetricNetworkSentPercent)
}

// GetNetworkSentBps will return network sent bytes per second
func (psh *PresenterStatusHandler) GetNetworkSentBps() uint64 {
	return psh.getFromCacheAsUint64(common.MetricNetworkSentBps)
}

// GetNetworkSentBpsPeak will return sent bytes per seconds peak
func (psh *PresenterStatusHandler) GetNetworkSentBpsPeak() uint64 {
	return psh.getFromCacheAsUint64(common.MetricNetworkSentBpsPeak)
}

// GetNetworkSentBytesInEpoch will return the number of bytes sent in current epoch
func (psh *PresenterStatusHandler) GetNetworkSentBytesInEpoch() uint64 {
	return psh.getFromCacheAsUint64(common.MetricNetworkSendBytesInCurrentEpochPerHost)
}

// GetNetworkReceivedBytesInEpoch will return the number of bytes received in current epoch
func (psh *PresenterStatusHandler) GetNetworkReceivedBytesInEpoch() uint64 {
	return psh.getFromCacheAsUint64(common.MetricNetworkRecvBytesInCurrentEpochPerHost)
}
