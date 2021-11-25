package sharding

import (
	"strconv"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
)

var log = logger.GetOrCreate("sharding")

// SerializableValidatorsToValidators creates the validator map from serializable validator map
func SerializableValidatorsToValidators(nodeRegistryValidators map[string][]*SerializableValidator) (map[uint32][]nodesCoordinator.Validator, error) {
	validators := make(map[uint32][]nodesCoordinator.Validator)
	for shardId, shardValidators := range nodeRegistryValidators {
		newValidators, err := SerializableShardValidatorListToValidatorList(shardValidators)
		if err != nil {
			return nil, err
		}
		shardIdInt, err := strconv.ParseUint(shardId, 10, 32)
		if err != nil {
			return nil, err
		}
		validators[uint32(shardIdInt)] = newValidators
	}

	return validators, nil
}

// SerializableShardValidatorListToValidatorList creates the validator list from serializable validator list
func SerializableShardValidatorListToValidatorList(shardValidators []*SerializableValidator) ([]nodesCoordinator.Validator, error) {
	newValidators := make([]nodesCoordinator.Validator, len(shardValidators))
	for i, shardValidator := range shardValidators {
		v, err := nodesCoordinator.NewValidator(shardValidator.PubKey, shardValidator.Chances, shardValidator.Index)
		if err != nil {
			return nil, err
		}
		newValidators[i] = v
	}
	return newValidators, nil
}
