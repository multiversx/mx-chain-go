package txcache

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
)

// txListForSender represents a sorted list of transactions of a particular sender
type txListForSender struct {
	sender          string
	list            *orderedTransactionsList
	totalBytes      atomic.Counter
	constraints     *senderConstraints
	selectionOffset int // Index from which selection should start (transactions before are in proposed blocks)

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

	insertionIndex := listForSender.list.findInsertionIndex(tx)
	added := listForSender.list.insertAt(tx, insertionIndex)
	if !added {
		return false, nil
	}

	// If transaction was inserted before the selection offset, increment offset to maintain position
	if insertionIndex < listForSender.selectionOffset {
		listForSender.selectionOffset++
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

			// If removal is at index < offset, decrement offset
			if i < listForSender.selectionOffset {
				listForSender.selectionOffset--
			}
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

	// Decrement offset by number of removed transactions (clamped to 0)
	listForSender.decrementSelectionOffset(len(removed))

	return evictedTxHashes
}

func (listForSender *txListForSender) removeTransactionsWithHigherOrEqualNonce(givenNonce uint64) {
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	removed := listForSender.list.removeAfterNonce(givenNonce)
	for _, tx := range removed {
		listForSender.onRemovedTransaction(tx)
	}
	if listForSender.selectionOffset > listForSender.list.len() {
		listForSender.selectionOffset = listForSender.list.len()
	}
}

// getTxsForSelection returns the transactions of the sender starting from the selection offset.
// Transactions before the offset are already in proposed blocks and should be skipped during selection.
func (listForSender *txListForSender) getTxsForSelection() []*WrappedTransaction {
	listForSender.mutex.RLock()
	defer listForSender.mutex.RUnlock()

	return listForSender.list.getAllFromIndex(listForSender.selectionOffset)
}

// incrementSelectionOffset increases the selection offset by the given count.
// This is called during OnProposed to skip transactions that are now in a proposed block.
// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) incrementSelectionOffset(count int) {
	listForSender.selectionOffset += count
	// Clamp to list length to avoid going beyond available transactions
	if listForSender.selectionOffset > listForSender.list.len() {
		listForSender.selectionOffset = listForSender.list.len()
	}
}

// decrementSelectionOffset decreases the selection offset by the given count.
// This is called when transactions are removed from the front of the list (executed).
// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) decrementSelectionOffset(count int) {
	listForSender.selectionOffset -= count
	// Clamp to 0 to avoid negative offset
	if listForSender.selectionOffset < 0 {
		listForSender.selectionOffset = 0
	}
}

// resetSelectionOffsetByNonce resets the selection offset to point to the first transaction
// with nonce >= startNonce. Uses binary search for efficiency.
// This is called during block replacement to re-enable transactions for selection.
func (listForSender *txListForSender) resetSelectionOffsetByNonce(startNonce uint64) {
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	listForSender.selectionOffset = listForSender.list.findIndexByNonce(startNonce)
}

// getSelectionOffset returns the current selection offset.
// This is primarily used for testing.
func (listForSender *txListForSender) getSelectionOffset() int {
	listForSender.mutex.RLock()
	defer listForSender.mutex.RUnlock()

	return listForSender.selectionOffset
}
