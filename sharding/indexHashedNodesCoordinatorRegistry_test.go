package sharding

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sameValidatorsMaps(map1, map2 map[uint32][]nodesCoordinator.Validator) bool {
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

func sameValidatorsDifferentMapTypes(map1 map[uint32][]nodesCoordinator.Validator, map2 map[string][]*SerializableValidator) bool {
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

func sameValidators(list1 []nodesCoordinator.Validator, list2 []nodesCoordinator.Validator) bool {
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

func validatorsEqualSerializableValidators(validators []nodesCoordinator.Validator, sValidators []*SerializableValidator) bool {
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
	nCoordinator, _ := NewIndexHashedNodesCoordinator(args)

	expectedConfig, _ := nCoordinator.GetNodesConfigPerEpoch(0)

	key := []byte("config")
	err := nCoordinator.saveState(key)
	assert.Nil(t, err)

	nCoordinator.RemoveNodesConfigEpochs(0)
	err = nCoordinator.LoadState(key)
	assert.Nil(t, err)

	actualConfig, _ := nCoordinator.GetNodesConfigPerEpoch(0)

	assert.Equal(t, expectedConfig.ShardID, actualConfig.ShardID)
	assert.Equal(t, expectedConfig.NbShards, actualConfig.NbShards)
	assert.True(t, sameValidatorsMaps(expectedConfig.EligibleMap, actualConfig.EligibleMap))
	assert.True(t, sameValidatorsMaps(expectedConfig.WaitingMap, actualConfig.WaitingMap))
}

func TestIndexHashedNodesCooridinator_nodesCoordinatorToRegistry(t *testing.T) {
	args := createArguments()
	nodesCoordinator, _ := NewIndexHashedNodesCoordinator(args)

	ncr := nodesCoordinator.NodesCoordinatorToRegistry()
	nc := nodesCoordinator.GetNodesConfig()

	assert.Equal(t, nodesCoordinator.GetCurrentEpoch(), ncr.CurrentEpoch)
	assert.Equal(t, len(nodesCoordinator.GetNodesConfig()), len(ncr.EpochsConfig))

	for epoch, config := range nc {
		assert.True(t, sameValidatorsDifferentMapTypes(config.EligibleMap, ncr.EpochsConfig[fmt.Sprint(epoch)].EligibleValidators))
		assert.True(t, sameValidatorsDifferentMapTypes(config.WaitingMap, ncr.EpochsConfig[fmt.Sprint(epoch)].WaitingValidators))
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

	assert.Equal(t, len(nodesCoordinator1.GetNodesConfig()), len(nodesConfig))
	for epoch, config := range nodesCoordinator1.GetNodesConfig() {
		assert.True(t, sameValidatorsMaps(config.EligibleMap, nodesConfig[epoch].EligibleMap))
		assert.True(t, sameValidatorsMaps(config.WaitingMap, nodesConfig[epoch].WaitingMap))
	}
}

func TestIndexHashedNodesCooridinator_nodesCoordinatorToRegistryLimitNumEpochsInRegistry(t *testing.T) {
	args := createArguments()
	args.Epoch = 100
	nCoordinator, _ := NewIndexHashedNodesCoordinator(args)
	for e := uint32(0); e < args.Epoch; e++ {
		eligibleMap := createDummyNodesMap(10, args.NbShards, "eligible")
		waitingMap := createDummyNodesMap(3, args.NbShards, "waiting")

		nCoordinator.SetNodesConfigPerEpoch(e, &nodesCoordinator.EpochNodesConfig{
			NbShards:    args.NbShards,
			ShardID:     args.ShardIDAsObserver,
			EligibleMap: eligibleMap,
			WaitingMap:  waitingMap,
			Selectors:   make(map[uint32]nodesCoordinator.RandomSelector),
			LeavingMap:  make(map[uint32][]nodesCoordinator.Validator),
			NewList:     make([]nodesCoordinator.Validator, 0),
		})
	}

	ncr := nCoordinator.NodesCoordinatorToRegistry()
	nc := nCoordinator.GetNodesConfig()

	require.Equal(t, nCoordinator.GetCurrentEpoch(), ncr.CurrentEpoch)
	require.Equal(t, nodesCoordinatorStoredEpochs, len(ncr.EpochsConfig))

	for epochStr := range ncr.EpochsConfig {
		epoch, err := strconv.Atoi(epochStr)
		require.Nil(t, err)
		require.True(t, sameValidatorsDifferentMapTypes(nc[uint32(epoch)].EligibleMap, ncr.EpochsConfig[epochStr].EligibleValidators))
		require.True(t, sameValidatorsDifferentMapTypes(nc[uint32(epoch)].WaitingMap, ncr.EpochsConfig[epochStr].WaitingValidators))
	}
}

func TestIndexHashedNodesCoordinator_epochNodesConfigToEpochValidators(t *testing.T) {
	args := createArguments()
	nc, _ := NewIndexHashedNodesCoordinator(args)

	for _, nodesConfig := range nc.GetNodesConfig() {
		epochValidators := epochNodesConfigToEpochValidators(nodesConfig)
		assert.True(t, sameValidatorsDifferentMapTypes(nodesConfig.EligibleMap, epochValidators.EligibleValidators))
		assert.True(t, sameValidatorsDifferentMapTypes(nodesConfig.WaitingMap, epochValidators.WaitingValidators))
	}
}

func TestIndexHashedNodesCoordinator_epochValidatorsToEpochNodesConfig(t *testing.T) {
	args := createArguments()
	nc, _ := NewIndexHashedNodesCoordinator(args)

	for _, nodesConfig := range nc.GetNodesConfig() {
		epochValidators := epochNodesConfigToEpochValidators(nodesConfig)
		epochNodesConfig, err := epochValidatorsToEpochNodesConfig(epochValidators)
		assert.Nil(t, err)
		assert.True(t, sameValidatorsDifferentMapTypes(epochNodesConfig.EligibleMap, epochValidators.EligibleValidators))
		assert.True(t, sameValidatorsDifferentMapTypes(epochNodesConfig.WaitingMap, epochValidators.WaitingValidators))
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
