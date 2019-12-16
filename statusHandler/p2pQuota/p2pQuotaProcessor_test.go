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

func TestNewP2pQuotaProcessor_NilStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	pqp, err := p2pQuota.NewP2pQuotaProcessor(nil)
	assert.True(t, check.IfNil(pqp))
	assert.Equal(t, statusHandler.ErrNilAppStatusHandler, err)
}

func TestNewP2pQuotaProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	pqp, err := p2pQuota.NewP2pQuotaProcessor(&mock.AppStatusHandlerStub{})
	assert.False(t, check.IfNil(pqp))
	assert.Nil(t, err)
}

//------- AddQuota

func TestP2pQuotaProcessor_AddQuotaShouldWork(t *testing.T) {
	t.Parallel()

	pqp, _ := p2pQuota.NewP2pQuotaProcessor(&mock.AppStatusHandlerStub{})
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

func TestP2pQuotaProcessor_ResetStatisticsShouldEmptyStatsAndCallSetOnAllMetrics(t *testing.T) {
	t.Parallel()

	identifier1 := "identifier"
	numReceived1 := uint64(1)
	sizeReceived1 := uint64(2)
	numProcessed1 := uint64(3)
	sizeProcessed1 := uint64(4)

	status := mock.NewAppStatusHandlerMock()
	pqp, _ := p2pQuota.NewP2pQuotaProcessor(status)
	pqp.AddQuota(identifier1, uint32(numReceived1), sizeReceived1, uint32(numProcessed1), sizeProcessed1)

	pqp.ResetStatistics()

	assert.Nil(t, pqp.GetQuota(identifier1))

	numReceivers := uint64(1)
	checkSumMetrics(t, status, numReceived1, sizeReceived1, numProcessed1, sizeProcessed1)
	checkTopSumMetrics(t, status, numReceived1, sizeReceived1, numProcessed1, sizeProcessed1)
	checkMaxMetrics(t, status, numReceived1, sizeReceived1, numProcessed1, sizeProcessed1)
	checkTopMaxMetrics(t, status, numReceived1, sizeReceived1, numProcessed1, sizeProcessed1)
	checkNumReceivers(t, status, numReceivers, numReceivers)
}

func TestP2pQuotaProcessor_ResetStatisticsShouldSetTops(t *testing.T) {
	t.Parallel()

	identifier1 := "identifier"
	numReceived1 := uint64(1)
	sizeReceived1 := uint64(2)
	numProcessed1 := uint64(3)
	sizeProcessed1 := uint64(4)

	identifier2 := "identifier"
	numReceived2 := uint64(10)
	sizeReceived2 := uint64(20)
	numProcessed2 := uint64(30)
	sizeProcessed2 := uint64(40)

	status := mock.NewAppStatusHandlerMock()
	pqp, _ := p2pQuota.NewP2pQuotaProcessor(status)
	pqp.AddQuota(identifier1, uint32(numReceived1), sizeReceived1, uint32(numProcessed1), sizeProcessed1)
	pqp.ResetStatistics()
	pqp.AddQuota(identifier2, uint32(numReceived2), sizeReceived2, uint32(numProcessed2), sizeProcessed2)

	pqp.ResetStatistics()

	numReceivers := uint64(1)
	checkSumMetrics(t, status, numReceived2, sizeReceived2, numProcessed2, sizeProcessed2)
	checkTopSumMetrics(t, status, numReceived2, sizeReceived2, numProcessed2, sizeProcessed2)
	checkMaxMetrics(t, status, numReceived2, sizeReceived2, numProcessed2, sizeProcessed2)
	checkTopMaxMetrics(t, status, numReceived2, sizeReceived2, numProcessed2, sizeProcessed2)
	checkNumReceivers(t, status, numReceivers, numReceivers)
}

func checkSumMetrics(
	t *testing.T,
	status *mock.AppStatusHandlerMock,
	numReceived uint64,
	sizeReceived uint64,
	numProcessed uint64,
	sizeProcessed uint64,
) {

	value := status.GetUint64(core.MetricP2pSumNumReceivedMessages)
	assert.Equal(t, value, numReceived)

	value = status.GetUint64(core.MetricP2pSumSizeReceivedMessages)
	assert.Equal(t, value, sizeReceived)

	value = status.GetUint64(core.MetricP2pSumNumProcessedMessages)
	assert.Equal(t, value, numProcessed)

	value = status.GetUint64(core.MetricP2pSumSizeProcessedMessages)
	assert.Equal(t, value, sizeProcessed)
}

func checkTopSumMetrics(
	t *testing.T,
	status *mock.AppStatusHandlerMock,
	numReceived uint64,
	sizeReceived uint64,
	numProcessed uint64,
	sizeProcessed uint64,
) {

	value := status.GetUint64(core.MetricP2pTopSumNumReceivedMessages)
	assert.Equal(t, value, numReceived)

	value = status.GetUint64(core.MetricP2pTopSumSizeReceivedMessages)
	assert.Equal(t, value, sizeReceived)

	value = status.GetUint64(core.MetricP2pTopSumNumProcessedMessages)
	assert.Equal(t, value, numProcessed)

	value = status.GetUint64(core.MetricP2pTopSumSizeProcessedMessages)
	assert.Equal(t, value, sizeProcessed)
}

func checkMaxMetrics(
	t *testing.T,
	status *mock.AppStatusHandlerMock,
	numReceived uint64,
	sizeReceived uint64,
	numProcessed uint64,
	sizeProcessed uint64,
) {

	value := status.GetUint64(core.MetricP2pMaxNumReceivedMessages)
	assert.Equal(t, value, numReceived)

	value = status.GetUint64(core.MetricP2pMaxSizeReceivedMessages)
	assert.Equal(t, value, sizeReceived)

	value = status.GetUint64(core.MetricP2pMaxNumProcessedMessages)
	assert.Equal(t, value, numProcessed)

	value = status.GetUint64(core.MetricP2pMaxSizeProcessedMessages)
	assert.Equal(t, value, sizeProcessed)
}

func checkTopMaxMetrics(
	t *testing.T,
	status *mock.AppStatusHandlerMock,
	numReceived uint64,
	sizeReceived uint64,
	numProcessed uint64,
	sizeProcessed uint64,
) {

	value := status.GetUint64(core.MetricP2pTopMaxNumReceivedMessages)
	assert.Equal(t, value, numReceived)

	value = status.GetUint64(core.MetricP2pTopMaxSizeReceivedMessages)
	assert.Equal(t, value, sizeReceived)

	value = status.GetUint64(core.MetricP2pTopMaxNumProcessedMessages)
	assert.Equal(t, value, numProcessed)

	value = status.GetUint64(core.MetricP2pTopMaxSizeProcessedMessages)
	assert.Equal(t, value, sizeProcessed)
}

func checkNumReceivers(
	t *testing.T,
	status *mock.AppStatusHandlerMock,
	numReceiverPeers uint64,
	topNumReceiverPeers uint64,
) {
	value := status.GetUint64(core.MetricP2pNumReceiverPeers)
	assert.Equal(t, value, numReceiverPeers)

	value = status.GetUint64(core.MetricP2pTopNumReceiverPeers)
	assert.Equal(t, value, topNumReceiverPeers)
}
