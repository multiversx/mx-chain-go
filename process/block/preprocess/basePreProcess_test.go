package preprocess

import (
	"bytes"
	"testing"

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
