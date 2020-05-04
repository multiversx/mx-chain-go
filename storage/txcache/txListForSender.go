package txcache

import (
	"container/list"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/storage/txcache/maps"
)

const gracePeriodLowerBound = 5
const gracePeriodUpperBound = 7

// txListForSender represents a sorted list of transactions of a particular sender
type txListForSender struct {
	copyDetectedGap     bool
	copyPreviousNonce   uint64
	sender              string
	items               *list.List
	copyBatchIndex      *list.Element
	cacheConfig         *CacheConfig
	scoreChunk          *maps.MapChunkPointer
	mutex               sync.RWMutex
	accountNonceKnown   atomic.Flag
	sweepable           atomic.Flag
	lastComputedScore   atomic.Uint32
	accountNonce        atomic.Uint64
	totalBytes          atomic.Counter
	totalGas            atomic.Counter
	totalFee            atomic.Counter
	numFailedSelections atomic.Counter
	onScoreChange       scoreChangeCallback
}

type scoreChangeCallback func(value *txListForSender)

// newTxListForSender creates a new (sorted) list of transactions
func newTxListForSender(sender string, cacheConfig *CacheConfig, onScoreChange scoreChangeCallback) *txListForSender {
	return &txListForSender{
		items:         list.New(),
		sender:        sender,
		cacheConfig:   cacheConfig,
		onScoreChange: onScoreChange,
		scoreChunk:    &maps.MapChunkPointer{},
	}
}

// AddTx adds a transaction in sender's list
// This is a "sorted" insert
func (listForSender *txListForSender) AddTx(tx *WrappedTransaction) {
	// We don't allow concurrent interceptor goroutines to mutate a given sender's list
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	nonce := tx.Tx.GetNonce()
	gasPrice := tx.Tx.GetGasPrice()
	insertionPlace := listForSender.findInsertionPlace(nonce, gasPrice)

	if insertionPlace == nil {
		listForSender.items.PushFront(tx)
	} else {
		listForSender.items.InsertAfter(tx, insertionPlace)
	}

	listForSender.onAddedTransaction(tx)
}

func (listForSender *txListForSender) onAddedTransaction(tx *WrappedTransaction) {
	listForSender.totalBytes.Add(int64(estimateTxSize(tx)))
	listForSender.totalGas.Add(int64(estimateTxGas(tx)))
	listForSender.totalFee.Add(int64(estimateTxFee(tx)))
	listForSender.onScoreChange(listForSender)
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) findInsertionPlace(incomingNonce uint64, incomingGasPrice uint64) *list.Element {
	for element := listForSender.items.Back(); element != nil; element = element.Prev() {
		tx := element.Value.(*WrappedTransaction).Tx
		nonce := tx.GetNonce()
		gasPrice := tx.GetGasPrice()

		if nonce == incomingNonce && gasPrice > incomingGasPrice {
			return element
		}

		if nonce < incomingNonce {
			return element
		}
	}

	return nil
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
	value := element.Value.(*WrappedTransaction)

	listForSender.totalBytes.Subtract(int64(estimateTxSize(value)))
	listForSender.totalGas.Subtract(int64(estimateTxGas(value)))
	listForSender.totalGas.Subtract(int64(estimateTxFee(value)))
	listForSender.onScoreChange(listForSender)
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
		value := element.Value.(*WrappedTransaction)
		removedTxHashes[index] = value.TxHash

		index++
	}

	return removedTxHashes
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) findListElementWithTx(txToFind *WrappedTransaction) *list.Element {
	for element := listForSender.items.Front(); element != nil; element = element.Next() {
		value := element.Value.(*WrappedTransaction)

		if value == txToFind {
			return element
		}

		// Optimization: stop search at this point, since the list is sorted by nonce
		if value.Tx.GetNonce() > txToFind.Tx.GetNonce() {
			break
		}
	}

	return nil
}

// HasMoreThan checks whether the list has more items than specified
func (listForSender *txListForSender) HasMoreThan(count uint32) bool {
	return uint32(listForSender.countTxWithLock()) > count
}

// IsEmpty checks whether the list is empty
func (listForSender *txListForSender) IsEmpty() bool {
	return listForSender.countTxWithLock() == 0
}

// selectBatchTo copies a batch (usually small) of transactions to a destination slice
// It also updates the internal state used for copy operations
func (listForSender *txListForSender) selectBatchTo(isFirstBatch bool, destination []*WrappedTransaction, batchSize int) int {
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

	// If a nonce gap is detected, no transaction is returned in this read.
	// There is an exception though: if this is the first read operation for the sender in the current selection process and the sender is in the grace period,
	// then one transaction will be returned. But subsequent reads for this sender will return nothing.
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

		value := element.Value.(*WrappedTransaction)
		txNonce := value.Tx.GetNonce()

		if previousNonce > 0 && txNonce > previousNonce+1 {
			listForSender.copyDetectedGap = true
			break
		}

		destination[copied] = value
		element = element.Next()
		previousNonce = txNonce
	}

	listForSender.copyBatchIndex = element
	listForSender.copyPreviousNonce = previousNonce
	return copied
}

// getTxHashes returns the hashes of transactions in the list
func (listForSender *txListForSender) getTxHashes() [][]byte {
	listForSender.mutex.RLock()
	defer listForSender.mutex.RUnlock()

	result := make([][]byte, 0, listForSender.countTx())

	for element := listForSender.items.Front(); element != nil; element = element.Next() {
		value := element.Value.(*WrappedTransaction)
		result = append(result, value.TxHash)
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

	firstTxNonce := firstTx.Tx.GetNonce()
	accountNonce := listForSender.accountNonce.Get()
	hasGap := firstTxNonce > accountNonce
	return hasGap
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) getLowestNonceTx() *WrappedTransaction {
	front := listForSender.items.Front()
	if front == nil {
		return nil
	}

	value := front.Value.(*WrappedTransaction)
	return value
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
