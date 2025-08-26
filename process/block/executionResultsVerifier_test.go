package block

import (
	"fmt"
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

func TestExecutionResultsVerifier_verifyLastExecutionResultInfoMatchesLastExecutionResult(t *testing.T) {
	t.Parallel()

	t.Run("nil last execution result info", func(t *testing.T) {
		t.Parallel()
		header := createDummyPrevShardHeaderV2()
		blockchain := &testscommon.ChainHandlerStub{}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult([]byte("headerHash"), header)
		require.ErrorIs(t, err, process.ErrNilLastExecutionResultHandler)
	})
	t.Run("prev header get error", func(t *testing.T) {
		t.Parallel()
		header := createShardHeaderV3WithExecutionResults()

		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return nil
			},
		}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult([]byte("headerHash"), header)
		require.Equal(t, process.ErrNilHeaderHandler, err)
	})
	t.Run("no execution results, keeps last execution results", func(t *testing.T) {
		t.Parallel()
		prevHeaderHash := []byte("prevHeaderHash")
		prevHeader := createShardHeaderV3WithExecutionResults()
		header := &block.HeaderV3{
			PrevHash:            prevHeader.GetPrevHash(),
			Nonce:               prevHeader.GetNonce() + 1,
			Round:               prevHeader.GetRound() + 1,
			ShardID:             prevHeader.GetShardID(),
			LastExecutionResult: prevHeader.LastExecutionResult,
		}
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

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult([]byte("headerHash"), header)
		require.NoError(t, err)
	})
	t.Run("last execution result info does not match previous block,  nil execution results", func(t *testing.T) {
		t.Parallel()
		prevHeaderHash := []byte("prevHeaderHash")
		prevHeader := createShardHeaderV3WithExecutionResults()
		lastExecutionResult := createLastExecutionResultShard()
		lastExecutionResult.ExecutionResult.HeaderHash = []byte("differentHeaderHash") // Modify to ensure mismatch
		header := &block.HeaderV3{
			PrevHash:            prevHeaderHash,
			Nonce:               prevHeader.GetNonce() + 1,
			Round:               prevHeader.GetRound() + 1,
			ShardID:             prevHeader.GetShardID(),
			LastExecutionResult: lastExecutionResult,
		}

		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
		}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult([]byte("headerHash"), header)
		require.Equal(t, process.ErrExecutionResultDoesNotMatch, err)
	})
	t.Run("last execution result info (shard) matches previous block, nil execution results", func(t *testing.T) {
		t.Parallel()
		prevHeaderHash := []byte("prevHeaderHash")
		prevHeader := createShardHeaderV3WithExecutionResults()
		lastExecutionResult := createLastExecutionResultShard()
		header := &block.HeaderV3{
			PrevHash:            prevHeaderHash,
			Nonce:               prevHeader.GetNonce() + 1,
			Round:               prevHeader.GetRound() + 1,
			ShardID:             prevHeader.GetShardID(),
			LastExecutionResult: lastExecutionResult,
		}

		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
		}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult([]byte("headerHash"), header)
		require.NoError(t, err)
	})
	t.Run("last execution result info (meta) matches previous block, nil execution results", func(t *testing.T) {
		t.Parallel()

		prevHeaderHash := []byte("prevHeaderHash")
		prevHeader := createMetaHeaderV3WithExecutionResults()
		lastExecutionResult := createLastExecutionResultMeta()
		header := &block.MetaBlockV3{
			PrevHash:            prevHeaderHash,
			Nonce:               prevHeader.GetNonce() + 1,
			Round:               prevHeader.GetRound() + 1,
			LastExecutionResult: lastExecutionResult,
		}

		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return createMetaHeaderV3WithExecutionResults()
			},
		}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult([]byte("headerHash"), header)
		require.NoError(t, err)
	})
	t.Run("create last execution result from execution result error (nil execution result)", func(t *testing.T) {
		t.Parallel()
		header := &block.HeaderV3{
			PrevHash:            []byte("prevHash"),
			Nonce:               1,
			Round:               2,
			ShardID:             0,
			ExecutionResults:    make([]*block.ExecutionResult, 4),
			LastExecutionResult: &block.ExecutionResultInfo{}, // No last execution result info
		}
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return header
			},
		}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult([]byte("headerHash"), header)
		require.Equal(t, process.ErrNilExecutionResultHandler, err)
	})
	t.Run("execution results mismatch", func(t *testing.T) {
		t.Parallel()
		header := createShardHeaderV3WithExecutionResults()
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return header
			},
		}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult([]byte("headerHash"), header)
		require.Equal(t, process.ErrExecutionResultDoesNotMatch, err)
	})
	t.Run("first execution result not consecutive to prevHeader last execution result", func(t *testing.T) {
		t.Parallel()
		prevHeader := createShardHeaderV3WithMultipleExecutionResults(1, 5)
		prevHeader.LastExecutionResult.ExecutionResult.HeaderNonce = big.NewInt(10).Uint64()
		header := createShardHeaderV3WithMultipleExecutionResults(3, 6)
		headerHash := []byte("headerHash")
		header.ExecutionResults[0].BaseExecutionResult.HeaderNonce = 100 // Modify to ensure mismatch

		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return header.GetPrevHash()
			},
		}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult(headerHash, header)
		require.ErrorIs(t, err, process.ErrExecutionResultsNonConsecutive)
	})
	t.Run("valid execution results for shard", func(t *testing.T) {
		t.Parallel()
		prevHeader := createShardHeaderV3WithExecutionResults()
		prevHeader.LastExecutionResult.ExecutionResult.HeaderNonce = prevHeader.LastExecutionResult.ExecutionResult.HeaderNonce - 1
		header := createShardHeaderV3WithExecutionResults()
		headerHash := []byte("headerHash")
		header.ExecutionResults[len(header.ExecutionResults)-1] = createDummyShardExecutionResult()
		lastExecutionResult, err := createLastExecutionResultInfoFromExecutionResult(headerHash, header.ExecutionResults[len(header.ExecutionResults)-1], header.GetShardID())
		require.Nil(t, err)
		header.LastExecutionResult = lastExecutionResult.(*block.ExecutionResultInfo)

		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return header.GetPrevHash()
			},
		}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult(headerHash, header)
		require.NoError(t, err)
	})
}

