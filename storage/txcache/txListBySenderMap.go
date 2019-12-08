package txcache

import "github.com/ElrondNetwork/elrond-go/data/transaction"

// TxListBySenderMap is
type TxListBySenderMap struct {
	Map     *ConcurrentMap
	Counter AtomicCounter
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

func (txMap *TxListBySenderMap) addTransaction(txHash []byte, tx *transaction.Transaction) {
	sender := string(tx.SndAddr)
	listForSender := txMap.getOrAddListForSender(sender)
	listForSender.AddTransaction(txHash, tx)
}

func (txMap *TxListBySenderMap) removeTransaction(tx *transaction.Transaction) {
	sender := string(tx.SndAddr)
	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		// This should never happen (eviction should never cause this kind of inconsistency between the two internal maps)
		return
	}

	listForSender.RemoveTransaction(tx)
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
