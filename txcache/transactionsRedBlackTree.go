package txcache

import (
	"bytes"
	"container/list"

	"github.com/emirpasic/gods/trees/redblacktree"
)

// TransactionsComparator defines the comparator used for a red black tree.
// Inside txListForSender, the transactions should be ordered:
// in ascending order of the nonce.
// in descending order of the gas price, in case two or more txs have same nonce.
// in ascending order of hash, in case two or more txs have same nonce and same gas price.
// TransactionsComparator should reflect the order of the txListForSender.
func TransactionsComparator(tx1 interface{}, tx2 interface{}) int {
	wrappedTx1 := tx1.(*list.Element).Value.(*WrappedTransaction)
	wrappedTx2 := tx2.(*list.Element).Value.(*WrappedTransaction)

	// a transaction with lower nonce should be placed before the one with bigger nonce
	if wrappedTx1.Tx.GetNonce() != wrappedTx2.Tx.GetNonce() {
		if wrappedTx1.Tx.GetNonce() > wrappedTx2.Tx.GetNonce() {
			return 1
		}

		return -1
	}

	// a transactions with less gas price should be placed after the one with more gas price
	if wrappedTx1.Tx.GetGasPrice() != wrappedTx2.Tx.GetGasPrice() {
		if wrappedTx1.Tx.GetGasPrice() > wrappedTx2.Tx.GetGasPrice() {
			return -1
		}

		return 1
	}

	// a transaction with lower txHash should be placed before the one with bigger txHash
	return bytes.Compare(wrappedTx1.TxHash, wrappedTx2.TxHash)
}

// transactionsRedBlackTree defines a wrapper structure over a red black tree.
// The tree contains nodes where the key is a pointer to an element from the txListForSender.
// This helps to find faster where a new element should be inserted.
// Before adding the new element in the txListForSender, the FindInsertionPlace of the transactionsRedBlackTree should be called.
// The FindInsertionPlace will return the potential parent of the new element in the tree.
// The new element must be added in the txListForSender right after the returned element.
// Each time an element is added or removed from the txListForSender, it should be added or removed from the tree too.
type transactionsRedBlackTree struct {
	tree *redblacktree.Tree
}

// NewTransactionsRedBlackTree should create a new TransactionsRedBlackTree.
func NewTransactionsRedBlackTree() *transactionsRedBlackTree {
	return &transactionsRedBlackTree{
		tree: redblacktree.NewWith(TransactionsComparator),
	}
}

// Insert should insert the element in the tree.
func (trb *transactionsRedBlackTree) Insert(e *list.Element) {
	trb.tree.Put(e, struct{}{})
}

// Remove should remove an element from the tree.
func (trb *transactionsRedBlackTree) Remove(e *list.Element) {
	trb.tree.Remove(e)
}

// contains should return an error if an element is already in the tree.
func (trb *transactionsRedBlackTree) contains(e *list.Element) error {
	_, found := trb.tree.Get(e)
	if found {
		return errItemAlreadyInCache
	}

	return nil
}

// FindInsertionPlace should return the insertion place of a new element.
// If the new element is already present, an error will be returned.
// If the new element is lower than all the other ones, the FindInsertionPlace will return nil.
// This means that inside txListForSender, the element will become the new head.
// However, in the other cases, the parent will be returned and inside txListForSender the new element should be placed right after the parent.
func (trb *transactionsRedBlackTree) FindInsertionPlace(e *list.Element) (*list.Element, error) {
	err := trb.contains(e)
	if err != nil {
		return nil, err
	}

	node, found := trb.tree.Floor(e)
	if found {
		// our new element should be placed after this one
		return (node.Key).(*list.Element), nil
	}

	return nil, nil
}

// Len should return the size of the tree.
func (trb *transactionsRedBlackTree) Len() uint64 {
	return uint64(trb.tree.Size())
}
