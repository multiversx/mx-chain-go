package txcache

import (
	"sort"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

// TxListBySenderMap is a map-like structure for holding and accessing transactions by sender
type TxListBySenderMap struct {
	backingMap      *ConcurrentMap
	counter         core.AtomicCounter
	nextOrderNumber core.AtomicCounter
}

// NewTxListBySenderMap creates a new instance of TxListBySenderMap
func NewTxListBySenderMap(size uint32, noChunksHint uint32) TxListBySenderMap {
	// We'll hold at most "size" lists of at least 1 transaction
	backingMap := NewConcurrentMap(size, noChunksHint)

	return TxListBySenderMap{
		backingMap: backingMap,
		counter:    0,
	}
}

// AddTx adds a transaction in the map, in the corresponding list (selected by its sender)
func (txMap *TxListBySenderMap) AddTx(txHash []byte, tx data.TransactionHandler) {
	sender := string(tx.GetSndAddress())
	listForSender := txMap.getOrAddListForSender(sender)
	listForSender.AddTx(txHash, tx)
}

func (txMap *TxListBySenderMap) getOrAddListForSender(sender string) *TxListForSender {
	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		listForSender = txMap.addSender(sender)
	}

	return listForSender
}

func (txMap *TxListBySenderMap) getListForSender(sender string) (*TxListForSender, bool) {
	listForSenderUntyped, ok := txMap.backingMap.Get(sender)
	if !ok {
		return nil, false
	}

	listForSender := listForSenderUntyped.(*TxListForSender)
	return listForSender, true
}

func (txMap *TxListBySenderMap) addSender(sender string) *TxListForSender {
	orderNumber := txMap.nextOrderNumber.Get()
	listForSender := NewTxListForSender(sender, orderNumber)

	txMap.backingMap.Set(sender, listForSender)
	txMap.counter.Increment()
	txMap.nextOrderNumber.Increment()

	return listForSender
}

// RemoveTx removes a transaction from the map
func (txMap *TxListBySenderMap) RemoveTx(tx data.TransactionHandler) bool {
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

func (txMap *TxListBySenderMap) removeSender(sender string) {
	if !txMap.backingMap.Has(sender) {
		return
	}

	txMap.backingMap.Remove(sender)
	txMap.counter.Decrement()
}

// RemoveSendersBulk removes senders, in bulk
func (txMap *TxListBySenderMap) RemoveSendersBulk(senders []string) uint32 {
	oldCount := uint32(txMap.counter.Get())

	for _, senderKey := range senders {
		txMap.removeSender(senderKey)
	}

	newCount := uint32(txMap.counter.Get())
	noRemoved := oldCount - newCount
	return noRemoved
}

// GetListsSortedByOrderNumber gets the list of sender addreses, sorted by the global order number
func (txMap *TxListBySenderMap) GetListsSortedByOrderNumber() []*TxListForSender {
	lists := make([]*TxListForSender, txMap.counter.Get())

	index := 0
	txMap.backingMap.IterCb(func(key string, item interface{}) {
		lists[index] = item.(*TxListForSender)
		index++
	})

	sort.Slice(lists, func(i, j int) bool {
		return lists[i].orderNumber < lists[j].orderNumber
	})

	return lists
}

// ForEachSender is an iterator callback
type ForEachSender func(key string, value *TxListForSender)

// ForEach iterates over the senders
func (txMap *TxListBySenderMap) ForEach(function ForEachSender) {
	txMap.backingMap.IterCb(func(key string, item interface{}) {
		txList := item.(*TxListForSender)
		function(key, txList)
	})
}
