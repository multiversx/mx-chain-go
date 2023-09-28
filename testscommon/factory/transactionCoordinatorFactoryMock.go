package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/coordinator"
)

// TransactionCoordinatorFactoryMock -
type TransactionCoordinatorFactoryMock struct {
	CreateTransactionCoordinatorCalled func(args coordinator.ArgTransactionCoordinator) (process.TransactionCoordinator, error)
}

// CreateTransactionCoordinator -
func (t *TransactionCoordinatorFactoryMock) CreateTransactionCoordinator(args coordinator.ArgTransactionCoordinator) (process.TransactionCoordinator, error) {
	if t.CreateTransactionCoordinatorCalled != nil {
		return t.CreateTransactionCoordinatorCalled(args)
	}
	return nil, nil
}

// IsInterfaceNil -
func (t *TransactionCoordinatorFactoryMock) IsInterfaceNil() bool {
	return t == nil
}
