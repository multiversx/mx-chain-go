package txcache

import (
	"bytes"
	"container/list"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-storage-go/common"
)

// txListForSender represents a sorted list of transactions of a particular sender
type txListForSender struct {
	sender      string
	items       *list.List
	totalBytes  atomic.Counter
	constraints *senderConstraints

	mutex sync.RWMutex
}

// newTxListForSender creates a new (sorted) list of transactions
func newTxListForSender(sender string, constraints *senderConstraints) *txListForSender {
	return &txListForSender{
		items:       list.New(),
		sender:      sender,
		constraints: constraints,
	}
}

// AddTx adds a transaction in sender's list
// This is a "sorted" insert
func (listForSender *txListForSender) AddTx(tx *WrappedTransaction) (bool, [][]byte) {
	// We don't allow concurrent interceptor goroutines to mutate a given sender's list
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	insertionPlace, err := listForSender.findInsertionPlace(tx)
	if err != nil {
		return false, nil
	}

	if insertionPlace == nil {
		listForSender.items.PushFront(tx)
	} else {
		listForSender.items.InsertAfter(tx, insertionPlace)
	}

	listForSender.onAddedTransaction(tx)

	evicted := listForSender.applySizeConstraints()
	return true, evicted
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) applySizeConstraints() [][]byte {
	evictedTxHashes := make([][]byte, 0)

	// Iterate back to front
	for element := listForSender.items.Back(); element != nil; element = element.Prev() {
		if !listForSender.isCapacityExceeded() {
			break
		}

		listForSender.items.Remove(element)
		listForSender.onRemovedListElement(element)

		// Keep track of removed transactions
		value := element.Value.(*WrappedTransaction)
		evictedTxHashes = append(evictedTxHashes, value.TxHash)
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

// This function should only be used in critical section (listForSender.mutex).
// When searching for the insertion place, we consider the following rules:
// - transactions are sorted by nonce in ascending order.
// - transactions with the same nonce are sorted by gas price in descending order.
// - transactions with the same nonce and gas price are sorted by hash in ascending order.
// - duplicates are not allowed.
// - "PPU" measurement is not relevant in this context. Competition among transactions of the same sender (and nonce) is based on gas price.
func (listForSender *txListForSender) findInsertionPlace(incomingTx *WrappedTransaction) (*list.Element, error) {
	incomingNonce := incomingTx.Tx.GetNonce()
	incomingGasPrice := incomingTx.Tx.GetGasPrice()

	// The loop iterates from the back to the front of the list.
	// Starting from the back allows the function to quickly find the insertion point for transactions with higher nonces, which are more likely to be added.
	for element := listForSender.items.Back(); element != nil; element = element.Prev() {
		currentTx := element.Value.(*WrappedTransaction)
		currentTxNonce := currentTx.Tx.GetNonce()
		currentTxGasPrice := currentTx.Tx.GetGasPrice()

		if currentTxNonce == incomingNonce {
			if currentTxGasPrice > incomingGasPrice {
				// The case of same nonce, lower gas price.
				// We've found an insertion place: right after "element".
				return element, nil
			}

			if currentTxGasPrice == incomingGasPrice {
				// The case of same nonce, same gas price.

				comparison := bytes.Compare(currentTx.TxHash, incomingTx.TxHash)
				if comparison == 0 {
					// The incoming transaction will be discarded, since it's already in the cache.
					return nil, common.ErrItemAlreadyInCache
				}
				if comparison < 0 {
					// We've found an insertion place: right after "element".
					return element, nil
				}

				// We allow the search loop to continue, since the incoming transaction has a "higher hash".
			}

			// We allow the search loop to continue, since the incoming transaction has a higher gas price.
			continue
		}

		if currentTxNonce < incomingNonce {
			// We've found the first transaction with a lower nonce than the incoming one,
			// thus the incoming transaction will be placed right after this one.
			return element, nil
		}

		// We allow the search loop to continue, since the incoming transaction has a higher nonce.
	}

	// The incoming transaction will be inserted at the head of the list.
	return nil, nil
}

func (listForSender *txListForSender) onRemovedListElement(element *list.Element) {
	tx := element.Value.(*WrappedTransaction)
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

	result := make([]*WrappedTransaction, 0, listForSender.countTx())

	for element := listForSender.items.Front(); element != nil; element = element.Next() {
		value := element.Value.(*WrappedTransaction)
		result = append(result, value)
	}

	return result
}

// getTxsReversed returns the transactions of the sender, in reverse nonce order
func (listForSender *txListForSender) getTxsReversed() []*WrappedTransaction {
	listForSender.mutex.RLock()
	defer listForSender.mutex.RUnlock()

	result := make([]*WrappedTransaction, 0, listForSender.countTx())

	for element := listForSender.items.Back(); element != nil; element = element.Prev() {
		value := element.Value.(*WrappedTransaction)
		result = append(result, value)
	}

	return result
}

// getSequentialTxs returns the transactions of the sender, in the context of transactions selection.
// Middle gaps and duplicates are handled (affected transactions are excluded).
// Initial gaps and lower nonces are not handled (not enough information); they are detected a bit later, within the selection loop.
func (listForSender *txListForSender) getSequentialTxs() []*WrappedTransaction {
	listForSender.mutex.RLock()
	defer listForSender.mutex.RUnlock()

	result := make([]*WrappedTransaction, 0, listForSender.countTx())
	previousNonce := uint64(0)

	for element := listForSender.items.Front(); element != nil; element = element.Next() {
		value := element.Value.(*WrappedTransaction)
		nonce := value.Tx.GetNonce()

		isFirstTx := len(result) == 0
		if !isFirstTx {
			// Handle duplicates (only transactions with the highest gas price are included; see "findInsertionPlace").
			if nonce == previousNonce {
				log.Trace("txListForSender.getSequentialTxs, duplicate", "sender", listForSender.sender, "nonce", nonce)
				continue
			}

			// Handle middle gaps.
			if nonce != previousNonce+1 {
				log.Trace("txListForSender.getSequentialTxs, middle gap", "sender", listForSender.sender, "nonce", nonce, "previousNonce", previousNonce)
				break
			}
		}

		result = append(result, value)
		previousNonce = nonce
	}

	return result
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) countTx() uint64 {
	return uint64(listForSender.items.Len())
}

func (listForSender *txListForSender) countTxWithLock() uint64 {
	listForSender.mutex.RLock()
	defer listForSender.mutex.RUnlock()
	return uint64(listForSender.items.Len())
}

// removeTransactionsWithLowerOrEqualNonceReturnHashes removes transactions with nonces lower or equal to the given nonce
func (listForSender *txListForSender) removeTransactionsWithLowerOrEqualNonceReturnHashes(targetNonce uint64) [][]byte {
	evictedTxHashes := make([][]byte, 0)

	// We don't allow concurrent goroutines to mutate a given sender's list
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	for element := listForSender.items.Front(); element != nil; {
		tx := element.Value.(*WrappedTransaction)
		txNonce := tx.Tx.GetNonce()

		if txNonce > targetNonce {
			break
		}

		nextElement := element.Next()
		_ = listForSender.items.Remove(element)
		listForSender.onRemovedListElement(element)
		element = nextElement

		// Keep track of removed transactions
		evictedTxHashes = append(evictedTxHashes, tx.TxHash)
	}

	return evictedTxHashes
}

func (listForSender *txListForSender) removeTransactionsWithHigherOrEqualNonce(givenNonce uint64) {
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	for element := listForSender.items.Back(); element != nil; {
		tx := element.Value.(*WrappedTransaction)
		txNonce := tx.Tx.GetNonce()

		if txNonce < givenNonce {
			break
		}

		prevElement := element.Prev()
		_ = listForSender.items.Remove(element)
		listForSender.onRemovedListElement(element)
		element = prevElement
	}
}

// GetKey returns the key
func (listForSender *txListForSender) GetKey() string {
	return listForSender.sender
}
