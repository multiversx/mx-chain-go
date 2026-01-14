package txcache

import (
	"bytes"
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
)

// txListForSender represents a sorted list of transactions of a particular sender
type txListForSender struct {
	sender      string
	items       []*WrappedTransaction
	totalBytes  atomic.Counter
	constraints *senderConstraints

	mutex sync.RWMutex
}

// newTxListForSender creates a new (sorted) list of transactions
func newTxListForSender(sender string, constraints *senderConstraints) *txListForSender {
	return &txListForSender{
		items:       make([]*WrappedTransaction, 0),
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

	// Optimization: Check if we can append (Common Case: ordered nonce)
	if len(listForSender.items) == 0 {
		listForSender.items = append(listForSender.items, tx)
		listForSender.onAddedTransaction(tx)
		return true, nil // No eviction possible on first item usually, or we check constraints after?
		// Must check constraints even if 1 item (if strictly constrained)
		// usage: evicted := listForSender.applySizeConstraints(tracker)
		// return true, evicted
		// Fixed below
	} else {
		lastItem := listForSender.items[len(listForSender.items)-1]
		// Check simple append condition:
		// New Nonce > Last Nonce
		// OR (New Nonce == Last Nonce AND New Price <= Last Price (since desc price)) -> Wait.
		// Sort order: Nonce ASC, Price DESC, Hash ASC.
		// So if New Nonce > Last Nonce: Append.
		// If New Nonce == Last Nonce:
		//    If New Price < Last Price: Append.
		//    If New Price == Last Price:
		//         If New Hash > Last Hash: Append.

		isAppend := false
		if tx.Tx.GetNonce() > lastItem.Tx.GetNonce() {
			isAppend = true
		} else if tx.Tx.GetNonce() == lastItem.Tx.GetNonce() {
			if tx.Tx.GetGasPrice() < lastItem.Tx.GetGasPrice() {
				isAppend = true
			} else if tx.Tx.GetGasPrice() == lastItem.Tx.GetGasPrice() {
				if bytes.Compare(tx.TxHash, lastItem.TxHash) > 0 {
					isAppend = true
				}
			}
		}

		if isAppend {
			listForSender.items = append(listForSender.items, tx)
		} else {
			// Binary Search for insertion point
			index := sort.Search(len(listForSender.items), func(i int) bool {
				// Return true if items[i] should be AFTER or AT tx
				// We want to find the first element that is "Greater" than tx
				// Logic: items[i] >= tx
				return compareTxs(listForSender.items[i], tx) >= 0
			})

			// Check for duplicate
			if index < len(listForSender.items) && compareTxs(listForSender.items[index], tx) == 0 {
				return false, nil // Duplicate
			}

			// Insert at index
			listForSender.items = append(listForSender.items, nil) // grow
			copy(listForSender.items[index+1:], listForSender.items[index:])
			listForSender.items[index] = tx
		}
	}

	listForSender.onAddedTransaction(tx)

	evicted := listForSender.applySizeConstraints(tracker)
	return true, evicted
}

// compareTxs returns:
// -1 if tx1 < tx2
// 0 if tx1 == tx2
// 1 if tx1 > tx2
func compareTxs(tx1 *WrappedTransaction, tx2 *WrappedTransaction) int {
	// Nonce ASC
	if tx1.Tx.GetNonce() < tx2.Tx.GetNonce() {
		return -1
	}
	if tx1.Tx.GetNonce() > tx2.Tx.GetNonce() {
		return 1
	}

	// Gas Price DESC
	if tx1.Tx.GetGasPrice() > tx2.Tx.GetGasPrice() {
		return -1
	}
	if tx1.Tx.GetGasPrice() < tx2.Tx.GetGasPrice() {
		return 1
	}

	// Hash ASC
	return bytes.Compare(tx1.TxHash, tx2.TxHash)
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) applySizeConstraints(tracker *selectionTracker) [][]byte {
	evictedTxHashes := make([][]byte, 0)

	// Iterate back to front
	// With slice, if we remove from back, efficiently:
	// We can just find the split point where constraints satisfy?
	// No, we might skip some tracked transactions if we just truncate?
	// "tracker.IsTransactionTracked(value) -> remove"
	// The requirement says: if capacity exceeded, remove untracked items from back.

	// Since we modify the slice while iterating, it is trickier in-place?
	// Actually, typical "filter in place" approach works.
	// Or since we only remove from the END, we can iterate backwards.

	for i := len(listForSender.items) - 1; i >= 0; i-- {
		if !listForSender.isCapacityExceeded() {
			break
		}

		tx := listForSender.items[i]
		if !tracker.IsTransactionTracked(tx) {
			// Remove ith element
			// Since it is the end of the active list (effectively), and we are iterating backwards...
			// Wait, if we remove item i, items[i+1...] shift left.
			// But we are at the end.
			// If i is exactly len-1, it is O(1).
			// If we skipped some tracked items (so i < len-1), then we have to shift.
			// But usually we just drop tail.

			// Optimization: Remove
			listForSender.removeAt(i)
			listForSender.onRemovedTransaction(tx)
			evictedTxHashes = append(evictedTxHashes, tx.TxHash)

			// Since we removed current i, the indices >= i shift. But we go to i-1.
			// So we don't need to adjust i.
		}
	}

	return evictedTxHashes
}

func (listForSender *txListForSender) removeAt(index int) {
	copy(listForSender.items[index:], listForSender.items[index+1:])
	listForSender.items[len(listForSender.items)-1] = nil // Avoid memory leak
	listForSender.items = listForSender.items[:len(listForSender.items)-1]
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

	// Return a copy to be safe? The original returned copy.
	result := make([]*WrappedTransaction, len(listForSender.items))
	copy(result, listForSender.items)
	return result
}

// getTxsReversed returns the transactions of the sender, in reverse nonce order
func (listForSender *txListForSender) getTxsReversed() []*WrappedTransaction {
	listForSender.mutex.RLock()
	defer listForSender.mutex.RUnlock()

	result := make([]*WrappedTransaction, 0, len(listForSender.items))
	for i := len(listForSender.items) - 1; i >= 0; i-- {
		result = append(result, listForSender.items[i])
	}
	return result
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) countTx() uint64 {
	return uint64(len(listForSender.items))
}

func (listForSender *txListForSender) countTxWithLock() uint64 {
	listForSender.mutex.RLock()
	defer listForSender.mutex.RUnlock()
	return uint64(len(listForSender.items))
}

// removeTransactionsWithLowerOrEqualNonceReturnHashes removes transactions with nonces lower or equal to the given nonce
func (listForSender *txListForSender) removeTransactionsWithLowerOrEqualNonceReturnHashes(targetNonce uint64) [][]byte {
	evictedTxHashes := make([][]byte, 0)

	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	// Find the first element that has nonce > targetNonce
	// All elements BEFORE that should be removed.
	// Optimization: Since sorted by Nonce, we can find the cutoff index.

	// cutoffIndex is the first index where nonce > targetNonce.
	// So items[0...cutoffIndex-1] have nonce <= targetNonce -> REMOVE.

	cutoffIndex := 0
	for i, tx := range listForSender.items {
		if tx.Tx.GetNonce() > targetNonce {
			cutoffIndex = i
			break
		}
		// If we reached end, cutoffIndex should effectively be len
		cutoffIndex = i + 1
	}

	if cutoffIndex > 0 {
		// Collect evicted hashes
		for i := 0; i < cutoffIndex; i++ {
			tx := listForSender.items[i]
			evictedTxHashes = append(evictedTxHashes, tx.TxHash)
			listForSender.onRemovedTransaction(tx)
		}

		// Remove from front: Reslice
		// Potential memory leak if we don't nil pointers?
		// Since we are discarding the backing array part, eventually GC handles it if we copy?
		// Better: copy remaining to front? Or just reslice?
		// If we just reslice `items = items[k:]`, the underlying array keeps references to old ptrs?
		// Yes. To avoid mem leak of WrappedTransaction, we should nil them?
		// It's safer to copy if we want to release memory, OR nil out the dropped elements manually?
		// But we can't nil them if we lost access.

		// Correct approach for standard slice queue:
		// copy(s, s[k:])
		// for i := len(s)-k; i < len(s); i++ { s[i] = nil }
		// s = s[:len(s)-k]

		remain := len(listForSender.items) - cutoffIndex
		copy(listForSender.items, listForSender.items[cutoffIndex:])
		// Nil out the rest
		for i := remain; i < len(listForSender.items); i++ {
			listForSender.items[i] = nil
		}
		listForSender.items = listForSender.items[:remain]
	}

	return evictedTxHashes
}

func (listForSender *txListForSender) removeTransactionsWithHigherOrEqualNonce(givenNonce uint64) {
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	// Find first element with nonce >= givenNonce.
	// Everything from there onwards should be removed.

	cutoffIndex := -1
	for i, tx := range listForSender.items {
		if tx.Tx.GetNonce() >= givenNonce {
			cutoffIndex = i
			break
		}
	}

	// If found, remove from cutoffIndex to end
	if cutoffIndex != -1 {
		for i := cutoffIndex; i < len(listForSender.items); i++ {
			listForSender.onRemovedTransaction(listForSender.items[i])
			listForSender.items[i] = nil // Help GC
		}
		listForSender.items = listForSender.items[:cutoffIndex]
	}
}
