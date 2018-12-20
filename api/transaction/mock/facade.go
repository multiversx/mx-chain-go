package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"math/big"
)

type Facade struct {
	GenerateTransactionHandler func(sender string, receiver string, amount big.Int,
		code string) (*transaction.Transaction, error)
	GetTransactionHandler func(hash string) (*transaction.Transaction, error)
}

func (f *Facade) GenerateTransaction(sender string, receiver string, amount big.Int,
	code string) (*transaction.Transaction, error) {
	return f.GenerateTransactionHandler(sender, receiver, amount, code)
}

func (f *Facade) GetTransaction(hash string) (*transaction.Transaction, error) {
	return f.GetTransactionHandler(hash)
}

type WrongFacade struct {
}
