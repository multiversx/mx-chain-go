package block

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
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

func Test_createLastExecutionResultInfoFromExecutionResult(t *testing.T) {
	t.Parallel()

	t.Run("nil notarizedOnHeaderHash", func(t *testing.T) {
		t.Parallel()
		lastExecutionResultInfo, err := createLastExecutionResultInfoFromExecutionResult(nil, nil, 0)
		require.Equal(t, process.ErrNilNotarizedOnHeaderHash, err)
		require.Nil(t, lastExecutionResultInfo)
	})
	t.Run("nil executionResult", func(t *testing.T) {
		t.Parallel()
		notarizedOnHash := []byte("notarizedOnHash")
		lastExecutionResultInfo, err := createLastExecutionResultInfoFromExecutionResult(notarizedOnHash, nil, 0)
		require.Equal(t, process.ErrNilExecutionResultHandler, err)
		require.Nil(t, lastExecutionResultInfo)
	})
	t.Run("invalid shard executionResult type", func(t *testing.T) {
		t.Parallel()
		notarizedOnHash := []byte("notarizedOnHash")
		executionResult := createDummyMetaExecutionResult() // This is a valid type, but we will use it incorrectly
		lastExecutionResultInfo, err := createLastExecutionResultInfoFromExecutionResult(notarizedOnHash, executionResult, 0)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
		require.Nil(t, lastExecutionResultInfo)
	})
	t.Run("valid executionResult for shard", func(t *testing.T) {
		t.Parallel()
		notarizedOnHash := []byte("notarizedOnHash")
		executionResult := createDummyShardExecutionResult()
		expectedLastExecutionResult := &block.ExecutionResultInfo{
			NotarizedOnHeaderHash: notarizedOnHash,
			ExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  executionResult.GetHeaderHash(),
				HeaderNonce: executionResult.GetHeaderNonce(),
				HeaderRound: executionResult.GetHeaderRound(),
				RootHash:    executionResult.GetRootHash(),
			},
		}

		lastExecutionResultHandler, err := createLastExecutionResultInfoFromExecutionResult(notarizedOnHash, executionResult, 0)
		lastExecutionResultInfo := lastExecutionResultHandler.(*block.ExecutionResultInfo)
		require.NoError(t, err)
		require.NotNil(t, lastExecutionResultInfo)
		require.Equal(t, expectedLastExecutionResult, lastExecutionResultInfo)
	})
	t.Run("invalid metaChain executionResult", func(t *testing.T) {
		t.Parallel()
		notarizedOnHash := []byte("notarizedOnHash")
		executionResult := createDummyShardExecutionResult()
		lastExecutionResultHandler, err := createLastExecutionResultInfoFromExecutionResult(notarizedOnHash, executionResult, core.MetachainShardId)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
		require.Nil(t, lastExecutionResultHandler)
	})
	t.Run("valid executionResult for metaChain", func(t *testing.T) {
		t.Parallel()
		notarizedOnHash := []byte("notarizedOnHash")
		executionResult := createDummyMetaExecutionResult()
		expectedLastExecutionResult := &block.MetaExecutionResultInfo{
			NotarizedOnHeaderHash: notarizedOnHash,
			ExecutionResult: &block.BaseMetaExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash:  executionResult.GetHeaderHash(),
					HeaderNonce: executionResult.GetHeaderNonce(),
					HeaderRound: executionResult.GetHeaderRound(),
					RootHash:    executionResult.GetRootHash(),
				},
				ValidatorStatsRootHash: executionResult.GetValidatorStatsRootHash(),
				AccumulatedFeesInEpoch: executionResult.GetAccumulatedFeesInEpoch(),
				DevFeesInEpoch:         executionResult.GetDevFeesInEpoch(),
			},
		}

		lastExecutionResultHandler, err := createLastExecutionResultInfoFromExecutionResult(notarizedOnHash, executionResult, core.MetachainShardId)
		lastExecutionResultInfo, ok := lastExecutionResultHandler.(*block.MetaExecutionResultInfo)
		require.True(t, ok)
		require.NoError(t, err)
		require.NotNil(t, lastExecutionResultInfo)
		require.Equal(t, expectedLastExecutionResult, lastExecutionResultInfo)
	})
}

