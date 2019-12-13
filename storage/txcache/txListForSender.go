package txcache

import (
	"container/list"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
)

// txListForSender represents a sorted list of transactions of a particular sender
type txListForSender struct {
	items          *list.List
	mutex          sync.Mutex
	copyBatchIndex *list.Element
	orderNumber    int64
	sender         string
}

// txListForSenderNode is a node of the linked list
type txListForSenderNode struct {
	txHash []byte
	tx     data.TransactionHandler
}

// newTxListForSender creates a new (sorted) list of transactions
func newTxListForSender(sender string, globalIndex int64) *txListForSender {
	return &txListForSender{
		items:       list.New(),
		orderNumber: globalIndex,
		sender:      sender,
	}
}

// AddTx adds a transaction in sender's list
// This is a "sorted" insert
func (listForSender *txListForSender) AddTx(txHash []byte, tx data.TransactionHandler) {
	// We don't allow concurent interceptor goroutines to mutate a given sender's list
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	nonce := tx.GetNonce()
	mark := listForSender.findTxWithLargerNonce(nonce)
	newNode := txListForSenderNode{txHash, tx}

	if mark == nil {
		listForSender.items.PushBack(newNode)
	} else {
		listForSender.items.InsertBefore(newNode, mark)
	}
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) findTxWithLargerNonce(nonce uint64) *list.Element {
	for element := listForSender.items.Front(); element != nil; element = element.Next() {
		value := element.Value.(txListForSenderNode)
		if value.tx.GetNonce() > nonce {
			return element
		}
	}

	return nil
}

// RemoveTx removes a transaction from the sender's list
func (listForSender *txListForSender) RemoveTx(tx data.TransactionHandler) bool {
	// We don't allow concurent interceptor goroutines to mutate a given sender's list
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	marker := listForSender.findListElementWithTx(tx)
	isFound := marker != nil
	if isFound {
		listForSender.items.Remove(marker)
	}

	return isFound
}

// RemoveHighNonceTxs removes "count" transactions from the back of the list
func (listForSender *txListForSender) RemoveHighNonceTxs(count uint32) [][]byte {
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	removedTxHashes := make([][]byte, count)

	index := uint32(0)
	var previous *list.Element
	for element := listForSender.items.Back(); element != nil && count > index; element = previous {
		// Remove node
		previous = element.Prev()
		listForSender.items.Remove(element)

		// Keep track of removed transaction
		value := element.Value.(txListForSenderNode)
		removedTxHashes[index] = value.txHash

		index++
	}

	return removedTxHashes
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) findListElementWithTx(txToFind data.TransactionHandler) *list.Element {
	for element := listForSender.items.Front(); element != nil; element = element.Next() {
		value := element.Value.(txListForSenderNode)

		if value.tx == txToFind {
			return element
		}

		// Optimization: stop search at this point, since the list is sorted by nonce
		if value.tx.GetNonce() > txToFind.GetNonce() {
			break
		}
	}

	return nil
}

// HasMoreThan checks whether the list has more items than specified
func (listForSender *txListForSender) HasMoreThan(count uint32) bool {
	return uint32(listForSender.items.Len()) > count
}

// IsEmpty checks whether the list is empty
func (listForSender *txListForSender) IsEmpty() bool {
	return listForSender.items.Len() == 0
}

// copyBatchTo copies a batch (usually small) of transactions to a destination slice
// It also updates the internal state used for copy operations
func (listForSender *txListForSender) copyBatchTo(withReset bool, destination []data.TransactionHandler, batchSize int) int {
	// We can't read from multiple goroutines at the same time
	// And we can't mutate the sender's list while reading it
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	// Reset the internal state used for copy operations
	if withReset {
		listForSender.copyBatchIndex = listForSender.items.Front()
	}

	element := listForSender.copyBatchIndex
	availableSpace := len(destination)

	if element == nil {
		return 0
	}

	copied := 0
	for ; ; copied++ {
		if element == nil || copied == batchSize || copied == availableSpace {
			break
		}

		value := element.Value.(txListForSenderNode)
		destination[copied] = value.tx
		element = element.Next()
	}

	listForSender.copyBatchIndex = element
	return copied
}

// getTxHashes returns the hashes of transactions in the list
func (listForSender *txListForSender) getTxHashes() [][]byte {
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	result := make([][]byte, listForSender.items.Len())

	index := 0
	for element := listForSender.items.Front(); element != nil; element = element.Next() {
		value := element.Value.(txListForSenderNode)
		result[index] = value.txHash
		index++
	}

	return result
}

func (listForSender *txListForSender) getHighestNonceTx() data.TransactionHandler {
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	back := listForSender.items.Back()

	if back == nil {
		return nil
	}

	value := back.Value.(txListForSenderNode)
	return value.tx
}
