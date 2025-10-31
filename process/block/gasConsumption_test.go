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
		BlockCapacityOverestimationFactor: 200,
		PercentDecreaseLimitsStep:         10,
	}
}

func generateTxs(gasLimitPerTx uint64, numTxs uint32) ([][]byte, []data.TransactionHandler) {
	txs := make([]data.TransactionHandler, 0, numTxs)
	txHashes := make([][]byte, 0, numTxs)
	for i := uint32(0); i < numTxs; i++ {
		txs = append(txs, &transaction.Transaction{
			GasLimit: gasLimitPerTx,
		})
		txHashes = append(txHashes, []byte(fmt.Sprintf("hash_%d", i)))
	}

	return txHashes, txs
}

func generateTxsForMb(gasLimitPerTx uint64, numTxs uint32) []data.TransactionHandler {
	_, txs := generateTxs(gasLimitPerTx, numTxs)
	return txs
}

func generateTxsForMiniBlocks(mbs []data.MiniBlockHeaderHandler) map[string][]data.TransactionHandler {
	txs := make(map[string][]data.TransactionHandler, len(mbs))
	for _, mb := range mbs {
		_, txs[string(mb.GetHash())] = generateTxs(maxGasLimitPerTx, mb.GetTxCount())
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
	t.Run("invalid block overestimation factor should error", func(t *testing.T) {
		t.Parallel()

		args := getMockArgsGasConsumption()
		args.BlockCapacityOverestimationFactor = 5
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

	t.Run("empty mini blocks should early exit", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(nil, nil)
		require.NoError(t, err)
		require.Zero(t, pendingMbs)
		require.Equal(t, -1, lastMbIndex)
	})
	t.Run("missing txs for mini block", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks([]data.MiniBlockHeaderHandler{&coreBlock.MiniBlockHeader{}}, map[string][]data.TransactionHandler{"mb1": generateTxsForMb(maxGasLimitPerTx, 1)})
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
		txs[string(mbs[0].GetHash())][0] = generateTxsForMb(maxGasLimitPerTx+1, 1)[0] // overwrite first tx
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
	t.Run("should work within limits and multiple calls", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		mbs := generateMiniBlocks(3, 5)
		txs := generateTxsForMiniBlocks(mbs)
		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(mbs, txs)
		require.NoError(t, err)
		require.Zero(t, pendingMbs) // no pending mini blocks, we are within limits
		require.Equal(t, len(mbs)-1, lastMbIndex)

		// should allow multiple calls
		lastMbIndex, pendingMbs, err = gc.CheckIncomingMiniBlocks(mbs, txs)
		require.NoError(t, err)
		require.Zero(t, pendingMbs) // no pending mini blocks, we are still within limits
		require.Equal(t, len(mbs)-1, lastMbIndex)

		require.Equal(t, 30*maxGasLimitPerTx, gc.TotalGasConsumed()) // 3 mbs of 5 txs from 2 different calls
	})
	t.Run("should work and save pending mini blocks", func(t *testing.T) {
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

		// full space for txs left although mini blocks reached the limit
		bandwidthForTxs := gc.GetBandwidthForTransactions()
		require.Equal(t, maxGasLimitPerBlock, bandwidthForTxs)
	})
}

func TestGasConsumption_CheckOutgoingTransactions(t *testing.T) {
	t.Parallel()

	t.Run("no transactions should early exit", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		_, _, err := gc.CheckOutgoingTransactions(nil, nil)
		require.NoError(t, err)
	})
	t.Run("different lengths should error", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		txHashes, txs := generateTxs(maxGasLimitPerTx, 10)
		txHashes = txHashes[:len(txs)-1]
		_, _, err := gc.CheckOutgoingTransactions(txHashes, txs)
		require.Equal(t, process.ErrInvalidValue, err)
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

		addedTxs, addedPendingMbs, err := gc.CheckOutgoingTransactions(generateTxs(maxGasLimitPerTx, 1))
		require.NoError(t, err)
		require.Zero(t, len(addedTxs))
		require.Zero(t, len(addedPendingMbs))
	})
	t.Run("one tx exceeds the maximum gas limit per tx", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		addedTxs, addedPendingMbs, err := gc.CheckOutgoingTransactions(generateTxs(maxGasLimitPerTx+1, 1))
		require.NoError(t, err)
		require.Zero(t, len(addedTxs))
		require.Zero(t, len(addedPendingMbs))
	})
	t.Run("should work within limits, no pending mbs, multiple calls", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		txHashes, txs := generateTxs(maxGasLimitPerTx, 10)
		addedTxs, addedPendingMbs, err := gc.CheckOutgoingTransactions(txHashes, txs)
		require.NoError(t, err)
		require.Equal(t, len(txs), len(addedTxs))
		require.Zero(t, len(addedPendingMbs))

		require.Equal(t, 10*maxGasLimitPerTx, gc.TotalGasConsumed())

		addedTxs, addedPendingMbs, err = gc.CheckOutgoingTransactions(txHashes, txs)
		require.NoError(t, err)
		require.Equal(t, len(txs), len(addedTxs))

		require.Equal(t, 20*maxGasLimitPerTx, gc.TotalGasConsumed())
		require.Zero(t, len(addedPendingMbs))
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

		pending := gc.GetPendingMiniBlocks()
		require.Len(t, pending, 2)

		// maxGasLimitPerBlock = 400
		// half of it * factor (200% by default) will be used for txs
		// thus 400 is the total max limit for transactions
		// will add all as there is space left from mini blocks
		// adding 30 txs will lead to an empty space of 100 worth of gas (enough for 2 more blocks)
		txHashes, txs := generateTxs(maxGasLimitPerTx, 30)
		addedTxs, addedPendingMbs, err := gc.CheckOutgoingTransactions(txHashes, txs)
		require.NoError(t, err)
		require.Equal(t, len(txs), len(addedTxs)) // added all
		require.Equal(t, 2, len(addedPendingMbs)) // added all pending mbs

		require.Equal(t, maxGasLimitPerBlock*2, gc.TotalGasConsumed()) // *2 due to the 200% factor

		pending = gc.GetPendingMiniBlocks()
		require.Len(t, pending, 0)
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
		// thus 400 is the total max limit for transactions
		// will add all as there is space left from mini blocks
		txHashes, txs := generateTxs(maxGasLimitPerTx, 50)
		addedTxs, _, err := gc.CheckOutgoingTransactions(txHashes, txs)
		require.NoError(t, err)
		require.Equal(t, len(txs), len(addedTxs))

		require.Equal(t, 50*maxGasLimitPerTx, gc.TotalGasConsumed())
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
	// thus 400 is the total max limit for transactions
	// will add all as there is space left from mini blocks
	// adding 30 txs will lead to an empty space of 100 worth of gas (enough for 2 more blocks)
	txHashes, txs := generateTxs(maxGasLimitPerTx, 30)
	addedTxs, _, err := gc.CheckOutgoingTransactions(txHashes, txs)
	require.NoError(t, err)
	require.Equal(t, len(txs), len(addedTxs)) // added all

	// call reset, block is full
	gc.Reset()
	require.Equal(t, uint64(0), gc.TotalGasConsumed())
}

