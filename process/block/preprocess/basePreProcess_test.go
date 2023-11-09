package preprocess

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/stretchr/testify/assert"
)

func TestBasePreProcess_handleProcessTransactionInit(t *testing.T) {
	t.Parallel()

	txHash := []byte("tx hash")
	initProcessedTxsCalled := false

	preProcessorExecutionInfoHandler := &testscommon.PreProcessorExecutionInfoHandlerMock{
		InitProcessedTxsResultsCalled: func(key []byte) {
			if !bytes.Equal(key, txHash) {
				return
			}

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

	recoveredJournalLen := bp.handleProcessTransactionInit(preProcessorExecutionInfoHandler, txHash)
	assert.Equal(t, journalLen, recoveredJournalLen)
	assert.True(t, initProcessedTxsCalled)
}
