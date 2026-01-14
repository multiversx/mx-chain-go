package txcache

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
)

// txListForSender represents a sorted list of transactions of a particular sender
type txListForSender struct {
	sender      string
	list        *orderedTransactionsList
	totalBytes  atomic.Counter
	constraints *senderConstraints

	mutex sync.RWMutex
}

// newTxListForSender creates a new (sorted) list of transactions
func newTxListForSender(sender string, constraints *senderConstraints) *txListForSender {
	return &txListForSender{
		list:        newOrderedTransactionsList(),
		sender:      sender,
		constraints: constraints,
	}
}

// AddTx adds a transaction in sender's list
// This is a "sorted" insert
func (listForSender *txListForSender) AddTx(tx *WrappedTransaction, tracker *selectionTracker) (bool, [][]byte) {
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	if tracker == nil {
		return false, nil
	}

	added := listForSender.list.insert(tx)
	if !added {
		return false, nil
	}

	listForSender.onAddedTransaction(tx)

	evicted := listForSender.applySizeConstraints(tracker)
	return true, evicted
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) applySizeConstraints(tracker *selectionTracker) [][]byte {
	evictedTxHashes := make([][]byte, 0)

	// Iterate back to front
	for i := listForSender.list.len() - 1; i >= 0; i-- {
		if !listForSender.isCapacityExceeded() {
			break
		}

		tx := listForSender.list.get(i)
		if !tracker.IsTransactionTracked(tx) {
			_ = listForSender.list.removeAt(i)
			listForSender.onRemovedTransaction(tx)
			evictedTxHashes = append(evictedTxHashes, tx.TxHash)
		}
	}

	return evictedTxHashes
}

func (listForSender *txListForSender) isCapacityExceeded() bool {
	maxBytes := int64(listForSender.constraints.maxNumBytes)
	maxNumTxs := uint64(listForSender.constraints.maxNumTxs)
	tooManyBytes := listForSender.totalBytes.Get() > maxBytes
	tooManyTxs := listForSender.countTx() > maxNumTxs

	return tooManyBytes || tooManyTxs
}

func (listForSender *txListForSender) onAddedTransaction(tx *WrappedTransaction) {
	listForSender.totalBytes.Add(tx.Size)
}

func (listForSender *txListForSender) onRemovedTransaction(tx *WrappedTransaction) {
	listForSender.totalBytes.Subtract(tx.Size)
}

// IsEmpty checks whether the list is empty
func (listForSender *txListForSender) IsEmpty() bool {
	return listForSender.countTxWithLock() == 0
}

// getTxs returns the transactions of the sender
func (listForSender *txListForSender) getTxs() []*WrappedTransaction {
	listForSender.mutex.RLock()
	defer listForSender.mutex.RUnlock()

	return listForSender.list.getAll()
}

// getTxsReversed returns the transactions of the sender, in reverse nonce order
func (listForSender *txListForSender) getTxsReversed() []*WrappedTransaction {
	listForSender.mutex.RLock()
	defer listForSender.mutex.RUnlock()

	items := listForSender.list.items
	result := make([]*WrappedTransaction, 0, len(items))
	for i := len(items) - 1; i >= 0; i-- {
		result = append(result, items[i])
	}
	return result
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) countTx() uint64 {
	return uint64(listForSender.list.len())
}

func (listForSender *txListForSender) countTxWithLock() uint64 {
	listForSender.mutex.RLock()
	defer listForSender.mutex.RUnlock()
	return uint64(listForSender.list.len())
}

// removeTransactionsWithLowerOrEqualNonceReturnHashes removes transactions with nonces lower or equal to the given nonce
func (listForSender *txListForSender) removeTransactionsWithLowerOrEqualNonceReturnHashes(targetNonce uint64) [][]byte {
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	removed := listForSender.list.removeBeforeNonce(targetNonce)
	evictedTxHashes := make([][]byte, len(removed))

	for i, tx := range removed {
		listForSender.onRemovedTransaction(tx)
		evictedTxHashes[i] = tx.TxHash
	}

	return evictedTxHashes
}

func (listForSender *txListForSender) removeTransactionsWithHigherOrEqualNonce(givenNonce uint64) {
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	removed := listForSender.list.removeAfterNonce(givenNonce)
	for _, tx := range removed {
		listForSender.onRemovedTransaction(tx)
	}
}
