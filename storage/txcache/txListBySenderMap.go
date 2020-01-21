package txcache

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

const numberOfScoreChunks = uint32(100)

// txListBySenderMap is a map-like structure for holding and accessing transactions by sender
type txListBySenderMap struct {
	backingMap *AlmostSortedMap
	counter    core.AtomicCounter
}

// newTxListBySenderMap creates a new instance of TxListBySenderMap
func newTxListBySenderMap(nChunksHint uint32) txListBySenderMap {
	backingMap := NewAlmostSortedMap(nChunksHint, numberOfScoreChunks)

	return txListBySenderMap{
		backingMap: backingMap,
		counter:    0,
	}
}

// addTx adds a transaction in the map, in the corresponding list (selected by its sender)
func (txMap *txListBySenderMap) addTx(txHash []byte, tx data.TransactionHandler) {
	sender := string(tx.GetSndAddress())
	listForSender := txMap.getOrAddListForSender(sender)
	listForSender.AddTx(txHash, tx)
	txMap.backingMap.OnScoreChange(listForSender)
}

func (txMap *txListBySenderMap) getOrAddListForSender(sender string) *txListForSender {
	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		listForSender = txMap.addSender(sender)
	}

	return listForSender
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
	listForSender := newTxListForSender(sender)

	txMap.backingMap.Set(listForSender)
	txMap.counter.Increment()

	return listForSender
}

// removeTx removes a transaction from the map
func (txMap *txListBySenderMap) removeTx(tx data.TransactionHandler) bool {
	sender := string(tx.GetSndAddress())

	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		txMap.onRemoveTxInconsistency(sender)
		return false
	}

	isFound := listForSender.RemoveTx(tx)

	if listForSender.IsEmpty() {
		txMap.removeSender(sender)
	} else {
		txMap.backingMap.OnScoreChange(listForSender)
	}

	return isFound
}

func (txMap *txListBySenderMap) removeSender(sender string) {
	if !txMap.backingMap.Has(sender) {
		return
	}

	txMap.backingMap.Remove(sender)
	txMap.counter.Decrement()
}

// RemoveSendersBulk removes senders, in bulk
func (txMap *txListBySenderMap) RemoveSendersBulk(senders []string) uint32 {
	oldCount := uint32(txMap.counter.Get())

	for _, senderKey := range senders {
		txMap.removeSender(senderKey)
	}

	newCount := uint32(txMap.counter.Get())
	nRemoved := oldCount - newCount
	return nRemoved
}

func (txMap *txListBySenderMap) getSnapshotAscending() []*txListForSender {
	itemsSnapshot := txMap.backingMap.GetSnapshotAscending()
	listsSnapshot := make([]*txListForSender, len(itemsSnapshot))

	for i, item := range itemsSnapshot {
		listsSnapshot[i] = item.(*txListForSender)
	}

	return listsSnapshot
}

// ForEachSender is an iterator callback
type ForEachSender func(key string, value *txListForSender)

func (txMap *txListBySenderMap) forEachAscending(function ForEachSender) {
	txMap.backingMap.IterCbSortedAscending(func(key string, item ScoredItem) {
		txList := item.(*txListForSender)
		function(key, txList)
	})
}

func (txMap *txListBySenderMap) forEachDescending(function ForEachSender) {
	txMap.backingMap.IterCbSortedDescending(func(key string, item ScoredItem) {
		txList := item.(*txListForSender)
		function(key, txList)
	})
}

func (txMap *txListBySenderMap) clear() {
	txMap.backingMap.Clear()
	txMap.counter.Set(0)
}

func (txMap *txListBySenderMap) countTxBySender(sender string) int64 {
	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		return 0
	}

	return listForSender.countTx()
}
