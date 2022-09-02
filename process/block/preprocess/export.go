package preprocess

import (
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
	"time"
)

// SetScheduledTXContinueFunc sets a new scheduled tx verifier function
func (txs *transactions) SetScheduledTXContinueFunc(newFunc func(isShardStuck func(uint32) bool, wrappedTx *txcache.WrappedTransaction, mapSCTxs map[string]struct{}, mbInfo *createScheduledMiniBlocksInfo) (*transaction.Transaction, *block.MiniBlock, bool)) {
	if newFunc != nil {
		txs.scheduledTXContinueFunc = newFunc
	}
}

// ProcessTxsFromMe exported function
func (txs *transactions) ProcessTxsFromMe(
	body *block.Body,
	haveTime func() bool,
	randomness []byte,
) (block.MiniBlockSlice, map[string]struct{}, error) {
	return txs.processTxsFromMe(body, haveTime, randomness)
}

// CreateScheduledMiniBlocks only the scheduled miniblocks
func (txs *transactions) CreateScheduledMiniBlocks(haveTime func() bool, randomness []byte, gasBandwidth uint64) (block.MiniBlockSlice, error) {
	startTime := time.Now()

	sortedTxs, remainingTxsForScheduled, err := txs.computeSortedTxs(txs.shardCoordinator.SelfId(), txs.shardCoordinator.SelfId(), gasBandwidth, randomness)
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

	if txs.blockTracker.ShouldSkipMiniBlocksCreationFromSelf() {
		log.Debug("CreateAndProcessMiniBlocks global stuck")
		return make(block.MiniBlockSlice, 0), nil
	}

	sortedTxsForScheduled := append(sortedTxs, remainingTxsForScheduled...)
	sortedTxsForScheduled, _ = txs.prefilterTransactions(nil, sortedTxsForScheduled, 0, gasBandwidth)
	txs.sortTransactionsBySenderAndNonce(sortedTxsForScheduled, randomness)

	haveAdditionalTime := process.HaveAdditionalTime()
	scheduledMiniBlocks, err := txs.createAndProcessScheduledMiniBlocksFromMeAsProposer(
		haveTime,
		haveAdditionalTime,
		sortedTxsForScheduled,
		make(map[string]struct{}),
	)
	if err != nil {
		log.Debug("createAndProcessScheduledMiniBlocksFromMeAsProposer", "error", err.Error())
		return make(block.MiniBlockSlice, 0), nil
	}

	return scheduledMiniBlocks, nil
}