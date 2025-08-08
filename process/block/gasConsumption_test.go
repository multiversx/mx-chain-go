package block_test

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	coreBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
)

const (
	maxGasLimitPerBlock     = uint64(400)
	maxGasLimitPerMiniBlock = uint64(100)
	maxGasLimitPerTx        = uint64(10)
)

var expectedError = errors.New("expected error")

func getMockArgsGasConsumption() block.ArgsGasConsumption {
	return block.ArgsGasConsumption{
		EconomicsFee: &economicsmocks.EconomicsHandlerMock{
			MaxGasLimitPerBlockCalled: func(shardID uint32) uint64 {
				return maxGasLimitPerBlock
			},
			MaxGasLimitPerMiniBlockCalled: func(shardID uint32) uint64 {
				return maxGasLimitPerMiniBlock
			},
			MaxGasLimitPerTxCalled: func() uint64 {
				return maxGasLimitPerTx
			},
			MaxGasLimitPerBlockForSafeCrossShardCalled: func() uint64 {
				return maxGasLimitPerBlock
			},
		},
		ShardCoordinator: &mock.ShardCoordinatorStub{},
		GasHandler: &mock.GasHandlerMock{
			ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return txHandler.GetGasLimit(), txHandler.GetGasLimit(), nil
			},
		},
		InitialLimitsFactor:       200,
		PercentDecreaseLimitsStep: 10,
	}
}

func generateTxs(gasLimitPerTx uint64, numTxs uint32) []data.TransactionHandler {
	txs := make([]data.TransactionHandler, 0, numTxs)
	for i := uint32(0); i < numTxs; i++ {
		txs = append(txs, &transaction.Transaction{
			GasLimit: gasLimitPerTx,
		})
	}

	return txs
}

func generateTxsForMiniBlocks(mbs []data.MiniBlockHeaderHandler) map[string][]data.TransactionHandler {
	txs := make(map[string][]data.TransactionHandler, len(mbs))
	for _, mb := range mbs {
		txs[string(mb.GetHash())] = generateTxs(maxGasLimitPerTx, mb.GetTxCount())
	}

	return txs
}

func generateMiniBlocks(numOfMiniBlocks int, numTxsInMiniBlock uint32) []data.MiniBlockHeaderHandler {
	mbs := make([]data.MiniBlockHeaderHandler, numOfMiniBlocks)
	for i := 0; i < numOfMiniBlocks; i++ {
		mbs[i] = &coreBlock.MiniBlockHeader{
			Hash:    []byte(fmt.Sprintf("mb%d", i)),
			TxCount: numTxsInMiniBlock,
		}
	}
	return mbs
}

func TestNewGasConsumption(t *testing.T) {
	t.Parallel()

	t.Run("nil economics handler should error", func(t *testing.T) {
		t.Parallel()

		args := getMockArgsGasConsumption()
		args.EconomicsFee = nil
		gc, err := block.NewGasConsumption(args)
		require.Nil(t, gc)
		require.Equal(t, process.ErrNilEconomicsFeeHandler, err)
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := getMockArgsGasConsumption()
		args.ShardCoordinator = nil
		gc, err := block.NewGasConsumption(args)
		require.Nil(t, gc)
		require.Equal(t, process.ErrNilShardCoordinator, err)
	})
	t.Run("nil gas handler should error", func(t *testing.T) {
		t.Parallel()

		args := getMockArgsGasConsumption()
		args.GasHandler = nil
		gc, err := block.NewGasConsumption(args)
		require.Nil(t, gc)
		require.Equal(t, process.ErrNilGasHandler, err)
	})
	t.Run("invalid initial limits factor should error", func(t *testing.T) {
		t.Parallel()

		args := getMockArgsGasConsumption()
		args.InitialLimitsFactor = 5
		gc, err := block.NewGasConsumption(args)
		require.Nil(t, gc)
		require.True(t, errors.Is(err, process.ErrInvalidValue))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		gc, err := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)
		require.NoError(t, err)
	})
}

