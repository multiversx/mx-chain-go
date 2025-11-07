package block

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func TestNewExecutionResultsVerifier(t *testing.T) {
	t.Parallel()

	blockchain := &testscommon.ChainHandlerMock{}
	executionManager := &processMocks.ExecutionManagerMock{}
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
		erv, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)
		require.NotNil(t, erv)
		require.Equal(t, blockchain, erv.blockChain)
		require.Equal(t, executionManager, erv.executionManager)
	})
}

func TestExecutionResultsVerifier_verifyLastExecutionResultInfoMatchesLastExecutionResult(t *testing.T) {
	t.Parallel()

	notarizedInRound := uint64(3)
	t.Run("nil last execution result info", func(t *testing.T) {
		t.Parallel()
		header := createDummyPrevShardHeaderV2()
		blockchain := &testscommon.ChainHandlerStub{}
		executionManager := &processMocks.ExecutionManagerMock{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult(header)
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
		executionManager := &processMocks.ExecutionManagerMock{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult(header)
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
		executionManager := &processMocks.ExecutionManagerMock{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult(header)
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
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return prevHeaderHash
			},
		}
		executionManager := &processMocks.ExecutionManagerMock{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult(header)
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
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return prevHeaderHash
			},
		}
		executionManager := &processMocks.ExecutionManagerMock{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult(header)
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
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return prevHeaderHash
			},
		}
		executionManager := &processMocks.ExecutionManagerMock{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult(header)
		require.NoError(t, err)
	})
	t.Run("create last execution result from execution result error (nil execution result)", func(t *testing.T) {
		t.Parallel()
		prevHeaderHash := []byte("prevHash")
		header := &block.HeaderV3{
			PrevHash:            prevHeaderHash,
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
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return prevHeaderHash
			},
		}
		executionManager := &processMocks.ExecutionManagerMock{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult(header)
		require.Equal(t, process.ErrNilExecutionResultHandler, err)
	})
	t.Run("execution results mismatch", func(t *testing.T) {
		t.Parallel()
		header := createShardHeaderV3WithExecutionResults()
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return header
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return []byte("headerHash")
			},
		}
		executionManager := &processMocks.ExecutionManagerMock{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult(header)
		require.Equal(t, process.ErrExecutionResultDoesNotMatch, err)
	})
	t.Run("first execution result not consecutive to prevHeader last execution result", func(t *testing.T) {
		t.Parallel()
		prevHeader := createShardHeaderV3WithMultipleExecutionResults(1, 5)
		prevHeader.LastExecutionResult.ExecutionResult.HeaderNonce = big.NewInt(10).Uint64()
		header := createShardHeaderV3WithMultipleExecutionResults(3, 6)
		header.ExecutionResults[0].BaseExecutionResult.HeaderNonce = 100 // Modify to ensure mismatch

		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return header.GetPrevHash()
			},
		}
		executionManager := &processMocks.ExecutionManagerMock{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult(header)
		require.ErrorIs(t, err, process.ErrExecutionResultsNonConsecutive)
	})
	t.Run("valid execution results for shard", func(t *testing.T) {
		t.Parallel()
		prevHeader := createShardHeaderV3WithExecutionResults()
		prevHeader.LastExecutionResult.ExecutionResult.HeaderNonce = prevHeader.LastExecutionResult.ExecutionResult.HeaderNonce - 1
		header := createShardHeaderV3WithExecutionResults()
		header.ExecutionResults[len(header.ExecutionResults)-1] = createDummyShardExecutionResult()
		lastExecutionResult, err := process.CreateLastExecutionResultInfoFromExecutionResult(notarizedInRound, header.ExecutionResults[len(header.ExecutionResults)-1], header.GetShardID())
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
		executionManager := &processMocks.ExecutionManagerMock{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.verifyLastExecutionResultInfoMatchesLastExecutionResult(header)
		require.NoError(t, err)
	})
}