func Test_createLastExecutionResultFromPrevHeader(t *testing.T) {
	t.Parallel()

	t.Run("nil prevHeader", func(t *testing.T) {
		t.Parallel()
		lastExecutionResult, err := createLastExecutionResultFromPrevHeader(nil, []byte("prevHeaderHash"))
		require.Equal(t, process.ErrNilBlockHeader, err)
		require.Nil(t, lastExecutionResult)
	})
	t.Run("nil prevHeaderHash", func(t *testing.T) {
		t.Parallel()
		prevHeader := createDummyPrevShardHeaderV2()
		lastExecutionResult, err := createLastExecutionResultFromPrevHeader(prevHeader, nil)
		require.Equal(t, process.ErrInvalidHash, err)
		require.Nil(t, lastExecutionResult)
	})
	t.Run("invalid shard prevHeader type", func(t *testing.T) {
		t.Parallel()
		prevHeaderHash := []byte("prevHeaderHash")
		// the expected header type is HeaderV2
		prevHeader := &block.Header{
			Nonce:    1,
			Round:    2,
			RootHash: []byte("prevRootHash"),
			ShardID:  0,
		}
		lastExecutionResult, err := createLastExecutionResultFromPrevHeader(prevHeader, prevHeaderHash)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
		require.Nil(t, lastExecutionResult)
	})
	t.Run("valid shard prevHeader type", func(t *testing.T) {
		t.Parallel()
		prevHeaderHash := []byte("prevHeaderHash")
		prevHeader := createDummyPrevShardHeaderV2()
		expectedLastExecutionResult := &block.ExecutionResultInfo{
			NotarizedOnHeaderHash: prevHeaderHash,
			ExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  prevHeaderHash,
				HeaderNonce: prevHeader.GetNonce(),
				HeaderRound: prevHeader.GetRound(),
				RootHash:    prevHeader.GetRootHash(),
			},
		}

		lastExecutionResultHandler, err := createLastExecutionResultFromPrevHeader(prevHeader, prevHeaderHash)
		lastExecutionResultInfo := lastExecutionResultHandler.(*block.ExecutionResultInfo)
		require.NoError(t, err)
		require.NotNil(t, lastExecutionResultInfo)
		require.Equal(t, expectedLastExecutionResult, lastExecutionResultInfo)
	})
	t.Run("invalid metaChain prevHeader type", func(t *testing.T) {
		t.Parallel()
		prevHeaderHash := []byte("prevHeaderHash")
		prevMetaHeader := createDummyInvalidMetaHeader()
		lastExecutionResultHandler, err := createLastExecutionResultFromPrevHeader(prevMetaHeader, prevHeaderHash)
		require.Equal(t, process.ErrWrongTypeAssertion, err)
		require.Nil(t, lastExecutionResultHandler)
	})
	t.Run("valid metaChain prevHeader type", func(t *testing.T) {
		t.Parallel()
		prevHeaderHash := []byte("prevHeaderHash")
		prevMetaHeader := createDummyPrevMetaHeader()
		expectedLastExecutionResult := &block.MetaExecutionResultInfo{
			NotarizedOnHeaderHash: prevHeaderHash,
			ExecutionResult: &block.BaseMetaExecutionResult{
				BaseExecutionResult: &block.BaseExecutionResult{
					HeaderHash:  prevHeaderHash,
					HeaderNonce: prevMetaHeader.GetNonce(),
					HeaderRound: prevMetaHeader.GetRound(),
					RootHash:    prevMetaHeader.GetRootHash(),
				},
				ValidatorStatsRootHash: prevMetaHeader.GetValidatorStatsRootHash(),
				AccumulatedFeesInEpoch: prevMetaHeader.GetAccumulatedFeesInEpoch(),
				DevFeesInEpoch:         prevMetaHeader.GetDevFeesInEpoch(),
			},
		}

		lastExecutionResultHandler, err := createLastExecutionResultFromPrevHeader(prevMetaHeader, prevHeaderHash)
		lastExecutionResultInfo, ok := lastExecutionResultHandler.(*block.MetaExecutionResultInfo)
		require.True(t, ok)
		require.NoError(t, err)
		require.NotNil(t, lastExecutionResultInfo)
		require.Equal(t, expectedLastExecutionResult, lastExecutionResultInfo)
	})
}

func createDummyShardExecutionResult() *block.ExecutionResult {
	return &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("headerHash"),
			HeaderNonce: 1,
			HeaderRound: 2,
			RootHash:    []byte("rootHash"),
		},
		ReceiptsHash:     []byte("receiptsHash"),
		MiniBlockHeaders: nil,
		DeveloperFees:    big.NewInt(100),
		AccumulatedFees:  big.NewInt(200),
		GasUsed:          10000000,
		ExecutedTxCount:  100,
	}
}

func createDummyMetaExecutionResult() *block.MetaExecutionResult {
	return &block.MetaExecutionResult{
		ExecutionResult: &block.BaseMetaExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("headerHash"),
				HeaderNonce: 1,
				HeaderRound: 2,
				RootHash:    []byte("rootHash"),
			},
			ValidatorStatsRootHash: []byte("validatorStatsRootHash"),
			AccumulatedFeesInEpoch: big.NewInt(300),
			DevFeesInEpoch:         big.NewInt(400),
		},
		ReceiptsHash:    []byte("receiptsHash"),
		DeveloperFees:   big.NewInt(500),
		AccumulatedFees: big.NewInt(600),
		GasUsed:         20000000,
		ExecutedTxCount: 200,
	}
}

func createDummyPrevShardHeaderV2() *block.HeaderV2 {
	return &block.HeaderV2{
		Header: &block.Header{
			Nonce:    1,
			Round:    2,
			RootHash: []byte("prevRootHash"),
		},
	}
}
func createDummyInvalidMetaHeader() data.HeaderHandler {
	return &block.HeaderV2{
		Header: &block.Header{
			Nonce:    1,
			Round:    2,
			RootHash: []byte("prevRootHash"),
			ShardID:  core.MetachainShardId,
		},
	}
}

func createDummyPrevMetaHeader() *block.MetaBlock {
	return &block.MetaBlock{
		Nonce:                  4,
		Round:                  5,
		RootHash:               []byte("prevRootHash"),
		ValidatorStatsRootHash: []byte("prevValidatorStatsRootHash"),
		DevFeesInEpoch:         big.NewInt(300),
		AccumulatedFeesInEpoch: big.NewInt(400),
	}
}
