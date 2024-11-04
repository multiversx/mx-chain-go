package txcache

import (
	"container/heap"

	"github.com/multiversx/mx-chain-core-go/core"
)

// evictionJournal keeps a short journal about the eviction process
// This is useful for debugging and reasoning about the eviction
type evictionJournal struct {
	numEvicted       int
	numEvictedByPass []int
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
		"evicted txs", evictionJournal.numEvicted,
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

// Eviction tolerates concurrent transaction additions / removals.
func (cache *TxCache) evictLeastLikelyToSelectTransactions() *evictionJournal {
	senders := cache.getSenders()
	bunches := make([]BunchOfTransactions, 0, len(senders))

	for _, sender := range senders {
		// Include transactions after gaps, as well (important), unlike when selecting transactions for processing.
		bunch := sender.getTxs()
		// Reverse the order of transactions (will come in handy later, when creating the min-heap).
		reverseSlice(bunch)

		bunches = append(bunches, bunch)
	}

	journal := &evictionJournal{}

	// Heap is reused among passes.
	// Items popped from the heap are added to "transactionsToEvict" (slice is re-created in each pass).
	transactionsHeap := &TransactionsMinHeap{}
	heap.Init(transactionsHeap)

	// Initialize the heap with the first transaction of each bunch
	for i, bunch := range bunches {
		if len(bunch) == 0 {
			// Some senders may have no transaction anymore (hazardous concurrent removals).
			continue
		}

		// Items will be reused (see below). Each sender gets one (and only one) item in the heap.
		heap.Push(transactionsHeap, &TransactionsHeapItem{
			senderIndex:      i,
			transactionIndex: 0,
			transaction:      bunch[0],
		})
	}

	for pass := 0; cache.isCapacityExceeded(); pass++ {
		transactionsToEvict := make(BunchOfTransactions, 0, cache.config.NumItemsToPreemptivelyEvict)
		transactionsToEvictHashes := make([][]byte, 0, cache.config.NumItemsToPreemptivelyEvict)

		// Select transactions (sorted).
		for transactionsHeap.Len() > 0 {
			// Always pick the "worst" transaction.
			item := heap.Pop(transactionsHeap).(*TransactionsHeapItem)

			if len(transactionsToEvict) >= int(cache.config.NumItemsToPreemptivelyEvict) {
				// We have enough transactions to evict in this pass.
				break
			}

			transactionsToEvict = append(transactionsToEvict, item.transaction)
			transactionsToEvictHashes = append(transactionsToEvictHashes, item.transaction.TxHash)

			// If there are more transactions in the same bunch (same sender as the popped item),
			// add the next one to the heap (to compete with the others in being "the worst").
			item.transactionIndex++

			if item.transactionIndex < len(bunches[item.senderIndex]) {
				// Item is reused (same originating sender), pushed back on the heap.
				item.transaction = bunches[item.senderIndex][item.transactionIndex]
				heap.Push(transactionsHeap, item)
			}
		}

		if len(transactionsToEvict) == 0 {
			// No more transactions to evict.
			break
		}

		// For each sender, find the "lowest" (in nonce) transaction to evict.
		lowestToEvictBySender := make(map[string]uint64)

		for _, tx := range transactionsToEvict {
			sender := string(tx.Tx.GetSndAddr())
			lowestToEvictBySender[sender] = tx.Tx.GetNonce()
		}

		// Remove those transactions from "txListBySender".
		for sender, nonce := range lowestToEvictBySender {
			cache.txListBySender.evictTransactionsWithHigherOrEqualNonces([]byte(sender), nonce)
		}

		// Remove those transactions from "txByHash".
		_ = cache.txByHash.RemoveTxsBulk(transactionsToEvictHashes)

		journal.numEvictedByPass = append(journal.numEvictedByPass, len(transactionsToEvict))
		journal.numEvicted += len(transactionsToEvict)
	}

	return journal
}
