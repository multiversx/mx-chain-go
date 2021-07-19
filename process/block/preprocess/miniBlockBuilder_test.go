package preprocess

import (
	"encoding/hex"
	"math/big"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newMiniBlockBuilderWithError(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.accounts = nil

	mbb, err := newMiniBlockBuilder(args)
	require.Equal(t, process.ErrNilAccountsAdapter, err)
	require.Nil(t, mbb)
}

func Test_newMiniBlockBuilderOK(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()

	mbb, err := newMiniBlockBuilder(args)
	require.Nil(t, err)
	require.NotNil(t, mbb)
}

func Test_checkMiniBlocksBuilderArgsNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.gasTracker.shardCoordinator = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilShardCoordinator, err)
}

func Test_checkMiniBlocksBuilderArgsNilGasHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.gasTracker.gasHandler = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilGasHandler, err)
}

func Test_checkMiniBlocksBuilderArgsNilFeeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.gasTracker.economicsFee = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func Test_checkMiniBlocksBuilderArgsNilBlockSizeComputationHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.blockSizeComputation = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilBlockSizeComputationHandler, err)
}

func Test_checkMiniBlocksBuilderArgsNilAccountsTxsPerShardsShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.accountTxsShards = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilAccountTxsPerShard, err)
}

func Test_checkMiniBlocksBuilderArgsNilBalanceComputationHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.balanceComputationHandler = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilBalanceComputationHandler, err)
}

func Test_checkMiniBlocksBuilderArgsNilHaveTimeHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.haveTime = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilHaveTimeHandler, err)
}

func Test_checkMiniBlocksBuilderArgsNilShardStuckHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.isShardStuck = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilIsShardStuckHandler, err)
}

func Test_checkMiniBlocksBuilderArgsNilMaxBlockSizeReachedHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.isMaxBlockSizeReached = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilIsMaxBlockSizeReachedHandler, err)
}

func Test_checkMiniBlocksBuilderArgsNilTxMaxTotalCostHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.getTxMaxTotalCost = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilTxMaxTotalCostHandler, err)
}

func Test_checkMiniBlocksBuilderArgsOK(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()

	err := checkMiniBlocksBuilderArgs(args)
	require.Nil(t, err)
}

func Test_updateAccountShardsInfo(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()

	mbb, _ := newMiniBlockBuilder(args)
	senderAddr := []byte("senderAddr")
	receiverAddr := []byte("receiverAddr")
	tx := createDefaultTx(senderAddr, receiverAddr, 50000)

	senderShardID := uint32(0)
	receiverShardID := uint32(0)
	wtx := createWrappedTransaction(tx, senderShardID, receiverShardID)

	mbb.updateAccountShardsInfo(tx, wtx)

	addrShardInfo, ok := mbb.accountTxsShards.accountsInfo[string(tx.SndAddr)]
	require.True(t, ok)

	require.Equal(t, senderShardID, addrShardInfo.senderShardID)
	require.Equal(t, receiverShardID, addrShardInfo.receiverShardID)
}

func Test_MiniBlocksBuilderHandleGasRefundIntraShard(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	mbb, _ := newMiniBlockBuilder(args)
	senderAddr := []byte("senderAddr")
	receiverAddr := []byte("receiverAddr")
	tx := createDefaultTx(senderAddr, receiverAddr, 70000)

	senderShardID := uint32(0)
	receiverShardID := uint32(0)
	wtx := createWrappedTransaction(tx, senderShardID, receiverShardID)

	initGas := uint64(200000)
	refund := uint64(20000)
	mbb.gasInfo.gasConsumedByMiniBlocksInSenderShard = initGas
	mbb.gasInfo.totalGasConsumedInSelfShard = initGas
	mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID] = initGas
	mbb.handleGasRefund(wtx, refund)

	require.Equal(t, initGas-refund, mbb.gasInfo.gasConsumedByMiniBlocksInSenderShard)
	require.Equal(t, initGas-refund, mbb.gasInfo.totalGasConsumedInSelfShard)
	require.Equal(t, initGas-refund, mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID])
}

func Test_MiniBlocksBuilderHandleGasRefundCrossShard(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	mbb, _ := newMiniBlockBuilder(args)
	senderAddr := []byte("senderAddr")
	receiverAddr := []byte("receiverAddr")
	tx := createDefaultTx(senderAddr, receiverAddr, 70000)

	senderShardID := uint32(0)
	receiverShardID := uint32(1)
	wtx := createWrappedTransaction(tx, senderShardID, receiverShardID)

	initGas := uint64(200000)
	refund := uint64(20000)
	mbb.gasInfo.gasConsumedByMiniBlocksInSenderShard = initGas
	mbb.gasInfo.totalGasConsumedInSelfShard = initGas
	mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID] = initGas
	mbb.handleGasRefund(wtx, refund)

	require.Equal(t, initGas, mbb.gasInfo.gasConsumedByMiniBlocksInSenderShard)
	require.Equal(t, initGas, mbb.gasInfo.totalGasConsumedInSelfShard)
	require.Equal(t, initGas, mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID])
}

