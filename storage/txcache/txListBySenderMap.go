package txcache

import (
	"sort"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

// txListBySenderMap is a map-like structure for holding and accessing transactions by sender
type txListBySenderMap struct {
	backingMap      *ConcurrentMap
	counter         core.AtomicCounter
	nextOrderNumber core.AtomicCounter
}

// newTxListBySenderMap creates a new instance of TxListBySenderMap
func newTxListBySenderMap(nChunksHint uint32) txListBySenderMap {
	backingMap := NewConcurrentMap(nChunksHint)

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
	orderNumber := txMap.nextOrderNumber.Get()
	listForSender := newTxListForSender(sender, orderNumber)

	txMap.backingMap.Set(sender, listForSender)
	txMap.counter.Increment()
	txMap.nextOrderNumber.Increment()

	return listForSender
}

// removeTx removes a transaction from the map
func (txMap *txListBySenderMap) removeTx(tx data.TransactionHandler) bool {
	sender := string(tx.GetSndAddress())

	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		return false
	}

	isFound := listForSender.RemoveTx(tx)

	if listForSender.IsEmpty() {
		txMap.removeSender(sender)
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

// GetListsSortedByOrderNumber gets the list of sender addreses, sorted by the global order number
func (txMap *txListBySenderMap) GetListsSortedByOrderNumber() []*txListForSender {
	lists := make([]*txListForSender, txMap.counter.Get())

	index := 0
	txMap.backingMap.IterCb(func(key string, item interface{}) {
		lists[index] = item.(*txListForSender)
		index++
	})

	sort.Slice(lists, func(i, j int) bool {
		return lists[i].orderNumber < lists[j].orderNumber
	})

	return lists
}

// ForEachSender is an iterator callback
type ForEachSender func(key string, value *txListForSender)

// forEach iterates over the senders
func (txMap *txListBySenderMap) forEach(function ForEachSender) {
	txMap.backingMap.IterCb(func(key string, item interface{}) {
		txList := item.(*txListForSender)
		function(key, txList)
	})
}
