package spos

import (
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

func TestConsensusMetrics_NewConsensusMetrics(t *testing.T) {
	t.Parallel()
	_ = logger.SetLogLevel("*:TRACE")

	t.Run("nil appStatusHandler", func(t *testing.T) {
		t.Parallel()
		cm := NewConsensusMetrics(nil)
		assert.Nil(t, cm)
		assert.True(t, cm.IsInterfaceNil(), "NewConsensusMetrics(nil) should return nil")
	})

	t.Run("normal operation", func(t *testing.T) {
		t.Parallel()
		appStatusHandler := &statusHandlerMock.AppStatusHandlerMock{}
		cm := NewConsensusMetrics(appStatusHandler)
		assert.NotNil(t, cm)
		assert.False(t, cm.IsInterfaceNil(), "NewConsensusMetrics(non-nil) should return non-nil")
	})
}

func TestConsensusMetrics_ResetAverages(t *testing.T) {
	t.Parallel()
	_ = logger.SetLogLevel("*:TRACE")

	t.Run("normal operation", func(t *testing.T) {
		t.Parallel()
		appStatusHandler := &statusHandlerMock.AppStatusHandlerMock{}
		cm := NewConsensusMetrics(appStatusHandler)
		if cm == nil {
			t.Errorf("NewConsensusMetrics() = nil, want non-nil")
			return
		}

		cm.blockReceivedDelaySum = 100
		cm.blockReceivedCount = 10
		cm.blockSignedDelaySum = 200
		cm.blockSignedCount = 20

		cm.ResetAverages()

		assert.Equal(t, uint64(0), cm.blockReceivedDelaySum, "blockReceivedDelaySum should be reset to 0")
		assert.Equal(t, uint64(0), cm.blockReceivedCount, "blockReceivedCount should be reset to 0")
		assert.Equal(t, uint64(0), cm.blockSignedDelaySum, "blockSignedDelaySum should be reset to 0")
		assert.Equal(t, uint64(0), cm.blockSignedCount, "blockSignedCount should be reset to 0")
	})
}

func TestConsensusMetrics_resetInstanceValues(t *testing.T) {
	t.Parallel()
	_ = logger.SetLogLevel("*:TRACE")

	t.Run("normal operation", func(t *testing.T) {
		t.Parallel()
		appStatusHandler := &statusHandlerMock.AppStatusHandlerMock{}
		cm := NewConsensusMetrics(appStatusHandler)
		if cm == nil {
			t.Errorf("NewConsensusMetrics() = nil, want non-nil")
			return
		}

		cm.blockHeaderReceivedOrSentDelay = 100
		cm.blockBodyReceivedOrSentDelay = 200
		cm.blockHash = []byte{0x01, 0x02, 0x03}

		cm.resetInstanceValues()

		assert.Equal(t, uint64(0), cm.blockHeaderReceivedOrSentDelay, "blockHeaderReceivedOrSentDelay should be reset to 0")
		assert.Equal(t, uint64(0), cm.blockBodyReceivedOrSentDelay, "blockBodyReceivedOrSentDelay should be reset to 0")
		assert.Nil(t, cm.blockHash, "blockHash should be reset to nil")
	})
}

func TestConsensusMetrics_SetBlockHeaderAndBodyReceived(t *testing.T) {
	t.Parallel()
	_ = logger.SetLogLevel("*:TRACE")

	t.Run("with header received first", func(t *testing.T) {
		t.Parallel()
		appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()
		cm := NewConsensusMetrics(appStatusHandler)

		appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, 0)
		blockHash := []byte{0x01, 0x02, 0x03}
		headerDelay := uint64(100)
		bodyDelay := uint64(200)

		cm.SetBlockHeaderReceived(blockHash, headerDelay)
		assert.Equal(t, headerDelay, cm.blockHeaderReceivedOrSentDelay, "blockHeaderReceivedOrSentDelay should be set correctly")
		assert.Equal(t, blockHash, cm.blockHash, "blockHash should be set correctly")
		assert.Equal(t, uint64(0), cm.blockBodyReceivedOrSentDelay, "blockBodyReceivedOrSentDelay should be 0")
		assert.Equal(t, uint64(0), appStatusHandler.GetUint64(common.MetricReceivedProposedBlockBody), "blockReceivedDelay metric should not be set")

		cm.SetBlockBodyReceived(blockHash, bodyDelay)
		assert.Equal(t, bodyDelay, cm.blockBodyReceivedOrSentDelay, "blockBodyReceivedOrSentDelay should be reset to 0")
		assert.Equal(t, headerDelay, cm.blockHeaderReceivedOrSentDelay, "blockHeaderReceivedOrSentDelay should be reset to 0")
		assert.Equal(t, blockHash, cm.blockHash, "blockHash should be set correctly")
		assert.Equal(t, uint64(200), appStatusHandler.GetUint64(common.MetricReceivedProposedBlockBody))
	})

	t.Run("with body received first", func(t *testing.T) {
		t.Parallel()

		appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()
		cm := NewConsensusMetrics(appStatusHandler)

		appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, 0)
		blockHash := []byte{0x01, 0x02, 0x04}
		headerDelay := uint64(150)
		bodyDelay := uint64(50)

		cm.SetBlockBodyReceived(blockHash, bodyDelay)
		assert.Equal(t, bodyDelay, cm.blockBodyReceivedOrSentDelay, "blockBodyReceivedOrSentDelay should be set correctly")
		assert.Equal(t, blockHash, cm.blockHash, "blockHash should be set correctly")
		assert.Equal(t, uint64(0), cm.blockHeaderReceivedOrSentDelay, "blockHeaderReceivedOrSentDelay should be 0")
		assert.Equal(t, uint64(0), appStatusHandler.GetUint64(common.MetricReceivedProposedBlockBody), "blockReceivedDelay metric should be 0")

		cm.SetBlockHeaderReceived(blockHash, headerDelay)
		assert.Equal(t, blockHash, cm.blockHash, "blockHash should be set correctly")
		assert.Equal(t, bodyDelay, cm.blockBodyReceivedOrSentDelay, "blockBodyReceivedOrSentDelay should be reset to 0")
		assert.Equal(t, headerDelay, cm.blockHeaderReceivedOrSentDelay, "blockHeaderReceivedOrSentDelay should be reset to 0")
		assert.Equal(t, uint64(150), appStatusHandler.GetUint64(common.MetricReceivedProposedBlockBody))
	})

	t.Run("with body for a different hash", func(t *testing.T) {
		t.Parallel()
		appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()
		cm := NewConsensusMetrics(appStatusHandler)

		appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, 0)
		headerHash := []byte{0x01, 0x02, 0x03}
		bodyHash := []byte{0x01, 0x02, 0x04}
		headerDelay := uint64(100)
		bodyDelay := uint64(200)

		cm.SetBlockHeaderReceived(headerHash, headerDelay)
		assert.Equal(t, headerDelay, cm.blockHeaderReceivedOrSentDelay, "blockHeaderReceivedOrSentDelay should be set correctly")
		assert.Equal(t, headerHash, cm.blockHash, "blockHash should be set correctly")
		assert.Equal(t, uint64(0), cm.blockBodyReceivedOrSentDelay, "blockBodyReceivedOrSentDelay should be 0")
		assert.Equal(t, uint64(0), appStatusHandler.GetUint64(common.MetricReceivedProposedBlockBody), "blockReceivedDelay metric should not be set")

		cm.SetBlockBodyReceived(bodyHash, bodyDelay)
		assert.Equal(t, uint64(0), cm.blockBodyReceivedOrSentDelay, "blockBodyReceivedOrSentDelay should be reset to 0")
		assert.Equal(t, headerDelay, cm.blockHeaderReceivedOrSentDelay, "blockHeaderReceivedOrSentDelay should be reset to 0")
		assert.Equal(t, headerHash, cm.blockHash, "blockHash should be set correctly")
		assert.Equal(t, uint64(0), appStatusHandler.GetUint64(common.MetricReceivedProposedBlockBody), "blockReceivedDelay metric should not be set")
	})
}

