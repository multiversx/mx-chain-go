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

// This function should only be used in critical section (listForSender.mutex).
// When searching for the insertion place, we consider the following rules:
// - transactions are sorted by nonce in ascending order.
// - transactions with the same nonce are sorted by gas price in descending order.
// - transactions with the same nonce and gas price are sorted by hash in ascending order.
// - duplicates are not allowed.
// - "PPU" measurement is not relevant in this context. Competition among transactions of the same sender (and nonce) is based on gas price.
// Returns success (true) or failure (false) if duplicate.
func (otl *orderedTransactionsList) insert(tx *WrappedTransaction) bool {
	// Optimization: Check if we can append (Common Case: ordered nonce)
	if len(otl.items) == 0 {
		otl.items = append(otl.items, tx)
		return true
	}

	lastItem := otl.items[len(otl.items)-1]
	if isGreater(tx, lastItem) {
		otl.items = append(otl.items, tx)
		return true
	}

	// Binary Search for insertion point
	index := sort.Search(len(otl.items), func(i int) bool {
		// Return true if items[i] >= tx
		return compareTxs(otl.items[i], tx) >= 0
	})

	// Check for duplicate
	if index < len(otl.items) && compareTxs(otl.items[index], tx) == 0 {
		return false // Duplicate
	}

	// Insert at index
	otl.items = append(otl.items, nil) // grow
	copy(otl.items[index+1:], otl.items[index:])
	otl.items[index] = tx
	return true
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