func TestGasConsumption_DecreaseOutgoingLimit(t *testing.T) {
	t.Parallel()

	gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
	require.NotNil(t, gc)

	// calling a lot of times to reach min limit and simulate possible overflow
	for i := 0; i < 100; i++ {
		gc.DecreaseOutgoingLimit()
	}

	// outgoing limit should be at lowest, 0
	txHashes, txs := generateTxs(maxGasLimitPerTx, 3)
	addedTxs, _, err := gc.CheckOutgoingTransactions(txHashes, txs)
	require.True(t, errors.Is(err, process.ErrZeroLimit))
	require.Zero(t, len(addedTxs)) // nothing added

	// calling reset should not reset the block limit
	gc.Reset()

	// calling reset should reset the limit
	gc.ResetOutgoingLimit()
	gc.Reset() // required to reset the state

	txHashes, txs = generateTxs(maxGasLimitPerTx, 30)
	addedTxs, _, err = gc.CheckOutgoingTransactions(txHashes, txs)
	require.NoError(t, err)
	require.Equal(t, len(txs), len(addedTxs)) // added all
}

func TestGasConsumption_DecreaseIncomingLimit(t *testing.T) {
	t.Parallel()

	gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
	require.NotNil(t, gc)

	// calling a lot of times to reach min limit and simulate possible overflow
	for i := 0; i < 100; i++ {
		gc.DecreaseIncomingLimit()
	}

	// incoming limit should be at lowest, 0
	mbs := generateMiniBlocks(2, 2)
	txsForMBs := generateTxsForMiniBlocks(mbs)
	lastMBIndex, pendingMBs, err := gc.CheckIncomingMiniBlocks(mbs, txsForMBs)
	require.True(t, errors.Is(err, process.ErrZeroLimit))
	require.Equal(t, -1, lastMBIndex) // nothing added
	require.Zero(t, pendingMBs)       // nothing pending

	// calling reset should not reset the block limit
	gc.Reset()

	// should be ok to add txs, only the limit for incoming was decreased
	txHashes, txs := generateTxs(maxGasLimitPerTx, 30)
	addedTxs, _, err := gc.CheckOutgoingTransactions(txHashes, txs)
	require.NoError(t, err)
	require.Equal(t, len(txs), len(addedTxs)) // added all

	// calling reset should reset the limit
	gc.ResetIncomingLimit()
	gc.Reset() // required to reset the state

	mbs = generateMiniBlocks(10, 5)
	txsForMBs = generateTxsForMiniBlocks(mbs)
	lastMBIndex, pendingMBs, err = gc.CheckIncomingMiniBlocks(mbs, txsForMBs)
	require.NoError(t, err)
	require.Equal(t, 7, lastMBIndex) // added 7
	require.Equal(t, 2, pendingMBs)  // 2 pending

	// zeroing the limit should not allow adding more pending mini blocks after transactions
	gc.ZeroIncomingLimit()
	txHashes, txs = generateTxs(maxGasLimitPerTx, 30)
	addedTxs, _, err = gc.CheckOutgoingTransactions(txHashes, txs)
	require.NoError(t, err)
	require.Equal(t, len(txs), len(addedTxs))           // added all
	require.Equal(t, 2, len(gc.GetPendingMiniBlocks())) // still have the 2 mini blocks as pending
}