func TestConsensusMetrics_SetProof(t *testing.T) {
	//t.Parallel()
	_ = logger.SetLogLevel("*:TRACE")
	t.Run("with no header or body received", func(t *testing.T) {
		//t.Parallel()
		appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()
		cm := NewConsensusMetrics(appStatusHandler)

		appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, 0)
		blockHash := []byte{0x01, 0x02, 0x03}
		proofDelay := uint64(50)

		err := cm.SetProofReceived(blockHash, proofDelay)
		assert.NotNil(t, err, "SetProofReceived should return error when no header or body received")
	})

	t.Run("with header received first", func(t *testing.T) {
		//t.Parallel()
		appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()
		cm := NewConsensusMetrics(appStatusHandler)

		appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, 0)
		appStatusHandler.SetUInt64Value(common.MetricReceivedProof, 0)

		blockHash := []byte{0x01, 0x02, 0x04}
		headerDelay := uint64(100)
		bodyDelay := uint64(200)
		proofDelay := uint64(250)

		cm.SetBlockHeaderReceived(blockHash, headerDelay)
		cm.SetBlockBodyReceived(blockHash, bodyDelay)
		cm.SetProofReceived(blockHash, proofDelay)

		assert.Equal(t, uint64(0), cm.blockHeaderReceivedOrSentDelay, "blockHeaderReceivedOrSentDelay should be reset to 0")
		assert.Equal(t, uint64(0), cm.blockBodyReceivedOrSentDelay, "blockBodyReceivedOrSentDelay should be reset to 0")
		assert.Nil(t, cm.blockHash, "blockHash should be reset to nil")
		assert.Equal(t, bodyDelay, appStatusHandler.GetUint64(common.MetricReceivedProposedBlockBody), "blockReceivedDelay metric should be updated correctly")
		assert.Equal(t, proofDelay-bodyDelay, appStatusHandler.GetUint64(common.MetricReceivedProof), "blockReceivedProof metric should be updated correctly")
	})

	t.Run("with body received first", func(t *testing.T) {
		//t.Parallel()
		appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()
		cm := NewConsensusMetrics(appStatusHandler)

		appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, 0)
		appStatusHandler.SetUInt64Value(common.MetricReceivedProof, 0)

		blockHash := []byte{0x01, 0x02, 0x04}
		headerDelay := uint64(200)
		bodyDelay := uint64(100)
		proofDelay := uint64(250)

		cm.SetBlockBodyReceived(blockHash, headerDelay)
		cm.SetBlockHeaderReceived(blockHash, bodyDelay)
		cm.SetProofReceived(blockHash, proofDelay)

		assert.Equal(t, uint64(0), cm.blockHeaderReceivedOrSentDelay, "blockHeaderReceivedOrSentDelay should be reset to 0")
		assert.Equal(t, uint64(0), cm.blockBodyReceivedOrSentDelay, "blockBodyReceivedOrSentDelay should be reset to 0")
		assert.Nil(t, cm.blockHash, "blockHash should be reset to nil")
		assert.Equal(t, headerDelay, appStatusHandler.GetUint64(common.MetricReceivedProposedBlockBody), "blockReceivedDelay metric should be updated correctly")
		assert.Equal(t, proofDelay-headerDelay, appStatusHandler.GetUint64(common.MetricReceivedProof), "blockReceivedProof metric should be updated correctly")
	})

	t.Run("with only header received", func(t *testing.T) {
		//t.Parallel()
		appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()
		cm := NewConsensusMetrics(appStatusHandler)

		appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, 0)
		appStatusHandler.SetUInt64Value(common.MetricReceivedProof, 0)

		blockHash := []byte{0x01, 0x02, 0x04}
		headerDelay := uint64(200)
		proofDelay := uint64(250)

		cm.SetBlockHeaderReceived(blockHash, headerDelay)
		cm.SetProofReceived(blockHash, proofDelay)

		assert.Equal(t, uint64(0), cm.blockHeaderReceivedOrSentDelay, "blockHeaderReceivedOrSentDelay should be reset to 0")
		assert.Equal(t, uint64(0), cm.blockBodyReceivedOrSentDelay, "blockBodyReceivedOrSentDelay should be reset to 0")
		assert.Nil(t, cm.blockHash, "blockHash should be reset to nil")
		assert.Equal(t, headerDelay, appStatusHandler.GetUint64(common.MetricReceivedProposedBlockBody), "blockReceivedDelay metric should be updated correctly")
		assert.Equal(t, proofDelay-headerDelay, appStatusHandler.GetUint64(common.MetricReceivedProof), "blockReceivedProof metric should be updated correctly")
	})

	t.Run("with only body received", func(t *testing.T) {
		//t.Parallel()
		appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()
		cm := NewConsensusMetrics(appStatusHandler)

		appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, 0)
		appStatusHandler.SetUInt64Value(common.MetricReceivedProof, 0)

		blockHash := []byte{0x01, 0x02, 0x04}
		bodyDelay := uint64(200)
		proofDelay := uint64(250)

		cm.SetBlockBodyReceived(blockHash, bodyDelay)
		cm.SetProofReceived(blockHash, proofDelay)

		assert.Equal(t, uint64(0), cm.blockHeaderReceivedOrSentDelay, "blockHeaderReceivedOrSentDelay should be reset to 0")
		assert.Equal(t, uint64(0), cm.blockBodyReceivedOrSentDelay, "blockBodyReceivedOrSentDelay should be reset to 0")
		assert.Nil(t, cm.blockHash, "blockHash should be reset to nil")
		assert.Equal(t, bodyDelay, appStatusHandler.GetUint64(common.MetricReceivedProposedBlockBody), "blockReceivedDelay metric should be updated correctly")
		assert.Equal(t, proofDelay-bodyDelay, appStatusHandler.GetUint64(common.MetricReceivedProof), "blockReceivedProof metric should be updated correctly")
	})

	t.Run("with proof for different hash", func(t *testing.T) {
		//t.Parallel()
		appStatusHandler := statusHandlerMock.NewAppStatusHandlerMock()
		cm := NewConsensusMetrics(appStatusHandler)

		appStatusHandler.SetUInt64Value(common.MetricReceivedProposedBlockBody, 0)
		appStatusHandler.SetUInt64Value(common.MetricReceivedProof, 0)

		blockHash := []byte{0x01, 0x02, 0x03}
		proofHash := []byte{0x01, 0x02, 0x04}
		headerDelay := uint64(100)
		bodyDelay := uint64(200)
		proofDelay := uint64(250)

		cm.SetBlockHeaderReceived(blockHash, headerDelay)
		cm.SetBlockBodyReceived(blockHash, bodyDelay)
		cm.SetProofReceived(proofHash, proofDelay)

		assert.NotEqual(t, uint64(0), cm.blockHeaderReceivedOrSentDelay, "blockHeaderReceivedOrSentDelay should not be reset to 0")
		assert.NotEqual(t, uint64(0), cm.blockBodyReceivedOrSentDelay, "blockBodyReceivedOrSentDelay should not be reset to 0")
		assert.NotNil(t, cm.blockHash, "blockHash should not be reset to nil")
		assert.Equal(t, bodyDelay, appStatusHandler.GetUint64(common.MetricReceivedProposedBlockBody), "blockReceivedDelay metric should be updated correctly")
		assert.Equal(t, uint64(0), appStatusHandler.GetUint64(common.MetricReceivedProof), "blockReceivedProof metric should be updated correctly")
	})

}
