package sharding_test

import (
	"bytes"
	"fmt"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"testing"

	"github.com/stretchr/testify/assert"
)

func sameValidatorsMaps(map1, map2 map[uint32][]sharding.Validator) bool {
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

func sameValidatorsDifferentMapTypes(map1 map[uint32][]sharding.Validator, map2 map[string][]*sharding.SerializableValidator) bool {
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

func sameValidators(list1 []sharding.Validator, list2 []sharding.Validator) bool {
	if len(list1) != len(list2) {
		return false
	}

	for i, validator := range list1 {
		if !bytes.Equal(validator.Address(), list2[i].Address()) {
			return false
		}
		if !bytes.Equal(validator.PubKey(), list2[i].PubKey()) {
			return false
		}
	}

	return true
}

func validatorsEqualSerializableValidators(validators []sharding.Validator, sValidators []*sharding.SerializableValidator) bool {
	if len(validators) != len(sValidators) {
		return false
	}

	for i, validator := range validators {
		if !bytes.Equal(validator.Address(), sValidators[i].Address) {
			return false
		}
		if !bytes.Equal(validator.PubKey(), sValidators[i].PubKey) {
			return false
		}
	}

	return true
}

func TestIndexHashedNodesCoordinator_LoadStateAfterSave(t *testing.T) {
	args := createArguments()
	nodesCoordinator, _ := sharding.NewIndexHashedNodesCoordinator(args)

	expectedConfig := nodesCoordinator.NodesConfig()[0]

	key := []byte("config")
	err := nodesCoordinator.SaveState(key)
	assert.Nil(t, err)

	delete(nodesCoordinator.NodesConfig(), 0)
	err = nodesCoordinator.LoadState(key)
	assert.Nil(t, err)

	actualConfig := nodesCoordinator.NodesConfig()[0]

	assert.Equal(t, expectedConfig.ShardId(), actualConfig.ShardId())
	assert.Equal(t, expectedConfig.NbShards(), actualConfig.NbShards())
	assert.True(t, sameValidatorsMaps(expectedConfig.EligibleMap(), actualConfig.EligibleMap()))
	assert.True(t, sameValidatorsMaps(expectedConfig.WaitingMap(), actualConfig.WaitingMap()))
}

func TestIndexHashedNodesCooridinator_nodesCoordinatorToRegistry(t *testing.T) {
	args := createArguments()
	nodesCoordinator, _ := sharding.NewIndexHashedNodesCoordinator(args)

	ncr := nodesCoordinator.NodesCoordinatorToRegistry()
	nc := nodesCoordinator.NodesConfig()

	assert.Equal(t, nodesCoordinator.CurrentEpoch(), ncr.CurrentEpoch)
	assert.Equal(t, len(nodesCoordinator.NodesConfig()), len(ncr.EpochsConfig))

	for epoch, config := range nc {
		assert.True(t, sameValidatorsDifferentMapTypes(config.EligibleMap(), ncr.EpochsConfig[fmt.Sprint(epoch)].EligibleValidators))
		assert.True(t, sameValidatorsDifferentMapTypes(config.WaitingMap(), ncr.EpochsConfig[fmt.Sprint(epoch)].WaitingValidators))
	}
}

func TestIndexHashedNodesCoordinator_registryToNodesCoordinator(t *testing.T) {
	args := createArguments()
	nodesCoordinator1, _ := sharding.NewIndexHashedNodesCoordinator(args)
	ncr := nodesCoordinator1.NodesCoordinatorToRegistry()

	args = createArguments()
	nodesCoordinator2, _ := sharding.NewIndexHashedNodesCoordinator(args)

	nodesConfig, err := nodesCoordinator2.RegistryToNodesCoordinator(ncr)
	assert.Nil(t, err)

	assert.Equal(t, len(nodesCoordinator1.NodesConfig()), len(nodesConfig))
	for epoch, config := range nodesCoordinator1.NodesConfig() {
		assert.True(t, sameValidatorsMaps(config.EligibleMap(), nodesConfig[epoch].EligibleMap()))
		assert.True(t, sameValidatorsMaps(config.WaitingMap(), nodesConfig[epoch].WaitingMap()))
	}
}

func TestIndexHashedNodesCoordinator_epochNodesConfigToEpochValidators(t *testing.T) {
	args := createArguments()
	nc, _ := sharding.NewIndexHashedNodesCoordinator(args)

	for _, nodesConfig := range nc.NodesConfig() {
		epochValidators := sharding.EpochNodesConfigToEpochValidators(nodesConfig)
		assert.True(t, sameValidatorsDifferentMapTypes(nodesConfig.EligibleMap(), epochValidators.EligibleValidators))
		assert.True(t, sameValidatorsDifferentMapTypes(nodesConfig.WaitingMap(), epochValidators.WaitingValidators))
	}
}

func TestIndexHashedNodesCoordinator_epochValidatorsToEpochNodesConfig(t *testing.T) {
	args := createArguments()
	nc, _ := sharding.NewIndexHashedNodesCoordinator(args)

	for _, nodesConfig := range nc.NodesConfig() {
		epochValidators := sharding.EpochNodesConfigToEpochValidators(nodesConfig)
		epochNodesConfig, err := sharding.EpochValidatorsToEpochNodesConfig(epochValidators)
		assert.Nil(t, err)
		assert.True(t, sameValidatorsDifferentMapTypes(epochNodesConfig.EligibleMap(), epochValidators.EligibleValidators))
		assert.True(t, sameValidatorsDifferentMapTypes(epochNodesConfig.WaitingMap(), epochValidators.WaitingValidators))
	}
}

func TestIndexHashedNodesCoordinator_validatorArrayToSerializableValidatorArray(t *testing.T) {
	validatorsMap := createDummyNodesMap(5, 2, "dummy")

	for _, validatorsArray := range validatorsMap {
		sValidators := sharding.ValidatorArrayToSerializableValidatorArray(validatorsArray)
		assert.True(t, validatorsEqualSerializableValidators(validatorsArray, sValidators))
	}
}

func TestIndexHashedNodesCoordinator_serializableValidatorsMapToValidatorsMap(t *testing.T) {
	validatorsMap := createDummyNodesMap(5, 2, "dummy")
	sValidatorsMap := make(map[string][]*sharding.SerializableValidator)

	for k, validatorsArray := range validatorsMap {
		sValidators := sharding.ValidatorArrayToSerializableValidatorArray(validatorsArray)
		sValidatorsMap[fmt.Sprint(k)] = sValidators
	}

	assert.True(t, sameValidatorsDifferentMapTypes(validatorsMap, sValidatorsMap))
}

func TestIndexHashedNodesCoordinator_serializableValidatorArrayToValidatorArray(t *testing.T) {
	validatorsMap := createDummyNodesMap(5, 2, "dummy")

	for _, validatorsArray := range validatorsMap {
		sValidators := sharding.ValidatorArrayToSerializableValidatorArray(validatorsArray)
		valArray, err := sharding.SerializableValidatorArrayToValidatorArray(sValidators)
		assert.Nil(t, err)
		assert.True(t, sameValidators(validatorsArray, valArray))
	}
}