func TestGasConsumption_ZeroOutgoingLimit(t *testing.T) {
	t.Parallel()

	gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
	require.NotNil(t, gc)

	// set the outgoing limit to zero
	gc.ZeroOutgoingLimit()

	// outgoing limit should be at 0, no txs should be added
	txHashes, txs := generateTxs(maxGasLimitPerTx, 3)
	addedTxs, _, err := gc.CheckOutgoingTransactions(txHashes, txs)
	require.True(t, errors.Is(err, process.ErrZeroLimit))
	require.Contains(t, err.Error(), "outgoing transactions")
	require.Equal(t, 0, len(addedTxs)) // no txs added

	// calling reset should not reset the block limit
	gc.Reset()

	// adding txs should still not be allowed
	txHashes, txs = generateTxs(maxGasLimitPerTx, 3)
	addedTxs, _, err = gc.CheckOutgoingTransactions(txHashes, txs)
	require.True(t, errors.Is(err, process.ErrZeroLimit))
	require.Contains(t, err.Error(), "outgoing transactions")
	require.Equal(t, 0, len(addedTxs)) // no txs added

	// calling ResetOutgoingLimit should restore the limit
	gc.ResetOutgoingLimit()
	gc.Reset() // required to reset the state

	txHashes, txs = generateTxs(maxGasLimitPerTx, 30)
	addedTxs, _, err = gc.CheckOutgoingTransactions(txHashes, txs)
	require.NoError(t, err)
	require.Equal(t, len(txs), len(addedTxs)) // added all
}

func TestGasConsumption_ZeroIncomingLimit(t *testing.T) {
	t.Parallel()

	gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
	require.NotNil(t, gc)

	// set the incoming limit to zero
	gc.ZeroIncomingLimit()

	// incoming limit should be at 0, no mini blocks should be added
	mbs := generateMiniBlocks(2, 2)
	txsForMBs := generateTxsForMiniBlocks(mbs)
	lastMBIndex, pendingMBs, err := gc.CheckIncomingMiniBlocks(mbs, txsForMBs)
	require.True(t, errors.Is(err, process.ErrZeroLimit))
	require.Contains(t, err.Error(), "incoming mini blocks")
	require.Equal(t, -1, lastMBIndex) // no mbs added
	require.Zero(t, pendingMBs)       // no pending mbs

	// calling reset should not reset the block limit
	gc.Reset()

	// adding mbs should still not be allowed
	mbs = generateMiniBlocks(2, 2)
	txsForMBs = generateTxsForMiniBlocks(mbs)
	lastMBIndex, pendingMBs, err = gc.CheckIncomingMiniBlocks(mbs, txsForMBs)
	require.True(t, errors.Is(err, process.ErrZeroLimit))
	require.Contains(t, err.Error(), "incoming mini blocks")
	require.Equal(t, -1, lastMBIndex) // no mbs added
	require.Zero(t, pendingMBs)       // no pending mbs

	// should be ok to add txs, only the limit for incoming was zeroed
	txHashes, txs := generateTxs(maxGasLimitPerTx, 30)
	addedTxs, _, err := gc.CheckOutgoingTransactions(txHashes, txs)
	require.NoError(t, err)
	require.Equal(t, len(txs), len(addedTxs)) // added all

	// calling ResetIncomingLimit should restore the limit
	gc.ResetIncomingLimit()
	gc.Reset() // required to reset the state

	mbs = generateMiniBlocks(2, 10)
	txsForMBs = generateTxsForMiniBlocks(mbs)
	lastMBIndex, pendingMBs, err = gc.CheckIncomingMiniBlocks(mbs, txsForMBs)
	require.NoError(t, err)
	require.Equal(t, 1, lastMBIndex) // added all
	require.Zero(t, pendingMBs)      // added all
}

