package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"math/big"
)

type TxProcessorMock struct {
	ProcessTransactionCalled func(transaction *transaction.Transaction) error
	SetBalancesToTrieCalled  func(accBalance map[string]big.Int) (rootHash []byte, err error)
}

func (etm *TxProcessorMock) SCHandler() func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
	panic("implement me")
}

func (etm *TxProcessorMock) SetSCHandler(func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error) {
	panic("implement me")
}

func (etm *TxProcessorMock) ProcessTransaction(transaction *transaction.Transaction) error {
	return etm.ProcessTransactionCalled(transaction)
}

func (etm *TxProcessorMock) SetBalancesToTrie(accBalance map[string]big.Int) (rootHash []byte, err error) {
	return etm.SetBalancesToTrieCalled(accBalance)
}
