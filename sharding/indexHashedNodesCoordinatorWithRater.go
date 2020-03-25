package sharding

import (
	"encoding/json"

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
	ihncr.epochStartRegistrationHandler.UnregisterHandler(indexNodesCoordinator)
	ihncr.epochStartRegistrationHandler.RegisterHandler(ihncr)

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
	updateList bool,
) error {
	err := ihgs.indexHashedNodesCoordinator.SetNodesPerShards(eligible, waiting, epoch, updateList)
	if err != nil {
		return err
	}

	return ihgs.expandAllLists(epoch)
}

// ComputeLeaving - computes the validators that have a threshold below the minimum rating
func (ihgs *indexHashedNodesCoordinatorWithRater) ComputeLeaving(allValidators []Validator) []Validator {
	leavingValidators := make([]Validator, 0)
	minChances := ihgs.GetChance(0)
	for _, val := range allValidators {
		pk := val.PubKey()
		rating := ihgs.GetRating(string(pk))
		chances := ihgs.GetChance(rating)
		if chances < minChances {
			leavingValidators = append(leavingValidators, val)
		}
	}

	return leavingValidators
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
		return ErrEpochNodesConfigDoesNotExist
	}

	shardsExpanded := make(map[uint32][]Validator)

	err := ihgs.expandListsForEpochConfig(nodesConfig, shardsExpanded)
	if err != nil {
		return err
	}

	return nil
}

func (ihgs *indexHashedNodesCoordinatorWithRater) expandListsForEpochConfig(nodesConfig *epochNodesConfig, shardsExpanded map[uint32][]Validator) error {
	nodesConfig.mutNodesMaps.Lock()
	defer nodesConfig.mutNodesMaps.Unlock()

	nodesConfig.expandedEligibleMap = make(map[uint32][]Validator)

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
	log.Trace("Expanding eligible list", "shardID", shardId)
	return ihgs.expandEligibleList(validators)
}

func (ihgs *indexHashedNodesCoordinatorWithRater) expandEligibleList(validators []Validator) ([]Validator, error) {
	minChance := ihgs.GetChance(0)
	minSize := len(validators) * int(minChance)
	validatorList := make([]Validator, 0, minSize)

	for _, validatorInShard := range validators {
		pk := validatorInShard.PubKey()
		rating := ihgs.GetRating(string(pk))
		chances := ihgs.GetChance(rating)
		if chances < minChance {
			//default chance if all validators need to be selected
			chances = minChance
		}
		log.Trace("Computing chances for validator", "pk", pk, "rating", rating, "chances", chances)

		for i := uint32(0); i < chances; i++ {
			validatorList = append(validatorList, validatorInShard)
		}
	}

	return validatorList, nil
}

// LoadState loads the nodes coordinator state from the used boot storage
func (ihgs *indexHashedNodesCoordinatorWithRater) LoadState(key []byte) error {
	ncInternalkey := append([]byte(keyPrefix), key...)

	log.Debug("getting nodes coordinator config", "key", ncInternalkey)

	data, err := ihgs.bootStorer.Get(ncInternalkey)
	if err != nil {
		return err
	}

	config := &NodesCoordinatorRegistry{}
	err = json.Unmarshal(data, config)
	if err != nil {
		return err
	}

	ihgs.mutSavedStateKey.Lock()
	ihgs.savedStateKey = key
	ihgs.mutSavedStateKey.Unlock()

	err = ihgs.SaveNodesCoordinatorRegistry(config)
	if err != nil {
		return err
	}

	err = ihgs.expandAllLists(config.CurrentEpoch)
	if err != nil {
		return err
	}

	return nil
}
