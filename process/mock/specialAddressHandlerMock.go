package mock

import (
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

	addresses []string
	epoch     uint32
	round     uint64
}

func NewSpecialAddressHandlerMock(
	addrConv state.AddressConverter,
	shardCoordinator sharding.Coordinator,
) *SpecialAddressHandlerMock {
	return &SpecialAddressHandlerMock{
		ElrondCommunityAddressCalled: nil,
		LeaderAddressCalled:          nil,
		BurnAddressCalled:            nil,
		ShardIdForAddressCalled:      nil,
		AdrConv:                      addrConv,
		ShardCoordinator:             shardCoordinator,
		addresses:                    nil,
		epoch:                        0,
		round:                        0,
	}
}

func (sh *SpecialAddressHandlerMock) SetElrondCommunityAddress(elrond []byte) {
}

func (sh *SpecialAddressHandlerMock) SetConsensusData(consensusRewardAddresses []string, round uint64, epoch uint32) {
	sh.addresses = consensusRewardAddresses
	sh.epoch = epoch
	sh.round = round
}

func (sh *SpecialAddressHandlerMock) ConsensusRewardAddresses() []string {
	return sh.addresses
}

func (sh *SpecialAddressHandlerMock) BurnAddress() []byte {
	if sh.BurnAddressCalled == nil {
		return []byte("burn")
	}

	return sh.BurnAddressCalled()
}

func (sh *SpecialAddressHandlerMock) ElrondCommunityAddress() []byte {
	if sh.ElrondCommunityAddressCalled == nil {
		return []byte("elrond")
	}

	return sh.ElrondCommunityAddressCalled()
}

func (sh *SpecialAddressHandlerMock) Round() uint64 {
	return sh.round
}

func (sh *SpecialAddressHandlerMock) Epoch() uint32 {
	return sh.epoch
}

func (sh *SpecialAddressHandlerMock) LeaderAddress() []byte {
	if sh.LeaderAddressCalled == nil {
		return []byte("leader")
	}

	return sh.LeaderAddressCalled()
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
