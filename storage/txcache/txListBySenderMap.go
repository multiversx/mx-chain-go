package txcache

import "github.com/ElrondNetwork/elrond-go/data/transaction"

// TxListBySenderMap is
type TxListBySenderMap struct {
	Map     *ConcurrentMap
	Counter AtomicCounter
}

// NewTxListBySenderMap creates a new map-like structure for holding and accessing transactions by sender
func NewTxListBySenderMap(size int, shardsHint int) TxListBySenderMap {
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
	listForSender := NewTxListForSender()
	txMap.Map.Set(sender, listForSender)
	txMap.Counter.Increment()
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
	txMap.Map.Remove(sender)
	txMap.Counter.Decrement()
}

func (txMap *TxListBySenderMap) removeSenders(senders []string) {
	for _, senderKey := range senders {
		txMap.removeSender(senderKey)
	}
}
