package sharding

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
)

type indexHashedNodesCoordinatorWithRater struct {
	*indexHashedNodesCoordinator
	RatingReader
	ChanceComputer
}

// NewIndexHashedNodesCoordinatorWithRater creates a new index hashed group selector
func NewIndexHashedNodesCoordinatorWithRater(
	indexNodesCoordinator *indexHashedNodesCoordinator,
	rater RatingReaderWithChanceComputer,
) (*indexHashedNodesCoordinatorWithRater, error) {
	if check.IfNil(indexNodesCoordinator) {
		return nil, ErrNilNodesCoordinator
	}
	if check.IfNil(rater) {
		return nil, ErrNilRater
	}

	ihncr := &indexHashedNodesCoordinatorWithRater{
		indexHashedNodesCoordinator: indexNodesCoordinator,
		RatingReader:                rater,
		ChanceComputer:              rater,
	}

	ihncr.nodesPerShardSetter = ihncr

	ihncr.epochStartSubscriber.UnregisterHandler(indexNodesCoordinator)
	ihncr.epochStartSubscriber.RegisterHandler(ihncr)

	err := ihncr.expandAllLists(indexNodesCoordinator.currentEpoch)
	if err != nil {
		return nil, err
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
		log.Warn(fmt.Sprintf("Nodes config for epoch %v does not exist", epoch))
		return fmt.Errorf("For epoch %v there was err: %w", epoch, ErrEpochNodesConfigDesNotExist)
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

	nodesConfig.expandedEligibleMap[core.MetachainShardId] = metaChainExpanded
	for shardId := uint32(0); shardId < uint32(nrShards-1); shardId++ {
		nodesConfig.expandedEligibleMap[shardId] = shardsExpanded[shardId]
	}

	return nil
}

func (ihgs *indexHashedNodesCoordinatorWithRater) expandList(nodesConfig *epochNodesConfig, shardId uint32, consensusGroupSize int) ([]Validator, error) {
	validators := nodesConfig.eligibleMap[shardId]
	mut := &nodesConfig.mutNodesMaps
	log.Trace("Expanding eligible list", "shardId", shardId)
	return ihgs.expandEligibleList(validators, mut, consensusGroupSize)
}

func (ihgs *indexHashedNodesCoordinatorWithRater) expandEligibleList(validators []Validator, mut *sync.RWMutex, consensusGroupSize int) ([]Validator, error) {
	mut.RLock()
	defer mut.RUnlock()

	validatorList := make([]Validator, 0)
	zeroChanceValidators := make([]Validator, 0)
	totalValidatorsSelected := 0

	for _, validatorInShard := range validators {
		pk := validatorInShard.PubKey()
		//ihgs.UpdateRatingFromTempRating([]string{string(pk)})
		rating := ihgs.GetRating(string(pk))
		chances := ihgs.GetChance(rating)
		log.Trace("Computing chances for validator", "pk", pk, "rating", rating, "chances", chances)
		if chances > 0 {
			totalValidatorsSelected += 1
		} else {
			zeroChanceValidators = append(zeroChanceValidators, validatorInShard)
		}
		for i := uint32(0); i < chances; i++ {
			validatorList = append(validatorList, validatorInShard)
		}
	}

	if totalValidatorsSelected < consensusGroupSize {
		log.Trace("Adding Zerochance validators")
		for _, validator := range zeroChanceValidators {
			log.Trace("ZeroChancevalidator", "pk", validator.PubKey())
			validatorList = append(validatorList, validator)
		}
	}

	return validatorList, nil
}
