package sovereign

import (
	"time"

	"github.com/multiversx/mx-chain-go/dataRetriever/dataPool/sovereign"
	"github.com/multiversx/mx-chain-go/process/block"
)

type sovereignOutGoingOperationPoolFactory struct {
	timeToWait time.Duration
}

// NewSovereignOutGoingOperationPoolFactory create a new sovereign outgoing operation pool factory
func NewSovereignOutGoingOperationPoolFactory(duration uint32) *sovereignOutGoingOperationPoolFactory {
	return &sovereignOutGoingOperationPoolFactory{
		timeToWait: time.Second * time.Duration(duration),
	}
}

// CreateOutGoingOperationPool creates a new outgoing operation pool for the chain run type sovereign
func (soopf *sovereignOutGoingOperationPoolFactory) CreateOutGoingOperationPool() block.OutGoingOperationsPool {
	return sovereign.NewOutGoingOperationPool(soopf.timeToWait)
}

// IsInterfaceNil returns true if there is no value under the interface
func (soopf *sovereignOutGoingOperationPoolFactory) IsInterfaceNil() bool {
	return soopf == nil
}
