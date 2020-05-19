package txcache

import (
	"fmt"
)

func (cache *TxCache) addTxDebug(correlation string, tx *WrappedTransaction) (ok bool, added bool) {
	ok = true
	added, _ = cache.txListBySender.addTxDebug(correlation, tx)
	if added {
		cache.txByHash.addTx(tx)
		fmt.Println(correlation, "added")
	} else {
		fmt.Println(correlation, "not added")
	}

	return
}

func (txMap *txListBySenderMap) addTxDebug(correlation string, tx *WrappedTransaction) (bool, txHashes) {
	sender := string(tx.Tx.GetSndAddr())
	listForSender := txMap.getOrAddListForSender(sender)
	return listForSender.addTxDebug(correlation, tx)
}

func (listForSender *txListForSender) addTxDebug(correlation string, tx *WrappedTransaction) (bool, txHashes) {
	// We don't allow concurrent interceptor goroutines to mutate a given sender's list
	listForSender.mutex.Lock()
	defer listForSender.mutex.Unlock()

	insertionPlace, err := listForSender.findInsertionPlace(tx)
	if err != nil {
		fmt.Println(correlation, "duplicated")
		return false, nil
	}

	if insertionPlace == nil {
		listForSender.items.PushFront(tx)
	} else {
		listForSender.items.InsertAfter(tx, insertionPlace)
	}

	return true, make([][]byte, 0)
}
