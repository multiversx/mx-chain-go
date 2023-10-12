package preprocess

import (
	"errors"
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage/txcache"
)

type sovereignChainTransactions struct {
	*transactions
}

// NewSovereignChainTransactionPreprocessor creates a new sovereign chain transaction preprocessor object
func NewSovereignChainTransactionPreprocessor(
	transactions *transactions,
) (*sovereignChainTransactions, error) {
	if transactions == nil {
		return nil, process.ErrNilPreProcessor
	}

	sct := &sovereignChainTransactions{
		transactions,
	}

	sct.scheduledTXContinueFunc = sct.shouldContinueProcessingScheduledTx
	sct.shouldSkipMiniBlockFunc = sct.shouldSkipMiniBlock
	sct.isTransactionEligibleForExecutionFunc = sct.isTransactionEligibleForExecution

	return sct, nil
}

// ProcessBlockTransactions processes all the transactions "from me" from the given body
func (sct *sovereignChainTransactions) ProcessBlockTransactions(
	header data.HeaderHandler,
	body *block.Body,
	haveTime func() bool,
) (block.MiniBlockSlice, error) {
	calculatedMiniBlocks, _, err := sct.processTxsFromMe(body, haveTime, header.GetPrevRandSeed())
	return calculatedMiniBlocks, err
}

// CreateAndProcessMiniBlocks creates mini blocks from the selected transactions
func (sct *sovereignChainTransactions) CreateAndProcessMiniBlocks(haveTime func() bool, randomness []byte) (block.MiniBlockSlice, error) {
	startTime := time.Now()

	//TODO: Check through "allin system tests" if this value of 2x (normal txs + scheduled txs) is enough or could be raised,
	//as with the new mechanism of proposing and parallel processing for proposer and validators, we are not limited anymore
	//for a subround block duration of maximum 25% from the whole round. Actually the processing time could / will go up to 60% - 80%.

	// TODO: maybe re-add the 2.0 multiplier
	gasBandwidth := sct.economicsFee.MaxGasLimitPerBlock(sct.shardCoordinator.SelfId())

	sortedTxs, err := sct.computeSortedTxs(sct.shardCoordinator.SelfId(), sct.shardCoordinator.SelfId(), randomness)
	elapsedTime := time.Since(startTime)
	if err != nil {
		log.Debug("computeSortedTxs", "error", err.Error())
		return make(block.MiniBlockSlice, 0), nil
	}

	if len(sortedTxs) == 0 {
		log.Trace("no transaction found after computeSortedTxs",
			"time [s]", elapsedTime,
		)
		return make(block.MiniBlockSlice, 0), nil
	}

	if !haveTime() {
		log.Debug("time is up after computeSortedTxs",
			"num txs", len(sortedTxs),
			"time [s]", elapsedTime,
		)
		return make(block.MiniBlockSlice, 0), nil
	}

	log.Debug("elapsed time to computeSortedTxs",
		"num txs", len(sortedTxs),
		"time [s]", elapsedTime,
	)

	selectedTxs, remainingTxs, gasEstimation := sct.addTxsWithinBandwidth(nil, sortedTxs, 0, gasBandwidth)
	log.Debug("maxGasUsage by proposer in BIL", "gasBandwidth", gasBandwidth/process.Billion, "gasEstimation", gasEstimation/process.Billion, "num txs", len(selectedTxs), "num remaining", len(remainingTxs), "num sorted txs", len(sortedTxs))

	scheduledMiniBlocks, err := sct.createScheduledMiniBlocksFromMeAsProposer(
		haveTime,
		haveAdditionalTimeFalse,
		selectedTxs,
		make(map[string]struct{}),
	)
	if err != nil {
		log.Debug("createScheduledMiniBlocksFromMeAsProposer", "error", err.Error())
		return make(block.MiniBlockSlice, 0), nil
	}

	return scheduledMiniBlocks, nil
}

