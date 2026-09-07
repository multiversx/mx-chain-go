package preprocess

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/state"
)

func TestBasePreProcessGetMaxGasLimitUsedForDestMeTxs(t *testing.T) {
	t.Parallel()

	const (
		supernovaEpoch  = uint32(2)
		supernovaRound  = uint64(260)
		currentGasLimit = uint64(600_000_000)
		legacyGasLimit  = uint64(1_500_000_000)
	)

	economicsFee := &economicsmocks.EconomicsHandlerMock{
		MaxGasLimitPerBlockCalled: func(_ uint32) uint64 {
			return currentGasLimit
		},
		MaxGasLimitPerBlockInEpochCalled: func(_ uint32, _ uint32) uint64 {
			return legacyGasLimit
		},
	}
	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
			return flag == common.SupernovaFlag && epoch >= supernovaEpoch
		},
		GetActivationEpochCalled: func(flag core.EnableEpochFlag) uint32 {
			return supernovaEpoch
		},
	}
	enableRoundsHandler := &testscommon.EnableRoundsHandlerStub{
		IsFlagEnabledInRoundCalled: func(flag common.EnableRoundFlag, round uint64) bool {
			return flag == common.SupernovaRoundFlag && round >= supernovaRound
		},
	}
	policy, err := process.ResolveGasProcessingPolicy(
		&testscommon.HeaderHandlerStub{EpochField: supernovaEpoch, RoundField: supernovaRound - 1},
		enableEpochsHandler,
		enableRoundsHandler,
		economicsFee,
		0,
	)
	require.NoError(t, err)

	bp := &basePreProcess{
		gasTracker: gasTracker{
			shardCoordinator: &testscommon.ShardsCoordinatorMock{},
			economicsFee:     economicsFee,
		},
	}

	require.Equal(t, currentGasLimit, bp.getMaxGasLimitUsedForDestMeTxs(0, process.GasProcessingPolicy{}))
	require.Equal(t, currentGasLimit/2, bp.getMaxGasLimitUsedForDestMeTxs(1, process.GasProcessingPolicy{}))
	require.Equal(t, legacyGasLimit, bp.getMaxGasLimitUsedForDestMeTxs(0, policy))
	require.Equal(t, legacyGasLimit, bp.getMaxGasLimitUsedForDestMeTxs(1, policy))
}

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
