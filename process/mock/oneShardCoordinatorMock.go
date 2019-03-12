package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type oneShardCoordinatorMock struct {
	noShards uint32
}

func NewOneShardCoordinatorMock() *oneShardCoordinatorMock {
	return &oneShardCoordinatorMock{noShards: 1}
}

func (scm *oneShardCoordinatorMock) NoShards() uint32 {
	return scm.noShards
}

func (scm *oneShardCoordinatorMock) SetNoShards(shards uint32) {
	scm.noShards = shards
}

func (scm *oneShardCoordinatorMock) ComputeShardForAddress(
	address state.AddressContainer,
	addressConverter state.AddressConverter) uint32 {

	return uint32(0)
}

func (scm *oneShardCoordinatorMock) ShardForCurrentNode() uint32 {
	return 0
}

func (scm *oneShardCoordinatorMock) CommunicationIdentifier(destShardID uint32) string {
	return "_0"
}
