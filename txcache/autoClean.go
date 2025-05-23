package txcache

import (
	"sort"
	"time"
)

/*
Previously, when not-executable transactions were selected from the mempool, they were detected as bad in the processing flow.
There, they were collected and “targeted for deletion”. Then, deleted upon Node’s request.
Now, however, the not-executable ones remain in pool, until they are removed due to processing a transaction with equal or higher nonce.
Maybe we should implement a self-cleaning mechanism, in addition to the rarely-executing, generic tx pool cleaner.

Self-cleaning: mempool, please simulate a selection and remove all not-executable transactions that you hold.

- when is autoclean executed?
-
Adrian Dobrita
6:33 PM
func (sp *shardProcessor) CommitBlock(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {
func (mp *metaProcessor) CommitBlock(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {

ANDREI:
txCache - Cleanup(accountDB, / - bandwidth de timp, de counter? - /) ->
	txs = self.selectSweepableTransactions(...) // seamana cu selectia, se uita lower & duplicates
	self.removeTxs(txs)

// de citit ce face: removeTransactionsWithLowerOrEqualNonceReturnHashes

// sa retinem mici contoare la spetele de stergere, sa le afisam (logging, debugging), plus timpi (durate)

(*) selectSweepableTransactions sa porneasca din "locuri" diferite, la fiecare rulare. Cumva random (dar la toate nodurile la fel - e.g. formula pe baza de adrese).

Loc de apel, prin commit block, la final.

// CommitBlock commits the block in the blockchain if everything was checked successfully
func (sp *shardProcessor) CommitBlock(
	> reper "removeTxsFromPools()".

//
System test - nonce-uri duplicate. spete cu scheduled, spete de concurență fină (receive in interceptor, accept, commit happens in processing flow, mempool insertion happens in interception flow)
------
process/block/preprocess/transactions.go (transactions pre-processor)
txs.txPool - ăsta e, de fapt, dataRetriever/txpool/shardedTxPool.go (ăsta e un wrapper peste TxCache)

// RemoveTxsFromPools removes transactions from associated pools
func (txs *transactions) RemoveTxsFromPools(body *block.Body) error {
	return txs.removeTxsFromPools(body, txs.txPool, txs.isMiniBlockCorrect)
	// (cred)
	// txs.txPool.SourceIsMeThatIsTheRealMempoolCleanup()
}

În shardedTxPool.go cred că punem acea funcție, și ea apelează, într-un final, operațiunea de cleanup din structura TxCache, căci shardedTxPool are referință la TxCache.

*/
func (cache *TxCache) Cleanup(session SelectionSession, nonce uint64, maxNum int, selectionLoopMaximumDuration time.Duration) (uint64) {

	return cache.RemoveSweepableTxs(session, nonce, maxNum, selectionLoopMaximumDuration)
}

func (cache *TxCache) getOrderedSenders() []*txListForSender {
	senders := make([]*txListForSender, 0, cache.txListBySender.counter.Get())
	sender_addresses := cache.txListBySender.backingMap.Keys()
	sort.Slice(sender_addresses, func(i, j int) bool {
		return sender_addresses[i] < sender_addresses[j]
	})

	for _, sender := range sender_addresses {
		listForSender, ok := cache.txListBySender.backingMap.Get(sender)
		if ok {
			senders = append(senders, listForSender.(*txListForSender))
		}
	}
	return senders
}

func (cache *TxCache) RemoveSweepableTxs(session SelectionSession, nonce uint64, maxNum int, selectionLoopMaximumDuration time.Duration) uint64 {
	cache.mutTxOperation.Lock()
	defer cache.mutTxOperation.Unlock()
	selectionLoopStartTime := time.Now()
	evicted := make([][] byte, 0, cache.txByHash.counter.Get())

	senders := cache.getOrderedSenders()
	for _, sender := range senders{
		sessionWrapper := newSelectionSessionWrapper(session)
		sender_address := []byte (sender.sender)
		nonce:= sessionWrapper.getNonce(sender_address)

		if len(evicted) >= maxNum || time.Since(selectionLoopStartTime) > selectionLoopMaximumDuration{
			break
		}
		evicted = append(evicted, sender.removeSweepableTransactionsReturnHashes(nonce)...)
	}
	if len(evicted) > 0 {
		cache.txByHash.RemoveTxsBulk(evicted)
	}
	logRemove.Trace("TxCache.Cleanup", "nonce", nonce, "len(cleaned)", len(evicted), "duration", time.Since(selectionLoopStartTime))
	return uint64(len(evicted))
}

func (listForSender *txListForSender) removeSweepableTransactionsReturnHashes(targetNonce uint64) [][]byte {
	evictedTxHashes := make([][]byte, 0)

	// We don't allow concurrent goroutines to mutate a given sender's list
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	lastTxNonceForDuplicatesProcessing:= targetNonce
	for element := listForSender.items.Front(); element != nil; {
		tx := element.Value.(*WrappedTransaction)
		txNonce := tx.Tx.GetNonce()
		txNonceForDuplicateProcessing:= lastTxNonceForDuplicatesProcessing
		lastTxNonceForDuplicatesProcessing = txNonce

		if txNonce > targetNonce && txNonce != txNonceForDuplicateProcessing{
			//fmt.Println("Skipped txNonce:", txNonce, ", targetNonce = ",targetNonce, " txNonceForDuplicateProcessing:", txNonceForDuplicateProcessing)	
			element = element.Next()
			continue
		}

		//fmt.Println("Not skipped txNonce:", txNonce, ", targetNonce = ",targetNonce, " txNonceForDuplicateProcessing:", txNonceForDuplicateProcessing)
		nextElement := element.Next()
		_ = listForSender.items.Remove(element)
		listForSender.onRemovedListElement(element)
		element = nextElement

		// Keep track of removed transactions
		evictedTxHashes = append(evictedTxHashes, tx.TxHash)
	}

	return evictedTxHashes
}