package transaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type TransactionNewer struct {
	*transaction.Transaction
}

func NewTransactionNewer() *TransactionNewer {
	return &TransactionNewer{Transaction: &transaction.Transaction{}}
}

func (tn *TransactionNewer) New() p2p.Newer {
	return NewTransactionNewer()
}

func (tn *TransactionNewer) ID() string {
	return string(tn.Transaction.Signature)
}
