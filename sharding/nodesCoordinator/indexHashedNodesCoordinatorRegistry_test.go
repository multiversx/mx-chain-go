package nodesCoordinator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever/dataPool"
	"github.com/multiversx/mx-chain-go/sharding/mock"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	err = nodesCoordinator.LoadState(key, 0)
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

func TestIndexHashedNodesCooridinator_nodesCoordinatorToRegistryLimitNumEpochsInRegistry(t *testing.T) {
	args := createArguments()
	args.Epoch = 100
	nodesCoordinator, _ := NewIndexHashedNodesCoordinator(args)
	for e := uint32(0); e < args.Epoch; e++ {
		eligibleMap := createDummyNodesMap(10, args.NbShards, "eligible")
		waitingMap := createDummyNodesMap(3, args.NbShards, "waiting")

		nodesCoordinator.nodesConfig[e] = &epochNodesConfig{
			nbShards:    args.NbShards,
			shardID:     args.ShardIDAsObserver,
			eligibleMap: eligibleMap,
			waitingMap:  waitingMap,
			selectors:   make(map[uint32]RandomSelector),
			leavingMap:  make(map[uint32][]Validator),
			newList:     make([]Validator, 0),
		}
	}

	ncr := nodesCoordinator.NodesCoordinatorToRegistry()
	nc := nodesCoordinator.nodesConfig

	require.Equal(t, nodesCoordinator.currentEpoch, ncr.CurrentEpoch)
	require.Equal(t, numStoredEpochs, len(ncr.EpochsConfig))

	for epochStr := range ncr.EpochsConfig {
		epoch, err := strconv.Atoi(epochStr)
		require.Nil(t, err)
		require.True(t, sameValidatorsDifferentMapTypes(nc[uint32(epoch)].eligibleMap, ncr.EpochsConfig[epochStr].EligibleValidators))
		require.True(t, sameValidatorsDifferentMapTypes(nc[uint32(epoch)].waitingMap, ncr.EpochsConfig[epochStr].WaitingValidators))
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

func TestIndexHashedNodesCoordinator_GetNodesCoordinatorRegistry(t *testing.T) {
	t.Parallel()

	t.Run("nil storer, should fail", func(t *testing.T) {
		t.Parallel()

		nodesConfig, err := GetNodesCoordinatorRegistry([]byte("key"), nil, 1, numStoredEpochs)
		require.Nil(t, nodesConfig)
		require.Equal(t, ErrNilBootStorer, err)
	})

	t.Run("getting from old key, should work", func(t *testing.T) {
		t.Parallel()

		nodesConfigRegistry := &NodesCoordinatorRegistry{
			EpochsConfig: map[string]*EpochValidators{
				"10": {
					EligibleValidators: map[string][]*SerializableValidator{
						"0": {
							{
								PubKey: []byte("pubKey1"),
							},
						},
						"1": {
							{
								PubKey: []byte("pubKey1"),
							},
						},
						"2": {
							{
								PubKey: []byte("pubKey1"),
							},
						},
					},
				},
			},
			CurrentEpoch: 10,
		}

		epoch10Key := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), []byte(fmt.Sprint(10))...)

		storer := &mock.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				switch {
				case bytes.Equal(epoch10Key, key):
					return nil, errors.New("first get error")
				default:
					nodesConfigRegistryBytes, _ := json.Marshal(nodesConfigRegistry)
					return nodesConfigRegistryBytes, nil
				}
			},
		}

		nodesConfig, err := GetNodesCoordinatorRegistry([]byte("key"), storer, 10, numStoredEpochs)
		require.Nil(t, err)
		require.Equal(t, nodesConfigRegistry, nodesConfig)
	})

	t.Run("getting from old key, should also save configuration with new keys", func(t *testing.T) {
		t.Parallel()

		nodesConfigRegistry := &NodesCoordinatorRegistry{
			EpochsConfig: map[string]*EpochValidators{
				"10": {
					EligibleValidators: map[string][]*SerializableValidator{
						"0": {{PubKey: []byte("pubKey1")}},
						"1": {{PubKey: []byte("pubKey1")}},
						"2": {{PubKey: []byte("pubKey1")}},
					},
				},
				"9": {
					EligibleValidators: map[string][]*SerializableValidator{
						"0": {{PubKey: []byte("pubKey2")}},
						"1": {{PubKey: []byte("pubKey2")}},
						"2": {{PubKey: []byte("pubKey2")}},
					},
				},
				"8": {
					EligibleValidators: map[string][]*SerializableValidator{
						"0": {{PubKey: []byte("pubKey3")}},
						"1": {{PubKey: []byte("pubKey3")}},
						"2": {{PubKey: []byte("pubKey3")}},
					},
				},
			},
			CurrentEpoch: 10,
		}

		epoch10Key := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), []byte(fmt.Sprint(10))...)
		epoch9Key := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), []byte(fmt.Sprint(9))...)
		epoch8Key := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), []byte(fmt.Sprint(8))...)

		numPutCalls := uint32(0)
		storer := &mock.StorerStub{
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				switch {
				case bytes.Equal(epoch10Key, key):
					return nil, errors.New("first get error")
				default:
					nodesConfigRegistryBytes, _ := json.Marshal(nodesConfigRegistry)
					return nodesConfigRegistryBytes, nil
				}
			},
			PutCalled: func(key, data []byte) error {
				switch {
				case bytes.Equal(epoch10Key, key):
					numPutCalls++
					return nil
				case bytes.Equal(epoch9Key, key):
					numPutCalls++
					return nil
				case bytes.Equal(epoch8Key, key):
					numPutCalls++
					return nil
				default:
					return errors.New("put error")
				}
			},
		}

		numOfEpochsToKeep := uint32(3)

		nodesConfig, err := GetNodesCoordinatorRegistry([]byte("key"), storer, 10, numOfEpochsToKeep)
		require.Nil(t, err)
		require.Equal(t, nodesConfigRegistry, nodesConfig)
		require.Equal(t, numOfEpochsToKeep, numPutCalls)
	})

	t.Run("getting each key separatelly by epoch, should work", func(t *testing.T) {
		t.Parallel()

		nodesConfigRegistry := &NodesCoordinatorRegistry{
			EpochsConfig: map[string]*EpochValidators{
				"10": {
					EligibleValidators: map[string][]*SerializableValidator{
						"0": {
							{
								PubKey: []byte("pubKey1"),
							},
						},
						"1": {
							{
								PubKey: []byte("pubKey1"),
							},
						},
						"2": {
							{
								PubKey: []byte("pubKey1"),
							},
						},
					},
				},
			},
			CurrentEpoch: 10,
		}

		storer := &mock.StorerStub{
			GetCalled: func(key []byte) (b []byte, err error) {
				return nil, errors.New("get failed")
			},
			SearchFirstCalled: func(key []byte) ([]byte, error) {
				switch {
				case strings.Contains(string(key), common.NodesCoordinatorRegistryKeyPrefix):
					nodesConfigRegistryBytes, _ := json.Marshal(nodesConfigRegistry)
					return nodesConfigRegistryBytes, nil
				default:
					return nil, errors.New("invalid key")
				}
			},
		}

		nodesConfig, err := GetNodesCoordinatorRegistry([]byte("key"), storer, 10, numStoredEpochs)
		require.Nil(t, err)
		require.Equal(t, nodesConfigRegistry, nodesConfig)
	})
}

