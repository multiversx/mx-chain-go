package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

type TxProcessorMock struct {
	ProcessTransactionCalled func(transaction *transaction.Transaction, round uint32) error
	SetBalancesToTrieCalled  func(accBalance map[string]*big.Int) (rootHash []byte, err error)
}

func (etm *TxProcessorMock) SCHandler() func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
	panic("implement me")
}

func (etm *TxProcessorMock) SetSCHandler(func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error) {
	panic("implement me")
}

func (etm *TxProcessorMock) ProcessTransaction(transaction *transaction.Transaction, round uint32) error {
	return etm.ProcessTransactionCalled(transaction, round)
}

func (etm *TxProcessorMock) SetBalancesToTrie(accBalance map[string]*big.Int) (rootHash []byte, err error) {
	return etm.SetBalancesToTrieCalled(accBalance)
}
