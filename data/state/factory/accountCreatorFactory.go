package factory

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// NewAccountFactoryCreator returns an account factory depending on shard coordinator self id
func NewAccountFactoryCreator(coordinator sharding.Coordinator) (state.AccountFactory, error) {
	if coordinator == nil {
		return nil, state.ErrNilShardCoordinator
	}

	if coordinator.SelfId() < coordinator.NumberOfShards() {
		return NewAccountCreator(), nil
	}

	if coordinator.SelfId() == sharding.MetachainShardId {
		return NewMetaAccountCreator(), nil
	}

	return nil, state.ErrUnknownShardId
}