func TestIndexHashedNodesCoordinator_SaveNodesCoordinatorRegistry(t *testing.T) {
	t.Parallel()

	t.Run("nil nodes config, should fail", func(t *testing.T) {
		t.Parallel()

		err := SaveNodesCoordinatorRegistry(nil, &mock.StorerStub{})
		require.Equal(t, ErrNilNodesCoordinatorRegistry, err)
	})

	t.Run("nil storer, should fail", func(t *testing.T) {
		t.Parallel()

		nodesConfigRegistry := &NodesCoordinatorRegistry{
			CurrentEpoch: 10,
		}

		err := SaveNodesCoordinatorRegistry(nodesConfigRegistry, nil)
		require.Equal(t, ErrNilBootStorer, err)
	})

	t.Run("failed to put into storer", func(t *testing.T) {
		t.Parallel()

		nodesConfigRegistry := &NodesCoordinatorRegistry{
			CurrentEpoch: 10,
			EpochsConfig: map[string]*EpochValidators{
				"10": {
					EligibleValidators: map[string][]*SerializableValidator{
						"val1": {
							{
								PubKey: []byte("pubKey1"),
							},
						},
					},
				},
			},
		}

		expectedErr := errors.New("expected error")
		storer := &mock.StorerStub{
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
		}

		err := SaveNodesCoordinatorRegistry(nodesConfigRegistry, storer)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		nodesConfigRegistry := &NodesCoordinatorRegistry{
			CurrentEpoch: 10,
			EpochsConfig: map[string]*EpochValidators{
				"10": {
					EligibleValidators: map[string][]*SerializableValidator{
						"0": {{PubKey: []byte("pubKey1")}},
						"1": {{PubKey: []byte("pubKey1")}},
						"2": {{PubKey: []byte("pubKey1")}},
					},
				},
				"9": {
					EligibleValidators: map[string][]*SerializableValidator{
						"0": {{PubKey: []byte("pubKey2")}},
						"1": {{PubKey: []byte("pubKey2")}},
						"2": {{PubKey: []byte("pubKey2")}},
					},
				},
				"8": {
					EligibleValidators: map[string][]*SerializableValidator{
						"0": {{PubKey: []byte("pubKey3")}},
						"1": {{PubKey: []byte("pubKey3")}},
						"2": {{PubKey: []byte("pubKey3")}},
					},
				},
			},
		}

		putCalls := 0
		storer := &mock.StorerStub{
			PutCalled: func(key, data []byte) error {
				switch {
				case strings.Contains(string(key), common.NodesCoordinatorRegistryKeyPrefix):
					putCalls++
					return nil
				default:
					return errors.New("invalid key")
				}
			},
		}

		err := SaveNodesCoordinatorRegistry(nodesConfigRegistry, storer)
		require.Nil(t, err)
		require.Equal(t, 3, putCalls)
	})
}

