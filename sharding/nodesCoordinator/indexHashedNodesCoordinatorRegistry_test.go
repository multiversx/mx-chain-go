package nodesCoordinator

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
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
	t.Parallel()

	args := createArguments()
	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		GetActivationEpochCalled: func(flag core.EnableEpochFlag) uint32 {
			if flag == common.StakingV4Step2Flag {
				return stakingV4Epoch
			}
			return 0
		},
	}
	nodesCoordinator, _ := NewIndexHashedNodesCoordinator(args)

	expectedConfig := nodesCoordinator.nodesConfig[0]

	key := []byte("config")
	err := nodesCoordinator.saveState(key, 0)
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

func TestIndexHashedNodesCoordinator_LoadStateAfterSaveWithStakingV4(t *testing.T) {
	t.Parallel()

	args := createArguments()
	args.Epoch = stakingV4Epoch
	nodesCoordinator, _ := NewIndexHashedNodesCoordinator(args)

	nodesCoordinator.nodesConfig[stakingV4Epoch].leavingMap = createDummyNodesMap(3, 0, string(common.LeavingList))
	nodesCoordinator.nodesConfig[stakingV4Epoch].shuffledOutMap = createDummyNodesMap(3, 0, string(common.SelectedFromAuctionList))
	expectedConfig := nodesCoordinator.nodesConfig[stakingV4Epoch]

	key := []byte("config")
	err := nodesCoordinator.saveState(key, stakingV4Epoch)
	assert.Nil(t, err)

	delete(nodesCoordinator.nodesConfig, 0)
	err = nodesCoordinator.LoadState(key)
	assert.Nil(t, err)

	actualConfig := nodesCoordinator.nodesConfig[stakingV4Epoch]
	assert.Equal(t, expectedConfig.shardID, actualConfig.shardID)
	assert.Equal(t, expectedConfig.nbShards, actualConfig.nbShards)
	assert.True(t, sameValidatorsMaps(expectedConfig.eligibleMap, actualConfig.eligibleMap))
	assert.True(t, sameValidatorsMaps(expectedConfig.waitingMap, actualConfig.waitingMap))
	assert.True(t, sameValidatorsMaps(expectedConfig.shuffledOutMap, actualConfig.shuffledOutMap))
	assert.True(t, sameValidatorsMaps(expectedConfig.leavingMap, actualConfig.leavingMap))
}

func TestIndexHashedNodesCoordinator_nodesCoordinatorToRegistryWithStakingV4(t *testing.T) {
	args := createArguments()
	args.Epoch = stakingV4Epoch
	nodesCoordinator, _ := NewIndexHashedNodesCoordinator(args)

	nodesCoordinator.nodesConfig[stakingV4Epoch].leavingMap = createDummyNodesMap(3, 0, string(common.LeavingList))
	nodesCoordinator.nodesConfig[stakingV4Epoch].shuffledOutMap = createDummyNodesMap(3, 0, string(common.SelectedFromAuctionList))

	ncr := nodesCoordinator.NodesCoordinatorToRegistry(stakingV4Epoch)
	nc := nodesCoordinator.nodesConfig

	assert.Equal(t, nodesCoordinator.currentEpoch, ncr.GetCurrentEpoch())
	assert.Equal(t, len(nodesCoordinator.nodesConfig), len(ncr.GetEpochsConfig()))

	for epoch, config := range nc {
		ncrWithAuction := ncr.GetEpochsConfig()[fmt.Sprint(epoch)].(EpochValidatorsHandlerWithAuction)
		assert.True(t, sameValidatorsDifferentMapTypes(config.waitingMap, ncrWithAuction.GetWaitingValidators()))
		assert.True(t, sameValidatorsDifferentMapTypes(config.leavingMap, ncrWithAuction.GetLeavingValidators()))
		assert.True(t, sameValidatorsDifferentMapTypes(config.eligibleMap, ncrWithAuction.GetEligibleValidators()))
		assert.True(t, sameValidatorsDifferentMapTypes(config.shuffledOutMap, ncrWithAuction.GetShuffledOutValidators()))
	}
}

func TestIndexHashedNodesCoordinator_nodesCoordinatorToRegistry(t *testing.T) {
	args := createArguments()
	nodesCoordinator, _ := NewIndexHashedNodesCoordinator(args)

	ncr := nodesCoordinator.NodesCoordinatorToRegistry(args.Epoch)
	nc := nodesCoordinator.nodesConfig

	assert.Equal(t, nodesCoordinator.currentEpoch, ncr.GetCurrentEpoch())
	assert.Equal(t, len(nodesCoordinator.nodesConfig), len(ncr.GetEpochsConfig()))

	for epoch, config := range nc {
		assert.True(t, sameValidatorsDifferentMapTypes(config.eligibleMap, ncr.GetEpochsConfig()[fmt.Sprint(epoch)].GetEligibleValidators()))
		assert.True(t, sameValidatorsDifferentMapTypes(config.waitingMap, ncr.GetEpochsConfig()[fmt.Sprint(epoch)].GetWaitingValidators()))
	}
}

