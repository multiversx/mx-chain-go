package txcache

import (
	linkedList "container/list"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// TxListForSender is
type TxListForSender struct {
	CopyBatchIndex *linkedList.Element
	CopyBatchSize  int
	Items          *linkedList.List
	mutex          sync.Mutex
}

// NewTxListForSender creates a new (sorted) list of transactions
func NewTxListForSender() *TxListForSender {
	return &TxListForSender{
		Items: linkedList.New(),
	}
}

// AddTransaction adds a transaction in sender's list
// This is a "sorted" insert
func (list *TxListForSender) AddTransaction(tx *transaction.Transaction) {
	// We don't allow concurent interceptor goroutines to mutate a given sender's list
	list.mutex.Lock()

	nonce := tx.Nonce
	mark := list.findTransactionWithLargerNonce(nonce)
	if mark == nil {
		list.Items.PushBack(tx)
	} else {
		list.Items.InsertBefore(tx, mark)
	}

	list.mutex.Unlock()
}

func (list *TxListForSender) findTransactionWithLargerNonce(nonce uint64) *linkedList.Element {
	for element := list.Items.Front(); element != nil; element = element.Next() {
		tx := element.Value.(*transaction.Transaction)
		if tx.Nonce > nonce {
			return element
		}
	}

	return nil
}

// RemoveTransaction removes a transaction from the sender's list
func (list *TxListForSender) RemoveTransaction(tx *transaction.Transaction) {
	// We don't allow concurent interceptor goroutines to mutate a given sender's list
	list.mutex.Lock()

	marker := list.findTransaction(tx)

	if marker != nil {
		list.Items.Remove(marker)
	}

	list.mutex.Unlock()
}

// RemoveHighNonceTransactions removes "count" transactions from the back of the list
func (list *TxListForSender) RemoveHighNonceTransactions(count int) {
	list.mutex.Lock()

	for element := list.Items.Back(); element != nil && count > 0; element = element.Prev() {
		list.Items.Remove(element)
		count--
	}

	list.mutex.Unlock()
}

func (list *TxListForSender) findTransaction(txToFind *transaction.Transaction) *linkedList.Element {
	for element := list.Items.Front(); element != nil; element = element.Next() {
		tx := element.Value.(*transaction.Transaction)
		if tx == txToFind {
			return element
		}
	}

	return nil
}

// HasMoreThan checks whether the list has more items than specified
func (list *TxListForSender) HasMoreThan(count int) bool {
	return list.Items.Len() > count
}

// IsEmpty checks whether the list is empty
func (list *TxListForSender) IsEmpty() bool {
	return list.Items.Len() == 0
}

// RestartBatchCopying resets the internal state used for copy operations
func (list *TxListForSender) RestartBatchCopying(batchSize int) {
	// We cannot copy or start copy from multiple goroutines at the same time
	list.mutex.Lock()

	list.CopyBatchIndex = list.Items.Front()
	list.CopyBatchSize = batchSize

	list.mutex.Unlock()
}

// CopyBatchTo copies a batch (usually small) of transactions to a destination slice
// It also updates the internal state used for copy operations
func (list *TxListForSender) CopyBatchTo(destination []*transaction.Transaction) int {
	element := list.CopyBatchIndex
	batchSize := list.CopyBatchSize
	availableLength := len(destination)

	if element == nil {
		return 0
	}

	// We can't read from multiple goroutines at the same time
	// And we can't mutate the sender's list while reading it
	list.mutex.Lock()

	// todo rewrite loop, make it more readable
	copied := 0
	for true {
		if element == nil || copied == batchSize || availableLength == copied {
			break
		}

		tx := element.Value.(*transaction.Transaction)
		destination[copied] = tx
		copied++
		element = element.Next()
	}

	list.CopyBatchIndex = element

	list.mutex.Unlock()
	return copied
}
