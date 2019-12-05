package txcache

import (
	linkedList "container/list"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// TxListForSender is
type TxListForSender struct {
	CopyBatchIndex *linkedList.Element
	CopyBatchSize  int
	Items          *linkedList.List
}

// addTransaction is an inserted sort
func (list *TxListForSender) addTransaction(tx *transaction.Transaction) {
	if list.Items == nil {
		list.Items = linkedList.New()
	}

	nonce := tx.Nonce
	mark := list.findTransactionWithLargerNonce(nonce)
	if mark == nil {
		list.Items.PushBack(tx)
	} else {
		list.Items.InsertBefore(tx, mark)
	}
}

func (list *TxListForSender) findTransactionWithLargerNonce(nonce uint64) *linkedList.Element {
	for element := list.Items.Front(); element != nil; element = element.Next() {
		tx := element.Value.(*transaction.Transaction)
		if tx.Nonce > nonce {
			return element
		}
	}

	return nil
}

func (list *TxListForSender) removeTransaction(tx *transaction.Transaction) {
	marker := list.findTransaction(tx)
	list.Items.Remove(marker)
}

func (list *TxListForSender) findTransaction(txToFind *transaction.Transaction) *linkedList.Element {
	for element := list.Items.Front(); element != nil; element = element.Next() {
		tx := element.Value.(*transaction.Transaction)
		if tx == txToFind {
			return element
		}
	}

	return nil
}

func (list *TxListForSender) isEmpty() bool {
	return list.Items.Len() == 0
}

func (list *TxListForSender) restartBatchCopying(batchSize int) {
	list.CopyBatchIndex = list.Items.Front()
	list.CopyBatchSize = batchSize
}

func (list *TxListForSender) copyBatchTo(destination []*transaction.Transaction) int {
	element := list.CopyBatchIndex
	batchSize := list.CopyBatchSize
	availableLength := len(destination)

	if element == nil {
		return 0
	}

	// todo rewrite loop, make it more readable
	copied := 0
	for true {
		if element == nil || copied == batchSize || availableLength == copied {
			break
		}

		tx := element.Value.(*transaction.Transaction)
		destination[copied] = tx
		copied++
		element = element.Next()
	}

	list.CopyBatchIndex = element
	return copied
}