func TestExecutionResultsVerifier_VerifyHeaderExecutionResults(t *testing.T) {
	t.Parallel()

	t.Run("nil header hash", func(t *testing.T) {
		t.Parallel()
		header := createShardHeaderV3WithExecutionResults()
		blockchain := &testscommon.ChainHandlerStub{}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults(nil, header)
		require.Equal(t, process.ErrInvalidHash, err)
	})

	t.Run("nil header", func(t *testing.T) {
		t.Parallel()
		blockchain := &testscommon.ChainHandlerStub{}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults([]byte("headerHash"), nil)
		require.Equal(t, process.ErrNilHeaderHandler, err)
	})

	t.Run("no execution results", func(t *testing.T) {
		t.Parallel()
		header := &block.HeaderV3{
			PrevHash:            []byte("prevHash"),
			Nonce:               1,
			Round:               2,
			ShardID:             0,
			LastExecutionResult: createLastExecutionResultShard(),
			ExecutionResults:    nil,
		}
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return header
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("headerHash")
			},
		}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults([]byte("headerHash"), header)
		require.NoError(t, err)
	})
	t.Run("invalid header type", func(t *testing.T) {
		t.Parallel()
		header := &block.Header{
			Nonce:    1,
			Round:    2,
			RootHash: []byte("rootHash"),
			ShardID:  0,
		}
		blockchain := &testscommon.ChainHandlerStub{}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults([]byte("headerHash"), header)
		require.Equal(t, process.ErrInvalidHeader, err)
	})
	t.Run("always valid for genesis header (nonce < 0)", func(t *testing.T) {
		t.Parallel()
		header := &block.HeaderV3{
			PrevHash:            []byte("prevHash"),
			Nonce:               0,
			Round:               0,
			ShardID:             0,
			LastExecutionResult: createLastExecutionResultShard(),
			ExecutionResults:    nil,
		}
		blockchain := &testscommon.ChainHandlerStub{}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults([]byte("headerHash"), header)
		require.NoError(t, err)
	})
	t.Run("header with nil last execution results", func(t *testing.T) {
		t.Parallel()
		header := &block.HeaderV3{
			PrevHash:            []byte("prevHash"),
			Nonce:               1,
			Round:               2,
			ShardID:             0,
			LastExecutionResult: nil, // No last execution result info
			ExecutionResults:    nil,
		}
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return header
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("headerHash")
			},
		}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults([]byte("headerHash"), header)
		require.ErrorIs(t, err, process.ErrNilLastExecutionResultHandler)
	})
	t.Run("error getting pending execution results", func(t *testing.T) {
		t.Parallel()

		prevHeader := createPrevShardHeaderV3WithoutExecutionResults()
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
		}

		header := createShardHeaderV3WithExecutionResults()
		lastExecutionResult, err := createLastExecutionResultInfoFromExecutionResult(header.ExecutionResults[len(header.ExecutionResults)-1].GetHeaderHash(), header.ExecutionResults[len(header.ExecutionResults)-1], header.GetShardID())
		require.NoError(t, err)
		header.LastExecutionResult = lastExecutionResult.(*block.ExecutionResultInfo)

		expectedErr := fmt.Errorf("error getting pending execution results")
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{
			GetPendingExecutionResultsCalled: func() ([]data.ExecutionResultHandler, error) {
				return nil, expectedErr
			},
		}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults([]byte("headerHash"), header)
		require.Equal(t, expectedErr, err)
	})
	t.Run("execution results number mismatch", func(t *testing.T) {
		t.Parallel()
		prevHeader := createPrevShardHeaderV3WithoutExecutionResults()
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("prevHeaderHash")
			},
		}

		header := createShardHeaderV3WithExecutionResults()
		lastExecutionResult, err := createLastExecutionResultInfoFromExecutionResult(header.ExecutionResults[len(header.ExecutionResults)-1].GetHeaderHash(), header.ExecutionResults[len(header.ExecutionResults)-1], header.GetShardID())
		require.NoError(t, err)
		header.LastExecutionResult = lastExecutionResult.(*block.ExecutionResultInfo)
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{
			GetPendingExecutionResultsCalled: func() ([]data.ExecutionResultHandler, error) {
				return nil, nil // less than expected number of execution results
			},
		}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults([]byte("headerHash"), header)
		require.Equal(t, process.ErrExecutionResultsNumberMismatch, err)
	})
	t.Run("execution results mismatch", func(t *testing.T) {
		t.Parallel()

		prevHeader := createPrevShardHeaderV3WithoutExecutionResults()
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("prevHeaderHash")
			},
		}

		header := createShardHeaderV3WithExecutionResults()
		lastExecutionResult, err := createLastExecutionResultInfoFromExecutionResult(header.ExecutionResults[len(header.ExecutionResults)-1].GetHeaderHash(), header.ExecutionResults[len(header.ExecutionResults)-1], header.GetShardID())
		require.NoError(t, err)
		header.LastExecutionResult = lastExecutionResult.(*block.ExecutionResultInfo)

		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{
			GetPendingExecutionResultsCalled: func() ([]data.ExecutionResultHandler, error) {
				result := make([]data.ExecutionResultHandler, len(header.ExecutionResults))
				for i := range header.ExecutionResults {
					differentResult := *header.ExecutionResults[i]
					differentResult.GasUsed = 100000 // Modify to ensure mismatch
					result[i] = &differentResult
				}
				return result, nil
			},
		}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults([]byte("headerHash"), header)
		require.Equal(t, process.ErrExecutionResultDoesNotMatch, err)
	})
	t.Run("execution results not consecutive", func(t *testing.T) {
		t.Parallel()

		prevHeader := createShardHeaderV3WithMultipleExecutionResults(1, 1)
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("prevHeaderHash")
			},
		}

		header := createShardHeaderV3WithMultipleExecutionResults(3, 2)
		// remove middle execution results, so they become not consecutive
		header.ExecutionResults = []*block.ExecutionResult{header.ExecutionResults[0], header.ExecutionResults[2]}
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{
			GetPendingExecutionResultsCalled: func() ([]data.ExecutionResultHandler, error) {
				result := make([]data.ExecutionResultHandler, len(header.ExecutionResults))
				for i := range header.ExecutionResults {
					result[i] = header.ExecutionResults[i]
				}

				return result, nil
			},
		}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults([]byte("headerHash"), header)
		require.Equal(t, process.ErrExecutionResultsNonConsecutive, err)
	})
	t.Run("execution results consecutive", func(t *testing.T) {
		t.Parallel()

		prevHeader := createShardHeaderV3WithMultipleExecutionResults(1, 1)
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("prevHeaderHash")
			},
		}

		header := createShardHeaderV3WithMultipleExecutionResults(3, 2)
		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{
			GetPendingExecutionResultsCalled: func() ([]data.ExecutionResultHandler, error) {
				result := make([]data.ExecutionResultHandler, len(header.ExecutionResults))
				for i := range header.ExecutionResults {
					result[i] = header.ExecutionResults[i]
				}

				return result, nil
			},
		}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults([]byte("headerHash"), header)
		require.NoError(t, err)
	})
	t.Run("valid execution results for shard", func(t *testing.T) {
		t.Parallel()
		prevHeader := createPrevShardHeaderV3WithoutExecutionResults()
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("prevHeaderHash")
			},
		}

		header := createShardHeaderV3WithExecutionResults()
		lastExecutionResult, err := createLastExecutionResultInfoFromExecutionResult(header.ExecutionResults[len(header.ExecutionResults)-1].GetHeaderHash(), header.ExecutionResults[len(header.ExecutionResults)-1], header.GetShardID())
		require.NoError(t, err)
		header.LastExecutionResult = lastExecutionResult.(*block.ExecutionResultInfo)

		executionResultsTracker := &executionTrack.ExecutionResultsTrackerStub{
			GetPendingExecutionResultsCalled: func() ([]data.ExecutionResultHandler, error) {
				result := make([]data.ExecutionResultHandler, len(header.ExecutionResults))
				for i := range header.ExecutionResults {
					result[i] = header.ExecutionResults[i]
				}

				return result, nil
			},
		}
		erc, err := NewExecutionResultsVerifier(blockchain, executionResultsTracker)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults([]byte("headerHash"), header)
		require.NoError(t, err)
	})
}

