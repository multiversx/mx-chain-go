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
		log.Error("txListBySenderMap.removeTx() detected inconsistency: sender of tx not in cache", "sender", sender)
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

type txListBySenderSortKind string

// LRUCache is currently the only supported Cache type
const (
	SortByOrderNumberAsc txListBySenderSortKind = "SortByOrderNumberAsc"
	SortByTotalBytesDesc txListBySenderSortKind = "SortByTotalBytesDesc"
	SortByTotalGas       txListBySenderSortKind = "SortByTotalGas"
	SortBySmartScore     txListBySenderSortKind = "SortBySmartScore"
)

func (txMap *txListBySenderMap) GetListsSortedBy(sortKind txListBySenderSortKind) []*txListForSender {
	// TODO-TXCACHE: do partial sort? optimization.

	switch sortKind {
	case SortByOrderNumberAsc:
		return txMap.GetListsSortedByOrderNumber()
	case SortByTotalBytesDesc:
		return txMap.GetListsSortedByTotalBytes()
	case SortByTotalGas:
		return txMap.GetListsSortedByTotalGas()
	case SortBySmartScore:
		return txMap.GetListsSortedBySmartScore()
	default:
		return txMap.GetListsSortedByOrderNumber()
	}
}

// GetListsSortedByOrderNumber gets the list of sender addreses, sorted by the global order number, ascending
func (txMap *txListBySenderMap) GetListsSortedByOrderNumber() []*txListForSender {
	lists := txMap.getListsSortedByFunc(func(txListA, txListB *txListForSender) bool {
		return txListA.orderNumber < txListB.orderNumber
	})

	return lists
}

// GetListsSortedByTotalBytes gets the list of sender addreses, sorted by the total amount of bytes, descending
func (txMap *txListBySenderMap) GetListsSortedByTotalBytes() []*txListForSender {
	lists := txMap.getListsSortedByFunc(func(txListA, txListB *txListForSender) bool {
		return txListA.totalBytes > txListB.totalBytes
	})

	return lists
}

// GetListsSortedByTotalGas gets the list of sender addreses, sorted by the total amoung of gas, ascending
func (txMap *txListBySenderMap) GetListsSortedByTotalGas() []*txListForSender {
	lists := txMap.getListsSortedByFunc(func(txListA, txListB *txListForSender) bool {
		return txListA.totalGas < txListB.totalGas
	})

	return lists
}

// GetListsSortedBySmartScore gets the list of sender addreses, sorted by a smart score
// A low score means that the sender will be evicted
// A high score means that the sender will not be evicted, most probably
// Score is:
// - inversely proportional to sender's tx count
// - inversely proportional to sender's tx total size
// - directly proportional to sender's order number
// - directly proportional to sender's tx total gas
func (txMap *txListBySenderMap) GetListsSortedBySmartScore() []*txListForSender {
	// First get bounds for gas and size, in order to normalize the values when computing the score
	maxGas := int64(0)
	minGas := int64(0)
	maxSize := int64(0)
	minSize := int64(0)

	txMap.forEach(func(key string, item *txListForSender) {
		gas := item.totalGas.Get()
		size := item.totalBytes.Get()

		if gas > maxGas {
			maxGas = gas
		}
		if gas < minGas {
			minGas = gas
		}

		if size > maxSize {
			maxSize = gas
		}
		if size < minSize {
			minSize = size
		}
	})

	// Normalized values are quantized (in the range 0 - 100)
	// This way, sort is also a bit more optimized (less item movement)
	// And partial, approximate sort is sufficient
	gasRange := maxGas - minGas
	sizeRange := maxSize - minSize
	if gasRange == 0 {
		gasRange = 1
	}
	if sizeRange == 0 {
		sizeRange = 1
	}

	lists := txMap.getListsSortedByFunc(func(txListA, txListB *txListForSender) bool {
		gasA := float64(txListA.totalGas.Get()-minGas) / float64(gasRange)
		gasB := float64(txListB.totalGas.Get()-minGas) / float64(gasRange)

		scoreA := gasA * float64(txListA.orderNumber) * 100
		scoreB := gasB * float64(txListB.orderNumber) * 100

		return scoreA < scoreB
	})

	return lists
}

func (txMap *txListBySenderMap) getListsSortedByFunc(less func(txListA, txListB *txListForSender) bool) []*txListForSender {
	counter := txMap.counter.Get()
	if counter < 1 {
		return make([]*txListForSender, 0)
	}

	lists := make([]*txListForSender, 0, counter)

	txMap.forEach(func(key string, item *txListForSender) {
		lists = append(lists, item)
	})

	sort.Slice(lists, func(i, j int) bool {
		return less(lists[i], lists[j])
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

func (txMap *txListBySenderMap) clear() {
	txMap.backingMap.Clear()
	txMap.counter.Set(0)
}

func (txMap *txListBySenderMap) countTxBySender(sender string) int {
	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		return 0
	}

	return listForSender.countTx()
}
