package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

type ExecTransactionMock struct {
	ProcessTransactionCalled func(transaction *transaction.Transaction, round int32) error
}

func (etm *ExecTransactionMock) SChandler() func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
	panic("implement me")
}

func (etm *ExecTransactionMock) SetSChandler(func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error) {
	panic("implement me")
}

func (etm *ExecTransactionMock) ProcessTransaction(transaction *transaction.Transaction, round int32) error {
	return etm.ProcessTransactionCalled(transaction, round)
}