func TestGasConsumption_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	args := getMockArgsGasConsumption()
	args.EconomicsFee = nil
	gc, _ := block.NewGasConsumption(args)
	require.True(t, gc.IsInterfaceNil())

	gc, _ = block.NewGasConsumption(getMockArgsGasConsumption())
	require.False(t, gc.IsInterfaceNil())
}

func TestGasConsumption_CheckIncomingMiniBlocks(t *testing.T) {
	t.Parallel()

	t.Run("mini blocks selection done should error", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(nil, nil) // coverage only
		require.NoError(t, err)
		require.Zero(t, pendingMbs)
		require.Equal(t, -1, lastMbIndex)

		// first call with dummy data should set isMiniBlockSelectionDone
		_, _, _ = gc.CheckIncomingMiniBlocks([]data.MiniBlockHeaderHandler{&coreBlock.MiniBlockHeader{}}, map[string][]data.TransactionHandler{"mb1": generateTxs(maxGasLimitPerTx, 1)})

		// second call should early exit
		lastMbIndex, pendingMbs, err = gc.CheckIncomingMiniBlocks([]data.MiniBlockHeaderHandler{&coreBlock.MiniBlockHeader{}}, map[string][]data.TransactionHandler{"mb1": generateTxs(maxGasLimitPerTx, 1)})
		require.Equal(t, process.ErrMiniBlocksAlreadyProcessed, err)
		require.Zero(t, pendingMbs)
		require.Equal(t, -1, lastMbIndex)
	})
	t.Run("missing txs for mini block", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks([]data.MiniBlockHeaderHandler{&coreBlock.MiniBlockHeader{}}, map[string][]data.TransactionHandler{"mb1": generateTxs(maxGasLimitPerTx, 1)})
		require.True(t, errors.Is(err, process.ErrInvalidValue))
		require.True(t, strings.Contains(err.Error(), "could not find mini block hash in transactions map"))
		require.Zero(t, pendingMbs)
		require.Equal(t, -1, lastMbIndex)
	})
	t.Run("number of txs for mini block differ from the ones in mini block", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		mbs := generateMiniBlocks(1, 5)
		txs := generateTxsForMiniBlocks(mbs)
		txs[string(mbs[0].GetHash())] = txs[string(mbs[0].GetHash())][1:] // remove one tx
		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(mbs, txs)
		require.True(t, errors.Is(err, process.ErrInvalidValue))
		require.True(t, strings.Contains(err.Error(), "the provided mini block does not match the number of transactions provided"))
		require.Zero(t, pendingMbs)
		require.Equal(t, -1, lastMbIndex)
	})
	t.Run("ComputeGasProvidedByTx fails", func(t *testing.T) {
		t.Parallel()

		args := getMockArgsGasConsumption()
		args.GasHandler = &mock.GasHandlerMock{
			ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, expectedError
			},
		}
		gc, _ := block.NewGasConsumption(args)
		require.NotNil(t, gc)

		mbs := generateMiniBlocks(1, 5)
		txs := generateTxsForMiniBlocks(mbs)
		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(mbs, txs)
		require.Equal(t, expectedError, err)
		require.Zero(t, pendingMbs)
		require.Equal(t, -1, lastMbIndex)
	})
	t.Run("one tx exceeds the maximum gas limit per tx", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		mbs := generateMiniBlocks(1, 5)
		txs := generateTxsForMiniBlocks(mbs)
		txs[string(mbs[0].GetHash())][0] = generateTxs(maxGasLimitPerTx+1, 1)[0] // overwrite first tx
		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(mbs, txs)
		require.Equal(t, process.ErrMaxGasLimitPerTransactionIsReached, err)
		require.Zero(t, pendingMbs)
		require.Equal(t, -1, lastMbIndex)
	})
	t.Run("mini block limit exceeded", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		mbs := generateMiniBlocks(1, 21) // too many txs
		txs := generateTxsForMiniBlocks(mbs)
		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(mbs, txs)
		require.Equal(t, process.ErrMaxGasLimitPerMiniBlockIsReached, err)
		require.Zero(t, pendingMbs)
		require.Equal(t, -1, lastMbIndex)
	})
	t.Run("should work within limits, no pending txs", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		mbs := generateMiniBlocks(3, 5)
		txs := generateTxsForMiniBlocks(mbs)
		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(mbs, txs)
		require.NoError(t, err)
		require.Zero(t, pendingMbs) // no pending mini blocks, we are within limits
		require.Equal(t, len(mbs)-1, lastMbIndex)
	})
	t.Run("should work and save pending mini blocks, no pending txs", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		// maxGasLimitPerBlock = 400
		// half of it * factor (200% by default) will be used for mini blocks
		// thus 400 is the total max limit for mini blocks
		// 5 txs in each mb with a gas limit of 10 => gasLimitPerMb = 50
		// so the limit will be reached after 8 mini blocks
		// calling with 10 mini blocks should add 8 to the block and save 2 as pending
		mbs := generateMiniBlocks(10, 5)
		txs := generateTxsForMiniBlocks(mbs)
		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(mbs, txs)
		require.NoError(t, err)
		require.Equal(t, 2, pendingMbs)  // 2 pending mini blocks
		require.Equal(t, 7, lastMbIndex) // last index saved 7

		require.Equal(t, maxGasLimitPerTx*8*5, gc.TotalGasConsumed()) // 8 mbs of 5 txs each
	})
	t.Run("should work within limits and continue adding pending txs to fill the block", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		// maxGasLimitPerBlock = 400
		// half of it * factor (200% by default) will be used for txs
		// thus 400 is the total max limit for transactions (= 40 txs)
		initialTxs := generateTxs(maxGasLimitPerTx, 70) // will have 30 pending
		lastTxIndex, err := gc.CheckOutgoingTransactions(initialTxs)
		require.NoError(t, err)
		require.Equal(t, 39, lastTxIndex)

		// maxGasLimitPerBlock = 400
		// half of it * factor (200% by default) will be used for mini blocks
		// thus 400 is the total max limit for mini blocks
		// 5 txs in each mb with a gas limit of 10 => gasLimitPerMb = 50
		// adding 5 mbs will lead to an empty space left of 150 worth of gas (enough for 15 more pending outgoing txs)
		mbs := generateMiniBlocks(5, 5)
		txs := generateTxsForMiniBlocks(mbs)
		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(mbs, txs)
		require.NoError(t, err)
		require.Zero(t, pendingMbs)      // no pending mini blocks
		require.Equal(t, 4, lastMbIndex) // last index saved 4

		expectedFinalIndex := 54 // 40 added first, then 15 more
		finalLastTxIndex := gc.GetLastTransactionIndexIncluded()
		require.Equal(t, expectedFinalIndex, finalLastTxIndex)

		require.Equal(t, maxGasLimitPerBlock*2, gc.TotalGasConsumed()) // *2 due to the 200% factor
	})
}

