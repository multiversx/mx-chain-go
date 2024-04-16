package testscommon

import (
	"github.com/multiversx/mx-chain-go/sharding"
)

// MultiShardCoordinatorFactoryMock -
type MultiShardCoordinatorFactoryMock struct {
	CreateShardCoordinatorCalled func(numberOfShards, selfId uint32) (sharding.Coordinator, error)
}

// CreateShardCoordinator -
func (msc *MultiShardCoordinatorFactoryMock) CreateShardCoordinator(numberOfShards, selfId uint32) (sharding.Coordinator, error) {
	if msc.CreateShardCoordinatorCalled != nil {
		return msc.CreateShardCoordinator(numberOfShards, selfId)
	}
	return nil, nil
}

// IsInterfaceNil checks if the underlying pointer is nil
func (msc *MultiShardCoordinatorFactoryMock) IsInterfaceNil() bool {
	return msc == nil
}