func TestGasConsumption_RevertIncomingMiniBlocks(t *testing.T) {
	t.Parallel()

	t.Run("empty mini block hashes should early exit", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		gc.RevertIncomingMiniBlocks(nil)
	})

	t.Run("revert non-pending mini block should decrease total gas consumed", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		// add mini blocks that will be within limits
		mbs := generateMiniBlocks(3, 5)
		txs := generateTxsForMiniBlocks(mbs)
		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(mbs, txs)
		require.NoError(t, err)
		require.Equal(t, 2, lastMbIndex)
		require.Zero(t, pendingMbs)

		totalGasBeforeRevert := gc.TotalGasConsumed()
		expectedGasPerMb := maxGasLimitPerTx * 5 // 5 txs per mini block

		// revert the first mini block and one none existing for coverage
		gc.RevertIncomingMiniBlocks([][]byte{mbs[0].GetHash(), []byte("non-existent-hash")})

		// total gas should decrease by the gas consumed by the first mini block
		require.Equal(t, totalGasBeforeRevert-expectedGasPerMb, gc.TotalGasConsumed())

		// revert the second and third mini blocks
		gc.RevertIncomingMiniBlocks([][]byte{mbs[1].GetHash(), mbs[2].GetHash()})

		// total gas should be zero now
		require.Equal(t, uint64(0), gc.TotalGasConsumed())
	})

	t.Run("revert pending mini block at first position", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		// add more mini blocks than the limit to create pending ones
		mbs := generateMiniBlocks(10, 5)
		txs := generateTxsForMiniBlocks(mbs)
		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(mbs, txs)
		require.NoError(t, err)
		require.Equal(t, 7, lastMbIndex)
		require.Equal(t, 2, pendingMbs)

		// get pending mini blocks
		pendingMiniBlocks := gc.GetPendingMiniBlocks()
		require.Len(t, pendingMiniBlocks, 2)

		totalGasBeforeRevert := gc.TotalGasConsumed()

		// revert the first pending mini block
		firstPendingHash := mbs[8].GetHash()
		gc.RevertIncomingMiniBlocks([][]byte{firstPendingHash})

		// total gas should stay the same, only a pending mini block was removed
		require.Equal(t, totalGasBeforeRevert, gc.TotalGasConsumed())

		// check that pending mini blocks were updated
		updatedPendingMiniBlocks := gc.GetPendingMiniBlocks()
		require.Len(t, updatedPendingMiniBlocks, 1)
		require.Equal(t, mbs[9].GetHash(), updatedPendingMiniBlocks[0].GetHash())
	})

	t.Run("revert pending mini block at last position", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		// add more mini blocks than the limit to create pending ones
		mbs := generateMiniBlocks(10, 5)
		txs := generateTxsForMiniBlocks(mbs)
		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(mbs, txs)
		require.NoError(t, err)
		require.Equal(t, 7, lastMbIndex)
		require.Equal(t, 2, pendingMbs)

		// get pending mini blocks
		pendingMiniBlocks := gc.GetPendingMiniBlocks()
		require.Len(t, pendingMiniBlocks, 2)

		totalGasBeforeRevert := gc.TotalGasConsumed()

		// revert the last pending mini block (index 9, which is the last in pending)
		lastPendingHash := mbs[9].GetHash()
		gc.RevertIncomingMiniBlocks([][]byte{lastPendingHash})

		// total gas should stay the same, only a pending mini block was removed
		require.Equal(t, totalGasBeforeRevert, gc.TotalGasConsumed())

		// check that pending mini blocks were updated
		updatedPendingMiniBlocks := gc.GetPendingMiniBlocks()
		require.Len(t, updatedPendingMiniBlocks, 1)
		require.Equal(t, mbs[8].GetHash(), updatedPendingMiniBlocks[0].GetHash())
	})

	t.Run("revert pending mini block at middle position", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		// add more mini blocks than the limit to create 3 pending ones
		mbs := generateMiniBlocks(11, 5)
		txs := generateTxsForMiniBlocks(mbs)
		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(mbs, txs)
		require.NoError(t, err)
		require.Equal(t, 7, lastMbIndex)
		require.Equal(t, 3, pendingMbs)

		// get pending mini blocks
		pendingMiniBlocks := gc.GetPendingMiniBlocks()
		require.Len(t, pendingMiniBlocks, 3)

		totalGasBeforeRevert := gc.TotalGasConsumed()

		// revert the middle pending mini block (index 9, which is the middle in pending)
		middlePendingHash := mbs[9].GetHash()
		gc.RevertIncomingMiniBlocks([][]byte{middlePendingHash})

		// total gas should stay the same, only a pending mini block was removed
		require.Equal(t, totalGasBeforeRevert, gc.TotalGasConsumed())

		// check that pending mini blocks were updated
		updatedPendingMiniBlocks := gc.GetPendingMiniBlocks()
		require.Len(t, updatedPendingMiniBlocks, 2)
		require.Equal(t, mbs[8].GetHash(), updatedPendingMiniBlocks[0].GetHash())
		require.Equal(t, mbs[10].GetHash(), updatedPendingMiniBlocks[1].GetHash())
	})

	t.Run("revert multiple mini blocks with mixed pending and non-pending", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		// add mini blocks, some will be pending
		mbs := generateMiniBlocks(10, 5)
		txs := generateTxsForMiniBlocks(mbs)
		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(mbs, txs)
		require.NoError(t, err)
		require.Equal(t, 7, lastMbIndex)
		require.Equal(t, 2, pendingMbs)

		totalGasBeforeRevert := gc.TotalGasConsumed()
		expectedGasPerMb := maxGasLimitPerTx * 5

		// revert both a non-pending mini block (index 0) and a pending one (index 8)
		hashesToRevert := [][]byte{
			mbs[0].GetHash(), // non-pending
			mbs[8].GetHash(), // pending (first in pending)
		}
		gc.RevertIncomingMiniBlocks(hashesToRevert)

		// total gas should decrease only by the non-pending mini block
		require.Equal(t, totalGasBeforeRevert-expectedGasPerMb, gc.TotalGasConsumed())

		// pending mini blocks should be reduced by one
		updatedPendingMiniBlocks := gc.GetPendingMiniBlocks()
		require.Len(t, updatedPendingMiniBlocks, 1)
		require.Equal(t, mbs[9].GetHash(), updatedPendingMiniBlocks[0].GetHash())
	})

	t.Run("revert all pending mini blocks", func(t *testing.T) {
		t.Parallel()

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		// add mini blocks with pending ones
		mbs := generateMiniBlocks(10, 5)
		txs := generateTxsForMiniBlocks(mbs)
		lastMbIndex, pendingMbs, err := gc.CheckIncomingMiniBlocks(mbs, txs)
		require.NoError(t, err)
		require.Equal(t, 7, lastMbIndex)
		require.Equal(t, 2, pendingMbs)

		totalGasBeforeRevert := gc.TotalGasConsumed()

		// revert all pending mini blocks
		hashesToRevert := [][]byte{
			mbs[8].GetHash(),
			mbs[9].GetHash(),
		}
		gc.RevertIncomingMiniBlocks(hashesToRevert)

		// total gas should stay the same, only pending mini blocks were removed
		require.Equal(t, totalGasBeforeRevert, gc.TotalGasConsumed())

		// no pending mini blocks should remain
		updatedPendingMiniBlocks := gc.GetPendingMiniBlocks()
		require.Len(t, updatedPendingMiniBlocks, 0)
	})
}

