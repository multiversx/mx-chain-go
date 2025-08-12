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
	"github.com/multiversx/mx-chain-go/testscommon/executionTrack"
)

func TestNewExecutionResultsVerifier(t *testing.T) {
	t.Parallel()

	blockchain := &testscommon.ChainHandlerMock{}
	executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
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

func TestExecutionResultsVerifier_getPrevBlockLastExecutionResult(t *testing.T) {
	t.Parallel()

	t.Run("nil current block header", func(t *testing.T) {
		t.Parallel()
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return nil
			},
		}

		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		_, err = erc.getPrevBlockLastExecutionResult()
		require.Equal(t, process.ErrNilHeaderHandler, err)
	})

	t.Run("valid prev header - shard header v3", func(t *testing.T) {
		t.Parallel()

		prevHeader := createShardHeaderV3WithExecutionResults()
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
		}

		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		lastExecutionResult, err := erc.getPrevBlockLastExecutionResult()
		require.NoError(t, err)
		require.NotNil(t, lastExecutionResult)
		require.Equal(t, prevHeader.LastExecutionResult, lastExecutionResult)
	})
	t.Run("valid prev header - shard header v2", func(t *testing.T) {
		t.Parallel()

		prevHeader := createDummyPrevShardHeaderV2()
		prevHeaderHash := []byte("prevHeaderHash")
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return prevHeaderHash
			},
		}

		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		lastExecutionResult, err := erc.getPrevBlockLastExecutionResult()
		require.NoError(t, err)
		require.NotNil(t, lastExecutionResult)
		expectedLastExecutionResult, _ := createLastExecutionResultFromPrevHeader(prevHeader, prevHeaderHash)
		require.Equal(t, expectedLastExecutionResult, lastExecutionResult)
	})
	t.Run("valid prev header - meta header v3", func(t *testing.T) {
		t.Parallel()

		prevHeader := createMetaHeaderV3WithExecutionResults()
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
		}

		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		lastExecutionResult, err := erc.getPrevBlockLastExecutionResult()
		require.NoError(t, err)
		require.NotNil(t, lastExecutionResult)
		require.Equal(t, prevHeader.LastExecutionResult, lastExecutionResult)
	})
	t.Run("valid prev header - meta header", func(t *testing.T) {
		t.Parallel()

		prevHeader := createDummyPrevMetaHeader()
		prevHeaderHash := []byte("prevHeaderHash")
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return prevHeaderHash
			},
		}

		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		lastExecutionResult, err := erc.getPrevBlockLastExecutionResult()
		require.NoError(t, err)
		require.NotNil(t, lastExecutionResult)
		expectedLastExecutionResult, _ := createLastExecutionResultFromPrevHeader(prevHeader, prevHeaderHash)
		require.Equal(t, expectedLastExecutionResult, lastExecutionResult)
	})
}

func TestExecutionResultsVerifier_checkLastExecutionResultInfoAgainstPrevBlock(t *testing.T) {
	t.Parallel()

	t.Run("nil last execution result info", func(t *testing.T) {
		t.Parallel()
		blockchain := &testscommon.ChainHandlerStub{}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.checkLastExecutionResultInfoAgainstPrevBlock(nil)
		require.Equal(t, process.ErrNilLastExecutionResultHandler, err)
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

func createLastExecutionResultShard() *block.ExecutionResultInfo {
	return &block.ExecutionResultInfo{
		NotarizedOnHeaderHash: []byte("notarizedOnHeaderHash"),
		ExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  []byte("headerHash"),
			HeaderNonce: 1,
			HeaderRound: 2,
			RootHash:    []byte("rootHash"),
		},
	}
}

func createLastExecutionResultMeta() *block.MetaExecutionResultInfo {
	return &block.MetaExecutionResultInfo{
		NotarizedOnHeaderHash: []byte("notarizedOnHeaderHash"),
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
	}
}

func createShardHeaderV3WithExecutionResults() *block.HeaderV3 {
	executionResult := createDummyShardExecutionResult()
	lastExecutionResult := createLastExecutionResultShard()
	return &block.HeaderV3{
		Nonce:               2,
		Round:               3,
		ShardID:             0,
		ExecutionResults:    []*block.ExecutionResult{executionResult},
		LastExecutionResult: lastExecutionResult,
	}
}

func createMetaHeaderV3WithExecutionResults() *block.MetaBlockV3 {
	executionResults := createDummyMetaExecutionResult()
	lastExecutionResult := createLastExecutionResultMeta()
	return &block.MetaBlockV3{
		Nonce:               3,
		Round:               4,
		LastExecutionResult: lastExecutionResult,
		ExecutionResults:    []*block.MetaExecutionResult{executionResults},
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
