package p2pQuota_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/statusHandler/mock"
	"github.com/ElrondNetwork/elrond-go/statusHandler/p2pQuota"
	"github.com/stretchr/testify/assert"
)

func TestNewP2PQuotaProcessor_NilStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	pqp, err := p2pQuota.NewP2PQuotaProcessor(nil)
	assert.True(t, check.IfNil(pqp))
	assert.Equal(t, statusHandler.ErrNilAppStatusHandler, err)
}

func TestNewP2PQuotaProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	pqp, err := p2pQuota.NewP2PQuotaProcessor(&mock.AppStatusHandlerStub{})
	assert.False(t, check.IfNil(pqp))
	assert.Nil(t, err)
}

//------- AddQuota

func TestP2PQuotaProcessor_AddQuotaShouldWork(t *testing.T) {
	t.Parallel()

	pqp, _ := p2pQuota.NewP2PQuotaProcessor(&mock.AppStatusHandlerStub{})
	nonExistingIdentifier := "non existing identifier"
	identifier := "identifier"
	numReceived := uint32(1)
	sizeReceived := uint64(2)
	numProcessed := uint32(3)
	sizeProcessed := uint64(4)

	pqp.AddQuota(identifier, numReceived, sizeReceived, numProcessed, sizeProcessed)

	nonExistentQuota := pqp.GetQuota(nonExistingIdentifier)
	assert.Nil(t, nonExistentQuota)

	quota := pqp.GetQuota(identifier)
	assert.Equal(t, numReceived, quota.NumReceived())
	assert.Equal(t, sizeReceived, quota.SizeReceived())
	assert.Equal(t, numProcessed, quota.NumProcessed())
	assert.Equal(t, sizeProcessed, quota.SizeProcessed())
}

//------- ResetStatistics

func TestP2PQuotaProcessor_ResetStatisticsShouldEmptyStatsAndCallSetOnAllMetrics(t *testing.T) {
	t.Parallel()

	identifier := "identifier"
	numReceived := uint64(1)
	sizeReceived := uint64(2)
	numProcessed := uint64(3)
	sizeProcessed := uint64(4)

	numReceivedNetwork := uint64(5)
	sizeReceivedNetwork := uint64(6)
	numProcessedNetwork := uint64(7)
	sizeProcessedNetwork := uint64(8)

	status := mock.NewAppStatusHandlerMock()
	pqp, _ := p2pQuota.NewP2PQuotaProcessor(status)
	pqp.AddQuota(identifier, uint32(numReceived), sizeReceived, uint32(numProcessed), sizeProcessed)
	pqp.SetGlobalQuota(uint32(numReceivedNetwork), sizeReceivedNetwork, uint32(numProcessedNetwork), sizeProcessedNetwork)

	pqp.ResetStatistics()

	assert.Nil(t, pqp.GetQuota(identifier))

	numReceivers := uint64(1)
	checkNetworkMetrics(t, status, numReceivedNetwork, sizeReceivedNetwork, numProcessedNetwork, sizeProcessedNetwork)
	checkPeakNetworkMetrics(t, status, numReceivedNetwork, sizeReceivedNetwork, numProcessedNetwork, sizeProcessedNetwork)
	checkPeerMetrics(t, status, numReceived, sizeReceived, numProcessed, sizeProcessed)
	checkPeakPeerMetrics(t, status, numReceived, sizeReceived, numProcessed, sizeProcessed)
	checkNumReceivers(t, status, numReceivers, numReceivers)
}

func TestP2PQuotaProcessor_ResetStatisticsShouldSetPeerStatisticsTops(t *testing.T) {
	t.Parallel()

	identifier1 := "identifier"
	numReceived1 := uint64(10)
	sizeReceived1 := uint64(20)
	numProcessed1 := uint64(30)
	sizeProcessed1 := uint64(40)

	identifier2 := "identifier"
	numReceived2 := uint64(1)
	sizeReceived2 := uint64(2)
	numProcessed2 := uint64(3)
	sizeProcessed2 := uint64(4)

	status := mock.NewAppStatusHandlerMock()
	pqp, _ := p2pQuota.NewP2PQuotaProcessor(status)
	pqp.AddQuota(identifier1, uint32(numReceived1), sizeReceived1, uint32(numProcessed1), sizeProcessed1)
	pqp.ResetStatistics()
	pqp.AddQuota(identifier2, uint32(numReceived2), sizeReceived2, uint32(numProcessed2), sizeProcessed2)

	pqp.ResetStatistics()

	numReceivers := uint64(1)
	checkPeerMetrics(t, status, numReceived2, sizeReceived2, numProcessed2, sizeProcessed2)
	checkPeakPeerMetrics(t, status, numReceived1, sizeReceived1, numProcessed1, sizeProcessed1)
	checkNumReceivers(t, status, numReceivers, numReceivers)
}

