package address

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type specialAddresses struct {
	elrondAddress      []byte
	shardConsensusData *data.ConsensusRewardData
	metaConsensusData  []*data.ConsensusRewardData
	burnAddress        []byte

	adrConv          state.AddressConverter
	shardCoordinator sharding.Coordinator
}

// NewSpecialAddressHolder creates a special address holder
func NewSpecialAddressHolder(
	elrondAddress []byte,
	burnAddress []byte,
	adrConv state.AddressConverter,
	shardCoordinator sharding.Coordinator,
) (*specialAddresses, error) {
	if elrondAddress == nil {
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
		elrondAddress:     elrondAddress,
		burnAddress:       burnAddress,
		adrConv:           adrConv,
		shardCoordinator:  shardCoordinator,
		metaConsensusData: make([]*data.ConsensusRewardData, 0),
	}

	return sp, nil
}

// SetElrondCommunityAddress sets elrond address
func (sp *specialAddresses) SetElrondCommunityAddress(elrond []byte) {
	sp.elrondAddress = elrond
}

// ElrondCommunityAddress provides elrond address
func (sp *specialAddresses) ElrondCommunityAddress() []byte {
	return sp.elrondAddress
}

// BurnAddress provides burn address
func (sp *specialAddresses) BurnAddress() []byte {
	return sp.burnAddress
}

// SetConsensusData sets the consensus rewards addresses for the round
func (sp *specialAddresses) SetConsensusData(rewardAddresses []string, round uint64, epoch uint32) {
	sp.shardConsensusData = &data.ConsensusRewardData{
		Round:     round,
		Epoch:     epoch,
		Addresses: rewardAddresses,
	}
}

// ConsensusShardRewardAddresses provides the consensus reward addresses
func (sp *specialAddresses) ConsensusShardRewardData() *data.ConsensusRewardData {
	return sp.shardConsensusData
}

// SetMetaConsensusData sets the rewards addresses for the metachain nodes
func (sp *specialAddresses) SetMetaConsensusData(rewardAddresses []string, round uint64, epoch uint32) {
	sp.metaConsensusData = append(sp.metaConsensusData, &data.ConsensusRewardData{
		Round:     round,
		Epoch:     epoch,
		Addresses: rewardAddresses,
	})
}

// ClearMetaConsensusData clears the previously set addresses for rewarding metachain nodes
func (sp *specialAddresses) ClearMetaConsensusData() {
	sp.metaConsensusData = make([]*data.ConsensusRewardData, 0)
}

func (sp *specialAddresses) ConsensusMetaRewardData() []*data.ConsensusRewardData {
	return sp.metaConsensusData
}

// LeaderAddress provides leader address
func (sp *specialAddresses) LeaderAddress() []byte {
	if sp.shardConsensusData == nil || len(sp.shardConsensusData.Addresses) == 0 {
		return nil
	}

	return []byte(sp.shardConsensusData.Addresses[0])
}

// Round returns the round for the current block
func (sp *specialAddresses) Round() uint64 {
	if sp.shardConsensusData == nil {
		return 0
	}

	return sp.shardConsensusData.Round
}

// Epoch returns the epoch for the current block
func (sp *specialAddresses) Epoch() uint32 {
	if sp.shardConsensusData == nil {
		return 0
	}

	return sp.shardConsensusData.Epoch
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
