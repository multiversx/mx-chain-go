package sharding

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func sameValidatorsMaps(map1, map2 map[uint32][]Validator) bool {
	if len(map1) != len(map2) {
		return false
	}

	for k, v := range map1 {
		if !sameNodes(v, map2[k]) {
			return false
		}
	}

	return true
}

func sameNodes(list1 []Validator, list2 []Validator) bool {
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

func TestIndexHashedNodesCoordinator_LoadStateAfterSave(t *testing.T) {
	args := createArguments()
	nodesCoordinator, err := NewIndexHashedNodesCoordinator(args)
	assert.Nil(t, err)

	expectedConfig := nodesCoordinator.nodesConfig[0]

	key := []byte("config")
	err = nodesCoordinator.saveState(key)
	assert.Nil(t, err)

	delete(nodesCoordinator.nodesConfig, 0)
	err = nodesCoordinator.LoadState(key)
	assert.Nil(t, err)

	actualConfig := nodesCoordinator.nodesConfig[0]

	assert.Equal(t, expectedConfig.shardId, actualConfig.shardId)
	assert.Equal(t, expectedConfig.nbShards, actualConfig.nbShards)
	assert.True(t, sameValidatorsMaps(expectedConfig.eligibleMap, actualConfig.eligibleMap))
	assert.True(t, sameValidatorsMaps(expectedConfig.waitingMap, actualConfig.waitingMap))
}

func TestIndexHashedNodesCooridinator_nodesCoordinatorToRegistry(t *testing.T) {

}

func TestIndexHashedNodesCoordinator_registryToNodesCoordinator(t *testing.T) {

}

func TestIndexHashedNodesCoordinator_epochNodesConfigToEpochValidators(t *testing.T) {

}

func TestIndexHashedNodesCoordinator_epochValidatorsToEpochNodesConfig(t *testing.T) {

}

func TestIndexHashedNodesCoordinator_serializableValidatorsMapToValidatorsMap(t *testing.T) {

}

func TestIndexHashedNodesCoordinator_validatorArrayToSerializableValidatorArray(t *testing.T) {

}

func TestIndexHashedNodesCoordinator_serializableValidatorArrayToValidatorArray(t *testing.T) {

}
