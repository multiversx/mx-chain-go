package factory

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// TransactionCoordinatorFactoryStub -
type TransactionCoordinatorFactoryStub struct {
}

// CreateTransactionCoordinator -
func (t *TransactionCoordinatorFactoryStub) CreateTransactionCoordinator(args coordinator.ArgTransactionCoordinator) (process.TransactionCoordinator, error) {
	return &testscommon.TransactionCoordinatorMock{}, nil
}

// IsInterfaceNil -
func (t *TransactionCoordinatorFactoryStub) IsInterfaceNil() bool {
	return t == nil
}
