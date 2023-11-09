package preprocess

import (
	"encoding/hex"
	"errors"
	"math/big"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
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

func Test_MiniBlocksBuilderUpdateAccountShardsInfo(t *testing.T) {
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
	penalize := uint64(0)
	mbb.gasInfo.gasConsumedByMiniBlocksInSenderShard = initGas
	mbb.gasInfo.totalGasConsumedInSelfShard = initGas
	mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID] = initGas
	mbb.handleGasRefund(wtx, refund, penalize)

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
	penalize := uint64(20000)
	mbb.gasInfo.gasConsumedByMiniBlocksInSenderShard = initGas
	mbb.gasInfo.totalGasConsumedInSelfShard = initGas
	mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID] = initGas
	mbb.handleGasRefund(wtx, refund, penalize)

	require.Equal(t, initGas, mbb.gasInfo.gasConsumedByMiniBlocksInSenderShard)
	require.Equal(t, initGas, mbb.gasInfo.totalGasConsumedInSelfShard)
	require.Equal(t, initGas, mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID])
}

func Test_MiniBlocksBuilderHandleFailedTransactionAddsToCount(t *testing.T) {
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

func Test_MiniBlocksBuilderHandleCrossShardScCallOrSpecialTx(t *testing.T) {
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

func Test_MiniBlocksBuilderWouldExceedBlockSizeWithTxWithNormalTxWithinBlockLimit(t *testing.T) {
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

func Test_MiniBlocksBuilderWouldExceedBlockSizeWithTxWithNormalTxBlockLimitExceeded(t *testing.T) {
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

func Test_MiniBlocksBuilderWouldExceedBlockSizeWithTxWithSpecialTxWithinBlockLimit(t *testing.T) {
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

func Test_MiniBlocksBuilderWouldExceedBlockSizeWithTxWithSpecialTxBlockLimitExceededDueToSCR(t *testing.T) {
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

func Test_MiniBlocksBuilderAddTxAndUpdateBlockSize(t *testing.T) {
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

func Test_MiniBlocksBuilderAddTxAndUpdateBlockSizeCrossShardSmartContractCall(t *testing.T) {
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

func Test_MiniBlocksBuilderHandleBadTransactionWithSkip(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	mbb, _ := newMiniBlockBuilder(args)
	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	tx := createDefaultTx(sender, receiver, 50000)
	senderShardID := uint32(0)
	receiverShardID := uint32(1)
	wtx := createWrappedTransaction(tx, senderShardID, receiverShardID)
	gasInfo := gasConsumedInfo{
		prevGasConsumedInReceiverShard:        15,
		gasConsumedByMiniBlocksInSenderShard:  10,
		gasConsumedByMiniBlockInReceiverShard: 20,
		totalGasConsumedInSelfShard:           30,
	}
	prevGasInfo := gasConsumedInfo{
		prevGasConsumedInReceiverShard:        5,
		gasConsumedByMiniBlocksInSenderShard:  5,
		gasConsumedByMiniBlockInReceiverShard: 10,
		totalGasConsumedInSelfShard:           20,
	}
	mbb.gasInfo = gasInfo
	mbb.prevGasInfo = prevGasInfo

	mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID] = mbb.gasInfo.gasConsumedByMiniBlockInReceiverShard
	mbb.handleBadTransaction(process.ErrHigherNonceInTransaction, wtx, tx)

	require.Equal(t, tx.SndAddr, mbb.senderToSkip)
	require.Equal(t, prevGasInfo, mbb.gasInfo)
	require.Equal(t, gasInfo.prevGasConsumedInReceiverShard, mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID])
}

func Test_MiniBlocksBuilderHandleBadTransactionNoSkip(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	mbb, _ := newMiniBlockBuilder(args)
	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	tx := createDefaultTx(sender, receiver, 50000)
	senderShardID := uint32(0)
	receiverShardID := uint32(1)
	wtx := createWrappedTransaction(tx, senderShardID, receiverShardID)
	gasInfo := gasConsumedInfo{
		gasConsumedByMiniBlocksInSenderShard:  10,
		gasConsumedByMiniBlockInReceiverShard: 20,
		totalGasConsumedInSelfShard:           30,
	}
	prevGasInfo := gasConsumedInfo{
		gasConsumedByMiniBlocksInSenderShard:  5,
		gasConsumedByMiniBlockInReceiverShard: 10,
		totalGasConsumedInSelfShard:           20,
	}

	mbb.prevGasInfo = prevGasInfo
	mbb.gasInfo = gasInfo
	mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID] = mbb.gasInfo.gasConsumedByMiniBlockInReceiverShard
	mbb.handleBadTransaction(process.ErrAccountNotFound, wtx, tx)

	require.Equal(t, []byte(""), mbb.senderToSkip)
	require.Equal(t, prevGasInfo, mbb.gasInfo)
	require.Equal(t, gasInfo.prevGasConsumedInReceiverShard, mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID])
}

func Test_MiniBlocksBuilderShouldSenderBeSkippedNoConfiguredSenderToSkip(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	mbb, _ := newMiniBlockBuilder(args)
	mbb.senderToSkip = []byte("")
	shouldSkip := mbb.shouldSenderBeSkipped([]byte("currentSender"))
	require.False(t, shouldSkip)
}

func Test_MiniBlocksBuilderShouldSenderBeSkippedDifferentConfiguredSenderToSkip(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	mbb, _ := newMiniBlockBuilder(args)
	mbb.senderToSkip = []byte("differentSender")
	shouldSkip := mbb.shouldSenderBeSkipped([]byte("currentSender"))
	require.False(t, shouldSkip)
}

func Test_MiniBlocksBuilderShouldSenderBeSkippedSenderConfiguredToSkip(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	mbb, _ := newMiniBlockBuilder(args)
	mbb.senderToSkip = []byte("currentSender")
	shouldSkip := mbb.shouldSenderBeSkipped([]byte("currentSender"))
	require.True(t, shouldSkip)
}

func Test_MiniBlocksBuilderAccountGasForTxComputeGasProvidedWithErr(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	expectedErr := errors.New("expectedError")
	args.gasTracker = gasTracker{
		shardCoordinator: args.gasTracker.shardCoordinator,
		economicsFee:     args.gasTracker.economicsFee,
		gasHandler: &testscommon.GasHandlerStub{
			RemoveGasProvidedCalled:  func(hashes [][]byte) {},
			RemoveGasRefundedCalled:  func(hashes [][]byte) {},
			RemoveGasPenalizedCalled: func(hashes [][]byte) {},
			ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, expectedErr
			},
		},
	}
	mbb, _ := newMiniBlockBuilder(args)
	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	tx := createDefaultTx(sender, receiver, 50000)
	senderShardID := uint32(0)
	receiverShardID := uint32(1)
	wtx := createWrappedTransaction(tx, senderShardID, receiverShardID)
	expectedGasConsumedInReceiverShard := uint64(10)
	mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID] = expectedGasConsumedInReceiverShard

	_, err := mbb.accountGasForTx(tx, wtx)
	require.Equal(t, expectedErr, err)
	require.Equal(t, expectedGasConsumedInReceiverShard, mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID])
}

func Test_MiniBlocksBuilderAccountGasForTxComputeGasProvidedOK(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	gasProvidedByTxInReceiverShard := uint64(20)
	gasProvidedByTxInSenderShard := uint64(10)
	args.gasTracker = gasTracker{
		shardCoordinator: args.gasTracker.shardCoordinator,
		economicsFee:     args.gasTracker.economicsFee,
		gasHandler: &testscommon.GasHandlerStub{
			RemoveGasProvidedCalled: func(hashes [][]byte) {},
			RemoveGasRefundedCalled: func(hashes [][]byte) {},
			ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return gasProvidedByTxInSenderShard, gasProvidedByTxInReceiverShard, nil
			},
		},
	}
	mbb, _ := newMiniBlockBuilder(args)
	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	tx := createDefaultTx(sender, receiver, 50000)
	senderShardID := uint32(0)
	receiverShardID := uint32(1)
	wtx := createWrappedTransaction(tx, senderShardID, receiverShardID)

	gasConsumedByMiniBlockInReceiverShard := uint64(100)
	gasConsumedByMiniBlocksInSenderShard := uint64(200)
	mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID] = gasConsumedByMiniBlockInReceiverShard

	mbb.gasInfo = gasConsumedInfo{
		gasConsumedByMiniBlocksInSenderShard:  gasConsumedByMiniBlocksInSenderShard,
		gasConsumedByMiniBlockInReceiverShard: gasConsumedByMiniBlockInReceiverShard,
		totalGasConsumedInSelfShard:           gasConsumedByMiniBlocksInSenderShard,
	}

	_, err := mbb.accountGasForTx(tx, wtx)

	expectedConsumedReceiverShard := gasConsumedByMiniBlockInReceiverShard + gasProvidedByTxInReceiverShard
	expectedConsumedSenderShard := gasConsumedByMiniBlocksInSenderShard + gasProvidedByTxInSenderShard
	require.Nil(t, err)
	require.Equal(t, expectedConsumedReceiverShard, mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID])
	require.Equal(t, expectedConsumedReceiverShard, mbb.gasInfo.gasConsumedByMiniBlockInReceiverShard)
	require.Equal(t, expectedConsumedSenderShard, mbb.gasInfo.gasConsumedByMiniBlocksInSenderShard)
	require.Equal(t, expectedConsumedSenderShard, mbb.gasInfo.totalGasConsumedInSelfShard)
}