func (sct *sovereignChainTransactions) computeSortedTxs(
	sndShardId uint32,
	dstShardId uint32,
	randomness []byte,
) ([]*txcache.WrappedTransaction, error) {
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)
	txShardPool := sct.txPool.ShardDataStore(strCache)
	if check.IfNil(txShardPool) {
		return nil, process.ErrNilTxDataPool
	}

	sortedTransactionsProvider := createSortedTransactionsProvider(txShardPool)
	sortedTxs := sortedTransactionsProvider.GetSortedTransactions()
	sct.sortTransactionsBySenderAndNonce(sortedTxs, randomness)

	return sortedTxs, nil
}

// ProcessMiniBlock does nothing on sovereign chain
func (sct *sovereignChainTransactions) ProcessMiniBlock(
	_ *block.MiniBlock,
	_ func() bool,
	_ func() bool,
	_ bool,
	_ bool,
	_ int,
	_ process.PreProcessorExecutionInfoHandler,
) ([][]byte, int, bool, error) {
	return nil, 0, false, nil
}

func (sct *sovereignChainTransactions) shouldContinueProcessingScheduledTx(
	_ func(uint32) bool,
	wrappedTx *txcache.WrappedTransaction,
	_ map[string]struct{},
	mbInfo *createScheduledMiniBlocksInfo,
) (*transaction.Transaction, *block.MiniBlock, bool) {
	txHash := wrappedTx.TxHash
	senderShardID := wrappedTx.SenderShardID
	receiverShardID := wrappedTx.ReceiverShardID

	tx, ok := wrappedTx.Tx.(*transaction.Transaction)
	if !ok {
		log.Debug("wrong type assertion",
			"hash", txHash,
			"sender shard", senderShardID,
			"receiver shard", receiverShardID)
		return nil, nil, false
	}

	miniBlock, ok := mbInfo.mapMiniBlocks[receiverShardID]
	if !ok {
		log.Debug("scheduled mini block is not created", "shard", receiverShardID)
		return nil, nil, false
	}

	return tx, miniBlock, true
}

func (sct *sovereignChainTransactions) shouldSkipMiniBlock(miniBlock *block.MiniBlock) bool {
	return !sct.isMiniBlockCorrect(miniBlock.Type)
}

func (sct *sovereignChainTransactions) isTransactionEligibleForExecution(tx *transaction.Transaction, err error) bool {
	if isCriticalError(err) {
		log.Debug("sovereignChainTransactions.isTransactionEligibleForExecution: isCriticalError", "error", err)
		return false
	}

	senderAccount, _, errGetAccounts := sct.txProcessor.GetSenderAndReceiverAccounts(tx)
	if check.IfNil(senderAccount) {
		log.Debug("sovereignChainTransactions.isTransactionEligibleForExecution: GetSenderAndReceiverAccounts", "error", errGetAccounts)
		return false
	}

	accntInfo, found := sct.accntsTracker.getAccountInfo(tx.GetSndAddr())
	if !found {
		accntInfo = accountInfo{
			nonce:   senderAccount.GetNonce(),
			balance: big.NewInt(0).Set(senderAccount.GetBalance()),
		}
	}

	if accntInfo.nonce < tx.GetNonce() {
		log.Trace("sovereignChainTransactions.isTransactionEligibleForExecution", "error", process.ErrHigherNonceInTransaction,
			"account nonce", accntInfo.nonce,
			"tx nonce", tx.GetNonce())
		return false
	}

	txFee := sct.economicsFee.ComputeTxFee(tx)
	if accntInfo.balance.Cmp(txFee) < 0 {
		log.Trace("sovereignChainTransactions.isTransactionEligibleForExecution", "error", process.ErrInsufficientFee,
			"account balance", accntInfo.balance.String(),
			"fee needed", txFee.String())
		return false
	}

	cost := big.NewInt(0).Add(txFee, tx.GetValue())
	if accntInfo.balance.Cmp(cost) < 0 {
		log.Trace("sovereignChainTransactions.isTransactionEligibleForExecution", "error", process.ErrInsufficientFunds,
			"account balance", accntInfo.balance.String(),
			"cost", cost.String())
		return false
	}

	accntInfo.nonce++
	accntInfo.balance.Sub(accntInfo.balance, cost)
	sct.accntsTracker.setAccountInfo(tx.GetSndAddr(), accntInfo)

	return true
}

func isCriticalError(err error) bool {
	return err != nil && !errors.Is(err, process.ErrHigherNonceInTransaction)
}
