package txcache

import (
	"container/list"
	"sync"

	"github.com/ElrondNetwork/elrond-go/data"
)

// TxListForSender is
type TxListForSender struct {
	Items          *list.List
	mutex          sync.Mutex
	copyBatchIndex *list.Element
	copyBatchSize  int
	orderNumber    int64
	sender         string
}

// TxListForSenderNode is a node of the linked list
type TxListForSenderNode struct {
	TxHash []byte
	Tx     data.TransactionHandler
}

// NewTxListForSender creates a new (sorted) list of transactions
func NewTxListForSender(sender string, globalIndex int64) *TxListForSender {
	return &TxListForSender{
		Items:       list.New(),
		orderNumber: globalIndex,
		sender:      sender,
	}
}

// AddTx adds a transaction in sender's list
// This is a "sorted" insert
func (listForSender *TxListForSender) AddTx(txHash []byte, tx data.TransactionHandler) {
	// We don't allow concurent interceptor goroutines to mutate a given sender's list
	listForSender.mutex.Lock()

	nonce := tx.GetNonce()
	mark := listForSender.findTxWithLargerNonce(nonce)
	newNode := TxListForSenderNode{txHash, tx}

	if mark == nil {
		listForSender.Items.PushBack(newNode)
	} else {
		listForSender.Items.InsertBefore(newNode, mark)
	}

	listForSender.mutex.Unlock()
}

func (listForSender *TxListForSender) findTxWithLargerNonce(nonce uint64) *list.Element {
	for element := listForSender.Items.Front(); element != nil; element = element.Next() {
		value := element.Value.(TxListForSenderNode)
		if value.Tx.GetNonce() > nonce {
			return element
		}
	}

	return nil
}

// RemoveTx removes a transaction from the sender's list
func (listForSender *TxListForSender) RemoveTx(tx data.TransactionHandler) bool {
	// We don't allow concurent interceptor goroutines to mutate a given sender's list
	listForSender.mutex.Lock()

	marker := listForSender.findTx(tx)
	isFound := marker != nil
	if isFound {
		listForSender.Items.Remove(marker)
	}

	listForSender.mutex.Unlock()

	return isFound
}

// RemoveHighNonceTxs removes "count" transactions from the back of the list
func (listForSender *TxListForSender) RemoveHighNonceTxs(count uint32) [][]byte {
	removedTxHashes := make([][]byte, count)

	listForSender.mutex.Lock()

	index := uint32(0)
	var previous *list.Element
	for element := listForSender.Items.Back(); element != nil && count > index; element = previous {
		// Remove node
		previous = element.Prev()
		listForSender.Items.Remove(element)

		// Keep track of removed transaction
		value := element.Value.(TxListForSenderNode)
		removedTxHashes[index] = value.TxHash

		index++
	}

	listForSender.mutex.Unlock()

	return removedTxHashes
}

func (listForSender *TxListForSender) findTx(txToFind data.TransactionHandler) *list.Element {
	for element := listForSender.Items.Front(); element != nil; element = element.Next() {
		value := element.Value.(TxListForSenderNode)
		if value.Tx == txToFind {
			return element
		}
	}

	return nil
}

// HasMoreThan checks whether the list has more items than specified
func (listForSender *TxListForSender) HasMoreThan(count uint32) bool {
	return uint32(listForSender.Items.Len()) > count
}

// IsEmpty checks whether the list is empty
func (listForSender *TxListForSender) IsEmpty() bool {
	return listForSender.Items.Len() == 0
}

// StartBatchCopying resets the internal state used for copy operations
func (listForSender *TxListForSender) StartBatchCopying(batchSize int) {
	// We cannot copy or start copy from multiple goroutines at the same time
	listForSender.mutex.Lock()

	listForSender.copyBatchIndex = listForSender.Items.Front()
	listForSender.copyBatchSize = batchSize

	listForSender.mutex.Unlock()
}

// CopyBatchTo copies a batch (usually small) of transactions to a destination slice
// It also updates the internal state used for copy operations
func (listForSender *TxListForSender) CopyBatchTo(destination []data.TransactionHandler) int {
	element := listForSender.copyBatchIndex
	batchSize := listForSender.copyBatchSize
	availableSpace := len(destination)

	if element == nil {
		return 0
	}

	// We can't read from multiple goroutines at the same time
	// And we can't mutate the sender's list while reading it
	listForSender.mutex.Lock()

	copied := 0
	for ; ; copied++ {
		if element == nil || copied == batchSize || copied == availableSpace {
			break
		}

		value := element.Value.(TxListForSenderNode)
		destination[copied] = value.Tx
		element = element.Next()
	}

	listForSender.copyBatchIndex = element

	listForSender.mutex.Unlock()
	return copied
}

// GetTxHashes returns the hashes of transactions in the list
func (listForSender *TxListForSender) GetTxHashes() [][]byte {
	result := make([][]byte, listForSender.Items.Len())

	index := 0
	for element := listForSender.Items.Front(); element != nil; element = element.Next() {
		value := element.Value.(TxListForSenderNode)
		result[index] = value.TxHash
		index++
	}

	return result
}

func (listForSender *TxListForSender) getHighestNonceTx() data.TransactionHandler {
	back := listForSender.Items.Back()

	if back == nil {
		return nil
	}

	value := back.Value.(TxListForSenderNode)
	return value.Tx
}
