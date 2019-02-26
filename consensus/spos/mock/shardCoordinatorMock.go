package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type ShardCoordinatorMock struct {
}

func (scm ShardCoordinatorMock) NoShards() uint32 {
	panic("implement me")
}

func (scm ShardCoordinatorMock) SetNoShards(uint32) {
	panic("implement me")
}

func (scm ShardCoordinatorMock) ComputeShardForAddress(address state.AddressContainer, addressConverter state.AddressConverter) uint32 {
	panic("implement me")
}

func (scm ShardCoordinatorMock) ShardForCurrentNode() uint32 {
	return 0
}