func Test_MiniBlocksBuilderAccountHasEnoughBalanceAddressNotSet(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.balanceComputationHandler = &testscommon.BalanceComputationStub{
		IsAddressSetCalled: func(address []byte) bool {
			return false
		},
	}

	mbb, _ := newMiniBlockBuilder(args)
	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	tx := createDefaultTx(sender, receiver, 50000)

	enoughBalance := mbb.accountHasEnoughBalance(tx)
	require.True(t, enoughBalance)
}

func Test_MiniBlocksBuilderAccountHasEnoughBalanceAddressSetNotEnoughBalance(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.balanceComputationHandler = &testscommon.BalanceComputationStub{
		IsAddressSetCalled: func(address []byte) bool {
			return true
		},
		AddressHasEnoughBalanceCalled: func(address []byte, value *big.Int) bool {
			return false
		},
	}

	mbb, _ := newMiniBlockBuilder(args)
	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	tx := createDefaultTx(sender, receiver, 50000)

	enoughBalance := mbb.accountHasEnoughBalance(tx)
	require.False(t, enoughBalance)
}

func Test_MiniBlocksBuilderAccountHasEnoughBalanceAddressSetEnoughBalance(t *testing.T) {
	t.Parallel()

	args := createDefaultMiniBlockBuilderArgs()
	args.balanceComputationHandler = &testscommon.BalanceComputationStub{
		IsAddressSetCalled: func(address []byte) bool {
			return true
		},
		AddressHasEnoughBalanceCalled: func(address []byte, value *big.Int) bool {
			return true
		},
	}

	mbb, _ := newMiniBlockBuilder(args)
	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	tx := createDefaultTx(sender, receiver, 50000)

	enoughBalance := mbb.accountHasEnoughBalance(tx)
	require.True(t, enoughBalance)
}

