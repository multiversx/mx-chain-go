package disabled

import (
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/stretchr/testify/require"
)

func TestBalanceComputationHandler(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	handler := &BalanceComputationHandler{}
	handler.Init()
	handler.SetBalanceToAddress([]byte("addr"), big.NewInt(0))
	require.True(t, handler.AddBalanceToAddress([]byte("addr"), big.NewInt(0)))
	require.True(t, handler.SubBalanceFromAddress([]byte("addr"), big.NewInt(0)))
	require.True(t, handler.IsAddressSet([]byte("addr")))
	require.True(t, handler.AddressHasEnoughBalance([]byte("addr"), big.NewInt(0)))
	require.False(t, handler.IsInterfaceNil())
}

func TestBlockSizeComputationHandler(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	handler := &BlockSizeComputationHandler{}
	handler.Init()
	handler.AddNumMiniBlocks(0)
	handler.AddNumTxs(0)
	require.False(t, handler.IsMaxBlockSizeReached(0, 1))
	require.False(t, handler.IsMaxBlockSizeWithoutThrottleReached(0, 1))
	require.False(t, handler.IsInterfaceNil())
}

func TestBlockTracker(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	handler := &BlockTracker{}
	require.False(t, handler.IsShardStuck(0))
	require.False(t, handler.ShouldSkipMiniBlocksCreationFromSelf())
	require.False(t, handler.IsInterfaceNil())
}

func TestESDTGlobalSettingsHandler(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	handler := &ESDTGlobalSettingsHandler{}
	require.False(t, handler.IsPaused([]byte("addr")))
	require.False(t, handler.IsLimitedTransfer([]byte("addr")))
	require.False(t, handler.IsInterfaceNil())
}

func TestFeeHandler(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	handler := &FeeHandler{}
	require.Equal(t, 1.0, handler.GasPriceModifier())
	require.Equal(t, 0.0, handler.DeveloperPercentage())
	require.Equal(t, big.NewInt(0), handler.GenesisTotalSupply())
	require.Equal(t, uint64(0), handler.MinGasPrice())
	require.Equal(t, uint64(0), handler.MinGasLimit())
	require.Equal(t, uint64(0), handler.ExtraGasLimitGuardedTx())
	require.Equal(t, uint64(math.MaxUint64), handler.MaxGasPriceSetGuardian())
	require.Equal(t, uint64(math.MaxUint64), handler.MaxGasLimitPerBlock(0))
	require.Equal(t, uint64(math.MaxUint64), handler.MaxGasLimitPerMiniBlock(0))
	require.Equal(t, uint64(math.MaxUint64), handler.MaxGasLimitPerBlockForSafeCrossShard())
	require.Equal(t, uint64(math.MaxUint64), handler.MaxGasLimitPerMiniBlockForSafeCrossShard())
	require.Equal(t, uint64(math.MaxUint64), handler.MaxGasLimitPerTx())
	require.Equal(t, uint64(0), handler.ComputeGasLimit(nil))
	require.Equal(t, big.NewInt(0), handler.ComputeMoveBalanceFee(nil))
	require.Equal(t, big.NewInt(0), handler.ComputeFeeForProcessing(nil, 0))
	require.Equal(t, big.NewInt(0), handler.ComputeTxFee(nil))
	require.Nil(t, handler.CheckValidityTxValues(nil))
	handler.CreateBlockStarted(scheduled.GasAndFees{})
	require.Equal(t, big.NewInt(0), handler.GetAccumulatedFees())
	handler.ProcessTransactionFee(big.NewInt(0), big.NewInt(0), []byte(""))
	handler.ProcessTransactionFeeRelayedUserTx(big.NewInt(0), big.NewInt(0), []byte(""), []byte(""))
	handler.RevertFees([][]byte{})
	require.Equal(t, big.NewInt(0), handler.GetDeveloperFees())
	require.Equal(t, uint64(0), handler.GasPerDataByte())
	require.Equal(t, uint64(0), handler.GasPriceForProcessing(nil))
	require.Equal(t, uint64(0), handler.GasPriceForMove(nil))
	require.Equal(t, uint64(0), handler.MinGasPriceForProcessing())
	require.Equal(t, big.NewInt(0), handler.ComputeTxFeeBasedOnGasUsed(nil, 0))
	require.False(t, handler.IsInterfaceNil())

	used, fee := handler.ComputeGasUsedAndFeeBasedOnRefundValue(nil, big.NewInt(0))
	require.Equal(t, uint64(0), used)
	require.Equal(t, big.NewInt(0), fee)

	c1, c2 := handler.SplitTxGasInCategories(nil)
	require.Equal(t, uint64(0), c1)
	require.Equal(t, uint64(0), c2)

	limit, err := handler.ComputeGasLimitBasedOnBalance(nil, big.NewInt(0))
	require.Equal(t, uint64(0), limit)
	require.Nil(t, err)
}