func TestIndexHashedNodesCoordinator_nodesCoordinatorWithAuctionToRegistryAndBack(t *testing.T) {
	args := createArguments()
	nodesCoordinator, _ := NewIndexHashedNodesCoordinator(args)

	nodesConfigForEpoch := nodesCoordinator.nodesConfig[args.Epoch]
	nodesConfigForEpoch.shuffledOutMap = createDummyNodesMap(3, 0, string(common.WaitingList))
	nodesConfigForEpoch.lowWaitingList = true
	// leave only one epoch config in nc
	nodesCoordinator.nodesConfig = make(map[uint32]*epochNodesConfig)
	nodesCoordinator.nodesConfig[args.Epoch] = nodesConfigForEpoch

	ncr := nodesCoordinator.nodesCoordinatorToRegistryWithAuction()
	require.True(t, sameValidatorsDifferentMapTypes(nodesConfigForEpoch.eligibleMap, ncr.GetEpochsConfig()[fmt.Sprint(args.Epoch)].GetEligibleValidators()))
	require.True(t, sameValidatorsDifferentMapTypes(nodesConfigForEpoch.waitingMap, ncr.GetEpochsConfig()[fmt.Sprint(args.Epoch)].GetWaitingValidators()))
	require.True(t, sameValidatorsDifferentMapTypes(nodesConfigForEpoch.shuffledOutMap, ncr.GetEpochsConfigWithAuction()[fmt.Sprint(args.Epoch)].GetShuffledOutValidators()))
	require.Equal(t, nodesConfigForEpoch.lowWaitingList, ncr.GetEpochsConfigWithAuction()[fmt.Sprint(args.Epoch)].GetLowWaitingList())

	nodesConfig, err := nodesCoordinator.registryToNodesCoordinator(ncr)
	require.Nil(t, err)

	assert.Equal(t, len(nodesCoordinator.nodesConfig), len(nodesConfig))
	for epoch, config := range nodesCoordinator.nodesConfig {
		require.True(t, sameValidatorsMaps(config.eligibleMap, nodesConfig[epoch].eligibleMap))
		require.True(t, sameValidatorsMaps(config.waitingMap, nodesConfig[epoch].waitingMap))
		require.True(t, sameValidatorsMaps(config.shuffledOutMap, nodesConfig[epoch].shuffledOutMap))
		require.Equal(t, config.lowWaitingList, nodesConfig[epoch].lowWaitingList)
	}
}

func TestIndexHashedNodesCoordinator_registryToNodesCoordinator(t *testing.T) {
	args := createArguments()
	nodesCoordinator1, _ := NewIndexHashedNodesCoordinator(args)
	ncr := nodesCoordinator1.NodesCoordinatorToRegistry(args.Epoch)

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

	ncr := nodesCoordinator.NodesCoordinatorToRegistry(args.Epoch)
	nc := nodesCoordinator.nodesConfig

	require.Equal(t, nodesCoordinator.currentEpoch, ncr.GetCurrentEpoch())
	require.Equal(t, NodesCoordinatorStoredEpochs, len(ncr.GetEpochsConfig()))

	for epochStr := range ncr.GetEpochsConfig() {
		epoch, err := strconv.Atoi(epochStr)
		require.Nil(t, err)
		require.True(t, sameValidatorsDifferentMapTypes(nc[uint32(epoch)].eligibleMap, ncr.GetEpochsConfig()[epochStr].GetEligibleValidators()))
		require.True(t, sameValidatorsDifferentMapTypes(nc[uint32(epoch)].waitingMap, ncr.GetEpochsConfig()[epochStr].GetWaitingValidators()))
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

func createArgsForStateTests() ArgNodesCoordinator {
	args := createArguments()
	args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
		GetActivationEpochCalled: func(flag core.EnableEpochFlag) uint32 {
			if flag == common.StakingV4Step2Flag {
				return stakingV4Epoch
			}
			return 0
		},
	}
	return args
}

func TestIndexHashedNodesCoordinator_MergeStateAddsOnlyMissingEpochs(t *testing.T) {
	t.Parallel()

	args := createArgsForStateTests()
	nc, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)

	err = nc.saveState([]byte("old"), 0)
	require.Nil(t, err)

	nc.nodesConfig[5] = nc.nodesConfig[0]
	delete(nc.nodesConfig, 0)
	require.Nil(t, nc.nodesConfig[0])
	require.NotNil(t, nc.nodesConfig[5])

	err = nc.MergeState([]byte("old"))
	require.Nil(t, err)
	require.NotNil(t, nc.nodesConfig[0])
	require.NotNil(t, nc.nodesConfig[5])
}

func TestIndexHashedNodesCoordinator_MergeStateDoesNotOverwriteExisting(t *testing.T) {
	t.Parallel()

	args := createArgsForStateTests()
	nc, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)

	err = nc.saveState([]byte("backup"), 0)
	require.Nil(t, err)

	nc.nodesConfig[0].shardID = 99

	err = nc.MergeState([]byte("backup"))
	require.Nil(t, err)
	require.Equal(t, uint32(99), nc.nodesConfig[0].shardID)
}