func TestIndexHashedNodesCoordinator_SaveLoadNodesCoordinatorRegistry(t *testing.T) {
	t.Parallel()

	logger.SetLogLevel("*:DEBUG")

	epochKey := []byte(fmt.Sprint(1))

	args := createArguments()
	args.ValidatorInfoCacher = dataPool.NewCurrentEpochValidatorInfoPool()
	args.BootStorer = genericMocks.NewStorerMockWithEpoch(1)
	ihnc, _ := NewIndexHashedNodesCoordinator(args)

	eligibleMap := createDummyNodesMap(10, 3, "eligible")
	waitingMap := createDummyNodesMap(10, 3, "waiting")
	leavingMap := createDummyNodesMap(10, 3, "leaving")
	_ = ihnc.setNodesPerShards(eligibleMap, waitingMap, leavingMap, 1)
	_ = ihnc.setNodesPerShards(eligibleMap, waitingMap, leavingMap, 2)
	_ = ihnc.setNodesPerShards(eligibleMap, waitingMap, leavingMap, 3)
	_ = ihnc.setNodesPerShards(eligibleMap, waitingMap, leavingMap, 4)

	expectedNodesConfig, ok := ihnc.nodesConfig[1]
	require.True(t, ok)
	require.NotNil(t, expectedNodesConfig)

	err := ihnc.saveState(epochKey)
	assert.Nil(t, err)

	delete(ihnc.nodesConfig, 1)
	err = ihnc.LoadState(epochKey, 1)
	assert.Nil(t, err)

	actualConfig, ok := ihnc.nodesConfig[1]
	require.True(t, ok)

	assert.Equal(t, expectedNodesConfig.shardID, expectedNodesConfig.shardID)
	assert.Equal(t, expectedNodesConfig.nbShards, expectedNodesConfig.nbShards)
	assert.True(t, sameValidatorsMaps(expectedNodesConfig.eligibleMap, actualConfig.eligibleMap))
	assert.True(t, sameValidatorsMaps(expectedNodesConfig.waitingMap, actualConfig.waitingMap))
	assert.True(t, sameValidatorsMaps(expectedNodesConfig.leavingMap, actualConfig.leavingMap))
}
