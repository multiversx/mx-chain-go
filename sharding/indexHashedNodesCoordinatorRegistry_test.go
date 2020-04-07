package sharding

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func sameValidatorsMaps(map1, map2 map[uint32][]Validator) bool {
	if len(map1) != len(map2) {
		return false
	}

	for k, v := range map1 {
		if !sameValidators(v, map2[k]) {
			return false
		}
	}

	return true
}

func sameValidatorsDifferentMapTypes(map1 map[uint32][]Validator, map2 map[string][]*SerializableValidator) bool {
	if len(map1) != len(map2) {
		return false
	}

	for k, v := range map1 {
		if !validatorsEqualSerializableValidators(v, map2[fmt.Sprint(k)]) {
			return false
		}
	}

	return true
}

func sameValidators(list1 []Validator, list2 []Validator) bool {
	if len(list1) != len(list2) {
		return false
	}

	for i, validator := range list1 {
		if !bytes.Equal(validator.PubKey(), list2[i].PubKey()) {
			return false
		}
		if validator.Index() != list2[i].Index() {
			return false
		}
		if validator.Chances() != list2[i].Chances() {
			return false
		}
	}

	return true
}

func validatorsEqualSerializableValidators(validators []Validator, sValidators []*SerializableValidator) bool {
	if len(validators) != len(sValidators) {
		return false
	}

	for i, validator := range validators {
		if !bytes.Equal(validator.PubKey(), sValidators[i].PubKey) {
			return false
		}
	}

	return true
}

func TestIndexHashedNodesCoordinator_LoadStateAfterSave(t *testing.T) {
	args := createArguments()
	nodesCoordinator, _ := NewIndexHashedNodesCoordinator(args)

	expectedConfig := nodesCoordinator.nodesConfig[0]

	key := []byte("config")
	err := nodesCoordinator.saveState(key)
	assert.Nil(t, err)

	delete(nodesCoordinator.nodesConfig, 0)
	err = nodesCoordinator.LoadState(key)
	assert.Nil(t, err)

	actualConfig := nodesCoordinator.nodesConfig[0]

	assert.Equal(t, expectedConfig.shardID, actualConfig.shardID)
	assert.Equal(t, expectedConfig.nbShards, actualConfig.nbShards)
	assert.True(t, sameValidatorsMaps(expectedConfig.eligibleMap, actualConfig.eligibleMap))
	assert.True(t, sameValidatorsMaps(expectedConfig.waitingMap, actualConfig.waitingMap))
}

func TestIndexHashedNodesCooridinator_nodesCoordinatorToRegistry(t *testing.T) {
	args := createArguments()
	nodesCoordinator, _ := NewIndexHashedNodesCoordinator(args)

	ncr := nodesCoordinator.NodesCoordinatorToRegistry()
	nc := nodesCoordinator.nodesConfig

	assert.Equal(t, nodesCoordinator.currentEpoch, ncr.CurrentEpoch)
	assert.Equal(t, len(nodesCoordinator.nodesConfig), len(ncr.EpochsConfig))

	for epoch, config := range nc {
		assert.True(t, sameValidatorsDifferentMapTypes(config.eligibleMap, ncr.EpochsConfig[fmt.Sprint(epoch)].EligibleValidators))
		assert.True(t, sameValidatorsDifferentMapTypes(config.waitingMap, ncr.EpochsConfig[fmt.Sprint(epoch)].WaitingValidators))
	}
}

func TestIndexHashedNodesCoordinator_registryToNodesCoordinator(t *testing.T) {
	args := createArguments()
	nodesCoordinator1, _ := NewIndexHashedNodesCoordinator(args)
	ncr := nodesCoordinator1.NodesCoordinatorToRegistry()

	args = createArguments()
	nodesCoordinator2, _ := NewIndexHashedNodesCoordinator(args)

	nodesConfig, err := nodesCoordinator2.registryToNodesCoordinator(ncr)
	assert.Nil(t, err)

	assert.Equal(t, len(nodesCoordinator1.nodesConfig), len(nodesConfig))
	for epoch, config := range nodesCoordinator1.nodesConfig {
		assert.True(t, sameValidatorsMaps(config.eligibleMap, nodesConfig[epoch].eligibleMap))
		assert.True(t, sameValidatorsMaps(config.waitingMap, nodesConfig[epoch].waitingMap))
	}
}

func TestIndexHashedNodesCoordinator_epochNodesConfigToEpochValidators(t *testing.T) {
	args := createArguments()
	nc, _ := NewIndexHashedNodesCoordinator(args)

	for _, nodesConfig := range nc.nodesConfig {
		epochValidators := epochNodesConfigToEpochValidators(nodesConfig)
		assert.True(t, sameValidatorsDifferentMapTypes(nodesConfig.eligibleMap, epochValidators.EligibleValidators))
		assert.True(t, sameValidatorsDifferentMapTypes(nodesConfig.waitingMap, epochValidators.WaitingValidators))
	}
}

func TestIndexHashedNodesCoordinator_epochValidatorsToEpochNodesConfig(t *testing.T) {
	args := createArguments()
	nc, _ := NewIndexHashedNodesCoordinator(args)

	for _, nodesConfig := range nc.nodesConfig {
		epochValidators := epochNodesConfigToEpochValidators(nodesConfig)
		epochNodesConfig, err := epochValidatorsToEpochNodesConfig(epochValidators)
		assert.Nil(t, err)
		assert.True(t, sameValidatorsDifferentMapTypes(epochNodesConfig.eligibleMap, epochValidators.EligibleValidators))
		assert.True(t, sameValidatorsDifferentMapTypes(epochNodesConfig.waitingMap, epochValidators.WaitingValidators))
	}
}

func TestIndexHashedNodesCoordinator_validatorArrayToSerializableValidatorArray(t *testing.T) {
	validatorsMap := createDummyNodesMap(5, 2, "dummy")

	for _, validatorsArray := range validatorsMap {
		sValidators := ValidatorArrayToSerializableValidatorArray(validatorsArray)
		assert.True(t, validatorsEqualSerializableValidators(validatorsArray, sValidators))
	}
}

func TestIndexHashedNodesCoordinator_serializableValidatorsMapToValidatorsMap(t *testing.T) {
	validatorsMap := createDummyNodesMap(5, 2, "dummy")
	sValidatorsMap := make(map[string][]*SerializableValidator)

	for k, validatorsArray := range validatorsMap {
		sValidators := ValidatorArrayToSerializableValidatorArray(validatorsArray)
		sValidatorsMap[fmt.Sprint(k)] = sValidators
	}

	assert.True(t, sameValidatorsDifferentMapTypes(validatorsMap, sValidatorsMap))
}

func TestIndexHashedNodesCoordinator_serializableValidatorArrayToValidatorArray(t *testing.T) {
	validatorsMap := createDummyNodesMap(5, 2, "dummy")

	for _, validatorsArray := range validatorsMap {
		sValidators := ValidatorArrayToSerializableValidatorArray(validatorsArray)
		valArray, err := serializableValidatorArrayToValidatorArray(sValidators)
		assert.Nil(t, err)
		assert.True(t, sameValidators(validatorsArray, valArray))
	}
}
