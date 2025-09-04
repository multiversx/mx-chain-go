package spos

import (
	"testing"

	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

func TestConsensusMetrics_NewConsensusMetrics(t *testing.T) {
	t.Parallel()
	_ = logger.SetLogLevel("*:TRACE")

	t.Run("normal operation", func(t *testing.T) {
		t.Parallel()
		appStatusHandler := &statusHandlerMock.AppStatusHandlerStub{}
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
		appStatusHandler := &statusHandlerMock.AppStatusHandlerStub{}
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
		appStatusHandler := &statusHandlerMock.AppStatusHandlerStub{}
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