func createDummyShardExecutionResults(numResults int, firstNonce uint64) []*block.ExecutionResult {
	results := make([]*block.ExecutionResult, numResults)
	for i := 0; i < numResults; i++ {
		results[i] = &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("headerHash"),
				HeaderNonce: firstNonce + uint64(i),
				HeaderRound: firstNonce + uint64(i) + 1,
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
	return results
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

func createShardHeaderV3WithMultipleExecutionResults(numResults int, firstNonce uint64) *block.HeaderV3 {
	executionResults := createDummyShardExecutionResults(numResults, firstNonce)
	lastExecResult, _ := createLastExecutionResultInfoFromExecutionResult([]byte("headerHash"), executionResults[len(executionResults)-1], 0)

	return &block.HeaderV3{
		PrevHash:            []byte("prevHash"),
		Nonce:               2 + uint64(numResults) + firstNonce,
		Round:               3 + uint64(numResults) + firstNonce,
		ShardID:             0,
		ExecutionResults:    executionResults,
		LastExecutionResult: lastExecResult.(*block.ExecutionResultInfo),
	}
}

func createShardHeaderV3WithExecutionResults() *block.HeaderV3 {
	executionResult := createDummyShardExecutionResult()
	lastExecutionResult := createLastExecutionResultShard()
	return &block.HeaderV3{
		PrevHash:            []byte("prevHash"),
		Nonce:               2,
		Round:               3,
		ShardID:             0,
		ExecutionResults:    []*block.ExecutionResult{executionResult},
		LastExecutionResult: lastExecutionResult,
	}
}

func createPrevShardHeaderV3WithoutExecutionResults() *block.HeaderV3 {
	lastExecutionResults := createLastExecutionResultShard()
	lastExecutionResults.ExecutionResult.HeaderNonce = 0
	lastExecutionResults.ExecutionResult.HeaderRound = 1

	return &block.HeaderV3{
		PrevHash:            []byte("prevHash"),
		Nonce:               1,
		Round:               2,
		ShardID:             0,
		ExecutionResults:    nil,
		LastExecutionResult: lastExecutionResults,
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
