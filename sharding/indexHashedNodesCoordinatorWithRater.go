package sharding

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
)

type indexHashedNodesCoordinatorWithRater struct {
	indexHashedNodesCoordinator
	RatingReader
}

// NewIndexHashedNodesCoordinator creates a new index hashed group selector
func NewIndexHashedNodesCoordinatorWithRater(
	arguments ArgNodesCoordinator,
	rater RatingReader,
) (*indexHashedNodesCoordinatorWithRater,
	error) {
	indexNodesCoordinator, err := NewIndexHashedNodesCoordinator(arguments)

	if err != nil {
		return nil, err
	}

	if check.IfNil(rater) {
		return nil, ErrNilRater
	}

	return &indexHashedNodesCoordinatorWithRater{
		indexHashedNodesCoordinator: *indexNodesCoordinator,
		RatingReader:                rater,
	}, nil
}

func (ihgs *indexHashedNodesCoordinatorWithRater) ComputeValidatorsGroup(
	randomness []byte,
	round uint64,
	shardId uint32,
) (validatorsGroup []Validator, err error) {
	if randomness == nil {
		return nil, ErrNilRandomness
	}

	if shardId >= ihgs.nbShards && shardId != MetachainShardId {
		return nil, ErrInvalidShardId
	}

	if ihgs == nil {
		return nil, ErrNilRandomness
	}

	tempList := make([]Validator, 0)
	consensusSize := ihgs.consensusGroupSize(shardId)
	randomness = []byte(fmt.Sprintf("%d-%s", round, core.ToB64(randomness)))

	// TODO: pre-compute eligible list and update only on rating change.
	expandedList := ihgs.expandEligibleList(shardId)
	lenExpandedList := len(expandedList)

	for startIdx := 0; startIdx < consensusSize; startIdx++ {
		proposedIndex := ihgs.computeListIndex(startIdx, lenExpandedList, string(randomness))
		checkedIndex := ihgs.checkIndex(proposedIndex, expandedList, tempList)
		tempList = append(tempList, expandedList[checkedIndex])
	}

	return tempList, nil
}

func (ihgs *indexHashedNodesCoordinatorWithRater) expandEligibleList(shardId uint32) []Validator {
	validatorList := make([]Validator, 0)

	for _, validator := range ihgs.nodesMap[shardId] {
		pk := validator.Address()
		rating := ihgs.GetRating(string(pk))
		for i := uint32(0); i < rating; i++ {
			validatorList = append(validatorList, validator)
		}
	}

	return validatorList
}

// SetNodesPerShards loads the distribution of nodes per shard into the nodes management component
func (ihgs *indexHashedNodesCoordinatorWithRater) SetNodesPerShards(nodes map[uint32][]Validator) error {
	if nodes == nil {
		return ErrNilInputNodesMap
	}

	nodesList, ok := nodes[MetachainShardId]
	if ok && len(nodesList) < ihgs.metaConsensusGroupSize {
		return ErrSmallMetachainEligibleListSize
	}

	for shardId := uint32(0); shardId < ihgs.nbShards; shardId++ {
		nbNodesShard := len(nodes[shardId])
		if nbNodesShard < ihgs.shardConsensusGroupSize {
			return ErrSmallShardEligibleListSize
		}
	}

	ihgs.nodesMap = nodes

	return nil
}

// GetValidatorsPublicKeys calculates the validators consensus group for a specific shard, randomness and round number,
// returning their public keys
func (ihgs *indexHashedNodesCoordinatorWithRater) GetValidatorsPublicKeys(
	randomness []byte,
	round uint64,
	shardId uint32,
) ([]string, error) {
	consensusNodes, err := ihgs.ComputeValidatorsGroup(randomness, round, shardId)
	if err != nil {
		return nil, err
	}

	pubKeys := make([]string, 0)

	for _, v := range consensusNodes {
		pubKeys = append(pubKeys, string(v.PubKey()))
	}

	return pubKeys, nil
}

// GetValidatorsRewardsAddresses calculates the validator consensus group for a specific shard, randomness and round
// number, returning their staking/rewards addresses
func (ihgs *indexHashedNodesCoordinatorWithRater) GetValidatorsRewardsAddresses(
	randomness []byte,
	round uint64,
	shardId uint32,
) ([]string, error) {
	consensusNodes, err := ihgs.ComputeValidatorsGroup(randomness, round, shardId)
	if err != nil {
		return nil, err
	}

	addresses := make([]string, len(consensusNodes))
	for i, v := range consensusNodes {
		addresses[i] = string(v.Address())
	}

	return addresses, nil
}

func (ihgs *indexHashedNodesCoordinatorWithRater) IsInterfaceNil() bool {
	return ihgs == nil
}