func TestGasConsumption_ConcurrentOps(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	require.NotPanics(t, func() {
		mbs := generateMiniBlocks(1, 2)
		mbsHashes := make([][]byte, len(mbs))
		for _, mb := range mbs {
			mbsHashes = append(mbsHashes, mb.GetHash())
		}
		txsInMBs := generateTxsForMiniBlocks(mbs)

		txHashes, txs := generateTxs(maxGasLimitPerTx, 3)

		gc, _ := block.NewGasConsumption(getMockArgsGasConsumption())
		require.NotNil(t, gc)

		numCalls := 1000
		wg := sync.WaitGroup{}
		wg.Add(numCalls)

		for i := 0; i < numCalls; i++ {
			go func(idx int) {
				switch idx % 12 {
				case 0:
					_, _, _ = gc.CheckOutgoingTransactions(txHashes, txs)
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
					gc.TotalGasConsumed()
				case 8:
					_ = gc.GetPendingMiniBlocks()
				case 9:
					gc.ZeroIncomingLimit()
				case 10:
					gc.ZeroOutgoingLimit()
				case 11:
					gc.RevertIncomingMiniBlocks(mbsHashes)
				default:
					require.Fail(t, "should have not been called")
				}

				wg.Done()
			}(i)
		}

		wg.Wait()
	})
}
