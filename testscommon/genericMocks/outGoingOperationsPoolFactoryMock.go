package genericMocks

import (
	"github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
)

// OutGoingOperationsFactoryMock -
type OutGoingOperationsFactoryMock struct {
	CreateOutGoingOperationPoolCalled func() block.OutGoingOperationsPool
}

// CreateOutGoingOperationPool -
func (oof *OutGoingOperationsFactoryMock) CreateOutGoingOperationPool() block.OutGoingOperationsPool {
	if oof.CreateOutGoingOperationPoolCalled != nil {
		return oof.CreateOutGoingOperationPool()
	}
	return &sovereign.OutGoingOperationsPoolMock{}
}

// IsInterfaceNil -
func (oof *OutGoingOperationsFactoryMock) IsInterfaceNil() bool {
	return oof == nil
}
