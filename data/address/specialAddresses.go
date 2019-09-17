package address

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type specialAddresses struct {
	elrond                   []byte
	consensusRewardAddresses []string
	burnAddress              []byte

	epoch            uint32
	round            uint64
	adrConv          state.AddressConverter
	shardCoordinator sharding.Coordinator
}

// NewSpecialAddressHolder creates a special address holder
func NewSpecialAddressHolder(
	elrond []byte,
	burnAddress []byte,
	adrConv state.AddressConverter,
	shardCoordinator sharding.Coordinator,
) (*specialAddresses, error) {
	if elrond == nil {
		return nil, data.ErrNilElrondAddress
	}
	if burnAddress == nil {
		return nil, data.ErrNilBurnAddress
	}
	if adrConv == nil || adrConv.IsInterfaceNil() {
		return nil, data.ErrNilAddressConverter
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, data.ErrNilShardCoordinator
	}

	sp := &specialAddresses{
		elrond:           elrond,
		burnAddress:      burnAddress,
		adrConv:          adrConv,
		shardCoordinator: shardCoordinator,
	}

	return sp, nil
}

// SetElrondCommunityAddress sets elrond address
func (sp *specialAddresses) SetElrondCommunityAddress(elrond []byte) {
	sp.elrond = elrond
}

// ElrondCommunityAddress provides elrond address
func (sp *specialAddresses) ElrondCommunityAddress() []byte {
	return sp.elrond
}

// BurnAddress provides burn address
func (sp *specialAddresses) BurnAddress() []byte {
	return sp.burnAddress
}

// SetConsensusData sets the consensus rewards addresses for the round
func (sp *specialAddresses) SetConsensusData(consensusRewardAddresses []string, round uint64, epoch uint32) {
	sp.consensusRewardAddresses = consensusRewardAddresses
	sp.round = round
	sp.epoch = epoch
}

// LeaderAddress provides leader address
func (sp *specialAddresses) LeaderAddress() []byte {
	if len(sp.consensusRewardAddresses) == 0 {
		return nil
	}

	return []byte(sp.consensusRewardAddresses[0])
}

// ConsensusRewardAddresses provides the consensus reward addresses
func (sp *specialAddresses) ConsensusRewardAddresses() []string {
	return sp.consensusRewardAddresses
}

// Round returns the round for the current block
func (sp *specialAddresses) Round() uint64 {
	return sp.round
}

// Epoch returns the epoch for the current block
func (sp *specialAddresses) Epoch() uint32 {
	return sp.epoch
}

// ShardIdForAddress calculates shard id for address
func (sp *specialAddresses) ShardIdForAddress(pubKey []byte) (uint32, error) {
	convAdr, err := sp.adrConv.CreateAddressFromPublicKeyBytes(pubKey)
	if err != nil {
		return 0, err
	}

	return sp.shardCoordinator.ComputeId(convAdr), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sp *specialAddresses) IsInterfaceNil() bool {
	if sp == nil {
		return true
	}
	return false
}
