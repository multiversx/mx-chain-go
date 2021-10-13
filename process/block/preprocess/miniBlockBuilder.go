package preprocess

import (
	"bytes"
	"errors"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
)

type miniBlocksBuilderArgs struct {
	gasTracker                gasTracker
	accounts                  state.AccountsAdapter
	accountTxsShards          *accountTxsShards
	blockSizeComputation      BlockSizeComputationHandler
	balanceComputationHandler BalanceComputationHandler
	haveTime                  func() bool
	isShardStuck              func(uint32) bool
	isMaxBlockSizeReached     func(int, int) bool
	getTxMaxTotalCost         func(txHandler data.TransactionHandler) *big.Int
	getTotalGasConsumed       func() uint64
	txPool                    dataRetriever.ShardedDataCacherNotifier
}

type miniBlockBuilderStats struct {
	numTxsAdded                           uint32
	numTxsBad                             uint32
	numTxsSkipped                         uint32
	numTxsFailed                          uint32
	numTxsWithInitialBalanceConsumed      uint32
	numCrossShardSCCallsOrSpecialTxs      uint32
	numCrossShardTxsWithTooMuchGas        uint32
	totalProcessingTime                   time.Duration
	totalGasComputeTime                   time.Duration
	firstInvalidTxFound                   bool
	firstCrossShardScCallOrSpecialTxFound bool
}

type miniBlocksBuilder struct {
	gasTracker
	accounts                   state.AccountsAdapter
	accountTxsShards           *accountTxsShards
	balanceComputationHandler  BalanceComputationHandler
	blockSizeComputation       BlockSizeComputationHandler
	gasConsumedInReceiverShard map[uint32]uint64
	gasInfo                    gasConsumedInfo
	prevGasInfo                gasConsumedInfo
	senderToSkip               []byte
	miniBlocks                 map[uint32]*block.MiniBlock
	haveTime                   func() bool
	isShardStuck               func(uint32) bool
	isMaxBlockSizeReached      func(int, int) bool
	getTxMaxTotalCost          func(txHandler data.TransactionHandler) *big.Int
	stats                      miniBlockBuilderStats
	txPool                     dataRetriever.ShardedDataCacherNotifier
}

func newMiniBlockBuilder(args miniBlocksBuilderArgs) (*miniBlocksBuilder, error) {
	err := checkMiniBlocksBuilderArgs(args)
	if err != nil {
		return nil, err
	}

	return &miniBlocksBuilder{
		gasTracker:                 args.gasTracker,
		accounts:                   args.accounts,
		accountTxsShards:           args.accountTxsShards,
		balanceComputationHandler:  args.balanceComputationHandler,
		blockSizeComputation:       args.blockSizeComputation,
		miniBlocks:                 initializeMiniBlocksMap(args.gasTracker.shardCoordinator),
		gasConsumedInReceiverShard: make(map[uint32]uint64),
		haveTime:                   args.haveTime,
		isShardStuck:               args.isShardStuck,
		isMaxBlockSizeReached:      args.isMaxBlockSizeReached,
		getTxMaxTotalCost:          args.getTxMaxTotalCost,
		gasInfo: gasConsumedInfo{
			gasConsumedByMiniBlocksInSenderShard:  0,
			gasConsumedByMiniBlockInReceiverShard: 0,
			totalGasConsumedInSelfShard:           args.getTotalGasConsumed(),
		},
		stats:        miniBlockBuilderStats{},
		senderToSkip: []byte(""),
		txPool:       args.txPool,
	}, nil
}