func TestGasConsumption_CheckOutgoingTransactions(t *testing.T) {
	t.Parallel()

	t.Run("transactions selection done should error", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		lastTxIndex, err := gc.CheckOutgoingTransactions(nil) // coverage only
		require.NoError(t, err)
		require.Equal(t, -1, lastTxIndex)

		// first call with dummy data should set isTransactionSelectionDone
		_, _ = gc.CheckOutgoingTransactions(generateTxs(maxGasLimitPerTx, 1))

		// second call should early exit
		lastTxIndex, err = gc.CheckOutgoingTransactions(generateTxs(maxGasLimitPerTx, 1))
		require.Equal(t, process.ErrTransactionsAlreadyProcessed, err)
		require.Equal(t, -1, lastTxIndex)
	})
	t.Run("ComputeGasProvidedByTx fails", func(t *testing.T) {
		t.Parallel()

		args := getMockArgsGasConsumption()
		args.GasHandler = &mock.GasHandlerMock{
			ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, expectedError
			},
		}
		gc, _ := block.NewGasConsumption(args)
		require.NotNil(t, gc)

		lastTxIndex, err := gc.CheckOutgoingTransactions(generateTxs(maxGasLimitPerTx, 1))
		require.Equal(t, expectedError, err)
		require.Equal(t, -1, lastTxIndex)
	})
	t.Run("one tx exceeds the maximum gas limit per tx", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		lastTxIndex, err := gc.CheckOutgoingTransactions(generateTxs(maxGasLimitPerTx+1, 1))
		require.Equal(t, process.ErrMaxGasLimitPerTransactionIsReached, err)
		require.Equal(t, -1, lastTxIndex)
	})
	t.Run("should work within limits, no pending mbs", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		txs := generateTxs(maxGasLimitPerTx, 10)
		lastTxIndex, err := gc.CheckOutgoingTransactions(txs)
		require.NoError(t, err)
		require.Equal(t, len(txs)-1, lastTxIndex)

		require.Equal(t, 10*maxGasLimitPerTx, gc.TotalGasConsumed())
	})
	t.Run("should work and save pending txs, no pending mbs", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		// maxGasLimitPerBlock = 400
		// half of it * factor (200% by default) will be used for txs
		// thus 400 is the total max limit for transactions (= 40 txs)
		txs := generateTxs(maxGasLimitPerTx, 50)
		lastTxIndex, err := gc.CheckOutgoingTransactions(txs)
		require.NoError(t, err)
		require.Equal(t, 39, lastTxIndex)

		require.Equal(t, maxGasLimitPerBlock, gc.TotalGasConsumed()) // all space for txs
	})
	t.Run("should work within limits and continue adding pending mini blocks to fill the block", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		// maxGasLimitPerBlock = 400
		// half of it * factor (200% by default) will be used for mini blocks
		// thus 400 is the total max limit for mini blocks
		// 5 txs in each mb with a gas limit of 10 => gasLimitPerMb = 50
		// adding 10 mbs will lead to adding 8 and saving 2 as pending
		mbs := generateMiniBlocks(10, 5)
		txsInMBs := generateTxsForMiniBlocks(mbs)
		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(mbs, txsInMBs)
		require.NoError(t, err)
		require.Equal(t, 2, pendingMbs)  // 2 pending mini blocks
		require.Equal(t, 7, lastMbIndex) // last index saved 7

		// maxGasLimitPerBlock = 400
		// half of it * factor (200% by default) will be used for txs
		// thus 400 is the total max limit for transactions (= 40 txs)
		// adding 30 txs will lead to an empty space of 100 worth of gas (enough for 2 more blocks)
		txs := generateTxs(maxGasLimitPerTx, 30)
		lastTxIndex, err := gc.CheckOutgoingTransactions(txs)
		require.NoError(t, err)
		require.Equal(t, 29, lastTxIndex) // added all

		expectedFinalMBIndex := 9
		finalLastMBIndex := gc.GetLastMiniBlockIndexIncluded()
		require.Equal(t, expectedFinalMBIndex, finalLastMBIndex)

		require.Equal(t, maxGasLimitPerBlock*2, gc.TotalGasConsumed()) // *2 due to the 200% factor
	})
	t.Run("should work with multiple destination shards", func(t *testing.T) {
		t.Parallel()

		cnt := 0
		args := getMockArgsGasConsumption()
		args.ShardCoordinator = &mock.ShardCoordinatorStub{
			ComputeIdCalled: func(address []byte) uint32 {
				cnt++
				if cnt < 30 {
					return 1 // first 40 txs going to shard 1, won't exceed the limit
				}

				return 0
			},
		}
		args.GasHandler = &mock.GasHandlerMock{
			ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				if txSenderShardId == txReceiverSharedId {
					return maxGasLimitPerTx, maxGasLimitPerTx, nil
				}

				return maxGasLimitPerTx / 2, maxGasLimitPerTx / 2, nil
			},
		}
		gc, _ := block.NewGasConsumption(args)
		require.NotNil(t, gc)

		// maxGasLimitPerBlock = 400
		// half of it * factor (200% by default) will be used for txs
		// thus 400 is the total max limit for transactions (= 40 txs)
		txs := generateTxs(maxGasLimitPerTx, 50)
		lastTxIndex, err := gc.CheckOutgoingTransactions(txs)
		require.NoError(t, err)
		require.Equal(t, 39, lastTxIndex)

		require.Equal(t, maxGasLimitPerBlock, gc.TotalGasConsumed()) // all space for txs
	})
}

