package txcache

import (
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/storage/txcache/maps"
)

const numberOfScoreChunks = uint32(100)

// txListBySenderMap is a map-like structure for holding and accessing transactions by sender
type txListBySenderMap struct {
	backingMap  *maps.BucketSortedMap
	cacheConfig CacheConfig
	counter     atomic.Counter
}

// newTxListBySenderMap creates a new instance of TxListBySenderMap
func newTxListBySenderMap(nChunksHint uint32, cacheConfig CacheConfig) txListBySenderMap {
	backingMap := maps.NewBucketSortedMap(nChunksHint, numberOfScoreChunks)

	return txListBySenderMap{
		backingMap:  backingMap,
		cacheConfig: cacheConfig,
	}
}

// addTx adds a transaction in the map, in the corresponding list (selected by its sender)
func (txMap *txListBySenderMap) addTx(tx *WrappedTransaction) {
	sender := string(tx.Tx.GetSndAddress())
	listForSender := txMap.getOrAddListForSender(sender)
	listForSender.AddTx(tx)
	txMap.notifyScoreChange(listForSender)
}

func (txMap *txListBySenderMap) notifyScoreChange(txList *txListForSender) {
	txMap.backingMap.NotifyScoreChange(txList)
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
	listForSender := newTxListForSender(sender, &txMap.cacheConfig)

	txMap.backingMap.Set(listForSender)
	txMap.counter.Increment()

	return listForSender
}

// removeTx removes a transaction from the map
func (txMap *txListBySenderMap) removeTx(tx *WrappedTransaction) bool {
	sender := string(tx.Tx.GetSndAddress())

	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		txMap.onRemoveTxInconsistency(sender)
		return false
	}

	isFound := listForSender.RemoveTx(tx)

	if listForSender.IsEmpty() {
		txMap.removeSender(sender)
	} else {
		txMap.notifyScoreChange(listForSender)
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

func (txMap *txListBySenderMap) notifyAccountNonce(accountKey []byte, nonce uint64) {
	sender := string(accountKey)
	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		// Discard the notification
		return
	}

	listForSender.notifyAccountNonce(nonce)
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

func (txMap *txListBySenderMap) forEachDescending(function ForEachSender) {
	txMap.backingMap.IterCbSortedDescending(func(key string, item maps.BucketSortedMapItem) {
		txList := item.(*txListForSender)
		function(key, txList)
	})
}

func (txMap *txListBySenderMap) clear() {
	txMap.backingMap.Clear()
	txMap.counter.Set(0)
}
