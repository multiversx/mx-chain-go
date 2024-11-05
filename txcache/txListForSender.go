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
	sender            string
	accountNonce      atomic.Uint64
	accountNonceKnown atomic.Flag
	items             *list.List
	totalBytes        atomic.Counter
	constraints       *senderConstraints

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

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) findInsertionPlace(incomingTx *WrappedTransaction) (*list.Element, error) {
	incomingNonce := incomingTx.Tx.GetNonce()
	incomingGasPrice := incomingTx.Tx.GetGasPrice()

	for element := listForSender.items.Back(); element != nil; element = element.Prev() {
		currentTx := element.Value.(*WrappedTransaction)
		currentTxNonce := currentTx.Tx.GetNonce()
		currentTxGasPrice := currentTx.Tx.GetGasPrice()

		if currentTxNonce == incomingNonce {
			if currentTxGasPrice > incomingGasPrice {
				// The incoming transaction will be placed right after the existing one, which has same nonce but higher price.
				// If the nonces are the same, but the incoming gas price is higher or equal, the search loop continues.
				return element, nil
			}
			if currentTxGasPrice == incomingGasPrice {
				// The incoming transaction will be placed right after the existing one, which has same nonce and the same price.
				// (but different hash, because of some other fields like receiver, value or data)
				// This will order out the transactions having the same nonce and gas price

				comparison := bytes.Compare(currentTx.TxHash, incomingTx.TxHash)
				if comparison == 0 {
					// The incoming transaction will be discarded
					return nil, common.ErrItemAlreadyInCache
				}
				if comparison < 0 {
					return element, nil
				}
			}
		}

		if currentTxNonce < incomingNonce {
			// We've found the first transaction with a lower nonce than the incoming one,
			// thus the incoming transaction will be placed right after this one.
			return element, nil
		}
	}

	// The incoming transaction will be inserted at the head of the list.
	return nil, nil
}

// RemoveTx removes a transaction from the sender's list
func (listForSender *txListForSender) RemoveTx(tx *WrappedTransaction) bool {
	// We don't allow concurrent interceptor goroutines to mutate a given sender's list
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	marker := listForSender.findListElementWithTx(tx)
	isFound := marker != nil
	if isFound {
		listForSender.items.Remove(marker)
		listForSender.onRemovedListElement(marker)
	}

	return isFound
}

func (listForSender *txListForSender) onRemovedListElement(element *list.Element) {
	tx := element.Value.(*WrappedTransaction)
	listForSender.totalBytes.Subtract(tx.Size)
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) findListElementWithTx(txToFind *WrappedTransaction) *list.Element {
	txToFindHash := txToFind.TxHash
	txToFindNonce := txToFind.Tx.GetNonce()

	for element := listForSender.items.Front(); element != nil; element = element.Next() {
		value := element.Value.(*WrappedTransaction)
		nonce := value.Tx.GetNonce()

		// Optimization: first, compare nonces, then hashes.
		if nonce == txToFindNonce {
			if bytes.Equal(value.TxHash, txToFindHash) {
				return element
			}
		}

		// Optimization: stop search at this point, since the list is sorted by nonce
		if nonce > txToFindNonce {
			break
		}
	}

	return nil
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

// getTxsWithoutGaps returns the transactions of the sender (gaps are handled, affected transactions are excluded)
func (listForSender *txListForSender) getTxsWithoutGaps() []*WrappedTransaction {
	listForSender.mutex.RLock()
	defer listForSender.mutex.RUnlock()

	accountNonce := listForSender.accountNonce.Get()
	accountNonceKnown := listForSender.accountNonceKnown.IsSet()

	result := make([]*WrappedTransaction, 0, listForSender.countTx())
	previousNonce := uint64(0)

	for element := listForSender.items.Front(); element != nil; element = element.Next() {
		value := element.Value.(*WrappedTransaction)
		nonce := value.Tx.GetNonce()

		// Detect initial gaps.
		if len(result) == 0 && accountNonceKnown && accountNonce != nonce {
			log.Trace("txListForSender.getTxsWithoutGaps, initial gap", "sender", listForSender.sender, "nonce", nonce, "accountNonce", accountNonce)
			break
		}

		// Detect middle gaps.
		if len(result) > 0 && nonce != previousNonce+1 {
			log.Trace("txListForSender.getTxsWithoutGaps, middle gap", "sender", listForSender.sender, "nonce", nonce, "previousNonce", previousNonce)
			break
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

// notifyAccountNonceReturnEvictedTransactions sets the known account nonce, removes the transactions with lower nonces, and returns their hashes
func (listForSender *txListForSender) notifyAccountNonceReturnEvictedTransactions(nonce uint64) [][]byte {
	// Optimization: if nonce is the same, do nothing.
	if listForSender.accountNonce.Get() == nonce {
		return nil
	}

	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	listForSender.accountNonce.Set(nonce)
	_ = listForSender.accountNonceKnown.SetReturningPrevious()

	return listForSender.evictTransactionsWithLowerNoncesNoLockReturnEvicted(nonce)
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) evictTransactionsWithLowerNoncesNoLockReturnEvicted(givenNonce uint64) [][]byte {
	evictedTxHashes := make([][]byte, 0)

	for element := listForSender.items.Front(); element != nil; {
		tx := element.Value.(*WrappedTransaction)
		txNonce := tx.Tx.GetNonce()

		if txNonce >= givenNonce {
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

func (listForSender *txListForSender) evictTransactionsWithHigherOrEqualNonces(givenNonce uint64) {
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