func TestGasConsumption_Reset(t *testing.T) {
	t.Parallel()

	gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
	require.NotNil(t, gc)

	// maxGasLimitPerBlock = 400
	// half of it * factor (200% by default) will be used for mini blocks
	// thus 400 is the total max limit for mini blocks
	// 5 txs in each mb with a gas limit of 10 => gasLimitPerMb = 50
	// adding 10 mbs will lead to adding 8 and saving 2 as pending
	mbs := generateMiniBlocks(10, 5)
	txsInMBs := generateTxsForMiniBlocks(mbs)
	lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(mbs, txsInMBs)
	require.NoError(t, err)
	require.Equal(t, 2, pendingMbs)  // 2 pending mini blocks
	require.Equal(t, 7, lastMbIndex) // last index saved 7

	// maxGasLimitPerBlock = 400
	// half of it * factor (200% by default) will be used for txs
	// thus 400 is the total max limit for transactions (= 40 txs)
	// adding 30 txs will lead to an empty space of 100 worth of gas (enough for 2 more blocks)
	txs := generateTxs(maxGasLimitPerTx, 30)
	lastTxIndex, err := gc.CheckOutgoingTransactions(txs)
	require.NoError(t, err)
	require.Equal(t, 29, lastTxIndex) // added all

	// call reset, block is full
	gc.Reset()
	require.Equal(t, -1, gc.GetLastTransactionIndexIncluded())
	require.Equal(t, -1, gc.GetLastMiniBlockIndexIncluded())

	// should accept more txs now
	txs = generateTxs(maxGasLimitPerTx, 30)
	lastTxIndex, err = gc.CheckOutgoingTransactions(txs)
	require.NoError(t, err)
	require.Equal(t, 29, lastTxIndex) // added all
}

