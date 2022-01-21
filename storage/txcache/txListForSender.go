package txcache

import (
	"bytes"
	"container/list"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/txcache/maps"
)

var _ maps.BucketSortedMapItem = (*txListForSender)(nil)

// txListForSender represents a sorted list of transactions of a particular sender
type txListForSender struct {
	copyDetectedGap     bool
	lastComputedScore   atomic.Uint32
	accountNonceKnown   atomic.Flag
	sweepable           atomic.Flag
	copyPreviousNonce   uint64
	sender              string
	items               *list.List
	copyBatchIndex      *list.Element
	constraints         *senderConstraints
	scoreChunk          *maps.MapChunk
	accountNonce        atomic.Uint64
	totalBytes          atomic.Counter
	totalGas            atomic.Counter
	totalFeeScore       atomic.Counter
	numFailedSelections atomic.Counter
	onScoreChange       scoreChangeCallback

	scoreChunkMutex sync.RWMutex
	mutex           sync.RWMutex
}

type scoreChangeCallback func(value *txListForSender, scoreParams senderScoreParams)

// newTxListForSender creates a new (sorted) list of transactions
func newTxListForSender(sender string, constraints *senderConstraints, onScoreChange scoreChangeCallback) *txListForSender {
	return &txListForSender{
		items:         list.New(),
		sender:        sender,
		constraints:   constraints,
		onScoreChange: onScoreChange,
	}
}