func checkMiniBlocksBuilderArgs(args miniBlocksBuilderArgs) error {
	if check.IfNil(args.gasTracker.shardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(args.gasTracker.gasHandler) {
		return process.ErrNilGasHandler
	}
	if check.IfNil(args.gasTracker.economicsFee) {
		return process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.accounts) {
		return process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.balanceComputationHandler) {
		return process.ErrNilBalanceComputationHandler
	}
	if check.IfNil(args.blockSizeComputation) {
		return process.ErrNilBlockSizeComputationHandler
	}
	if check.IfNil(args.txPool) {
		return process.ErrNilTransactionPool
	}
	if args.accountTxsShards == nil {
		return process.ErrNilAccountTxsPerShard
	}
	if args.haveTime == nil {
		return process.ErrNilHaveTimeHandler
	}
	if args.isShardStuck == nil {
		return process.ErrNilIsShardStuckHandler
	}
	if args.isMaxBlockSizeReached == nil {
		return process.ErrNilIsMaxBlockSizeReachedHandler
	}
	if args.getTxMaxTotalCost == nil {
		return process.ErrNilTxMaxTotalCostHandler
	}

	return nil
}

func (mbb *miniBlocksBuilder) updateAccountShardsInfo(tx *transaction.Transaction, wtx *txcache.WrappedTransaction) {
	mbb.accountTxsShards.Lock()
	mbb.accountTxsShards.accountsInfo[string(tx.GetSndAddr())] = &txShardInfo{
		senderShardID:   wtx.SenderShardID,
		receiverShardID: wtx.ReceiverShardID,
	}
	mbb.accountTxsShards.Unlock()
}

// function returns through the first parameter if the given transaction can be added to the miniBlock
// second return values returns an error in case no more transactions can be added to the miniBlocks
func (mbb *miniBlocksBuilder) checkAddTransaction(wtx *txcache.WrappedTransaction) (canAddTx bool, canAddMore bool, tx *transaction.Transaction) {
	tx, ok := wtx.Tx.(*transaction.Transaction)
	if !ok {
		log.Debug("wrong type assertion",
			"hash", wtx.TxHash,
			"sender shard", wtx.SenderShardID,
			"receiver shard", wtx.ReceiverShardID)
		return false, true, nil
	}

	if !mbb.haveTime() {
		log.Debug("time is out")
		return false, false, tx
	}

	receiverShardID := wtx.ReceiverShardID
	miniBlock, ok := mbb.miniBlocks[receiverShardID]
	if !ok {
		log.Debug("mini block is not created", "shard", receiverShardID)
		return false, true, tx
	}

	if mbb.wouldExceedBlockSizeWithTx(tx, receiverShardID, miniBlock) {
		log.Debug("max txs accepted in one block is reached", "num txs added", mbb.stats.numTxsAdded)
		return false, false, tx
	}

	if mbb.isShardStuck(receiverShardID) {
		log.Trace("shard is stuck", "shard", receiverShardID)
		return false, true, tx
	}

	if mbb.shouldSenderBeSkipped(tx.GetSndAddr()) {
		return false, true, tx
	}

	if !mbb.accountHasEnoughBalance(tx) {
		return false, true, tx
	}

	if mbb.accountGasForTx(tx, wtx) != nil {
		return false, true, tx
	}

	return true, true, tx
}

func (mbb *miniBlocksBuilder) wouldExceedBlockSizeWithTx(tx *transaction.Transaction, receiverShardID uint32, miniBlock *block.MiniBlock) bool {
	numNewMiniBlocks := 0
	if len(miniBlock.TxHashes) == 0 {
		numNewMiniBlocks = 1
	}
	numNewTxs := 1

	if isCrossShardScCallOrSpecialTx(receiverShardID, mbb.shardCoordinator.SelfId(), tx) {
		if !mbb.stats.firstCrossShardScCallOrSpecialTxFound {
			numNewMiniBlocks++
		}
		numNewTxs += common.AdditionalScrForEachScCallOrSpecialTx
	}

	return mbb.isMaxBlockSizeReached(numNewMiniBlocks, numNewTxs)
}

func isCrossShardScCallOrSpecialTx(receiverShardID uint32, selfShardID uint32, tx *transaction.Transaction) bool {
	return receiverShardID != selfShardID && (core.IsSmartContractAddress(tx.RcvAddr) || len(tx.RcvUserName) > 0)
}

func (mbb *miniBlocksBuilder) shouldSenderBeSkipped(address []byte) bool {
	if len(mbb.senderToSkip) == 0 {
		return false
	}
	if bytes.Equal(mbb.senderToSkip, address) {
		mbb.stats.numTxsSkipped++
		return true
	}
	return false
}

func initializeMiniBlocksMap(shardCoordinator sharding.Coordinator) map[uint32]*block.MiniBlock {
	miniBlocksMap := make(map[uint32]*block.MiniBlock)
	for shardID := uint32(0); shardID < shardCoordinator.NumberOfShards(); shardID++ {
		miniBlocksMap[shardID] = createEmptyMiniBlock(shardCoordinator.SelfId(), shardID, block.TxBlock, nil)
	}

	miniBlocksMap[core.MetachainShardId] = createEmptyMiniBlock(shardCoordinator.SelfId(), core.MetachainShardId, block.TxBlock, nil)

	return miniBlocksMap
}

func createEmptyMiniBlock(
	senderShardID uint32,
	receiverShardID uint32,
	blockType block.Type,
	reserved []byte,
) *block.MiniBlock {

	miniBlock := &block.MiniBlock{
		Type:            blockType,
		SenderShardID:   senderShardID,
		ReceiverShardID: receiverShardID,
		TxHashes:        make([][]byte, 0),
		Reserved:        reserved,
	}

	return miniBlock
}

func (mbb *miniBlocksBuilder) accountHasEnoughBalance(tx *transaction.Transaction) bool {
	txMaxTotalCost := big.NewInt(0)
	isAddressSet := mbb.balanceComputationHandler.IsAddressSet(tx.GetSndAddr())
	if isAddressSet {
		txMaxTotalCost = mbb.getTxMaxTotalCost(tx)
		addressHasEnoughBalance := mbb.balanceComputationHandler.AddressHasEnoughBalance(tx.GetSndAddr(), txMaxTotalCost)
		if !addressHasEnoughBalance {
			mbb.stats.numTxsWithInitialBalanceConsumed++
			return false
		}
	}

	return true
}

func (mbb *miniBlocksBuilder) accountGasForTx(tx *transaction.Transaction, wtx *txcache.WrappedTransaction) error {
	mbb.prevGasInfo = mbb.gasInfo
	mbb.gasInfo.gasConsumedByMiniBlockInReceiverShard = mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID]
	startTime := time.Now()
	gasConsumedByTxInSelfShard, err := mbb.computeGasConsumed(
		wtx.SenderShardID,
		wtx.ReceiverShardID,
		tx,
		wtx.TxHash,
		&mbb.gasInfo)
	elapsedTime := time.Since(startTime)
	mbb.stats.totalGasComputeTime += elapsedTime
	if err != nil {
		log.Trace("miniBlocksBuilder.accountGasForTx", "error", err)
		isTxTargetedForDeletion := errors.Is(err, process.ErrMaxGasLimitPerOneTxInReceiverShardIsReached)
		if isTxTargetedForDeletion {
			mbb.stats.numCrossShardTxsWithTooMuchGas++
			strCache := process.ShardCacherIdentifier(wtx.SenderShardID, wtx.ReceiverShardID)
			mbb.txPool.RemoveData(wtx.TxHash, strCache)
		}
		return err
	}

	mbb.gasHandler.SetGasConsumed(gasConsumedByTxInSelfShard, wtx.TxHash)
	mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID] = mbb.gasInfo.gasConsumedByMiniBlockInReceiverShard
	return nil
}