func TestGasConsumption_DecreaseOutgoingLimit(t *testing.T) {
	t.Parallel()

	gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
	require.NotNil(t, gc)

	// calling a lot of times to reach min limit and simulate possible overflow
	for i := 0; i < 100; i++ {
		gc.DecreaseOutgoingLimit()
	}

	// outgoing limit should be at lowest, 10% of the half configured 400 => 20 (max 2 txs)
	txs := generateTxs(maxGasLimitPerTx, 3)
	lastTxIndex, err := gc.CheckOutgoingTransactions(txs)
	require.NoError(t, err)
	require.Equal(t, 1, lastTxIndex) // added only 2 txs, one should be pending

	// calling reset should not reset the block limit
	gc.Reset()

	// adding txs should be allowed again, but within the same limit
	txs = generateTxs(maxGasLimitPerTx, 3)
	lastTxIndex, err = gc.CheckOutgoingTransactions(txs)
	require.NoError(t, err)
	require.Equal(t, 1, lastTxIndex) // added only 2 txs, one should be pending

	// should be ok to add mini blocks, only the limit for outgoing was decreased
	mbs := generateMiniBlocks(1, 5)
	txsForMBs := generateTxsForMiniBlocks(mbs)
	lastMBIndex, pendingMBs, err := gc.CheckIncomingMiniBlocks(mbs, txsForMBs)
	require.NoError(t, err)
	require.Zero(t, lastMBIndex) // added all
	require.Zero(t, pendingMBs)  // added all

	// calling reset should reset the limit
	gc.ResetOutgoingLimit()
	gc.Reset() // required to reset the state

	txs = generateTxs(maxGasLimitPerTx, 30)
	lastTxIndex, err = gc.CheckOutgoingTransactions(txs)
	require.NoError(t, err)
	require.Equal(t, 29, lastTxIndex) // added all
}

