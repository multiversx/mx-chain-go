package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

// Facade is the mock implementation of a transaction router handler
type Facade struct {
	GenerateTransactionHandler func(sender string, receiver string, value *big.Int,
		code string) (*transaction.Transaction, error)
	GetTransactionHandler  func(hash string) (*transaction.Transaction, error)
	SendTransactionHandler func(nonce uint64, sender string, receiver string, value *big.Int, code string,
		signature []byte) (*transaction.Transaction, error)
	GenerateAndSendBulkTransactionsHandler         func(destination string, value *big.Int, nrTransactions uint64) error
	GenerateAndSendBulkTransactionsOneByOneHandler func(destination string, value *big.Int, nrTransactions uint64) error
}

// GenerateTransaction is the mock implementation of a handler's GenerateTransaction method
func (f *Facade) GenerateTransaction(sender string, receiver string, value *big.Int,
	code string) (*transaction.Transaction, error) {
	return f.GenerateTransactionHandler(sender, receiver, value, code)
}

// GetTransaction is the mock implementation of a handler's GetTransaction method
func (f *Facade) GetTransaction(hash string) (*transaction.Transaction, error) {
	return f.GetTransactionHandler(hash)
}

// SendTransaction is the mock implementation of a handler's SendTransaction method
func (f *Facade) SendTransaction(nonce uint64, sender string, receiver string, value *big.Int, code string, signature []byte) (*transaction.Transaction, error) {
	return f.SendTransactionHandler(nonce, sender, receiver, value, code, signature)
}

// GenerateAndSendBulkTransactions is the mock implementation of a handler's GenerateAndSendBulkTransactions method
func (f *Facade) GenerateAndSendBulkTransactions(destination string, value *big.Int, nrTransactions uint64) error {
	return f.GenerateAndSendBulkTransactionsHandler(destination, value, nrTransactions)
}

// GenerateAndSendBulkTransactionsOneByOne is the mock implementation of a handler's GenerateAndSendBulkTransactionsOneByOne method
func (f *Facade) GenerateAndSendBulkTransactionsOneByOne(destination string, value *big.Int, nrTransactions uint64) error {
	return f.GenerateAndSendBulkTransactionsOneByOneHandler(destination, value, nrTransactions)
}

// WrongFacade is a struct that can be used as a wrong implementation of the node router handler
type WrongFacade struct {
}