func TestIndexHashedNodesCoordinator_MergeStateKeyNotFound(t *testing.T) {
	t.Parallel()

	args := createArgsForStateTests()
	nc, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)

	err = nc.MergeState([]byte("nonexistent"))
	require.NotNil(t, err)
	require.NotNil(t, nc.nodesConfig[0])
}

func TestIndexHashedNodesCoordinator_MergeStatePreservesValidators(t *testing.T) {
	t.Parallel()

	args := createArgsForStateTests()
	nc, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)

	expectedEligible := nc.nodesConfig[0].eligibleMap
	expectedWaiting := nc.nodesConfig[0].waitingMap
	expectedNbShards := nc.nodesConfig[0].nbShards

	err = nc.saveState([]byte("key1"), 0)
	require.Nil(t, err)

	delete(nc.nodesConfig, 0)

	err = nc.MergeState([]byte("key1"))
	require.Nil(t, err)

	actual := nc.nodesConfig[0]
	require.NotNil(t, actual)
	require.Equal(t, expectedNbShards, actual.nbShards)
	require.True(t, sameValidatorsMaps(expectedEligible, actual.eligibleMap))
	require.True(t, sameValidatorsMaps(expectedWaiting, actual.waitingMap))
	require.NotNil(t, actual.selectors)
}

func TestIndexHashedNodesCoordinator_LoadStateThenMergeOlderEpochs(t *testing.T) {
	t.Parallel()

	args := createArgsForStateTests()
	args.Epoch = 10
	nc, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)

	for e := uint32(5); e <= 10; e++ {
		nc.nodesConfig[e] = &epochNodesConfig{
			nbShards:       args.NbShards,
			shardID:        args.ShardIDAsObserver,
			eligibleMap:    createDummyNodesMap(10, args.NbShards, fmt.Sprintf("eligible_%d", e)),
			waitingMap:     createDummyNodesMap(3, args.NbShards, fmt.Sprintf("waiting_%d", e)),
			selectors:      make(map[uint32]RandomSelector),
			leavingMap:     make(map[uint32][]Validator),
			shuffledOutMap: make(map[uint32][]Validator),
			newList:        make([]Validator, 0),
			auctionList:    make([]Validator, 0),
		}
		nc.nodesConfig[e].selectors, _ = nc.createSelectors(nc.nodesConfig[e])
	}

	nc.currentEpoch = 7
	err = nc.saveState([]byte("old_key"), 7)
	require.Nil(t, err)

	nc.currentEpoch = 10
	err = nc.saveState([]byte("current_key"), 10)
	require.Nil(t, err)

	nc.nodesConfig = make(map[uint32]*epochNodesConfig)
	err = nc.LoadState([]byte("current_key"))
	require.Nil(t, err)

	epochsBeforeMerge := nc.GetCachedEpochs()
	require.True(t, len(epochsBeforeMerge) > 0)

	err = nc.MergeState([]byte("old_key"))
	require.Nil(t, err)

	epochsAfterMerge := nc.GetCachedEpochs()
	require.True(t, len(epochsAfterMerge) >= len(epochsBeforeMerge))
}

func TestIndexHashedNodesCoordinator_SaveStateAfterMergeOnlyKeepsWindow(t *testing.T) {
	t.Parallel()

	args := createArgsForStateTests()
	args.Epoch = 15
	nc, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)

	for e := uint32(5); e <= 15; e++ {
		nc.nodesConfig[e] = &epochNodesConfig{
			nbShards:       args.NbShards,
			shardID:        args.ShardIDAsObserver,
			eligibleMap:    createDummyNodesMap(10, args.NbShards, fmt.Sprintf("eligible_%d", e)),
			waitingMap:     createDummyNodesMap(3, args.NbShards, fmt.Sprintf("waiting_%d", e)),
			selectors:      make(map[uint32]RandomSelector),
			leavingMap:     make(map[uint32][]Validator),
			shuffledOutMap: make(map[uint32][]Validator),
			newList:        make([]Validator, 0),
			auctionList:    make([]Validator, 0),
		}
		nc.nodesConfig[e].selectors, _ = nc.createSelectors(nc.nodesConfig[e])
	}
	nc.currentEpoch = 15

	err = nc.saveState([]byte("save_key"), 15)
	require.Nil(t, err)

	nc.nodesConfig = make(map[uint32]*epochNodesConfig)
	err = nc.LoadState([]byte("save_key"))
	require.Nil(t, err)

	loadedEpochs := nc.GetCachedEpochs()
	require.Equal(t, NodesCoordinatorStoredEpochs, len(loadedEpochs))

	for e := uint32(15 - NodesCoordinatorStoredEpochs + 1); e <= 15; e++ {
		_, exists := loadedEpochs[e]
		require.True(t, exists, fmt.Sprintf("epoch %d should be in saved state", e))
	}
}
