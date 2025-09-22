package spos

import (
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestConsensusMetrics_NewConsensusMetrics(t *testing.T) {
	t.Parallel()

	t.Run("nil appStatusHandler", func(t *testing.T) {
		t.Parallel()
		cm, _ := NewConsensusMetrics(nil)
		assert.Nil(t, cm)
		assert.True(t, cm.IsInterfaceNil(), "NewConsensusMetrics(nil) should return nil")
	})

	t.Run("normal operation", func(t *testing.T) {
		t.Parallel()
		appStatusHandler := &statusHandlerMock.AppStatusHandlerMock{}
		cm, _ := NewConsensusMetrics(appStatusHandler)
		assert.NotNil(t, cm)
		assert.False(t, cm.IsInterfaceNil(), "NewConsensusMetrics(non-nil) should return non-nil")
	})
}

func TestConsensusMetrics_ResetAverages(t *testing.T) {
	t.Parallel()

	t.Run("normal operation", func(t *testing.T) {
		t.Parallel()
		appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()
		cm, _ := NewConsensusMetrics(appStatusHandler)
		if cm == nil {
			t.Errorf("NewConsensusMetrics() = nil, want non-nil")
			return
		}

		cm.blockReceivedDelaySum = 100
		cm.blockReceivedCount = 10
		cm.proofReceivedDelaySum = 300
		cm.proofReceivedCount = 20

		cm.ResetAverages()

		assert.Equal(t, uint64(0), cm.blockReceivedDelaySum, "blockReceivedDelaySum should be reset to 0")
		assert.Equal(t, uint64(0), cm.blockReceivedCount, "blockReceivedCount should be reset to 0")
		assert.Equal(t, uint64(0), cm.proofReceivedDelaySum, "blockSignedDelaySum should be reset to 0")
		assert.Equal(t, uint64(0), cm.proofReceivedCount, "blockSignedCount should be reset to 0")
	})
}

func TestConsensusMetrics_resetInstanceValues(t *testing.T) {
	t.Parallel()

	t.Run("normal operation", func(t *testing.T) {
		t.Parallel()
		appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()

		cm, _ := NewConsensusMetrics(appStatusHandler)
		if cm == nil {
			t.Errorf("NewConsensusMetrics() = nil, want non-nil")
			return
		}

		appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, 0)
		appStatusHandler.SetUInt64Value(common.MetricReceivedProof, 0)
		cm.blockReceivedOrSentDelay = 200

		cm.ResetInstanceValues()

		assert.Equal(t, uint64(0), cm.blockReceivedOrSentDelay, "blockHeaderReceivedOrSentDelay should be reset to 0")
	})
}

func TestConsensusMetrics_SetBlockReceivedOrSent(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()
		appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()
		cm, _ := NewConsensusMetrics(appStatusHandler)

		appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, 0)

		delay_100 := uint64(100)
		cm.SetBlockReceivedOrSent(delay_100)

		assert.Equal(t, uint64(delay_100), cm.blockReceivedOrSentDelay, "blockHeaderReceivedOrSentDelay should be set correctly")
		assert.Equal(t, uint64(delay_100), appStatusHandler.GetUint64(common.MetricReceivedProposedBlockBody), "blockReceivedDelay metric should be set correctly")
		assert.Equal(t, uint64(delay_100), appStatusHandler.GetUint64(common.MetricAvgReceivedProposedBlockBody), "AvgReceivedProposedBlockBody should be set correctly")
		assert.Equal(t, uint64(1), cm.blockReceivedCount, "blockReceivedCount should be incremented")
		assert.Equal(t, uint64(delay_100), cm.blockReceivedDelaySum, "blockReceivedDelaySum should be updated correctly")

		delay_200 := uint64(200)

		cm.SetBlockReceivedOrSent(delay_200)

		assert.Equal(t, uint64(delay_200), cm.blockReceivedOrSentDelay, "blockHeaderReceivedOrSentDelay should be set correctly")
		assert.Equal(t, uint64(delay_200), appStatusHandler.GetUint64(common.MetricReceivedProposedBlockBody), "blockReceivedDelay metric should be set correctly")
		assert.Equal(t, uint64((delay_100+delay_200)/2), appStatusHandler.GetUint64(common.MetricAvgReceivedProposedBlockBody), "AvgReceivedProposedBlockBody should be set correctly")
		assert.Equal(t, uint64(2), cm.blockReceivedCount, "blockReceivedCount should be incremented")
		assert.Equal(t, uint64(delay_100+delay_200), cm.blockReceivedDelaySum, "blockReceivedDelaySum should be updated correctly")

	})
}

