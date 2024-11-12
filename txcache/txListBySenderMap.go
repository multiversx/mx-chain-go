package txcache

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-storage-go/txcache/maps"
)

// txListBySenderMap is a map-like structure for holding and accessing transactions by sender
type txListBySenderMap struct {
	backingMap        *maps.ConcurrentMap
	senderConstraints senderConstraints
	counter           atomic.Counter
	mutex             sync.Mutex
}

// newTxListBySenderMap creates a new instance of TxListBySenderMap
func newTxListBySenderMap(
	nChunksHint uint32,
	senderConstraints senderConstraints,
) *txListBySenderMap {
	backingMap := maps.NewConcurrentMap(nChunksHint)

	return &txListBySenderMap{
		backingMap:        backingMap,
		senderConstraints: senderConstraints,
	}
}

// addTxReturnEvicted adds a transaction in the map, in the corresponding list (selected by its sender).
// This function returns a boolean indicating whether the transaction was added, and a slice of evicted transaction hashes (upon applying sender-level constraints).
func (txMap *txListBySenderMap) addTxReturnEvicted(tx *WrappedTransaction) (bool, [][]byte) {
	sender := string(tx.Tx.GetSndAddr())
	listForSender := txMap.getOrAddListForSender(sender)

	added, evictedHashes := listForSender.AddTx(tx)
	return added, evictedHashes
}

// getOrAddListForSender gets or lazily creates a list (using double-checked locking pattern)
func (txMap *txListBySenderMap) getOrAddListForSender(sender string) *txListForSender {
	listForSender, ok := txMap.getListForSender(sender)
	if ok {
		return listForSender
	}

	txMap.mutex.Lock()
	defer txMap.mutex.Unlock()

	listForSender, ok = txMap.getListForSender(sender)
	if ok {
		return listForSender
	}

	return txMap.addSender(sender)
}

func (txMap *txListBySenderMap) getListForSender(sender string) (*txListForSender, bool) {
	listForSenderUntyped, ok := txMap.backingMap.Get(sender)
	if !ok {
		return nil, false
	}

	listForSender := listForSenderUntyped.(*txListForSender)
	return listForSender, true
}

func (txMap *txListBySenderMap) addSender(sender string) *txListForSender {
	listForSender := newTxListForSender(sender, &txMap.senderConstraints)

	txMap.backingMap.Set(sender, listForSender)
	txMap.counter.Increment()

	return listForSender
}

// removeTransactionsWithLowerOrEqualNonceReturnHashes removes transactions with nonces lower or equal to the given transaction's nonce.
func (txMap *txListBySenderMap) removeTransactionsWithLowerOrEqualNonceReturnHashes(tx *WrappedTransaction) [][]byte {
	sender := string(tx.Tx.GetSndAddr())

	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		// This happens when a sender whose transactions were selected for processing is removed from cache in the meantime.
		// When it comes to remove one if its transactions due to processing (commited / finalized block), they don't exist in cache anymore.
		log.Trace("txListBySenderMap.removeTxReturnEvicted detected slight inconsistency: sender of tx not in cache", "tx", tx.TxHash, "sender", []byte(sender))
		return nil
	}

	evicted := listForSender.removeTransactionsWithLowerOrEqualNonceReturnHashes(tx.Tx.GetNonce())
	txMap.removeSenderIfEmpty(listForSender)
	return evicted
}

func (txMap *txListBySenderMap) removeSenderIfEmpty(listForSender *txListForSender) {
	if listForSender.IsEmpty() {
		txMap.removeSender(listForSender.sender)
	}
}

// Important note: this doesn't remove the transactions from txCache.txByHash. That is the responsibility of the caller (of this function).
func (txMap *txListBySenderMap) removeSender(sender string) bool {
	logRemove.Trace("txListBySenderMap.removeSender", "sender", sender)

	_, removed := txMap.backingMap.Remove(sender)
	if removed {
		txMap.counter.Decrement()
	}

	return removed
}

// RemoveSendersBulk removes senders, in bulk
func (txMap *txListBySenderMap) RemoveSendersBulk(senders []string) uint32 {
	numRemoved := uint32(0)

	for _, senderKey := range senders {
		if txMap.removeSender(senderKey) {
			numRemoved++
		}
	}

	return numRemoved
}

func (txMap *txListBySenderMap) notifyAccountNonce(accountKey []byte, nonce uint64) {
	sender := string(accountKey)
	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		return
	}

	listForSender.notifyAccountNonce(nonce)
}

// removeTransactionsWithHigherOrEqualNonce removes transactions with nonces higher or equal to the given nonce.
// Useful for the eviction flow.
func (txMap *txListBySenderMap) removeTransactionsWithHigherOrEqualNonce(accountKey []byte, nonce uint64) {
	sender := string(accountKey)
	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		return
	}

	listForSender.removeTransactionsWithHigherOrEqualNonce(nonce)
	txMap.removeSenderIfEmpty(listForSender)
}

func (txMap *txListBySenderMap) getSenders() []*txListForSender {
	senders := make([]*txListForSender, 0, txMap.counter.Get())

	txMap.backingMap.IterCb(func(key string, item interface{}) {
		listForSender := item.(*txListForSender)
		senders = append(senders, listForSender)
	})

	return senders
}

func (txMap *txListBySenderMap) clear() {
	txMap.backingMap.Clear()
	txMap.counter.Set(0)
}