func TestGasConsumption_DecreaseIncomingLimit(t *testing.T) {
	t.Parallel()

	gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
	require.NotNil(t, gc)

	// calling a lot of times to reach min limit and simulate possible overflow
	for i := 0; i < 100; i++ {
		gc.DecreaseIncomingLimit()
	}

	// incoming limit should be at lowest, 10% of the half configured 400 => 20 (1 mb with 2 txs)
	mbs := generateMiniBlocks(2, 2)
	txsForMBs := generateTxsForMiniBlocks(mbs)
	lastMBIndex, pendingMBs, err := gc.CheckIncomingMiniBlocks(mbs, txsForMBs)
	require.NoError(t, err)
	require.Zero(t, lastMBIndex)    // only first one added
	require.Equal(t, 1, pendingMBs) // one mb left

	// calling reset should not reset the block limit
	gc.Reset()

	// adding mbs should be allowed again, but within the same limit
	mbs = generateMiniBlocks(2, 2)
	txsForMBs = generateTxsForMiniBlocks(mbs)
	lastMBIndex, pendingMBs, err = gc.CheckIncomingMiniBlocks(mbs, txsForMBs)
	require.NoError(t, err)
	require.Zero(t, lastMBIndex)    // only first one added
	require.Equal(t, 1, pendingMBs) // one mb left

	// should be ok to add txs, only the limit for incoming was decreased
	txs := generateTxs(maxGasLimitPerTx, 30)
	lastTxIndex, err := gc.CheckOutgoingTransactions(txs)
	require.NoError(t, err)
	require.Equal(t, 29, lastTxIndex) // added all

	// calling reset should reset the limit
	gc.ResetIncomingLimit()
	gc.Reset() // required to reset the state

	mbs = generateMiniBlocks(2, 10)
	txsForMBs = generateTxsForMiniBlocks(mbs)
	lastMBIndex, pendingMBs, err = gc.CheckIncomingMiniBlocks(mbs, txsForMBs)
	require.NoError(t, err)
	require.Equal(t, 1, lastMBIndex) // added all
	require.Zero(t, 0, pendingMBs)   // added all
}

func TestGasConsumption_ConcurrentOps(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	require.NotPanics(t, func() {
		mbs := generateMiniBlocks(1, 2)
		txsInMBs := generateTxsForMiniBlocks(mbs)

		txs := generateTxs(maxGasLimitPerTx, 3)

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		numCalls := 1000
		wg := sync.WaitGroup{}
		wg.Add(numCalls)

		for i := 0; i < numCalls; i++ {
			go func(idx int) {
				switch idx % 10 {
				case 0:
					_, _ = gc.CheckOutgoingTransactions(txs)
				case 1:
					_, _, _ = gc.CheckIncomingMiniBlocks(mbs, txsInMBs)
				case 2:
					gc.Reset()
				case 3:
					gc.DecreaseOutgoingLimit()
				case 4:
					gc.DecreaseIncomingLimit()
				case 5:
					gc.ResetOutgoingLimit()
				case 6:
					gc.ResetIncomingLimit()
				case 7:
					gc.GetLastMiniBlockIndexIncluded()
				case 8:
					gc.GetLastTransactionIndexIncluded()
				case 9:
					gc.TotalGasConsumed()
				default:
					require.Fail(t, "should have not been called")
				}

				wg.Done()
			}(i)
		}

		wg.Wait()
	})
}
