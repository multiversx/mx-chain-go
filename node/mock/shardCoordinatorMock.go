package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type ShardCoordinatorMock struct {
	SelfIdField uint32
}

func (scm ShardCoordinatorMock) NumberOfShards() uint32 {
	panic("implement me")
}

func (scm ShardCoordinatorMock) ComputeId(address state.AddressContainer) uint32 {
	panic("implement me")
}

func (scm ShardCoordinatorMock) SetSelfId(shardId uint32) error {
	scm.SelfIdField = shardId
	return nil
}

func (scm ShardCoordinatorMock) SelfId() uint32 {
	return scm.SelfIdField
}

func (scm ShardCoordinatorMock) SameShard(firstAddress, secondAddress state.AddressContainer) bool {
	return true
}

func (scm ShardCoordinatorMock) CommunicationIdentifier(destShardID uint32) string {
	return "_0"
}
