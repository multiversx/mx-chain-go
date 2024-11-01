package txcache

import (
	"github.com/multiversx/mx-chain-core-go/core"
)

// evictionJournal keeps a short journal about the eviction process
// This is useful for debugging and reasoning about the eviction
type evictionJournal struct {
	numTxs    int
	numPasses int
}

// doEviction does cache eviction.
// We do not allow more evictions to start concurrently.
func (cache *TxCache) doEviction() *evictionJournal {
	if cache.isEvictionInProgress.IsSet() {
		return nil
	}

	if !cache.isCapacityExceeded() {
		return nil
	}

	cache.evictionMutex.Lock()
	defer cache.evictionMutex.Unlock()

	_ = cache.isEvictionInProgress.SetReturningPrevious()
	defer cache.isEvictionInProgress.Reset()

	if !cache.isCapacityExceeded() {
		return nil
	}

	logRemove.Debug("doEviction(): before eviction",
		"num bytes", cache.NumBytes(),
		"num txs", cache.CountTx(),
		"num senders", cache.CountSenders(),
	)

	stopWatch := core.NewStopWatch()
	stopWatch.Start("eviction")

	evictionJournal := cache.evictLeastLikelyToSelectTransactions()

	stopWatch.Stop("eviction")

	logRemove.Debug(
		"doEviction(): after eviction",
		"num bytes", cache.NumBytes(),
		"num now", cache.CountTx(),
		"num senders", cache.CountSenders(),
		"duration", stopWatch.GetMeasurement("eviction"),
		"evicted txs", evictionJournal.numTxs,
	)

	return evictionJournal
}

func (cache *TxCache) isCapacityExceeded() bool {
	exceeded := cache.areThereTooManyBytes() || cache.areThereTooManySenders() || cache.areThereTooManyTxs()
	return exceeded
}

func (cache *TxCache) areThereTooManyBytes() bool {
	numBytes := cache.NumBytes()
	tooManyBytes := numBytes > int(cache.config.NumBytesThreshold)
	return tooManyBytes
}

func (cache *TxCache) areThereTooManySenders() bool {
	numSenders := cache.CountSenders()
	tooManySenders := numSenders > uint64(cache.config.CountThreshold)
	return tooManySenders
}

func (cache *TxCache) areThereTooManyTxs() bool {
	numTxs := cache.CountTx()
	tooManyTxs := numTxs > uint64(cache.config.CountThreshold)
	return tooManyTxs
}

func (cache *TxCache) evictLeastLikelyToSelectTransactions() *evictionJournal {
	senders := cache.getSenders()
	bunches := make([]BunchOfTransactions, 0, len(senders))

	for _, sender := range senders {
		// Include transactions after gaps, as well (important), unlike when selecting transactions for processing.
		bunches = append(bunches, sender.getTxs())
	}

	transactions := mergeBunchesInParallel(bunches, numJobsForMerging)
	transactionsHashes := make([][]byte, len(transactions))

	for i, tx := range transactions {
		transactionsHashes[i] = tx.TxHash
	}

	journal := &evictionJournal{}

	for pass := 1; cache.isCapacityExceeded(); pass++ {
		cutoffIndex := len(transactions) - int(cache.config.NumItemsToPreemptivelyEvict)*pass
		if cutoffIndex <= 0 {
			cutoffIndex = 0
		}

		transactionsToEvict := transactions[cutoffIndex:]
		transactionsToEvictHashes := transactionsHashes[cutoffIndex:]

		transactions = transactions[:cutoffIndex]
		transactionsHashes = transactionsHashes[:cutoffIndex]

		// For each sender, find the "lowest" (in nonce) transaction to evict.
		lowestToEvictBySender := make(map[string]uint64)

		for _, tx := range transactionsToEvict {
			transactionsToEvictHashes = append(transactionsToEvictHashes, tx.TxHash)
			sender := string(tx.Tx.GetSndAddr())

			if _, ok := lowestToEvictBySender[sender]; ok {
				continue
			}

			lowestToEvictBySender[sender] = tx.Tx.GetNonce()
		}

		// Remove those transactions from "txListBySender".
		for sender, nonce := range lowestToEvictBySender {
			list, ok := cache.txListBySender.getListForSender(sender)
			if !ok {
				continue
			}

			list.evictTransactionsWithHigherNonces(nonce - 1)
		}

		// Remove those transactions from "txByHash".
		cache.txByHash.RemoveTxsBulk(transactionsToEvictHashes)

		journal.numPasses = pass
		journal.numTxs += len(transactionsToEvict)
	}

	return journal
}
