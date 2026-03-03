package txcache

import (
	"bytes"
	"sort"
)

// orderedTransactionsList manages a slice of transactions sorted by execution order
type orderedTransactionsList struct {
	items []*WrappedTransaction
}

func newOrderedTransactionsList() *orderedTransactionsList {
	return &orderedTransactionsList{
		items: make([]*WrappedTransaction, 0),
	}
}

func (otl *orderedTransactionsList) removeAt(index int) *WrappedTransaction {
	if index < 0 || index >= len(otl.items) {
		return nil
	}
	tx := otl.items[index]
	copy(otl.items[index:], otl.items[index+1:])
	otl.items[len(otl.items)-1] = nil // Avoid memory leak
	otl.items = otl.items[:len(otl.items)-1]
	return tx
}

// removeAfterNonce removes all transactions with nonce >= givenNonce
func (otl *orderedTransactionsList) removeAfterNonce(givenNonce uint64) []*WrappedTransaction {
	cutoffIndex := -1
	for i, tx := range otl.items {
		if tx.Tx.GetNonce() >= givenNonce {
			cutoffIndex = i
			break
		}
	}

	if cutoffIndex == -1 {
		return nil
	}

	removed := make([]*WrappedTransaction, 0, len(otl.items)-cutoffIndex)
	for i := cutoffIndex; i < len(otl.items); i++ {
		removed = append(removed, otl.items[i])
		otl.items[i] = nil // Help GC
	}
	otl.items = otl.items[:cutoffIndex]
	return removed
}

// removeBeforeNonce removes all transactions with nonce <= targetNonce
// Returns the removed transactions
func (otl *orderedTransactionsList) removeBeforeNonce(targetNonce uint64) []*WrappedTransaction {
	cutoffIndex := 0
	for i, tx := range otl.items {
		if tx.Tx.GetNonce() > targetNonce {
			cutoffIndex = i
			break
		}
		cutoffIndex = i + 1
	}

	if cutoffIndex == 0 {
		return nil
	}

	removed := make([]*WrappedTransaction, cutoffIndex)
	copy(removed, otl.items[:cutoffIndex])

	remain := len(otl.items) - cutoffIndex
	copy(otl.items, otl.items[cutoffIndex:])

	for i := remain; i < len(otl.items); i++ {
		otl.items[i] = nil
	}
	otl.items = otl.items[:remain]

	return removed
}

func (otl *orderedTransactionsList) get(index int) *WrappedTransaction {
	if index < 0 || index >= len(otl.items) {
		return nil
	}
	return otl.items[index]
}

func (otl *orderedTransactionsList) len() int {
	return len(otl.items)
}

func (otl *orderedTransactionsList) getAll() []*WrappedTransaction {
	result := make([]*WrappedTransaction, len(otl.items))
	copy(result, otl.items)
	return result
}

// getAllFromIndex returns a copy of all transactions starting from the given index.
// This is used for selection to skip already-proposed transactions.
func (otl *orderedTransactionsList) getAllFromIndex(startIndex int) []*WrappedTransaction {
	if startIndex < 0 {
		startIndex = 0
	}
	if startIndex >= len(otl.items) {
		return make([]*WrappedTransaction, 0)
	}

	result := make([]*WrappedTransaction, len(otl.items)-startIndex)
	copy(result, otl.items[startIndex:])
	return result
}

// findIndexByNonce returns the index of the first transaction with nonce >= targetNonce.
// Uses binary search for efficiency. Returns len(items) if no such transaction exists.
// This is used for block replacement to reset the selection offset.
func (otl *orderedTransactionsList) findIndexByNonce(targetNonce uint64) int {
	return sort.Search(len(otl.items), func(i int) bool {
		return otl.items[i].Tx.GetNonce() >= targetNonce
	})
}

// findInsertionIndex returns the index where a transaction would be inserted.
// This function should only be used in critical section (listForSender.mutex).
// When searching for the insertion place, we consider the following rules:
// - transactions are sorted by nonce in ascending order.
// - transactions with the same nonce are sorted by gas price in descending order.
// - transactions with the same nonce and gas price are sorted by hash in ascending order.
// - "PPU" measurement is not relevant in this context. Competition among transactions of the same sender (and nonce) is based on gas price.
func (otl *orderedTransactionsList) findInsertionIndex(tx *WrappedTransaction) int {
	if len(otl.items) == 0 {
		return 0
	}

	lastItem := otl.items[len(otl.items)-1]
	if isGreater(tx, lastItem) {
		return len(otl.items)
	}

	return sort.Search(len(otl.items), func(i int) bool {
		return compareTxs(otl.items[i], tx) >= 0
	})
}

// insertAt inserts a transaction at the given index (pre-computed by findInsertionIndex).
// This function should only be used in critical section (listForSender.mutex).
// Returns false if the transaction is a duplicate (item at index has same nonce, gas price, and hash).
// Duplicates are not allowed.
func (otl *orderedTransactionsList) insertAt(tx *WrappedTransaction, index int) bool {
	// Check for duplicate at the insertion index
	if index < len(otl.items) && compareTxs(otl.items[index], tx) == 0 {
		return false // Duplicate
	}

	// Insert at index
	otl.items = append(otl.items, nil) // grow
	copy(otl.items[index+1:], otl.items[index:])
	otl.items[index] = tx
	return true
}

// isGreater returns true if tx1 > tx2
func isGreater(tx1, tx2 *WrappedTransaction) bool {
	return compareTxs(tx1, tx2) > 0
}

// compareTxs returns:
// -1 if tx1 < tx2, 0 if tx1 == tx2, 1 if tx1 > tx2
func compareTxs(tx1 *WrappedTransaction, tx2 *WrappedTransaction) int {
	// Nonce ASC
	if tx1.Tx.GetNonce() < tx2.Tx.GetNonce() {
		return -1
	}
	if tx1.Tx.GetNonce() > tx2.Tx.GetNonce() {
		return 1
	}

	// Gas Price DESC (Secondary sort key)
	if tx1.Tx.GetGasPrice() > tx2.Tx.GetGasPrice() {
		return -1
	}
	if tx1.Tx.GetGasPrice() < tx2.Tx.GetGasPrice() {
		return 1
	}

	// Hash ASC (Tie-breaker)
	return bytes.Compare(tx1.TxHash, tx2.TxHash)
}
