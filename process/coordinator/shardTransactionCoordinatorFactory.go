package coordinator

import "github.com/multiversx/mx-chain-go/process"

type shardTransactionCoordinatorFactory struct {
}

// NewShardTransactionCoordinatorFactory creates a new sovereign transaction coordinator factory
func NewShardTransactionCoordinatorFactory() (*shardTransactionCoordinatorFactory, error) {
	return &shardTransactionCoordinatorFactory{}, nil
}

// CreateTransactionCoordinator creates a new sovereign transaction coordinator for the chain run type sovereign
func (stcf *shardTransactionCoordinatorFactory) CreateTransactionCoordinator(argsTransactionCoordinator ArgTransactionCoordinator) (process.TransactionCoordinator, error) {
	return NewTransactionCoordinator(argsTransactionCoordinator)
}

// IsInterfaceNil returns true if there is no value under the interface
func (stcf *shardTransactionCoordinatorFactory) IsInterfaceNil() bool {
	return stcf == nil
}