func (mbb *miniBlocksBuilder) handleBadTransaction(err error, wtx *txcache.WrappedTransaction, tx *transaction.Transaction) {
	if errors.Is(err, process.ErrHigherNonceInTransaction) {
		mbb.senderToSkip = tx.GetSndAddr()
	}

	mbb.gasHandler.RemoveGasConsumed([][]byte{wtx.TxHash})
	mbb.gasHandler.RemoveGasRefunded([][]byte{wtx.TxHash})
	mbb.gasHandler.RemoveGasPenalized([][]byte{wtx.TxHash})

	mbb.gasInfo = mbb.prevGasInfo
	mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID] = mbb.prevGasInfo.gasConsumedByMiniBlockInReceiverShard
	mbb.stats.numTxsBad++
}

func (mbb *miniBlocksBuilder) handleGasRefund(wtx *txcache.WrappedTransaction, gasRefunded uint64, gasPenalized uint64) {
	if wtx.SenderShardID == wtx.ReceiverShardID {
		gasToBeSubtracted := gasRefunded + gasPenalized
		shouldDoTheSubtraction := gasToBeSubtracted <= mbb.gasInfo.gasConsumedByMiniBlocksInSenderShard &&
			gasToBeSubtracted <= mbb.gasInfo.totalGasConsumedInSelfShard &&
			gasToBeSubtracted <= mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID]
		if shouldDoTheSubtraction {
			mbb.gasInfo.gasConsumedByMiniBlocksInSenderShard -= gasToBeSubtracted
			mbb.gasInfo.totalGasConsumedInSelfShard -= gasToBeSubtracted
			mbb.gasConsumedInReceiverShard[wtx.ReceiverShardID] -= gasToBeSubtracted
		}
	}
}

func (mbb *miniBlocksBuilder) handleFailedTransaction() {
	if !mbb.stats.firstInvalidTxFound {
		mbb.stats.firstInvalidTxFound = true
		mbb.blockSizeComputation.AddNumMiniBlocks(1)
	}

	mbb.blockSizeComputation.AddNumTxs(1)
	mbb.stats.numTxsFailed++
}

func (mbb *miniBlocksBuilder) addTxAndUpdateBlockSize(tx *transaction.Transaction, wtx *txcache.WrappedTransaction) {
	miniBlock := mbb.miniBlocks[wtx.ReceiverShardID]

	if len(miniBlock.TxHashes) == 0 {
		mbb.blockSizeComputation.AddNumMiniBlocks(1)
	}

	miniBlock.TxHashes = append(miniBlock.TxHashes, wtx.TxHash)
	mbb.blockSizeComputation.AddNumTxs(1)
	if isCrossShardScCallOrSpecialTx(wtx.ReceiverShardID, mbb.shardCoordinator.SelfId(), tx) {
		mbb.handleCrossShardScCallOrSpecialTx()
	}
	mbb.stats.numTxsAdded++
}

func (mbb *miniBlocksBuilder) handleCrossShardScCallOrSpecialTx() {
	if !mbb.stats.firstCrossShardScCallOrSpecialTxFound {
		mbb.stats.firstCrossShardScCallOrSpecialTxFound = true
		mbb.blockSizeComputation.AddNumMiniBlocks(1)
	}
	//we need to increment this as to account for the corresponding SCR hash
	mbb.blockSizeComputation.AddNumTxs(common.AdditionalScrForEachScCallOrSpecialTx)
	mbb.stats.numCrossShardSCCallsOrSpecialTxs++
}
