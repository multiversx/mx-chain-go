package coordinator

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
)

type TransactionCoordinatorCreator interface {
	CreateTransactionCoordinator(argsTransactionCoordinator ArgTransactionCoordinator) (process.TransactionCoordinator, error)
	IsInterfaceNil() bool
}

type sovereignTransactionCoordinatorFactory struct {
	transactionCoordinatorCreator TransactionCoordinatorCreator
}

// NewSovereignTransactionCoordinatorFactory creates a new sovereign transaction coordinator factory
func NewSovereignTransactionCoordinatorFactory(tcf TransactionCoordinatorCreator) (*sovereignTransactionCoordinatorFactory, error) {
	if check.IfNil(tcf) {
		return nil, process.ErrNilTransactionCoordinatorCreator
	}

	return &sovereignTransactionCoordinatorFactory{
		transactionCoordinatorCreator: tcf,
	}, nil
}

// CreateTransactionCoordinator creates a new sovereign transaction coordinator for the chain run type sovereign
func (stcf *sovereignTransactionCoordinatorFactory) CreateTransactionCoordinator(argsTransactionCoordinator ArgTransactionCoordinator) (process.TransactionCoordinator, error) {
	tc, err := stcf.transactionCoordinatorCreator.CreateTransactionCoordinator(argsTransactionCoordinator)
	if err != nil {
		return nil, err
	}

	return NewSovereignChainTransactionCoordinator(tc)
}

// IsInterfaceNil returns true if there is no value under the interface
func (stcf *sovereignTransactionCoordinatorFactory) IsInterfaceNil() bool {
	return stcf == nil
}
