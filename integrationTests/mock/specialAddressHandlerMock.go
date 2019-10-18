package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type SpecialAddressHandlerMock struct {
	ElrondCommunityAddressCalled func() []byte
	LeaderAddressCalled          func() []byte
	BurnAddressCalled            func() []byte
	ShardIdForAddressCalled      func([]byte) (uint32, error)
	AdrConv                      state.AddressConverter
	ShardCoordinator             sharding.Coordinator
	NodesCoordinator             sharding.NodesCoordinator

	shardConsensusData *data.ConsensusRewardData
	metaConsensusData  []*data.ConsensusRewardData
}

func NewSpecialAddressHandlerMock(
	addrConv state.AddressConverter,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
) *SpecialAddressHandlerMock {
	return &SpecialAddressHandlerMock{
		ElrondCommunityAddressCalled: nil,
		LeaderAddressCalled:          nil,
		BurnAddressCalled:            nil,
		ShardIdForAddressCalled:      nil,
		AdrConv:                      addrConv,
		ShardCoordinator:             shardCoordinator,
		NodesCoordinator:             nodesCoordinator,
		shardConsensusData: &data.ConsensusRewardData{
			Round:     0,
			Epoch:     0,
			Addresses: nil,
		},
		metaConsensusData: make([]*data.ConsensusRewardData, 0),
	}
}

func (sh *SpecialAddressHandlerMock) SetElrondCommunityAddress(elrond []byte) {
}

func (sh *SpecialAddressHandlerMock) SetShardConsensusData(randomness []byte, round uint64, epoch uint32, shardId uint32) error {
	addresses, err := sh.NodesCoordinator.GetValidatorsRewardsAddresses(randomness, round, shardId)
	if err != nil {
		return err
	}

	sh.shardConsensusData = &data.ConsensusRewardData{
		Round:     round,
		Epoch:     epoch,
		Addresses: addresses,
	}

	return nil

}

func (sh *SpecialAddressHandlerMock) ConsensusShardRewardData() *data.ConsensusRewardData {
	return sh.shardConsensusData
}

func (sh *SpecialAddressHandlerMock) SetMetaConsensusData(randomness []byte, round uint64, epoch uint32) error {
	if sh.metaConsensusData == nil {
		sh.metaConsensusData = make([]*data.ConsensusRewardData, 0)
	}

	addresses, err := sh.NodesCoordinator.GetValidatorsRewardsAddresses(randomness, round, sharding.MetachainShardId)
	if err != nil {
		return err
	}

	sh.metaConsensusData = append(sh.metaConsensusData, &data.ConsensusRewardData{
		Round:     round,
		Epoch:     epoch,
		Addresses: addresses,
	})

	return nil

}

func (sh *SpecialAddressHandlerMock) ClearMetaConsensusData() {
	sh.metaConsensusData = make([]*data.ConsensusRewardData, 0)
}

func (sh *SpecialAddressHandlerMock) ConsensusMetaRewardData() []*data.ConsensusRewardData {
	return sh.metaConsensusData
}

func (sh *SpecialAddressHandlerMock) BurnAddress() []byte {
	if sh.BurnAddressCalled == nil {
		return []byte("burn0000000000000000000000000000")
	}

	return sh.BurnAddressCalled()
}

func (sh *SpecialAddressHandlerMock) ElrondCommunityAddress() []byte {
	if sh.ElrondCommunityAddressCalled == nil {
		return []byte("elrond00000000000000000000000000")
	}

	return sh.ElrondCommunityAddressCalled()
}

func (sh *SpecialAddressHandlerMock) LeaderAddress() []byte {
	if sh.LeaderAddressCalled == nil {
		if sh.shardConsensusData != nil && len(sh.shardConsensusData.Addresses) > 0 {
			return []byte(sh.shardConsensusData.Addresses[0])
		}
		return []byte("leader0000000000000000000000000000")
	}

	return sh.LeaderAddressCalled()
}

func (sh *SpecialAddressHandlerMock) Round() uint64 {
	if sh.shardConsensusData == nil {
		return 0
	}
	return sh.shardConsensusData.Round
}

func (sh *SpecialAddressHandlerMock) Epoch() uint32 {
	if sh.shardConsensusData == nil {
		return 0
	}
	return sh.shardConsensusData.Epoch
}

func (sh *SpecialAddressHandlerMock) ShardIdForAddress(addr []byte) (uint32, error) {
	convAdr, err := sh.AdrConv.CreateAddressFromPublicKeyBytes(addr)
	if err != nil {
		return 0, err
	}

	return sh.ShardCoordinator.ComputeId(convAdr), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sh *SpecialAddressHandlerMock) IsInterfaceNil() bool {
	if sh == nil {
		return true
	}
	return false
}

// IsCurrentNodeInConsensus calculate if current node was in consensus group
func (sh *SpecialAddressHandlerMock) IsCurrentNodeInConsensus() bool {
	return false
}

func (sh *SpecialAddressHandlerMock) NumConsensusParticipantMeta() uint64 {
	return 0
}