func TestConsensusMetrics_SetProof(t *testing.T) {
	t.Parallel()

	t.Run("with no header or body received", func(t *testing.T) {
		t.Parallel()
		appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()
		cm, _ := NewConsensusMetrics(appStatusHandler)

		appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, 0)
		appStatusHandler.SetUInt64Value(common.MetricReceivedProof, 0)

		proofDelay := uint64(50)

		cm.SetProofReceived(proofDelay)
		assert.Zero(t, appStatusHandler.GetUint64(common.MetricReceivedProof), "SetProofReceived should not be set when no header or body received")
	})

	t.Run("with body received first", func(t *testing.T) {
		t.Parallel()
		appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()
		cm, _ := NewConsensusMetrics(appStatusHandler)

		appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, 0)
		appStatusHandler.SetUInt64Value(common.MetricReceivedProof, 0)

		headerDelay := uint64(200)
		proofDelay := uint64(250)

		cm.SetBlockReceivedOrSent(headerDelay)
		cm.SetProofReceived(proofDelay)

		assert.Equal(t, headerDelay, appStatusHandler.GetUint64(common.MetricReceivedProposedBlockBody), "blockReceivedDelay metric should be updated correctly")
		assert.Equal(t, proofDelay-headerDelay, appStatusHandler.GetUint64(common.MetricReceivedProof), "blockReceivedProof metric should be updated correctly")
	})

	t.Run("with proof delay smaller than header and body delay", func(t *testing.T) {
		t.Parallel()
		appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()
		cm, _ := NewConsensusMetrics(appStatusHandler)

		appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, 0)
		appStatusHandler.SetUInt64Value(common.MetricReceivedProof, 0)

		bodyDelay := uint64(200)
		proofDelay := uint64(50)

		cm.SetBlockReceivedOrSent(bodyDelay)
		cm.SetProofReceived(proofDelay)

		assert.Equal(t, bodyDelay, appStatusHandler.GetUint64(common.MetricReceivedProposedBlockBody), "blockReceivedDelay metric should be updated correctly")
		assert.Zero(t, appStatusHandler.GetUint64(common.MetricReceivedProof), "blockReceivedProof metric should be updated correctly")
	})

}

func TestConsensusMetrics_UpdateAverages(t *testing.T) {
	t.Parallel()

	appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()
	cm, _ := NewConsensusMetrics(appStatusHandler)

	cm.blockReceivedDelaySum = 300
	cm.blockReceivedCount = 3
	cm.updateAverages(common.MetricReceivedProposedBlockBody, 500)
	assert.Equal(t, uint64(800), cm.blockReceivedDelaySum, "blockReceivedDelaySum should be updated correctly")
	assert.Equal(t, uint64(4), cm.blockReceivedCount, "blockReceivedCount should be updated correctly")
	assert.Equal(t, uint64(200), appStatusHandler.GetUint64(common.MetricAvgReceivedProposedBlockBody), "AvgReceivedProposedBlockBody should be updated correctly")

	cm.proofReceivedDelaySum = 600
	cm.proofReceivedCount = 4
	cm.updateAverages(common.MetricReceivedProof, 400)
	assert.Equal(t, uint64(1000), cm.proofReceivedDelaySum, "blockSignedDelaySum should be updated correctly")
	assert.Equal(t, uint64(5), cm.proofReceivedCount, "blockSignedCount should be updated correctly")
	assert.Equal(t, uint64(200), appStatusHandler.GetUint64(common.MetricAvgReceivedProof), "AvgReceivedSignatures should be updated correctly")
}

func TestConsensusMetrics_IsProofSet(t *testing.T) {
	t.Parallel()

	appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()
	cm, _ := NewConsensusMetrics(appStatusHandler)
	cm.ResetInstanceValues()
	appStatusHandler.SetUInt64Value(common.MetricReceivedProof, 0)

	assert.False(t, IsProofForCurrentConsensusSet(appStatusHandler), "isProofForCurrentConsensusSet should be false initially")

	cm.SetProofReceived(100)
	assert.False(t, IsProofForCurrentConsensusSet(appStatusHandler), "isProofForCurrentConsensusSet should be false after setting only proof")

	cm.SetBlockReceivedOrSent(200)

	cm.SetProofReceived(300)
	assert.True(t, IsProofForCurrentConsensusSet(appStatusHandler), "isProofForCurrentConsensusSet should be true after setting header, body and proof")
}

func IsProofForCurrentConsensusSet(appStatusHandler *statusHandlerMock.AppStatusHandlerMock) bool {
	return appStatusHandler.GetUint64(common.MetricReceivedProof) > 0
}
