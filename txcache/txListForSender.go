package txcache

import (
	"container/list"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
)

// txListForSender represents a sorted list of transactions of a particular sender
type txListForSender struct {
	sender          string
	items           *list.List
	txsRedBlackTree *transactionsRedBlackTree
	totalBytes      atomic.Counter
	constraints     *senderConstraints

	mutex sync.RWMutex
}

// newTxListForSender creates a new (sorted) list of transactions
func newTxListForSender(sender string, constraints *senderConstraints) *txListForSender {
	return &txListForSender{
		items:           list.New(),
		txsRedBlackTree: NewTransactionsRedBlackTree(),
		sender:          sender,
		constraints:     constraints,
	}
}

// AddTx adds a transaction in sender's list
// This is a "sorted" insert
func (listForSender *txListForSender) AddTx(tx *WrappedTransaction, tracker *selectionTracker) (bool, [][]byte) {
	// We don't allow concurrent interceptor goroutines to mutate a given sender's list
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	if tracker == nil {
		return false, nil
	}

	insertionPlace, err := listForSender.findInsertionPlace(tx)
	if err != nil {
		return false, nil
	}

	if insertionPlace == nil {
		element := listForSender.items.PushFront(tx)
		listForSender.txsRedBlackTree.Insert(element)
	} else {
		element := listForSender.items.InsertAfter(tx, insertionPlace)
		listForSender.txsRedBlackTree.Insert(element)
	}

	listForSender.onAddedTransaction(tx)

	evicted := listForSender.applySizeConstraints(tracker)
	return true, evicted
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) applySizeConstraints(tracker *selectionTracker) [][]byte {
	evictedTxHashes := make([][]byte, 0)

	// Iterate back to front
	for element := listForSender.items.Back(); element != nil; {
		if !listForSender.isCapacityExceeded() {
			break
		}

		value := element.Value.(*WrappedTransaction)

		prevElem := element.Prev()

		if !tracker.IsTransactionTracked(value) {
			listForSender.items.Remove(element)
			listForSender.onRemovedListElement(element)

			// Keep track of removed transactions
			evictedTxHashes = append(evictedTxHashes, value.TxHash)
		}

		element = prevElem
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
	return listForSender.txsRedBlackTree.FindInsertionPlace(&list.Element{
		Value: incomingTx,
	})
}

func (listForSender *txListForSender) onRemovedListElement(element *list.Element) {
	listForSender.txsRedBlackTree.Remove(element)

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
