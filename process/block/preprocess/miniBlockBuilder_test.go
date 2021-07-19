package preprocess

import (
	"math/big"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
	"github.com/stretchr/testify/require"
)

func Test_newMiniBlockBuilderWithError(t *testing.T) {
	args := createDefaultMiniBlockBuilderArgs()
	args.accounts = nil

	mbb, err := newMiniBlockBuilder(args)
	require.Equal(t, process.ErrNilAccountsAdapter, err)
	require.Nil(t, mbb)
}

func Test_newMiniBlockBuilderOK(t *testing.T) {
	args := createDefaultMiniBlockBuilderArgs()

	mbb, err := newMiniBlockBuilder(args)
	require.Nil(t, err)
	require.NotNil(t, mbb)
}

func Test_checkMiniBlocksBuilderArgsNilShardCoordinatorShouldErr(t *testing.T) {
	args := createDefaultMiniBlockBuilderArgs()
	args.gasTracker.shardCoordinator = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilShardCoordinator, err)
}

func Test_checkMiniBlocksBuilderArgsNilGasHandlerShouldErr(t *testing.T) {
	args := createDefaultMiniBlockBuilderArgs()
	args.gasTracker.gasHandler = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilGasHandler, err)
}

func Test_checkMiniBlocksBuilderArgsNilFeeHandlerShouldErr(t *testing.T) {
	args := createDefaultMiniBlockBuilderArgs()
	args.gasTracker.economicsFee = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilEconomicsFeeHandler, err)
}

func Test_checkMiniBlocksBuilderArgsNilBlockSizeComputationHandlerShouldErr(t *testing.T) {
	args := createDefaultMiniBlockBuilderArgs()
	args.blockSizeComputation = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilBlockSizeComputationHandler, err)
}

func Test_checkMiniBlocksBuilderArgsNilAccountsTxsPerShardsShouldErr(t *testing.T) {
	args := createDefaultMiniBlockBuilderArgs()
	args.accountTxsShards = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilAccountTxsPerShard, err)
}

func Test_checkMiniBlocksBuilderArgsNilBalanceComputationHandlerShouldErr(t *testing.T) {
	args := createDefaultMiniBlockBuilderArgs()
	args.balanceComputationHandler = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilBalanceComputationHandler, err)
}

func Test_checkMiniBlocksBuilderArgsNilHaveTimeHandlerShouldErr(t *testing.T) {
	args := createDefaultMiniBlockBuilderArgs()
	args.haveTime = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilHaveTimeHandler, err)
}

func Test_checkMiniBlocksBuilderArgsNilShardStuckHandlerShouldErr(t *testing.T) {
	args := createDefaultMiniBlockBuilderArgs()
	args.isShardStuck = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilIsShardStuckHandler, err)
}

func Test_checkMiniBlocksBuilderArgsNilMaxBlockSizeReachedHandlerShouldErr(t *testing.T) {
	args := createDefaultMiniBlockBuilderArgs()
	args.isMaxBlockSizeReached = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilIsMaxBlockSizeReachedHandler, err)
}

func Test_checkMiniBlocksBuilderArgsNilTxMaxTotalCostHandlerShouldErr(t *testing.T) {
	args := createDefaultMiniBlockBuilderArgs()
	args.getTxMaxTotalCost = nil

	err := checkMiniBlocksBuilderArgs(args)
	require.Equal(t, process.ErrNilTxMaxTotalCostHandler, err)
}

func Test_checkMiniBlocksBuilderArgsOK(t *testing.T) {
	args := createDefaultMiniBlockBuilderArgs()

	err := checkMiniBlocksBuilderArgs(args)
	require.Nil(t, err)
}

func Test_updateAccountShardsInfo(t *testing.T) {
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
	args := createDefaultMiniBlockBuilderArgs()
	mbb, _ := newMiniBlockBuilder(args)

	mbb.handleCrossShardScCallOrSpecialTx()
	require.True(t, mbb.stats.firstCrossShardScCallOrSpecialTxFound)
	require.Equal(t, uint32(1), mbb.stats.numCrossShardSCCallsOrSpecialTxs)

	mbb.handleCrossShardScCallOrSpecialTx()
	require.True(t, mbb.stats.firstCrossShardScCallOrSpecialTxFound)
	require.Equal(t, uint32(2), mbb.stats.numCrossShardSCCallsOrSpecialTxs)
}

func Test_isCrossShardScCallOrSpecialTx(t *testing.T){

}

func createDefaultMiniBlockBuilderArgs() miniBlocksBuilderArgs {
	return miniBlocksBuilderArgs{
		gasTracker: gasTracker{
			shardCoordinator: &testscommon.ShardsCoordinatorMock{},
			economicsFee:     &economicsmocks.EconomicsHandlerMock{},
			gasHandler:       &testscommon.GasHandlerStub{},
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