func Test_MiniBlocksBuilderCheckAddTransactionWrongTypeAssertion(t *testing.T) {
	t.Parallel()

	wtx := &txcache.WrappedTransaction{
		Tx:                   nil,
		TxHash:               nil,
		SenderShardID:        0,
		ReceiverShardID:      0,
		Size:                 0,
		TxFeeScoreNormalized: 0,
	}

	args := createDefaultMiniBlockBuilderArgs()
	mbb, _ := newMiniBlockBuilder(args)

	actions, tx := mbb.checkAddTransaction(wtx)
	require.False(t, actions.canAddTx)
	require.True(t, actions.canAddMore)
	require.Nil(t, tx)
}

func Test_MiniBlocksBuilderCheckAddTransactionNotEnoughTime(t *testing.T) {
	t.Parallel()

	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	txInitial := createDefaultTx(sender, receiver, 50000)
	senderShardID := uint32(0)
	receiverShardID := uint32(1)
	wtx := createWrappedTransaction(txInitial, senderShardID, receiverShardID)

	args := createDefaultMiniBlockBuilderArgs()
	args.haveTime = func() bool {
		return false
	}
	mbb, _ := newMiniBlockBuilder(args)
	actions, tx := mbb.checkAddTransaction(wtx)
	require.False(t, actions.canAddTx)
	require.False(t, actions.canAddMore)
	require.Equal(t, txInitial, tx)
}