func TestP2PQuotaProcessor_ResetStatisticsShouldSetNetworkStatisticsTops(t *testing.T) {
	t.Parallel()

	numReceivedNetwork1 := uint64(10)
	sizeReceivedNetwork1 := uint64(20)
	numProcessedNetwork1 := uint64(30)
	sizeProcessedNetwork1 := uint64(40)

	numReceivedNetwork2 := uint64(1)
	sizeReceivedNetwork2 := uint64(2)
	numProcessedNetwork2 := uint64(3)
	sizeProcessedNetwork2 := uint64(4)

	status := mock.NewAppStatusHandlerMock()
	pqp, _ := p2pQuota.NewP2PQuotaProcessor(status)
	pqp.SetGlobalQuota(uint32(numReceivedNetwork1), sizeReceivedNetwork1, uint32(numProcessedNetwork1), sizeProcessedNetwork1)
	pqp.ResetStatistics()
	pqp.SetGlobalQuota(uint32(numReceivedNetwork2), sizeReceivedNetwork2, uint32(numProcessedNetwork2), sizeProcessedNetwork2)

	pqp.ResetStatistics()

	checkNetworkMetrics(t, status, numReceivedNetwork2, sizeReceivedNetwork2, numProcessedNetwork2, sizeProcessedNetwork2)
	checkPeakNetworkMetrics(t, status, numReceivedNetwork1, sizeReceivedNetwork1, numProcessedNetwork1, sizeProcessedNetwork1)
}

func checkNetworkMetrics(
	t *testing.T,
	status *mock.AppStatusHandlerMock,
	numReceived uint64,
	sizeReceived uint64,
	numProcessed uint64,
	sizeProcessed uint64,
) {

	value := status.GetUint64(core.MetricP2PNetworkNumReceivedMessages)
	assert.Equal(t, value, numReceived)

	value = status.GetUint64(core.MetricP2PNetworkSizeReceivedMessages)
	assert.Equal(t, value, sizeReceived)

	value = status.GetUint64(core.MetricP2PNetworkNumProcessedMessages)
	assert.Equal(t, value, numProcessed)

	value = status.GetUint64(core.MetricP2PNetworkSizeProcessedMessages)
	assert.Equal(t, value, sizeProcessed)
}

func checkPeakNetworkMetrics(
	t *testing.T,
	status *mock.AppStatusHandlerMock,
	numReceived uint64,
	sizeReceived uint64,
	numProcessed uint64,
	sizeProcessed uint64,
) {

	value := status.GetUint64(core.MetricP2PPeakNetworkNumReceivedMessages)
	assert.Equal(t, value, numReceived)

	value = status.GetUint64(core.MetricP2PPeakNetworkSizeReceivedMessages)
	assert.Equal(t, value, sizeReceived)

	value = status.GetUint64(core.MetricP2PPeakNetworkNumProcessedMessages)
	assert.Equal(t, value, numProcessed)

	value = status.GetUint64(core.MetricP2PPeakNetworkSizeProcessedMessages)
	assert.Equal(t, value, sizeProcessed)
}

func checkPeerMetrics(
	t *testing.T,
	status *mock.AppStatusHandlerMock,
	numReceived uint64,
	sizeReceived uint64,
	numProcessed uint64,
	sizeProcessed uint64,
) {

	value := status.GetUint64(core.MetricP2PPeerNumReceivedMessages)
	assert.Equal(t, value, numReceived)

	value = status.GetUint64(core.MetricP2PPeerSizeReceivedMessages)
	assert.Equal(t, value, sizeReceived)

	value = status.GetUint64(core.MetricP2PPeerNumProcessedMessages)
	assert.Equal(t, value, numProcessed)

	value = status.GetUint64(core.MetricP2PPeerSizeProcessedMessages)
	assert.Equal(t, value, sizeProcessed)
}

func checkPeakPeerMetrics(
	t *testing.T,
	status *mock.AppStatusHandlerMock,
	numReceived uint64,
	sizeReceived uint64,
	numProcessed uint64,
	sizeProcessed uint64,
) {

	value := status.GetUint64(core.MetricP2PPeakPeerNumReceivedMessages)
	assert.Equal(t, value, numReceived)

	value = status.GetUint64(core.MetricP2PPeakPeerSizeReceivedMessages)
	assert.Equal(t, value, sizeReceived)

	value = status.GetUint64(core.MetricP2PPeakPeerxNumProcessedMessages)
	assert.Equal(t, value, numProcessed)

	value = status.GetUint64(core.MetricP2PPeakPeerSizeProcessedMessages)
	assert.Equal(t, value, sizeProcessed)
}

func checkNumReceivers(
	t *testing.T,
	status *mock.AppStatusHandlerMock,
	numReceiverPeers uint64,
	topNumReceiverPeers uint64,
) {
	value := status.GetUint64(core.MetricP2PNumReceiverPeers)
	assert.Equal(t, value, numReceiverPeers)

	value = status.GetUint64(core.MetricP2PPeakNumReceiverPeers)
	assert.Equal(t, value, topNumReceiverPeers)
}
