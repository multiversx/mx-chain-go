package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

type ExecTransactionMock struct {
	ProcessTransactionCalled   func(transaction *transaction.Transaction) error
	SetRegisterHandlerCalled   func(data []byte) error
	SetUnregisterHandlerCalled func(data []byte) error
}

func (etm *ExecTransactionMock) SChandler() func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error {
	panic("implement me")
}

func (etm *ExecTransactionMock) SetSChandler(func(accountsAdapter state.AccountsAdapter, transaction *transaction.Transaction) error) {
	panic("implement me")
}

func (etm *ExecTransactionMock) RegisterHandler() func(data []byte) error {
	panic("implement me")
}

func (etm *ExecTransactionMock) SetRegisterHandler(f func(data []byte) error) {
	etm.SetRegisterHandlerCalled = f
}

func (etm *ExecTransactionMock) UnregisterHandler() func(data []byte) error {
	panic("implement me")
}

func (etm *ExecTransactionMock) SetUnregisterHandler(f func(data []byte) error) {
	etm.SetUnregisterHandlerCalled = f
}

func (etm *ExecTransactionMock) ProcessTransaction(transaction *transaction.Transaction) error {
	return etm.ProcessTransactionCalled(transaction)
}