func TestProcessedMiniBlocksTracker(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	handler := &ProcessedMiniBlocksTracker{}
	handler.SetProcessedMiniBlockInfo([]byte{}, []byte{}, nil)
	handler.RemoveMetaBlockHash([]byte{})
	handler.RemoveMiniBlockHash([]byte{})
	require.Nil(t, handler.GetProcessedMiniBlocksInfo([]byte{}))
	info, hash := handler.GetProcessedMiniBlockInfo([]byte{})
	require.Nil(t, info)
	require.Nil(t, hash)
	require.False(t, handler.IsMiniBlockFullyProcessed([]byte{}, []byte{}))
	require.Nil(t, handler.ConvertProcessedMiniBlocksMapToSlice())
	handler.ConvertSliceToProcessedMiniBlocksMap(nil)
	handler.DisplayProcessedMiniBlocks()
	require.False(t, handler.IsInterfaceNil())
}

func TestRater(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	handler := &Rater{}
	require.Equal(t, uint32(0), handler.GetChance(0))
	require.False(t, handler.IsInterfaceNil())
}

func TestRequestHandler(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	handler := &RequestHandler{}
	handler.SetEpoch(0)
	handler.RequestShardHeader(0, []byte{})
	handler.RequestMetaHeader([]byte{})
	handler.RequestMetaHeaderByNonce(0)
	handler.RequestShardHeaderByNonce(0, 0)
	handler.RequestTransaction(0, [][]byte{})
	handler.RequestUnsignedTransactions(0, [][]byte{})
	handler.RequestRewardTransactions(0, [][]byte{})
	handler.RequestMiniBlock(0, []byte{})
	handler.RequestMiniBlocks(0, [][]byte{})
	handler.RequestTrieNodes(0, [][]byte{}, "")
	handler.RequestStartOfEpochMetaBlock(0)
	require.Equal(t, time.Second, handler.RequestInterval())
	require.Nil(t, handler.SetNumPeersToQuery("", 0, 0))
	handler.RequestTrieNode([]byte{}, "", 0)
	require.Equal(t, make([]byte, 0), handler.CreateTrieNodeIdentifier([]byte{}, 0))
	handler.RequestPeerAuthenticationsByHashes(0, [][]byte{})
	handler.RequestValidatorInfo([]byte{})
	handler.RequestValidatorsInfo([][]byte{})
	require.False(t, handler.IsInterfaceNil())

	intra, cross, err := handler.GetNumPeersToQuery("")
	require.Equal(t, 0, intra)
	require.Equal(t, 0, cross)
	require.Nil(t, err)
}

func TestRewardTxProcessor(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	handler := &RewardTxProcessor{}
	require.Nil(t, handler.ProcessRewardTransaction(nil))
	require.False(t, handler.IsInterfaceNil())
}

func TestScheduledTxsExecutionHandler(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	handler := &ScheduledTxsExecutionHandler{}
	handler.Init()
	require.True(t, handler.AddScheduledTx([]byte{}, nil))
	handler.AddScheduledMiniBlocks(nil)
	require.Nil(t, handler.Execute([]byte{}))
	require.Nil(t, handler.ExecuteAll(nil))
	require.Equal(t, make(map[block.Type][]data.TransactionHandler), handler.GetScheduledIntermediateTxs())
	require.Equal(t, make(block.MiniBlockSlice, 0), handler.GetScheduledMiniBlocks())
	require.Equal(t, scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(0),
		DeveloperFees:   big.NewInt(0),
		GasProvided:     0,
		GasPenalized:    0,
		GasRefunded:     0,
	}, handler.GetScheduledGasAndFees())
	handler.SetScheduledInfo(nil)
	require.Nil(t, handler.RollBackToBlock([]byte{}))
	handler.SaveStateIfNeeded(nil)
	handler.SaveState([]byte{}, nil)
	require.Nil(t, handler.GetScheduledRootHash())
	handler.SetScheduledRootHash([]byte{})
	handler.SetScheduledGasAndFees(scheduled.GasAndFees{})
	handler.SetTransactionProcessor(nil)
	handler.SetTransactionCoordinator(nil)
	require.False(t, handler.IsScheduledTx([]byte{}))
	require.False(t, handler.IsMiniBlockExecuted([]byte{}))
	require.False(t, handler.IsInterfaceNil())

	hash, err := handler.GetScheduledRootHashForHeader([]byte{})
	require.Equal(t, make([]byte, 0), hash)
	require.Nil(t, err)

	hash, err = handler.GetScheduledRootHashForHeaderWithEpoch([]byte{}, 0)
	require.Equal(t, make([]byte, 0), hash)
	require.Nil(t, err)
}

func TestSimpleNFTStorage(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	handler := &SimpleNFTStorage{}
	token, ok, err := handler.GetESDTNFTTokenOnDestination(nil, []byte{}, 0)
	require.Equal(t, &esdt.ESDigitalToken{Value: big.NewInt(0)}, token)
	require.True(t, ok)
	require.Nil(t, err)
	require.Nil(t, handler.SaveNFTMetaData(nil))
	require.False(t, handler.IsInterfaceNil())
}

func TestTxVersionChecker(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	handler := NewDisabledTxVersionChecker()
	require.False(t, handler.IsGuardedTransaction(nil))
	require.False(t, handler.IsSignedWithHash(nil))
	require.Nil(t, handler.CheckTxVersion(nil))
	require.False(t, handler.IsInterfaceNil())
}
