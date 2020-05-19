package txcache

import (
	"fmt"
)

func (cache *TxCache) addTxDebug(correlation string, tx *WrappedTransaction) (ok bool, added bool) {
	ok = true
	added, _ = cache.txListBySender.addTxDebug(correlation, tx)
	if added {
		fmt.Println(correlation, "now add to map by hash")
		addedInByHash := cache.txByHash.addTx(tx)
		if !addedInByHash {
			fmt.Println(correlation, "not added in map by hash, already there")
		} else {
			fmt.Println(correlation, "added in map by hash")
		}
	} else {
		fmt.Println(correlation, "not added at all (duplicated?)")
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

	fmt.Println(correlation, "added tx", fmt.Sprintf("%p", tx), "to list of sender", fmt.Sprintf("%p", listForSender), "of size", listForSender.countTx())
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
		fmt.Println(correlation, "remove: tx", fmt.Sprintf("%p", tx), "not found in sender")
	} else {
		fmt.Println(correlation, "removed from sender")
	}
}

func (txMap *txListBySenderMap) removeTxDebug(correlation string, tx *WrappedTransaction) bool {
	sender := string(tx.Tx.GetSndAddr())

	listForSender, ok := txMap.getListForSender(sender)
	if !ok {
		fmt.Println(correlation, "sender to remove not in cache")
		return false
	}

	fmt.Println(correlation, "will remove tx", fmt.Sprintf("%p", tx), "from sender", fmt.Sprintf("%p", listForSender))
	isFound := listForSender.RemoveTx(tx)
	isEmpty := listForSender.IsEmpty()
	if isEmpty {
		fmt.Println(correlation, "empty sender", fmt.Sprintf("%p", listForSender), "will be removed")
		// This removal is correlated to failure. Without this, no identity inconsistency.
		txMap.removeSender(sender)
	}

	return isFound
}