func Test_MiniBlocksBuilderCheckAddTransactionInitializedMiniBlockNotFound(t *testing.T) {
	t.Parallel()

	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	txInitial := createDefaultTx(sender, receiver, 50000)
	senderShardID := uint32(0)
	receiverShardID := uint32(1)
	wtx := createWrappedTransaction(txInitial, senderShardID, receiverShardID)

	args := createDefaultMiniBlockBuilderArgs()

	mbb, _ := newMiniBlockBuilder(args)
	delete(mbb.miniBlocks, receiverShardID)

	actions, tx := mbb.checkAddTransaction(wtx)
	require.False(t, actions.canAddTx)
	require.True(t, actions.canAddMore)
	require.Equal(t, txInitial, tx)
}

func Test_MiniBlocksBuilderCheckAddTransactionExceedsBlockSize(t *testing.T) {
	t.Parallel()

	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	txInitial := createDefaultTx(sender, receiver, 50000)
	senderShardID := uint32(0)
	receiverShardID := uint32(1)
	wtx := createWrappedTransaction(txInitial, senderShardID, receiverShardID)

	args := createDefaultMiniBlockBuilderArgs()
	args.isMaxBlockSizeReached = func(i int, i2 int) bool {
		return true
	}
	mbb, _ := newMiniBlockBuilder(args)

	actions, tx := mbb.checkAddTransaction(wtx)
	require.False(t, actions.canAddTx)
	require.False(t, actions.canAddMore)
	require.Equal(t, txInitial, tx)
}

func Test_MiniBlocksBuilderCheckAddTransactionStuckShard(t *testing.T) {
	t.Parallel()

	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	txInitial := createDefaultTx(sender, receiver, 50000)
	senderShardID := uint32(0)
	receiverShardID := uint32(1)
	wtx := createWrappedTransaction(txInitial, senderShardID, receiverShardID)

	args := createDefaultMiniBlockBuilderArgs()
	args.isShardStuck = func(u uint32) bool {
		return true
	}
	mbb, _ := newMiniBlockBuilder(args)

	actions, tx := mbb.checkAddTransaction(wtx)
	require.False(t, actions.canAddTx)
	require.True(t, actions.canAddMore)
	require.Equal(t, txInitial, tx)
}

func Test_MiniBlocksBuilderCheckAddTransactionWithSenderSkip(t *testing.T) {
	t.Parallel()

	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	txInitial := createDefaultTx(sender, receiver, 50000)
	senderShardID := uint32(0)
	receiverShardID := uint32(1)
	wtx := createWrappedTransaction(txInitial, senderShardID, receiverShardID)

	args := createDefaultMiniBlockBuilderArgs()
	mbb, _ := newMiniBlockBuilder(args)
	mbb.senderToSkip = sender

	actions, tx := mbb.checkAddTransaction(wtx)
	require.False(t, actions.canAddTx)
	require.True(t, actions.canAddMore)
	require.Equal(t, txInitial, tx)
}

