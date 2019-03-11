package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type ShardCoordinatorMock struct {
}

func (scm ShardCoordinatorMock) CurrentNumberOfShards() uint32 {
	panic("implement me")
}

func (scm ShardCoordinatorMock) ComputeShardForAddress(address state.AddressContainer) uint32 {
	panic("implement me")
}

func (scm ShardCoordinatorMock) SetCurrentShardId(shardId uint32) error {
	panic("implement me")
}

func (scm ShardCoordinatorMock) CurrentShardId() uint32 {
	return 0
}