func TestExecutionResultsVerifier_VerifyHeaderExecutionResults(t *testing.T) {
	t.Parallel()

	t.Run("nil header", func(t *testing.T) {
		t.Parallel()
		blockchain := &testscommon.ChainHandlerStub{}
		executionManager := &processMocks.ExecutionManagerMock{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults(nil)
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
		executionManager := &processMocks.ExecutionManagerMock{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults(header)
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
		executionManager := &processMocks.ExecutionManagerMock{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults(header)
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
		executionManager := &processMocks.ExecutionManagerMock{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults(header)
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
		executionManager := &processMocks.ExecutionManagerMock{}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults(header)
		require.ErrorIs(t, err, process.ErrNilLastExecutionResultHandler)
	})
	t.Run("error getting pending execution results", func(t *testing.T) {
		t.Parallel()

		header := createShardHeaderV3WithExecutionResults()
		prevHeader := createPrevShardHeaderV3WithoutExecutionResults()
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return header.GetPrevHash()
			},
		}

		lastExecutionResult, err := process.CreateLastExecutionResultInfoFromExecutionResult(header.GetRound(), header.ExecutionResults[len(header.ExecutionResults)-1], header.GetShardID())
		require.NoError(t, err)
		header.LastExecutionResult = lastExecutionResult.(*block.ExecutionResultInfo)

		expectedErr := fmt.Errorf("error getting pending execution results")
		executionManager := &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return nil, expectedErr
			},
		}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults(header)
		require.Equal(t, expectedErr, err)
	})
	t.Run("execution results number mismatch", func(t *testing.T) {
		t.Parallel()

		header := createShardHeaderV3WithExecutionResults()
		prevHeader := createPrevShardHeaderV3WithoutExecutionResults()
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return header.GetPrevHash()
			},
		}

		lastExecutionResult, err := process.CreateLastExecutionResultInfoFromExecutionResult(header.GetRound(), header.ExecutionResults[len(header.ExecutionResults)-1], header.GetShardID())
		require.NoError(t, err)
		header.LastExecutionResult = lastExecutionResult.(*block.ExecutionResultInfo)
		executionManager := &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				return nil, nil // less than expected number of execution results
			},
		}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults(header)
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
		lastExecutionResult, err := process.CreateLastExecutionResultInfoFromExecutionResult(header.GetRound(), header.ExecutionResults[len(header.ExecutionResults)-1], header.GetShardID())
		require.NoError(t, err)
		header.LastExecutionResult = lastExecutionResult.(*block.ExecutionResultInfo)

		executionManager := &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				result := make([]data.BaseExecutionResultHandler, len(header.ExecutionResults))
				for i := range header.ExecutionResults {
					differentResult := *header.ExecutionResults[i]
					differentBase := *differentResult.BaseExecutionResult
					differentBase.GasUsed = 100000 // Modify to ensure mismatch
					differentResult.BaseExecutionResult = &differentBase
					result[i] = &differentResult
				}
				return result, nil
			},
		}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults(header)
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
		executionManager := &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				result := make([]data.BaseExecutionResultHandler, len(header.ExecutionResults))
				for i := range header.ExecutionResults {
					result[i] = header.ExecutionResults[i]
				}

				return result, nil
			},
		}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults(header)
		require.Equal(t, process.ErrExecutionResultsNonConsecutive, err)
	})
	t.Run("execution results consecutive", func(t *testing.T) {
		t.Parallel()

		header := createShardHeaderV3WithMultipleExecutionResults(3, 2)
		prevHeader := createShardHeaderV3WithMultipleExecutionResults(1, 1)
		blockchain := &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return prevHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return header.GetPrevHash()
			},
		}

		executionManager := &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				result := make([]data.BaseExecutionResultHandler, len(header.ExecutionResults))
				for i := range header.ExecutionResults {
					result[i] = header.ExecutionResults[i]
				}

				return result, nil
			},
		}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults(header)
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
		lastExecutionResult, err := process.CreateLastExecutionResultInfoFromExecutionResult(header.GetRound(), header.ExecutionResults[len(header.ExecutionResults)-1], header.GetShardID())
		require.NoError(t, err)
		header.LastExecutionResult = lastExecutionResult.(*block.ExecutionResultInfo)

		executionManager := &processMocks.ExecutionManagerMock{
			GetPendingExecutionResultsCalled: func() ([]data.BaseExecutionResultHandler, error) {
				result := make([]data.BaseExecutionResultHandler, len(header.ExecutionResults))
				for i := range header.ExecutionResults {
					result[i] = header.ExecutionResults[i]
				}

				return result, nil
			},
		}
		erc, err := NewExecutionResultsVerifier(blockchain, executionManager)
		require.NoError(t, err)

		err = erc.VerifyHeaderExecutionResults(header)
		require.NoError(t, err)
	})
}

func TestExecutionResultsVerifier_IsInterfaceNil(t *testing.T) {
	t.Parallel()
	var erc ExecutionResultsVerifier
	require.True(t, check.IfNil(erc))

	executionManager := &processMocks.ExecutionManagerMock{}
	blockchain := &testscommon.ChainHandlerStub{}
	erc, err := NewExecutionResultsVerifier(blockchain, executionManager)

	require.Nil(t, err)
	require.False(t, check.IfNil(erc))
}

func createDummyShardExecutionResults(numResults int, firstNonce uint64) []*block.ExecutionResult {
	results := make([]*block.ExecutionResult, numResults)
	for i := 0; i < numResults; i++ {
		results[i] = &block.ExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  []byte("headerHash"),
				HeaderNonce: firstNonce + uint64(i),
				HeaderRound: firstNonce + uint64(i) + 1,
				GasUsed:     10000000,
				RootHash:    []byte("rootHash"),
			},
			ReceiptsHash:     []byte("receiptsHash"),
			MiniBlockHeaders: nil,
			DeveloperFees:    big.NewInt(100),
			AccumulatedFees:  big.NewInt(200),
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
			GasUsed:     10000000,
			RootHash:    []byte("rootHash"),
		},
		ReceiptsHash:     []byte("receiptsHash"),
		MiniBlockHeaders: nil,
		DeveloperFees:    big.NewInt(100),
		AccumulatedFees:  big.NewInt(200),
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
				GasUsed:     20000000,
				RootHash:    []byte("rootHash"),
			},
			ValidatorStatsRootHash: []byte("validatorStatsRootHash"),
			AccumulatedFeesInEpoch: big.NewInt(300),
			DevFeesInEpoch:         big.NewInt(400),
		},
		ReceiptsHash:    []byte("receiptsHash"),
		DeveloperFees:   big.NewInt(500),
		AccumulatedFees: big.NewInt(600),
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

func createLastExecutionResultShard() *block.ExecutionResultInfo {
	return &block.ExecutionResultInfo{
		NotarizedInRound: 3,
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
		NotarizedInRound: 3,
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
	lastExecResult, _ := process.CreateLastExecutionResultInfoFromExecutionResult(3+uint64(numResults)+firstNonce, executionResults[len(executionResults)-1], 0)

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
