package preprocess

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/state"
)

func TestBasePreProcess_handleProcessTransactionInit(t *testing.T) {
	t.Parallel()

	mbHash := []byte("mb hash")
	txHash := []byte("tx hash")
	initProcessedTxsCalled := false

	preProcessorExecutionInfoHandler := &testscommon.PreProcessorExecutionInfoHandlerMock{
		InitProcessedTxsResultsCalled: func(key []byte, parentKey []byte) {
			if !bytes.Equal(key, txHash) {
				return
			}
			require.Equal(t, mbHash, parentKey)

			initProcessedTxsCalled = true
		},
	}

	journalLen := 262845
	bp := &basePreProcess{
		accounts: &state.AccountsStub{
			JournalLenCalled: func() int {
				return journalLen
			},
		},
		gasTracker: gasTracker{
			gasHandler: &testscommon.GasHandlerStub{
				ResetCalled: func(hash []byte) {
					assert.Fail(t, "should have not called gasComputation.Reset")
				},
			},
		},
	}

	recoveredJournalLen := bp.handleProcessTransactionInit(preProcessorExecutionInfoHandler, txHash, mbHash)
	assert.Equal(t, journalLen, recoveredJournalLen)
	assert.True(t, initProcessedTxsCalled)
}

func TestBasePreProcess_getIndexesOfLastTxProcessedOnExecution(t *testing.T) {
	t.Parallel()

	mb := &block.MiniBlock{
		TxHashes: [][]byte{
			[]byte("tx1"),
			[]byte("tx2"),
			[]byte("tx3"),
		},
	}

	t.Run("for v3 header", func(t *testing.T) {
		var headerHandler data.HeaderHandler = &testscommon.HeaderHandlerStub{
			IsHeaderV3Called: func() bool {
				return true
			},
		}

		bp := &basePreProcess{}

		pi, err := bp.getIndexesOfLastTxProcessedOnExecution(mb, headerHandler)
		require.NoError(t, err)

		require.Equal(t, int32(-1), pi.indexOfLastTxProcessed)
		require.Equal(t, int32(len(mb.GetTxHashes())-1), pi.indexOfLastTxProcessedByProposer)

	})
}
