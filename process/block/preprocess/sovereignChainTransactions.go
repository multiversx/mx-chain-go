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
	"github.com/multiversx/mx-chain-go/state"
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
	sct.isVerifyTransactionFailedFunc = sct.isVerifyTransactionFailed

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
	gasBandwidth := sct.economicsFee.MaxGasLimitPerBlock(sct.shardCoordinator.SelfId()) * 2.0

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

	selectedTxs, _, _ := sct.addTxsWithinBandwidth(nil, sortedTxs, 0, gasBandwidth)

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

	addressHasEnoughBalance := sct.hasAddressEnoughInitialBalance(tx)
	if !addressHasEnoughBalance {
		log.Debug("address has not enough initial balance", "sender", sct.pubkeyConverter.Encode(tx.SndAddr))
		mbInfo.schedulingInfo.numScheduledTxsWithInitialBalanceConsumed++
		return nil, nil, false
	}

	return tx, miniBlock, true
}

func (sct *sovereignChainTransactions) shouldSkipMiniBlock(miniBlock *block.MiniBlock) bool {
	shouldSkipMiniBlock := miniBlock.SenderShardID != sct.shardCoordinator.SelfId() ||
		!sct.isMiniBlockCorrect(miniBlock.Type)

	return shouldSkipMiniBlock
}

func (sct *sovereignChainTransactions) isVerifyTransactionFailed(tx data.TransactionHandler, senderAccount state.UserAccountHandler, err error) bool {
	if err != nil && !errors.Is(err, process.ErrHigherNonceInTransaction) {
		log.Debug("sovereignChainTransactions.isVerifyTransactionFailed", "error", err)
		return true
	}

	if check.IfNil(senderAccount) {
		log.Debug("sovereignChainTransactions.isVerifyTransactionFailed", "error", process.ErrNilUserAccount)
		return true
	}

	sct.mutSenderInfo.Lock()
	defer sct.mutSenderInfo.Unlock()

	sndAddr := string(tx.GetSndAddr())
	si, ok := sct.senderInfo[sndAddr]
	if !ok {
		si = &senderInfo{
			nonce:   senderAccount.GetNonce(),
			balance: big.NewInt(0).Set(senderAccount.GetBalance()),
		}
	}

	if si.nonce < tx.GetNonce() {
		log.Debug("sovereignChainTransactions.isVerifyTransactionFailed", "error", process.ErrHigherNonceInTransaction,
			"account nonce", si.nonce,
			"tx nonce", tx.GetNonce())
		return true
	}

	txFee := sct.economicsFee.ComputeTxFee(tx)
	if si.balance.Cmp(txFee) < 0 {
		log.Debug("sovereignChainTransactions.isVerifyTransactionFailed", "error", process.ErrInsufficientFee,
			"account balance", si.balance.String(),
			"fee needed", txFee.String())
		return true
	}

	cost := big.NewInt(0).Add(txFee, tx.GetValue())
	if si.balance.Cmp(cost) < 0 {
		log.Debug("sovereignChainTransactions.isVerifyTransactionFailed", "error", process.ErrInsufficientFunds,
			"account balance", si.balance.String(),
			"cost", cost.String())
		return true
	}

	si.nonce++
	si.balance = big.NewInt(0).Sub(si.balance, cost)
	sct.senderInfo[sndAddr] = si

	return false
}
