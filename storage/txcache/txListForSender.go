package txcache

import (
	"container/list"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/storage/txcache/maps"
)

const gracePeriodLowerBound = 5
const gracePeriodUpperBound = 7

// txListForSender represents a sorted list of transactions of a particular sender
type txListForSender struct {
	items                 *list.List
	mutex                 sync.RWMutex
	copyBatchIndex        *list.Element
	copyPreviousNonce     uint64
	copyDetectedGap       bool
	totalBytes            atomic.Counter
	totalGas              atomic.Counter
	totalFee              atomic.Counter
	sender                string
	scoreChunk            *maps.MapChunk
	scoreChangeInProgress atomic.Flag
	lastComputedScore     atomic.Uint32
	cacheConfig           *CacheConfig
	accountNonce          atomic.Uint64
	accountNonceKnown     atomic.Flag
	numFailedSelections   atomic.Counter
	sweepable             atomic.Flag
}

// txListForSenderNode is a node of the linked list
type txListForSenderNode struct {
	txHash []byte
	tx     data.TransactionHandler
}

// newTxListForSender creates a new (sorted) list of transactions
func newTxListForSender(sender string, cacheConfig *CacheConfig) *txListForSender {
	return &txListForSender{
		items:       list.New(),
		sender:      sender,
		cacheConfig: cacheConfig,
	}
}

// AddTx adds a transaction in sender's list
// This is a "sorted" insert
func (listForSender *txListForSender) AddTx(txHash []byte, tx data.TransactionHandler) {
	// We don't allow concurent interceptor goroutines to mutate a given sender's list
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	nonce := tx.GetNonce()
	mark := listForSender.findTxWithLowerNonce(nonce)
	newNode := txListForSenderNode{txHash, tx}

	if mark == nil {
		listForSender.items.PushFront(newNode)
	} else {
		listForSender.items.InsertAfter(newNode, mark)
	}

	listForSender.onAddedTransaction(tx)
}

func (listForSender *txListForSender) onAddedTransaction(tx data.TransactionHandler) {
	listForSender.totalBytes.Add(int64(estimateTxSize(tx)))
	listForSender.totalGas.Add(int64(estimateTxGas(tx)))
	listForSender.totalFee.Add(int64(estimateTxFee(tx)))
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) findTxWithLowerNonce(nonce uint64) *list.Element {
	for element := listForSender.items.Back(); element != nil; element = element.Prev() {
		value := element.Value.(txListForSenderNode)
		if value.tx.GetNonce() < nonce {
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
		listForSender.onRemovedListElement(marker)
	}

	return isFound
}

func (listForSender *txListForSender) onRemovedListElement(element *list.Element) {
	value := element.Value.(txListForSenderNode)

	listForSender.totalBytes.Subtract(int64(estimateTxSize(value.tx)))
	listForSender.totalGas.Subtract(int64(estimateTxGas(value.tx)))
	listForSender.totalGas.Subtract(int64(estimateTxFee(value.tx)))
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
		listForSender.onRemovedListElement(element)

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
	return uint32(listForSender.countTx()) > count
}

// IsEmpty checks whether the list is empty
func (listForSender *txListForSender) IsEmpty() bool {
	return listForSender.countTx() == 0
}

// selectBatchTo copies a batch (usually small) of transactions to a destination slice
// It also updates the internal state used for copy operations
func (listForSender *txListForSender) selectBatchTo(isFirstBatch bool, destination []data.TransactionHandler, destinationHashes [][]byte, batchSize int) int {
	// We can't read from multiple goroutines at the same time
	// And we can't mutate the sender's list while reading it
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	// Reset the internal state used for copy operations
	if isFirstBatch {
		hasInitialGap := listForSender.verifyInitialGapOnSelectionStart()

		listForSender.copyBatchIndex = listForSender.items.Front()
		listForSender.copyPreviousNonce = 0
		listForSender.copyDetectedGap = hasInitialGap
	}

	element := listForSender.copyBatchIndex
	availableSpace := len(destination)
	detectedGap := listForSender.copyDetectedGap
	previousNonce := listForSender.copyPreviousNonce

	if detectedGap {
		if isFirstBatch && listForSender.isInGracePeriod() {
			batchSize = 1
		} else {
			batchSize = 0
		}
	}

	copied := 0
	for ; ; copied++ {
		if element == nil || copied == batchSize || copied == availableSpace {
			break
		}

		value := element.Value.(txListForSenderNode)
		tx := value.tx
		txNonce := tx.GetNonce()

		if previousNonce > 0 && txNonce > previousNonce+1 {
			listForSender.copyDetectedGap = true
			break
		}

		destination[copied] = tx
		destinationHashes[copied] = value.txHash
		element = element.Next()
		previousNonce = txNonce
	}

	listForSender.copyBatchIndex = element
	listForSender.copyPreviousNonce = previousNonce
	return copied
}

// getTxHashes returns the hashes of transactions in the list
func (listForSender *txListForSender) getTxHashes() [][]byte {
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	result := make([][]byte, 0, listForSender.countTx())

	for element := listForSender.items.Front(); element != nil; element = element.Next() {
		value := element.Value.(txListForSenderNode)
		result = append(result, value.txHash)
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

func approximatelyCountTxInLists(lists []*txListForSender) uint64 {
	count := uint64(0)

	for _, listForSender := range lists {
		count += listForSender.countTxWithLock()
	}

	return count
}

// notifyAccountNonce does not update the "sweepable" flag, nor the "numFailedSelections" counter,
// since the notification comes at a time when we cannot actually detect whether the initial gap still exists or it was resolved.
func (listForSender *txListForSender) notifyAccountNonce(nonce uint64) {
	listForSender.accountNonce.Set(nonce)
	listForSender.accountNonceKnown.Set()
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) verifyInitialGapOnSelectionStart() bool {
	hasInitialGap := listForSender.hasInitialGap()

	if hasInitialGap {
		listForSender.numFailedSelections.Increment()

		if listForSender.isGracePeriodExceeded() {
			listForSender.sweepable.Set()
		}
	} else {
		listForSender.numFailedSelections.Reset()
		listForSender.sweepable.Unset()
	}

	return hasInitialGap
}

// hasInitialGap should only be called at tx selection time, since only then we can detect initial gaps with certainty
// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) hasInitialGap() bool {
	accountNonceKnown := listForSender.accountNonceKnown.IsSet()
	if !accountNonceKnown {
		return false
	}

	firstTx := listForSender.getLowestNonceTx()
	if firstTx == nil {
		return false
	}

	firstTxNonce := firstTx.GetNonce()
	accountNonce := listForSender.accountNonce.Get()
	hasGap := firstTxNonce > accountNonce
	return hasGap
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) getLowestNonceTx() data.TransactionHandler {
	front := listForSender.items.Front()
	if front == nil {
		return nil
	}

	value := front.Value.(txListForSenderNode)
	return value.tx
}

// isInGracePeriod returns whether the sender is grace period due to a number of failed selections
func (listForSender *txListForSender) isInGracePeriod() bool {
	numFailedSelections := listForSender.numFailedSelections.Get()
	return numFailedSelections >= gracePeriodLowerBound && numFailedSelections <= gracePeriodUpperBound
}

func (listForSender *txListForSender) isGracePeriodExceeded() bool {
	numFailedSelections := listForSender.numFailedSelections.Get()
	return numFailedSelections > gracePeriodUpperBound
}
