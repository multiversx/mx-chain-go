package sharding

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
)

type indexHashedNodesCoordinatorWithRater struct {
	*indexHashedNodesCoordinator
	RatingReader
	ChanceComputer
}

// NewIndexHashedNodesCoordinator creates a new index hashed group selector
func NewIndexHashedNodesCoordinatorWithRater(
	indexNodesCoordinator *indexHashedNodesCoordinator,
	rater RatingReader,
) (*indexHashedNodesCoordinatorWithRater, error) {
	if check.IfNil(indexNodesCoordinator) {
		return nil, ErrNilNodesCoordinator
	}
	if check.IfNil(rater) {
		return nil, ErrNilRater
	}

	chanceComputer, ok := rater.(ChanceComputer)
	if !ok {
		return nil, ErrNilChanceComputer
	}

	ihncr := &indexHashedNodesCoordinatorWithRater{
		indexHashedNodesCoordinator: indexNodesCoordinator,
		RatingReader:                rater,
		ChanceComputer:              chanceComputer,
	}

	return ihncr, nil
}

// SetNodesPerShards loads the distribution of nodes per shard into the nodes management component
func (ihgs *indexHashedNodesCoordinatorWithRater) SetNodesPerShards(
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	epoch uint32,
) error {
	err := ihgs.indexHashedNodesCoordinator.SetNodesPerShards(eligible, waiting, epoch)
	if err != nil {
		return err
	}

	for _, validators := range eligible {
		for _, validator := range validators {
			pk := validator.PubKey()
			ihgs.UpdateRatingFromTempRating(string(pk))
		}
	}

	err = ihgs.expandAllLists(epoch)
	return err
}

//IsInterfaceNil verifies that the underlying value is nil
func (ihgs *indexHashedNodesCoordinatorWithRater) IsInterfaceNil() bool {
	return ihgs == nil
}

func (ihgs *indexHashedNodesCoordinatorWithRater) expandAllLists(epoch uint32) error {
	ihgs.mutNodesConfig.Lock()
	defer ihgs.mutNodesConfig.Unlock()

	nodesConfig, ok := ihgs.nodesConfig[epoch]
	if !ok {
		return ErrEpochNodesConfigDesNotExist
	}

	shardsExpanded := make(map[uint32][]Validator)

	nrShards := len(nodesConfig.eligibleMap)

	metaChainExpanded, err := ihgs.expandList(nodesConfig, core.MetachainShardId, ihgs.metaConsensusGroupSize)
	if err != nil {
		return err
	}

	for shardId := uint32(0); shardId < uint32(nrShards-1); shardId++ {
		shardsExpanded[shardId], err = ihgs.expandList(nodesConfig, shardId, ihgs.shardConsensusGroupSize)
		if err != nil {
			return err
		}
	}

	nodesConfig.eligibleMap[core.MetachainShardId] = metaChainExpanded
	for shardId := uint32(0); shardId < uint32(nrShards-1); shardId++ {
		nodesConfig.eligibleMap[shardId] = shardsExpanded[shardId]
	}

	return nil
}

func (ihgs *indexHashedNodesCoordinatorWithRater) expandList(nodesConfig *epochNodesConfig, shardId uint32, consensusGroupSize int) ([]Validator, error) {
	validators := nodesConfig.eligibleMap[shardId]
	mut := &nodesConfig.mutNodesMaps

	return ihgs.expandEligibleList(validators, mut, consensusGroupSize)
}

func (ihgs *indexHashedNodesCoordinatorWithRater) expandEligibleList(validators []Validator, mut *sync.RWMutex, consensusGroupSize int) ([]Validator, error) {
	mut.RLock()
	defer mut.RUnlock()

	validatorList := make([]Validator, 0)
	zeroChanceValidators := make([]Validator, 0)
	totalValidatorsSelected := 0

	for _, validator := range validators {
		pk := validator.PubKey()
		rating := ihgs.GetRating(string(pk))
		chances := ihgs.GetChance(rating)
		if chances > 0 {
			totalValidatorsSelected += totalValidatorsSelected
		} else {
			zeroChanceValidators = append(zeroChanceValidators, validator)
		}
		for i := uint32(0); i < chances; i++ {
			validatorList = append(validatorList, validator)
		}
	}

	if totalValidatorsSelected < consensusGroupSize {
		for _, validator := range zeroChanceValidators {
			validatorList = append(validatorList, validator)
		}
	}

	return validatorList, nil
}
