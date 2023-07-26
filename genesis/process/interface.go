package process

import (
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/coordinator"
)

// TransactionCoordinatorCreator defines the transaction coordinator factory creator
type TransactionCoordinatorCreator interface {
	CreateTransactionCoordinator(argsTransactionCoordinator coordinator.ArgTransactionCoordinator) (process.TransactionCoordinator, error)
	IsInterfaceNil() bool
}