func Test_handleFailedTransactionAddsToCount(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	mbb, _ := newMiniBlockBuilder(args)

	mbb.handleFailedTransaction()
	require.True(t, mbb.stats.firstInvalidTxFound)
	require.Equal(t, uint32(1), mbb.stats.numTxsFailed)

	mbb.handleFailedTransaction()
	require.True(t, mbb.stats.firstInvalidTxFound)
	require.Equal(t, uint32(2), mbb.stats.numTxsFailed)
}

func Test_handleCrossShardScCallOrSpecialTx(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	mbb, _ := newMiniBlockBuilder(args)

	mbb.handleCrossShardScCallOrSpecialTx()
	require.True(t, mbb.stats.firstCrossShardScCallOrSpecialTxFound)
	require.Equal(t, uint32(1), mbb.stats.numCrossShardSCCallsOrSpecialTxs)

	mbb.handleCrossShardScCallOrSpecialTx()
	require.True(t, mbb.stats.firstCrossShardScCallOrSpecialTxFound)
	require.Equal(t, uint32(2), mbb.stats.numCrossShardSCCallsOrSpecialTxs)
}

func Test_isCrossShardScCallOrSpecialTxIntraShard(t *testing.T) {
	t.Parallel()

	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString("bbbbbbbbbb" + suffixShard0)
	tx := createDefaultTx(sender, receiver, 50000)
	crossOrSpecial := isCrossShardScCallOrSpecialTx(0, 0, tx)
	assert.False(t, crossOrSpecial)
}

func Test_isCrossShardScCallOrSpecialTxCrossShardTransferTx(t *testing.T) {
	t.Parallel()

	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString("bbbbbbbbbb" + suffixShard1)
	tx := createDefaultTx(sender, receiver, 50000)
	crossOrSpecial := isCrossShardScCallOrSpecialTx(0, 1, tx)
	assert.False(t, crossOrSpecial)
}

func Test_isCrossShardScCallOrSpecialTxCrossShardScCall(t *testing.T) {
	t.Parallel()

	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	tx := createDefaultTx(sender, receiver, 50000)
	crossOrSpecial := isCrossShardScCallOrSpecialTx(0, 1, tx)
	assert.True(t, crossOrSpecial)
}

func Test_isCrossShardScCallOrSpecialTxCrossShardWithUsername(t *testing.T) {
	t.Parallel()

	sender := "aaaaaaaaaa.elrond"
	receiver := "bbbbbbbbbb.elrond"
	tx := createDefaultTx(nil, nil, 50000)
	tx.SndUserName = make([]byte, 64)
	tx.RcvUserName = make([]byte, 64)
	_ = hex.Encode(tx.RcvUserName, []byte(receiver))
	_ = hex.Encode(tx.SndUserName, []byte(sender))
	crossOrSpecial := isCrossShardScCallOrSpecialTx(0, 1, tx)
	assert.True(t, crossOrSpecial)
}

func Test_wouldExceedBlockSizeWithTxWithNormalTxWithinBlockLimit(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.isMaxBlockSizeReached = func(newMbs int, newTxs int) bool {
		return newMbs > 1 || newTxs > 1
	}
	mbb, _ := newMiniBlockBuilder(args)

	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString("bbbbbbbbbb" + suffixShard0)
	tx := createDefaultTx(sender, receiver, 50000)
	mbs := &block.MiniBlock{
		TxHashes: make([][]byte, 0),
	}
	result := mbb.wouldExceedBlockSizeWithTx(tx, 0, mbs)
	require.False(t, result)
}

func Test_wouldExceedBlockSizeWithTxWithNormalTxBlockLimitExceeded(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.isMaxBlockSizeReached = func(newMbs int, newTxs int) bool {
		return newTxs > 0 || newMbs > 0
	}
	mbb, _ := newMiniBlockBuilder(args)

	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString("bbbbbbbbbb" + suffixShard0)
	tx := createDefaultTx(sender, receiver, 50000)
	mbs := &block.MiniBlock{
		TxHashes: make([][]byte, 1),
	}
	result := mbb.wouldExceedBlockSizeWithTx(tx, 0, mbs)
	require.True(t, result)
}

