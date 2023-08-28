package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/coordinator"
)

// TransactionCoordinatorFactoryStub -
type TransactionCoordinatorFactoryStub struct {
	CreateTransactionCoordinatorCalled func(args coordinator.ArgTransactionCoordinator) (process.TransactionCoordinator, error)
}

// NewTransactionCoordinatorFactoryStub -
func NewTransactionCoordinatorFactoryStub() *TransactionCoordinatorFactoryStub {
	return &TransactionCoordinatorFactoryStub{}
}

// CreateTransactionCoordinator -
func (t *TransactionCoordinatorFactoryStub) CreateTransactionCoordinator(args coordinator.ArgTransactionCoordinator) (process.TransactionCoordinator, error) {
	if t.CreateTransactionCoordinatorCalled != nil {
		return t.CreateTransactionCoordinatorCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (t *TransactionCoordinatorFactoryStub) IsInterfaceNil() bool {
	return false
}
