package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

type ShardCoordinatorMock struct {
	SelfShardId uint32
}

func (scm ShardCoordinatorMock) NumberOfShards() uint32 {
	panic("implement me")
}

func (scm ShardCoordinatorMock) ComputeId(address state.AddressContainer) uint32 {
	return 0
}

func (scm ShardCoordinatorMock) SetSelfShardId(shardId uint32) error {
	scm.SelfShardId = shardId
	return nil
}

func (scm ShardCoordinatorMock) SelfId() uint32 {
	return scm.SelfShardId
}

func (scm ShardCoordinatorMock) SameShard(firstAddress, secondAddress state.AddressContainer) bool {
	return true
}

func (scm ShardCoordinatorMock) CommunicationIdentifier(destShardID uint32) string {
	if destShardID == sharding.MetachainShardId {
		return "_0_META"
	}

	return "_0"
}
