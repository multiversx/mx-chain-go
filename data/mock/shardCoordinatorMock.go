package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type ShardCoordinatorMock struct {
	SelfID     uint32
	NrOfShards uint32
}

func (scm ShardCoordinatorMock) NumberOfShards() uint32 {
	return scm.NrOfShards
}

func (scm ShardCoordinatorMock) ComputeId(address state.AddressContainer) uint32 {
	panic("implement me")
}

func (scm ShardCoordinatorMock) SetSelfId(shardId uint32) error {
	panic("implement me")
}

func (scm ShardCoordinatorMock) SelfId() uint32 {
	return scm.SelfID
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