func Test_wouldExceedBlockSizeWithTxWithSpecialTxWithinBlockLimit(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.isMaxBlockSizeReached = func(newMbs int, newTxs int) bool {
		return newMbs > 2 || newTxs > 4
	}
	mbb, _ := newMiniBlockBuilder(args)

	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	tx := createDefaultTx(sender, receiver, 50000)
	mbs := &block.MiniBlock{
		TxHashes: make([][]byte, 0),
	}
	result := mbb.wouldExceedBlockSizeWithTx(tx, 1, mbs)
	require.False(t, result)
}

func Test_wouldExceedBlockSizeWithTxWithSpecialTxBlockLimitExceededDueToSCR(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.isMaxBlockSizeReached = func(newMbs int, newTxs int) bool {
		return newTxs > 1
	}
	mbb, _ := newMiniBlockBuilder(args)

	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	tx := createDefaultTx(sender, receiver, 50000)
	mbs := &block.MiniBlock{
		TxHashes: make([][]byte, 1),
	}
	result := mbb.wouldExceedBlockSizeWithTx(tx, 1, mbs)
	require.True(t, result)
}

func Test_addTxAndUpdateBlockSize(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	mbb, _ := newMiniBlockBuilder(args)
	sender, _ := hex.DecodeString("aaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString("bbbbb" + suffixShard0)
	tx := createDefaultTx(sender, receiver, 50000)
	senderShardID := uint32(0)
	receiverShardID := uint32(0)
	wtx := createWrappedTransaction(tx, senderShardID, receiverShardID)

	require.NotNil(t, mbb.miniBlocks[receiverShardID])
	nbTxHashes := len(mbb.miniBlocks[receiverShardID].TxHashes)

	mbb.addTxAndUpdateBlockSize(tx, wtx)
	require.Equal(t, nbTxHashes+1, len(mbb.miniBlocks[receiverShardID].TxHashes))
	require.Equal(t, uint32(1), mbb.stats.numTxsAdded)
	require.Equal(t, uint32(0), mbb.stats.numCrossShardSCCallsOrSpecialTxs)
}

func Test_addTxAndUpdateBlockSizeCrossShardSmartContractCall(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	mbb, _ := newMiniBlockBuilder(args)
	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	tx := createDefaultTx(sender, receiver, 50000)
	senderShardID := uint32(0)
	receiverShardID := uint32(1)
	wtx := createWrappedTransaction(tx, senderShardID, receiverShardID)

	require.NotNil(t, mbb.miniBlocks[receiverShardID])
	nbTxHashes := len(mbb.miniBlocks[receiverShardID].TxHashes)

	mbb.addTxAndUpdateBlockSize(tx, wtx)
	require.Equal(t, nbTxHashes+1, len(mbb.miniBlocks[receiverShardID].TxHashes))
	require.Equal(t, uint32(1), mbb.stats.numTxsAdded)
	require.Equal(t, uint32(1), mbb.stats.numCrossShardSCCallsOrSpecialTxs)
}

func createDefaultMiniBlockBuilderArgs() miniBlocksBuilderArgs {
	return miniBlocksBuilderArgs{
		gasTracker: gasTracker{
			shardCoordinator: &testscommon.ShardsCoordinatorMock{
				NoShards:     3,
				CurrentShard: 0,
				ComputeIdCalled: func(address []byte) uint32 {
					return 0
				},
				SelfIDCalled: func() uint32 {
					return 0
				},
			},
			economicsFee: &economicsmocks.EconomicsHandlerMock{},
			gasHandler:   &testscommon.GasHandlerStub{},
		},
		accounts: &testscommon.AccountsStub{},
		accountTxsShards: &accountTxsShards{
			accountsInfo: make(map[string]*txShardInfo),
			RWMutex:      sync.RWMutex{},
		},
		blockSizeComputation:      &testscommon.BlockSizeComputationStub{},
		balanceComputationHandler: &testscommon.BalanceComputationStub{},
		haveTime:                  haveTimeTrue,
		isShardStuck:              isShardStuckFalse,
		isMaxBlockSizeReached:     isMaxBlockSizeReachedFalse,
		getTxMaxTotalCost: func(txHandler data.TransactionHandler) *big.Int {
			txMaxTotalCost, _ := big.NewInt(0).SetString("1500000000", 0)
			return txMaxTotalCost
		},
	}
}

func createWrappedTransaction(
	tx *transaction.Transaction,
	senderShardID uint32,
	receiverShardID uint32,
) *txcache.WrappedTransaction {
	hasher := &testscommon.HasherMock{}
	marshaller := &testscommon.MarshalizerMock{}
	txMarshalled, _ := marshaller.Marshal(tx)
	txHash := hasher.Compute(string(txMarshalled))

	return &txcache.WrappedTransaction{
		Tx:                   tx,
		TxHash:               txHash,
		SenderShardID:        senderShardID,
		ReceiverShardID:      receiverShardID,
		Size:                 int64(len(txMarshalled)),
		TxFeeScoreNormalized: 10,
	}
}
