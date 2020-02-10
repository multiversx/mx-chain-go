package sharding

import (
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

	return ihgs.expandAllLists(epoch)
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

	err2 := ihgs.expandListsForEpochConfig(nodesConfig, shardsExpanded)
	if err2 != nil {
		return err2
	}

	return nil
}

func (ihgs *indexHashedNodesCoordinatorWithRater) expandListsForEpochConfig(nodesConfig *epochNodesConfig, shardsExpanded map[uint32][]Validator) error {
	nodesConfig.mutNodesMaps.Lock()
	defer nodesConfig.mutNodesMaps.Unlock()

	nrShards := len(nodesConfig.eligibleMap)
	metaChainExpanded, err := ihgs.expandList(nodesConfig, core.MetachainShardId)
	if err != nil {
		return err
	}

	for shardId := uint32(0); shardId < uint32(nrShards-1); shardId++ {
		shardsExpanded[shardId], err = ihgs.expandList(nodesConfig, shardId)
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

func (ihgs *indexHashedNodesCoordinatorWithRater) expandList(nodesConfig *epochNodesConfig, shardId uint32) ([]Validator, error) {
	validators := nodesConfig.eligibleMap[shardId]
	log.Trace("Expanding eligible list", "shardId", shardId)
	return ihgs.expandEligibleList(validators)
}

func (ihgs *indexHashedNodesCoordinatorWithRater) expandEligibleList(validators []Validator) ([]Validator, error) {
	validatorList := make([]Validator, 0)

	for _, validatorInShard := range validators {
		pk := validatorInShard.PubKey()
		rating := ihgs.GetRating(string(pk))
		chances := ihgs.GetChance(rating)
		if chances == 0 {
			chances = ihgs.GetChance(0)
		}
		log.Trace("Computing chances for validator", "pk", pk, "rating", rating, "chances", chances)

		for i := uint32(0); i < chances; i++ {
			validatorList = append(validatorList, validatorInShard)
		}
	}

	return validatorList, nil
}
