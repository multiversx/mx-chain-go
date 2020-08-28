package presenter

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
)

func TestPresenterStatusHandler_GetCpuLoadPercent(t *testing.T) {
	t.Parallel()

	cpuPercentLoad := uint64(90)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricCpuLoadPercent, cpuPercentLoad)
	result := presenterStatusHandler.GetCpuLoadPercent()

	assert.Equal(t, cpuPercentLoad, result)
}

func TestPresenterStatusHandler_GetMemLoadPercent(t *testing.T) {
	t.Parallel()

	memPercentLoad := uint64(80)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricMemLoadPercent, memPercentLoad)
	result := presenterStatusHandler.GetMemLoadPercent()

	assert.Equal(t, memPercentLoad, result)
}

func TestPresenterStatusHandler_GetTotalMem(t *testing.T) {
	t.Parallel()

	totalMem := uint64(8000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricMemTotal, totalMem)
	result := presenterStatusHandler.GetTotalMem()

	assert.Equal(t, totalMem, result)
}

func TestPresenterStatusHandler_GetMemUsedByNode(t *testing.T) {
	t.Parallel()

	memUsedByNode := uint64(500)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricMemUsedGolang, memUsedByNode)
	result := presenterStatusHandler.GetMemUsedByNode()

	assert.Equal(t, memUsedByNode, result)
}

func TestPresenterStatusHandler_GetNetworkRecvPercent(t *testing.T) {
	t.Parallel()

	networkRecvPercent := uint64(10)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricNetworkRecvPercent, networkRecvPercent)
	result := presenterStatusHandler.GetNetworkRecvPercent()

	assert.Equal(t, networkRecvPercent, result)
}

func TestPresenterStatusHandler_GetNetworkRecvBps(t *testing.T) {
	t.Parallel()

	networkRecvBps := uint64(1000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricNetworkRecvBps, networkRecvBps)
	result := presenterStatusHandler.GetNetworkRecvBps()

	assert.Equal(t, networkRecvBps, result)
}

func TestPresenterStatusHandler_GetNetworkRecvBpsPeak(t *testing.T) {
	t.Parallel()

	networkRecvBpsPeak := uint64(2000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricNetworkRecvBpsPeak, networkRecvBpsPeak)
	result := presenterStatusHandler.GetNetworkRecvBpsPeak()

	assert.Equal(t, networkRecvBpsPeak, result)
}

func TestPresenterStatusHandler_GetNetworkSentPercent(t *testing.T) {
	t.Parallel()

	networkSentPercent := uint64(10)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricNetworkSentPercent, networkSentPercent)
	result := presenterStatusHandler.GetNetworkSentPercent()

	assert.Equal(t, networkSentPercent, result)
}

func TestPresenterStatusHandler_GetNetworkSentBps(t *testing.T) {
	t.Parallel()

	networkSentBps := uint64(1000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricNetworkSentBps, networkSentBps)
	result := presenterStatusHandler.GetNetworkSentBps()

	assert.Equal(t, networkSentBps, result)
}

func TestPresenterStatusHandler_GetNetworkSentBpsPeak(t *testing.T) {
	t.Parallel()

	networkSentBpsPeak := uint64(2000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricNetworkSentBpsPeak, networkSentBpsPeak)
	result := presenterStatusHandler.GetNetworkSentBpsPeak()

	assert.Equal(t, networkSentBpsPeak, result)
}

func TestPresenterStatusHandler_GetNetworkSendBytesInEpoch(t *testing.T) {
	t.Parallel()

	networkBytesSentInEpoch := uint64(10000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricNetworkSendBytesInCurrentEpochPerHost, networkBytesSentInEpoch)
	result := presenterStatusHandler.GetNetworkSentBytesInEpoch()

	assert.Equal(t, networkBytesSentInEpoch, result)
}

func TestPresenterStatusHandler_GetNetworkReceivedBytesInEpoch(t *testing.T) {
	t.Parallel()

	networkBytesSentInEpoch := uint64(15000)
	presenterStatusHandler := NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricNetworkRecvBytesInCurrentEpochPerHost, networkBytesSentInEpoch)
	result := presenterStatusHandler.GetNetworkReceivedBytesInEpoch()

	assert.Equal(t, networkBytesSentInEpoch, result)
}
