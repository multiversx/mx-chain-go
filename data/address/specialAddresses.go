package address

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type specialAddresses struct {
	shardConsensusData *data.ConsensusRewardData
	metaConsensusData  []*data.ConsensusRewardData
	elrondAddress      []byte
	burnAddress        []byte

	adrConv          state.AddressConverter
	shardCoordinator sharding.Coordinator
	nodesCoordinator sharding.NodesCoordinator
}

// NewSpecialAddressHolder creates a special address holder
func NewSpecialAddressHolder(
	elrondAddress []byte,
	burnAddress []byte,
	adrConv state.AddressConverter,
	shardCoordinator sharding.Coordinator,
	nodesCoordinator sharding.NodesCoordinator,
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
	if nodesCoordinator == nil || nodesCoordinator.IsInterfaceNil() {
		return nil, data.ErrNilNodesCoordinator
	}

	sp := &specialAddresses{
		elrondAddress:     elrondAddress,
		burnAddress:       burnAddress,
		adrConv:           adrConv,
		shardCoordinator:  shardCoordinator,
		nodesCoordinator:  nodesCoordinator,
		metaConsensusData: make([]*data.ConsensusRewardData, 0),
	}

	return sp, nil
}

// SetShardConsensusData - sets the reward addresses for the current consensus group
func (sp *specialAddresses) SetShardConsensusData(randomness []byte, round uint64, epoch uint32, shardID uint32) error {
	// give transaction coordinator the consensus group validators addresses where to send the rewards.
	consensusAddresses, err := sp.nodesCoordinator.GetValidatorsRewardsAddresses(
		randomness, round, shardID,
	)
	if err != nil {
		return err
	}

	sp.shardConsensusData = &data.ConsensusRewardData{
		Round:     round,
		Epoch:     epoch,
		Addresses: consensusAddresses,
	}

	return nil
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

// ConsensusShardRewardData provides the consensus data required for generating the rewards for shard nodes
func (sp *specialAddresses) ConsensusShardRewardData() *data.ConsensusRewardData {
	return sp.shardConsensusData
}

// SetMetaConsensusData sets the rewards addresses for the metachain nodes
func (sp *specialAddresses) SetMetaConsensusData(randomness []byte, round uint64, epoch uint32) error {
	rewardAddresses, err := sp.nodesCoordinator.GetValidatorsRewardsAddresses(
		randomness,
		round,
		sharding.MetachainShardId,
	)
	if err != nil {
		return err
	}

	sp.metaConsensusData = append(sp.metaConsensusData, &data.ConsensusRewardData{
		Round:     round,
		Epoch:     epoch,
		Addresses: rewardAddresses,
	})

	return nil
}

// ClearMetaConsensusData clears the previously set addresses for rewarding metachain nodes
func (sp *specialAddresses) ClearMetaConsensusData() {
	sp.metaConsensusData = make([]*data.ConsensusRewardData, 0)
}

// ConsensusMetaRewardData provides the consensus data required for generating the rewards for metachain nodes
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
