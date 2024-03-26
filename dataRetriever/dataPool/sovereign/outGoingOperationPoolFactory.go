package sovereign

import (
	"github.com/multiversx/mx-chain-go/common/disabled"
	"github.com/multiversx/mx-chain-go/process/block"
)

type outGoingOperationPoolFactory struct {
}

// NewOutGoingOperationPoolFactory create a new outgoing operation pool factory
func NewOutGoingOperationPoolFactory() *outGoingOperationPoolFactory {
	return &outGoingOperationPoolFactory{}
}

// CreateOutGoingOperationPool creates a new outgoing operation pool for the chain run type normal
func (oopf *outGoingOperationPoolFactory) CreateOutGoingOperationPool() block.OutGoingOperationsPool {
	return disabled.NewDisabledOutGoingOperationPool()
}

// IsInterfaceNil returns true if there is no value under the interface
func (oopf *outGoingOperationPoolFactory) IsInterfaceNil() bool {
	return oopf == nil
}
