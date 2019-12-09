package txcache

import (
	linkedList "container/list"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// TxListForSender is
type TxListForSender struct {
	Items          *linkedList.List
	mutex          sync.Mutex
	copyBatchIndex *linkedList.Element
	copyBatchSize  int
	orderNumber    int64
	sender         string
}

// TxListForSenderNode is a node of the linked list
type TxListForSenderNode struct {
	TxHash []byte
	Tx     *transaction.Transaction
}

// NewTxListForSender creates a new (sorted) list of transactions
func NewTxListForSender(sender string, globalIndex int64) *TxListForSender {
	return &TxListForSender{
		Items:       linkedList.New(),
		orderNumber: globalIndex,
		sender:      sender,
	}
}

// AddTx adds a transaction in sender's list
// This is a "sorted" insert
func (list *TxListForSender) AddTx(txHash []byte, tx *transaction.Transaction) {
	// We don't allow concurent interceptor goroutines to mutate a given sender's list
	list.mutex.Lock()

	nonce := tx.Nonce
	mark := list.findTxWithLargerNonce(nonce)
	newNode := TxListForSenderNode{txHash, tx}

	if mark == nil {
		list.Items.PushBack(newNode)
	} else {
		list.Items.InsertBefore(newNode, mark)
	}

	list.mutex.Unlock()
}

func (list *TxListForSender) findTxWithLargerNonce(nonce uint64) *linkedList.Element {
	for element := list.Items.Front(); element != nil; element = element.Next() {
		value := element.Value.(TxListForSenderNode)
		if value.Tx.Nonce > nonce {
			return element
		}
	}

	return nil
}

// RemoveTx removes a transaction from the sender's list
func (list *TxListForSender) RemoveTx(tx *transaction.Transaction) {
	// We don't allow concurent interceptor goroutines to mutate a given sender's list
	list.mutex.Lock()

	marker := list.findTx(tx)

	if marker != nil {
		list.Items.Remove(marker)
	}

	list.mutex.Unlock()
}

// RemoveHighNonceTxs removes "count" transactions from the back of the list
func (list *TxListForSender) RemoveHighNonceTxs(count int) [][]byte {
	removedTxHashes := make([][]byte, count)

	list.mutex.Lock()

	index := 0
	var previous *linkedList.Element
	for element := list.Items.Back(); element != nil && count > index; element = previous {
		// Remove node
		previous = element.Prev()
		list.Items.Remove(element)

		// Keep track of removed transaction
		value := element.Value.(TxListForSenderNode)
		removedTxHashes[index] = value.TxHash

		index++
	}

	list.mutex.Unlock()

	return removedTxHashes
}

func (list *TxListForSender) findTx(txToFind *transaction.Transaction) *linkedList.Element {
	for element := list.Items.Front(); element != nil; element = element.Next() {
		value := element.Value.(TxListForSenderNode)
		if value.Tx == txToFind {
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

// StartBatchCopying resets the internal state used for copy operations
func (list *TxListForSender) StartBatchCopying(batchSize int) {
	// We cannot copy or start copy from multiple goroutines at the same time
	list.mutex.Lock()

	list.copyBatchIndex = list.Items.Front()
	list.copyBatchSize = batchSize

	list.mutex.Unlock()
}

// CopyBatchTo copies a batch (usually small) of transactions to a destination slice
// It also updates the internal state used for copy operations
func (list *TxListForSender) CopyBatchTo(destination []*transaction.Transaction) int {
	element := list.copyBatchIndex
	batchSize := list.copyBatchSize
	availableSpace := len(destination)

	if element == nil {
		return 0
	}

	// We can't read from multiple goroutines at the same time
	// And we can't mutate the sender's list while reading it
	list.mutex.Lock()

	copied := 0
	for ; ; copied++ {
		if element == nil || copied == batchSize || copied == availableSpace {
			break
		}

		value := element.Value.(TxListForSenderNode)
		destination[copied] = value.Tx
		element = element.Next()
	}

	list.copyBatchIndex = element

	list.mutex.Unlock()
	return copied
}

// GetTxHashes returns the hashes of transactions in the list
func (list *TxListForSender) GetTxHashes() [][]byte {
	result := make([][]byte, list.Items.Len())

	index := 0
	for element := list.Items.Front(); element != nil; element = element.Next() {
		value := element.Value.(TxListForSenderNode)
		result[index] = value.TxHash
		index++
	}

	return result
}

func (list *TxListForSender) getHighestNonceTx() *transaction.Transaction {
	back := list.Items.Back()

	if back == nil {
		return nil
	}

	value := back.Value.(TxListForSenderNode)
	return value.Tx
}
