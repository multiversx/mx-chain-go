package nodesCoordinator

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

type baseNodesCoordinatorWithRater struct {
	chanceComputer ChanceComputer
}

func (bnc *baseNodesCoordinatorWithRater) GetChance(rating uint32) uint32 {
	return bnc.chanceComputer.GetChance(rating)
}

// ValidatorsWeights returns the weights/chances for each given validator
func (bnc *baseNodesCoordinatorWithRater) ValidatorsWeights(validators []Validator) ([]uint32, error) {
	minChance := bnc.GetChance(0)
	weights := make([]uint32, len(validators))

	for i, validatorInShard := range validators {
		weights[i] = validatorInShard.Chances()
		if weights[i] < minChance {
			//default weight if all validators need to be selected
			weights[i] = minChance
		}
	}

	return weights, nil
}

// ComputeAdditionalLeaving - computes the extra leaving validators that have a threshold below the minimum rating
func (bnc *baseNodesCoordinatorWithRater) ComputeAdditionalLeaving(allValidators []*state.ShardValidatorInfo) (map[uint32][]Validator, error) {
	extraLeavingNodesMap := make(map[uint32][]Validator)
	minChances := bnc.GetChance(0)
	for _, vInfo := range allValidators {
		if vInfo.List == string(common.InactiveList) || vInfo.List == string(common.JailedList) {
			continue
		}
		chances := bnc.GetChance(vInfo.TempRating)
		if chances < minChances {
			val, err := NewValidator(vInfo.PublicKey, chances, vInfo.Index)
			if err != nil {
				return nil, err
			}
			log.Debug("computed leaving based on rating for validator", "pk", vInfo.GetPublicKey())
			extraLeavingNodesMap[vInfo.ShardId] = append(extraLeavingNodesMap[vInfo.ShardId], val)
		}
	}

	return extraLeavingNodesMap, nil
}