func Test_MiniBlocksBuilderCheckAddTransactionNotEnoughBalance(t *testing.T) {
	t.Parallel()

	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	txInitial := createDefaultTx(sender, receiver, 50000)
	senderShardID := uint32(0)
	receiverShardID := uint32(1)
	wtx := createWrappedTransaction(txInitial, senderShardID, receiverShardID)

	args := createDefaultMiniBlockBuilderArgs()
	args.balanceComputationHandler = &testscommon.BalanceComputationStub{
		IsAddressSetCalled: func(address []byte) bool {
			return true
		},
		AddressHasEnoughBalanceCalled: func(address []byte, value *big.Int) bool {
			return false
		},
	}
	mbb, _ := newMiniBlockBuilder(args)

	actions, tx := mbb.checkAddTransaction(wtx)
	require.False(t, actions.canAddTx)
	require.True(t, actions.canAddMore)
	require.Equal(t, txInitial, tx)
}

func Test_MiniBlocksBuilderCheckAddTransactionGasAccountingError(t *testing.T) {
	t.Parallel()

	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	txInitial := createDefaultTx(sender, receiver, 50000)
	senderShardID := uint32(0)
	receiverShardID := uint32(1)
	wtx := createWrappedTransaction(txInitial, senderShardID, receiverShardID)

	args := createDefaultMiniBlockBuilderArgs()
	expectedErr := errors.New("expectedError")
	args.gasTracker = gasTracker{
		shardCoordinator: args.gasTracker.shardCoordinator,
		economicsFee:     args.gasTracker.economicsFee,
		gasHandler: &testscommon.GasHandlerStub{
			RemoveGasProvidedCalled: func(hashes [][]byte) {},
			RemoveGasRefundedCalled: func(hashes [][]byte) {},
			ComputeGasProvidedByTxCalled: func(txSenderShardId uint32, txReceiverSharedId uint32, txHandler data.TransactionHandler) (uint64, uint64, error) {
				return 0, 0, expectedErr
			},
		},
	}
	mbb, _ := newMiniBlockBuilder(args)

	actions, tx := mbb.checkAddTransaction(wtx)
	require.False(t, actions.canAddTx)
	require.True(t, actions.canAddMore)
	require.Equal(t, txInitial, tx)
}

func Test_MiniBlocksBuilderCheckAddTransactionOK(t *testing.T) {
	t.Parallel()

	sender, _ := hex.DecodeString("aaaaaaaaaa" + suffixShard0)
	receiver, _ := hex.DecodeString(smartContractAddressStart + suffixShard1)
	txInitial := createDefaultTx(sender, receiver, 50000)
	senderShardID := uint32(0)
	receiverShardID := uint32(1)
	wtx := createWrappedTransaction(txInitial, senderShardID, receiverShardID)

	args := createDefaultMiniBlockBuilderArgs()
	mbb, _ := newMiniBlockBuilder(args)

	actions, tx := mbb.checkAddTransaction(wtx)
	require.True(t, actions.canAddTx)
	require.True(t, actions.canAddMore)
	require.Equal(t, txInitial, tx)
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
			economicsFee: &economicsmocks.EconomicsHandlerStub{},
			gasHandler: &testscommon.GasHandlerStub{
				RemoveGasProvidedCalled: func(hashes [][]byte) {
				},
				RemoveGasRefundedCalled: func(hashes [][]byte) {
				},
				RemoveGasPenalizedCalled: func(hashes [][]byte) {
				},
			},
		},
		accounts: &stateMock.AccountsStub{},
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
		getTotalGasConsumed: getTotalGasConsumedZero,
		txPool:              shardedDataCacherNotifier(),
	}
}

func createWrappedTransaction(
	tx *transaction.Transaction,
	senderShardID uint32,
	receiverShardID uint32,
) *txcache.WrappedTransaction {
	hasher := &hashingMocks.HasherMock{}
	marshaller := &marshallerMock.MarshalizerMock{}
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
