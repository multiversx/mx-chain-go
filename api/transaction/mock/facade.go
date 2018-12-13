package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"math/big"
)

type Facade struct {
	GenerateTransactionCalled func(sender string, receiver string, amount big.Int,
		code string) (*transaction.Transaction, error)
	GetTransactionCalled func(hash string) (*transaction.Transaction, error)
}

func (f *Facade) GenerateTransaction(sender string, receiver string, amount big.Int,
	code string) (*transaction.Transaction, error) {
	return f.GenerateTransactionCalled(sender, receiver, amount, code)
}

func (f *Facade) GetTransaction(hash string) (*transaction.Transaction, error) {
	return f.GetTransactionCalled(hash)
}

type WrongFacade struct {
}
