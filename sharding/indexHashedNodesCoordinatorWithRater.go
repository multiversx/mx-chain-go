package sharding

import (
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

	ihncr.mutNodesConfig.Lock()
	defer ihncr.mutNodesConfig.Unlock()

	nodesConfig, ok := ihncr.nodesConfig[ihncr.currentEpoch]
	if !ok {
		nodesConfig = &epochNodesConfig{}
	}

	nodesConfig.mutNodesMaps.Lock()
	defer nodesConfig.mutNodesMaps.Unlock()

	var err error
	nodesConfig.selectors, err = ihncr.createSelectors(nodesConfig)
	if err != nil {
		return nil, err
	}

	ihncr.epochStartSubscriber.UnregisterHandler(indexNodesCoordinator)
	ihncr.epochStartSubscriber.RegisterHandler(ihncr)
	return ihncr, nil
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

// ValidatorsWeights returns the weights/chances for each given validator
func (ihgs *indexHashedNodesCoordinatorWithRater) ValidatorsWeights(validators []Validator) ([]uint32, error) {
	minChance := ihgs.GetChance(0)
	weights := make([]uint32, len(validators))

	loadingFromDisk := ihgs.loadingFromDisk.Load().(bool)
	for i, validatorInShard := range validators {
		pk := validatorInShard.PubKey()
		if !loadingFromDisk {
			rating := ihgs.GetRating(string(pk))
			weights[i] = ihgs.GetChance(rating)
			if weights[i] < minChance {
				//default weight if all validators need to be selected
				weights[i] = minChance
			}
			log.Debug("Computing chances for validator", "pk", pk, "rating", rating, "chances", weights[i])
		} else {
			weights[i] = validatorInShard.Chances()
		}
		log.Debug("Loading chances for validator", "pk", pk, "chances", weights[i])
	}

	return weights, nil
}