// AddTx adds a transaction in sender's list
// This is a "sorted" insert
func (listForSender *txListForSender) AddTx(tx *WrappedTransaction, gasHandler TxGasHandler, txFeeHelper feeHelper) (bool, [][]byte) {
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

	listForSender.onAddedTransaction(tx, gasHandler, txFeeHelper)
	evicted := listForSender.applySizeConstraints()
	listForSender.triggerScoreChange()
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

func (listForSender *txListForSender) onAddedTransaction(tx *WrappedTransaction, gasHandler TxGasHandler, txFeeHelper feeHelper) {
	listForSender.totalBytes.Add(tx.Size)
	listForSender.totalGas.Add(int64(estimateTxGas(tx)))
	listForSender.totalFeeScore.Add(int64(estimateTxFeeScore(tx, gasHandler, txFeeHelper)))
}

func (listForSender *txListForSender) triggerScoreChange() {
	scoreParams := listForSender.getScoreParams()
	listForSender.onScoreChange(listForSender, scoreParams)
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) getScoreParams() senderScoreParams {
	fee := listForSender.totalFeeScore.GetUint64()
	gas := listForSender.totalGas.GetUint64()
	count := listForSender.countTx()

	return senderScoreParams{count: count, feeScore: fee, gas: gas}
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) findInsertionPlace(incomingTx *WrappedTransaction) (*list.Element, error) {
	incomingNonce := incomingTx.Tx.GetNonce()
	incomingGasPrice := incomingTx.Tx.GetGasPrice()

	for element := listForSender.items.Back(); element != nil; element = element.Prev() {
		currentTx := element.Value.(*WrappedTransaction)
		currentTxNonce := currentTx.Tx.GetNonce()
		currentTxGasPrice := currentTx.Tx.GetGasPrice()

		if incomingTx.sameAs(currentTx) {
			// The incoming transaction will be discarded
			return nil, storage.ErrItemAlreadyInCache
		}

		if currentTxNonce == incomingNonce && currentTxGasPrice > incomingGasPrice {
			// The incoming transaction will be placed right after the existing one, which has same nonce but higher price.
			// If the nonces are the same, but the incoming gas price is higher or equal, the search loop continues.
			return element, nil
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
		listForSender.triggerScoreChange()
	}

	return isFound
}

func (listForSender *txListForSender) onRemovedListElement(element *list.Element) {
	value := element.Value.(*WrappedTransaction)

	listForSender.totalBytes.Subtract(value.Size)
	listForSender.totalGas.Subtract(int64(estimateTxGas(value)))
	listForSender.totalFeeScore.Subtract(int64(value.TxFeeScoreNormalized))
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) findListElementWithTx(txToFind *WrappedTransaction) *list.Element {
	txToFindHash := txToFind.TxHash
	txToFindNonce := txToFind.Tx.GetNonce()

	for element := listForSender.items.Front(); element != nil; element = element.Next() {
		value := element.Value.(*WrappedTransaction)

		if bytes.Equal(value.TxHash, txToFindHash) {
			return element
		}

		// Optimization: stop search at this point, since the list is sorted by nonce
		if value.Tx.GetNonce() > txToFindNonce {
			break
		}
	}

	return nil
}

// IsEmpty checks whether the list is empty
func (listForSender *txListForSender) IsEmpty() bool {
	return listForSender.countTxWithLock() == 0
}

// selectBatchTo copies a batch (usually small) of transactions of a limited gas bandwidth and limited number of transactions to a destination slice
// It also updates the internal state used for copy operations
func (listForSender *txListForSender) selectBatchTo(isFirstBatch bool, destination []*WrappedTransaction, batchSize int, bandwidth uint64) batchSelectionJournal {
	// We can't read from multiple goroutines at the same time
	// And we can't mutate the sender's list while reading it
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	journal := batchSelectionJournal{}

	// Reset the internal state used for copy operations
	if isFirstBatch {
		hasInitialGap := listForSender.verifyInitialGapOnSelectionStart()

		listForSender.copyBatchIndex = listForSender.items.Front()
		listForSender.copyPreviousNonce = 0
		listForSender.copyDetectedGap = hasInitialGap

		journal.isFirstBatch = true
		journal.hasInitialGap = hasInitialGap
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
			journal.isGracePeriod = true
			batchSize = 1
		} else {
			batchSize = 0
		}
	}

	copiedBandwidth := uint64(0)
	lastTxGasLimit := uint64(0)
	copied := 0
	for ; ; copied, copiedBandwidth = copied+1, copiedBandwidth+lastTxGasLimit {
		if element == nil || copied == batchSize || copied == availableSpace || copiedBandwidth >= bandwidth {
			break
		}

		value := element.Value.(*WrappedTransaction)
		txNonce := value.Tx.GetNonce()
		lastTxGasLimit = value.Tx.GetGasLimit()

		if previousNonce > 0 && txNonce > previousNonce+1 {
			listForSender.copyDetectedGap = true
			journal.hasMiddleGap = true
			break
		}

		destination[copied] = value
		element = element.Next()
		previousNonce = txNonce
	}

	listForSender.copyBatchIndex = element
	listForSender.copyPreviousNonce = previousNonce
	journal.copied = copied
	return journal
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

// notifyAccountNonce does not update the "numFailedSelections" counter,
// since the notification comes at a time when we cannot actually detect whether the initial gap still exists or it was resolved.
func (listForSender *txListForSender) notifyAccountNonce(nonce uint64) {
	listForSender.accountNonce.Set(nonce)
	_ = listForSender.accountNonceKnown.SetReturningPrevious()
}

// This function should only be used in critical section (listForSender.mutex)
func (listForSender *txListForSender) verifyInitialGapOnSelectionStart() bool {
	hasInitialGap := listForSender.hasInitialGap()

	if hasInitialGap {
		listForSender.numFailedSelections.Increment()

		if listForSender.isGracePeriodExceeded() {
			_ = listForSender.sweepable.SetReturningPrevious()
		}
	} else {
		listForSender.numFailedSelections.Reset()
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
	return numFailedSelections >= senderGracePeriodLowerBound && numFailedSelections <= senderGracePeriodUpperBound
}

func (listForSender *txListForSender) isGracePeriodExceeded() bool {
	numFailedSelections := listForSender.numFailedSelections.Get()
	return numFailedSelections > senderGracePeriodUpperBound
}

func (listForSender *txListForSender) getLastComputedScore() uint32 {
	return listForSender.lastComputedScore.Get()
}

func (listForSender *txListForSender) setLastComputedScore(score uint32) {
	listForSender.lastComputedScore.Set(score)
}

// GetKey returns the key
func (listForSender *txListForSender) GetKey() string {
	return listForSender.sender
}

// GetScoreChunk returns the score chunk the sender is currently in
func (listForSender *txListForSender) GetScoreChunk() *maps.MapChunk {
	listForSender.scoreChunkMutex.RLock()
	defer listForSender.scoreChunkMutex.RUnlock()

	return listForSender.scoreChunk
}

// SetScoreChunk returns the score chunk the sender is currently in
func (listForSender *txListForSender) SetScoreChunk(scoreChunk *maps.MapChunk) {
	listForSender.scoreChunkMutex.Lock()
	listForSender.scoreChunk = scoreChunk
	listForSender.scoreChunkMutex.Unlock()
}
