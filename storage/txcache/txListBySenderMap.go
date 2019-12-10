package txcache

import (
	"sort"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// TxListBySenderMap is a map-like structure for holding and accessing transactions by sender
type TxListBySenderMap struct {
	Map             *ConcurrentMap
	Counter         core.AtomicCounter
	nextOrderNumber core.AtomicCounter
}

// NewTxListBySenderMap creates a new instance of TxListBySenderMap
func NewTxListBySenderMap(size uint32, shardsHint uint32) TxListBySenderMap {
	// We'll hold at most "size" lists of at least 1 transaction
	backingMap := NewConcurrentMap(size, shardsHint)

	return TxListBySenderMap{
		Map:     backingMap,
		Counter: 0,
	}
}

// AddTx adds a transaction in the map, in the corresponding list (selected by its sender)
func (txMap *TxListBySenderMap) AddTx(txHash []byte, tx *transaction.Transaction) {
	sender := string(tx.SndAddr)
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
	listForSenderUntyped, ok := txMap.Map.Get(sender)
	if !ok {
		return nil, false
	}

	listForSender := listForSenderUntyped.(*TxListForSender)
	return listForSender, true
}

func (txMap *TxListBySenderMap) addSender(sender string) *TxListForSender {
	orderNumber := txMap.nextOrderNumber.Get()
	listForSender := NewTxListForSender(sender, orderNumber)

	txMap.Map.Set(sender, listForSender)
	txMap.Counter.Increment()
	txMap.nextOrderNumber.Increment()

	return listForSender
}

// RemoveTx removes a transaction from the map
func (txMap *TxListBySenderMap) RemoveTx(tx *transaction.Transaction) {
	sender := string(tx.SndAddr)
	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		// This should never happen (eviction should never cause this kind of inconsistency between the two internal maps)
		return
	}

	listForSender.RemoveTx(tx)
	if listForSender.IsEmpty() {
		txMap.removeSender(sender)
	}
}

func (txMap *TxListBySenderMap) removeSender(sender string) {
	if !txMap.Map.Has(sender) {
		return
	}

	txMap.Map.Remove(sender)
	txMap.Counter.Decrement()
}

// RemoveSendersBulk removes senders, in bulk
func (txMap *TxListBySenderMap) RemoveSendersBulk(senders []string) uint32 {
	oldCount := uint32(txMap.Counter.Get())

	for _, senderKey := range senders {
		txMap.removeSender(senderKey)
	}

	newCount := uint32(txMap.Counter.Get())
	noRemoved := oldCount - newCount
	return noRemoved
}

// GetListsSortedByOrderNumber gets the list of sender addreses, sorted by the global order number
func (txMap *TxListBySenderMap) GetListsSortedByOrderNumber() []*TxListForSender {
	lists := make([]*TxListForSender, txMap.Counter.Get())

	index := 0
	txMap.Map.IterCb(func(key string, item interface{}) {
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
	txMap.Map.IterCb(func(key string, item interface{}) {
		txList := item.(*TxListForSender)
		function(key, txList)
	})
}
