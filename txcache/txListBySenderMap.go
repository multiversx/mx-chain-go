package txcache

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-storage-go/txcache/maps"
)

// txListBySenderMap is a map-like structure for holding and accessing transactions by sender
type txListBySenderMap struct {
	backingMap        *maps.ConcurrentMap
	senderConstraints senderConstraints
	counter           atomic.Counter
	scoreComputer     scoreComputer
	txGasHandler      TxGasHandler
	mutex             sync.Mutex
}

// newTxListBySenderMap creates a new instance of TxListBySenderMap
func newTxListBySenderMap(
	nChunksHint uint32,
	senderConstraints senderConstraints,
	scoreComputer scoreComputer,
	txGasHandler TxGasHandler,
) *txListBySenderMap {
	backingMap := maps.NewConcurrentMap(nChunksHint)

	return &txListBySenderMap{
		backingMap:        backingMap,
		senderConstraints: senderConstraints,
		scoreComputer:     scoreComputer,
		txGasHandler:      txGasHandler,
	}
}

// addTx adds a transaction in the map, in the corresponding list (selected by its sender)
func (txMap *txListBySenderMap) addTx(tx *WrappedTransaction) (bool, [][]byte) {
	sender := string(tx.Tx.GetSndAddr())
	listForSender := txMap.getOrAddListForSender(sender)
	return listForSender.AddTx(tx, txMap.txGasHandler)
}

// getOrAddListForSender gets or lazily creates a list (using double-checked locking pattern)
func (txMap *txListBySenderMap) getOrAddListForSender(sender string) *txListForSender {
	listForSender, ok := txMap.getListForSender(sender)
	if ok {
		return listForSender
	}

	txMap.mutex.Lock()
	defer txMap.mutex.Unlock()

	listForSender, ok = txMap.getListForSender(sender)
	if ok {
		return listForSender
	}

	return txMap.addSender(sender)
}

func (txMap *txListBySenderMap) getListForSender(sender string) (*txListForSender, bool) {
	listForSenderUntyped, ok := txMap.backingMap.Get(sender)
	if !ok {
		return nil, false
	}

	listForSender := listForSenderUntyped.(*txListForSender)
	return listForSender, true
}

func (txMap *txListBySenderMap) addSender(sender string) *txListForSender {
	listForSender := newTxListForSender(sender, &txMap.senderConstraints, txMap.scoreComputer)

	txMap.backingMap.Set(sender, listForSender)
	txMap.counter.Increment()

	return listForSender
}

// removeTx removes a transaction from the map
func (txMap *txListBySenderMap) removeTx(tx *WrappedTransaction) bool {
	sender := string(tx.Tx.GetSndAddr())

	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		// This happens when a sender whose transactions were selected for processing is removed from cache in the meantime.
		// When it comes to remove one if its transactions due to processing (commited / finalized block), they don't exist in cache anymore.
		log.Debug("txListBySenderMap.removeTx() detected slight inconsistency: sender of tx not in cache", "tx", tx.TxHash, "sender", []byte(sender))
		return false
	}

	isFound := listForSender.RemoveTx(tx)
	isEmpty := listForSender.IsEmpty()
	if isEmpty {
		txMap.removeSender(sender)
	}

	return isFound
}

// Important: this doesn't remove the transactions from txCache.txByHash. That's done by the caller.
func (txMap *txListBySenderMap) removeSender(sender string) bool {
	_, removed := txMap.backingMap.Remove(sender)
	if removed {
		txMap.counter.Decrement()
	}

	return removed
}

// RemoveSendersBulk removes senders, in bulk
func (txMap *txListBySenderMap) RemoveSendersBulk(senders []string) uint32 {
	numRemoved := uint32(0)

	for _, senderKey := range senders {
		if txMap.removeSender(senderKey) {
			numRemoved++
		}
	}

	return numRemoved
}

func (txMap *txListBySenderMap) notifyAccountNonce(accountKey []byte, nonce uint64) [][]byte {
	sender := string(accountKey)
	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		return nil
	}

	return listForSender.notifyAccountNonce(nonce)
}

func (txMap *txListBySenderMap) getSnapshotAscending() []*txListForSender {
	scoreGroups := txMap.getSendersGroupedByScore()
	listsSnapshot := make([]*txListForSender, 0, txMap.counter.Get())

	for i := 0; i < len(scoreGroups); i++ {
		listsSnapshot = append(listsSnapshot, scoreGroups[i]...)
	}

	return listsSnapshot
}

func (txMap *txListBySenderMap) getSnapshotDescending() []*txListForSender {
	scoreGroups := txMap.getSendersGroupedByScore()
	listsSnapshot := make([]*txListForSender, 0, txMap.counter.Get())

	for i := len(scoreGroups) - 1; i >= 0; i-- {
		listsSnapshot = append(listsSnapshot, scoreGroups[i]...)
	}

	return listsSnapshot
}

func (txMap *txListBySenderMap) getSendersGroupedByScore() [][]*txListForSender {
	groups := make([][]*txListForSender, maxSenderScore+1)
	// Hint for pre-allocating slice for each group (imprecise, but reasonable).
	groupSizeHint := txMap.counter.Get() / int64(maxSenderScore) / 2

	txMap.backingMap.IterCb(func(key string, item interface{}) {
		listForSender := item.(*txListForSender)
		score := listForSender.getScore()

		if groups[score] == nil {
			groups[score] = make([]*txListForSender, 0, groupSizeHint)
		}

		// TODO (next PR) randomize / shuffle.
		groups[score] = append(groups[score], listForSender)
	})

	monitorSendersScoreHistogram(groups)

	return groups
}

func (txMap *txListBySenderMap) clear() {
	txMap.backingMap.Clear()
	txMap.counter.Set(0)
}
