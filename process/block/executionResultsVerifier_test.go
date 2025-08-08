package block

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	executionTrack2 "github.com/multiversx/mx-chain-go/testscommon/executionTrack"
)

func TestNewExecutionResultsVerifier(t *testing.T) {
	t.Parallel()

	blockchain := &testscommon.ChainHandlerMock{}
	executionResultsTracker := &executionTrack2.ExecutionResultsTrackerStub{}
	t.Run("nil blockchain", func(t *testing.T) {
		t.Parallel()
		erv, err := NewExecutionResultsVerifier(nil, nil)
		require.Equal(t, process.ErrNilBlockChain, err)
		require.Nil(t, erv)
	})
	t.Run("nil execution results tracker", func(t *testing.T) {
		t.Parallel()
		erv, err := NewExecutionResultsVerifier(blockchain, nil)
		require.Equal(t, process.ErrNilExecutionResultsTracker, err)
		require.Nil(t, erv)
	})
	t.Run("valid parameters", func(t *testing.T) {
		t.Parallel()
		erv, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)
		require.NotNil(t, erv)
		require.Equal(t, blockchain, erv.blockChain)
		require.Equal(t, executionResultsTracker, erv.executionResultsTracker)
	})
}

func Test_createLastExecutionResultInfoFromExecutionResults(t *testing.T) {
	t.Parallel()
	lastExecResult := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("headerHash"),
			HeaderNonce: 1,
		},
		ReceiptsHash:     []byte("receiptsHash"),
		MiniBlockHeaders: nil,
		DeveloperFees:    big.NewInt(100),
		AccumulatedFees:  big.NewInt(200),
		GasUsed:          1000,
		ExecutedTxCount:  2,
	}
	notarizedOnHeaderHash := []byte("notarizedOnHeaderHash")

	t.Run("nil notarized on header hash", func(t *testing.T) {
		t.Parallel()

		result, err := createLastExecutionResultInfoFromExecutionResult(nil, lastExecResult)
		require.Nil(t, result)
		require.Equal(t, process.ErrNilNotarizedOnHeaderHash, err)
	})

	t.Run("nil last execution result", func(t *testing.T) {
		t.Parallel()

		result, err := createLastExecutionResultInfoFromExecutionResult(notarizedOnHeaderHash, nil)
		require.Nil(t, result)
		require.Equal(t, process.ErrNilExecutionResultHandler, err)
	})
	t.Run("valid data", func(t *testing.T) {
		t.Parallel()

		result, err := createLastExecutionResultInfoFromExecutionResult(notarizedOnHeaderHash, lastExecResult)
		require.NoError(t, err)
		require.NotNil(t, result)

		require.Equal(t, notarizedOnHeaderHash, result.GetNotarizedOnHeaderHash())
		require.Equal(t, lastExecResult.GetHeaderHash(), result.ExecutionResult.GetHeaderHash())
		require.Equal(t, lastExecResult.GetHeaderNonce(), result.ExecutionResult.GetHeaderNonce())
		require.Equal(t, lastExecResult.GetRootHash(), result.ExecutionResult.GetRootHash())
		require.Equal(t, lastExecResult.GetHeaderRound(), result.ExecutionResult.GetHeaderRound())
	})
}
