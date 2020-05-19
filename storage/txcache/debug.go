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

func (cache *TxCache) removeDebug(correlation string, key []byte) {
	cache.removeTxByHashDebug(correlation, key)
}

func (cache *TxCache) removeTxByHashDebug(correlation string, txHash []byte) {
	tx, ok := cache.txByHash.removeTx(string(txHash))
	if !ok {
		fmt.Println(correlation, "remove: not found in txByHash")
		return
	}

	found := cache.txListBySender.removeTxDebug(correlation, tx)
	if !found {
		fmt.Println(correlation, "remove: not found in txListBySender")
	}
}

func (txMap *txListBySenderMap) removeTxDebug(correlation string, tx *WrappedTransaction) bool {
	sender := string(tx.Tx.GetSndAddr())

	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		fmt.Println(correlation, "sender to remove not in cache")
		return false
	}

	isFound := listForSender.RemoveTx(tx)
	isEmpty := listForSender.IsEmpty()
	if isEmpty {
		fmt.Println(correlation, "empty sender will be removed")
		txMap.removeSender(sender)
	}

	return isFound
}
