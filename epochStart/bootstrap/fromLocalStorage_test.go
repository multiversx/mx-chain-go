package bootstrap

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
)

func TestPrepareEpochFromStorage(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, err := NewEpochStartBootstrap(args)
	require.Nil(t, err)
	epochStartProvider.initializeFromLocalStorage()

	epochStartProvider.baseData.lastEpoch = 10
	_, err = epochStartProvider.prepareEpochFromStorage()
	assert.Error(t, err)
}

func TestGetEpochStartMetaFromStorage(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, err := NewEpochStartBootstrap(args)
	require.Nil(t, err)
	epochStartProvider.initializeFromLocalStorage()

	meta := &block.MetaBlock{Nonce: 1}
	metaBytes, _ := json.Marshal(meta)
	storer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) (bytes []byte, err error) {
			return metaBytes, nil
		},
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			return metaBytes, nil
		},
	}
	metaBlock, err := epochStartProvider.getEpochStartMetaFromStorage(storer)
	assert.Nil(t, err)
	assert.Equal(t, meta, metaBlock)
}

func TestGetEpochStartMetaFromStorageFallbackToPreviousEpoch(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, err := NewEpochStartBootstrap(args)
	require.Nil(t, err)
	epochStartProvider.initializeFromLocalStorage()
	epochStartProvider.baseData.lastEpoch = 10

	meta := &block.MetaBlock{Nonce: 1, Epoch: 9}
	metaBytes, _ := json.Marshal(meta)
	searchedKeys := make([][]byte, 0)
	storer := &storageStubs.StorerStub{
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			searchedKeys = append(searchedKeys, append([]byte(nil), key...))
			if bytes.Equal(key, []byte(core.EpochStartIdentifier(10))) {
				return nil, errors.New("missing epoch start metablock")
			}
			if bytes.Equal(key, []byte(core.EpochStartIdentifier(9))) {
				return metaBytes, nil
			}

			return nil, errors.New("unexpected epoch start metablock key")
		},
	}

	metaBlock, err := epochStartProvider.getEpochStartMetaFromStorage(storer)
	require.Nil(t, err)
	assert.Equal(t, meta, metaBlock)
	assert.Equal(t, uint32(9), epochStartProvider.baseData.lastEpoch)
	require.Len(t, searchedKeys, 2)
	assert.Equal(t, []byte(core.EpochStartIdentifier(10)), searchedKeys[0])
	assert.Equal(t, []byte(core.EpochStartIdentifier(9)), searchedKeys[1])
}

func TestGetEpochStartMetaFromStorageFallsBackMultipleEpochs(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, err := NewEpochStartBootstrap(args)
	require.Nil(t, err)
	epochStartProvider.initializeFromLocalStorage()
	epochStartProvider.baseData.lastEpoch = 10

	meta := &block.MetaBlock{Nonce: 1, Epoch: 7}
	metaBytes, _ := json.Marshal(meta)
	searchedKeys := make([][]byte, 0)
	storer := &storageStubs.StorerStub{
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			searchedKeys = append(searchedKeys, append([]byte(nil), key...))
			if bytes.Equal(key, []byte(core.EpochStartIdentifier(7))) {
				return metaBytes, nil
			}

			return nil, errors.New("missing epoch start metablock")
		},
	}

	metaBlock, err := epochStartProvider.getEpochStartMetaFromStorage(storer)
	require.NoError(t, err)
	require.Equal(t, meta, metaBlock)
	require.Equal(t, uint32(7), epochStartProvider.baseData.lastEpoch)
	require.Len(t, searchedKeys, 4)
	assert.Equal(t, []byte(core.EpochStartIdentifier(10)), searchedKeys[0])
	assert.Equal(t, []byte(core.EpochStartIdentifier(9)), searchedKeys[1])
	assert.Equal(t, []byte(core.EpochStartIdentifier(8)), searchedKeys[2])
	assert.Equal(t, []byte(core.EpochStartIdentifier(7)), searchedKeys[3])
}

func TestGetEpochStartMetaFromStorageReturnsErrorAfterSearchingToEpochZero(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, err := NewEpochStartBootstrap(args)
	require.Nil(t, err)
	epochStartProvider.initializeFromLocalStorage()
	epochStartProvider.baseData.lastEpoch = 2

	searchErr := errors.New("missing epoch start metablock")
	searchedKeys := make([][]byte, 0)
	storer := &storageStubs.StorerStub{
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			searchedKeys = append(searchedKeys, append([]byte(nil), key...))
			return nil, searchErr
		},
	}

	metaBlock, err := epochStartProvider.getEpochStartMetaFromStorage(storer)
	require.Equal(t, searchErr, err)
	require.Nil(t, metaBlock)
	require.Equal(t, uint32(2), epochStartProvider.baseData.lastEpoch)
	require.Len(t, searchedKeys, 3)
	assert.Equal(t, []byte(core.EpochStartIdentifier(2)), searchedKeys[0])
	assert.Equal(t, []byte(core.EpochStartIdentifier(1)), searchedKeys[1])
	assert.Equal(t, []byte(core.EpochStartIdentifier(0)), searchedKeys[2])
}

func TestGetEpochStartMetaFromStorageUnmarshalErrorDoesNotFallBack(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, err := NewEpochStartBootstrap(args)
	require.Nil(t, err)
	epochStartProvider.initializeFromLocalStorage()
	epochStartProvider.baseData.lastEpoch = 10

	searchedKeys := make([][]byte, 0)
	storer := &storageStubs.StorerStub{
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			searchedKeys = append(searchedKeys, append([]byte(nil), key...))
			return []byte("not a valid meta header"), nil
		},
	}

	metaBlock, err := epochStartProvider.getEpochStartMetaFromStorage(storer)
	require.Error(t, err)
	require.Nil(t, metaBlock)
	// a corrupt metablock found at the latest epoch must fail hard, not fall back to an older epoch
	require.Len(t, searchedKeys, 1)
	assert.Equal(t, []byte(core.EpochStartIdentifier(10)), searchedKeys[0])
	assert.Equal(t, uint32(10), epochStartProvider.baseData.lastEpoch)
}

func TestGetShardIDForLatestEpochFallbackWithMissingNodesConfigErrors(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, err := NewEpochStartBootstrap(args)
	require.Nil(t, err)
	epochStartProvider.initializeFromLocalStorage()
	epochStartProvider.baseData.lastEpoch = 10

	round := int64(10)
	roundBytes, _ := json.Marshal(&bootstrapStorage.RoundNum{Num: round})
	bootstrapData := bootstrapStorage.BootstrapData{NodesCoordinatorConfigKey: []byte("key")}
	bootstrapDataBytes, _ := json.Marshal(bootstrapData)
	nodesCoordinatorKey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), bootstrapData.NodesCoordinatorConfigKey...)

	// the registry only knows about the latest epoch 10, not the fallback epoch 9
	registryBytes, _ := json.Marshal(&nodesCoordinator.NodesCoordinatorRegistry{
		EpochsConfig: map[string]*nodesCoordinator.EpochValidators{
			"10": {},
		},
	})
	metaBytes, _ := json.Marshal(&block.MetaBlock{Nonce: 1, Epoch: 9})

	storer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			switch {
			case bytes.Equal([]byte(common.HighestRoundFromBootStorage), key):
				return roundBytes, nil
			case bytes.Equal([]byte(strconv.FormatInt(round, 10)), key):
				return bootstrapDataBytes, nil
			default:
				return nil, nil
			}
		},
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			switch {
			case bytes.Equal(nodesCoordinatorKey, key):
				return registryBytes, nil
			case bytes.Equal([]byte(core.EpochStartIdentifier(10)), key):
				return nil, errors.New("missing epoch start metablock")
			case bytes.Equal([]byte(core.EpochStartIdentifier(9)), key):
				return metaBytes, nil
			default:
				return nil, errors.New("unexpected key")
			}
		},
	}

	epochStartProvider.storageOpenerHandler = &storageStubs.UnitOpenerStub{
		GetMostRecentStorageUnitCalled: func(cfg config.DBConfig) (storage.Storer, error) {
			return storer, nil
		},
	}

	_, _, err = epochStartProvider.getShardIDForLatestEpoch()
	// the metablock lookup falls back from epoch 10 to 9, but the nodes config only contains epoch 10,
	// so the mixed-epoch parameters are rejected instead of starting with a stale validator set
	require.True(t, errors.Is(err, epochStart.ErrMissingNodesConfigForBootstrapEpoch))
	assert.Equal(t, uint32(9), epochStartProvider.baseData.lastEpoch)
}

func TestCheckNodesConfigForEpoch(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, err := NewEpochStartBootstrap(args)
	require.Nil(t, err)

	t.Run("nil nodes config should error", func(t *testing.T) {
		epochStartProvider.nodesConfig = nil
		errCheck := epochStartProvider.checkNodesConfigForEpoch(9)
		require.True(t, errors.Is(errCheck, epochStart.ErrMissingNodesConfigForBootstrapEpoch))
	})

	t.Run("epoch config missing should error", func(t *testing.T) {
		epochStartProvider.nodesConfig = &nodesCoordinator.NodesCoordinatorRegistry{
			EpochsConfig: map[string]*nodesCoordinator.EpochValidators{
				"10": {},
			},
		}
		errCheck := epochStartProvider.checkNodesConfigForEpoch(9)
		require.True(t, errors.Is(errCheck, epochStart.ErrMissingNodesConfigForBootstrapEpoch))
	})

	t.Run("epoch config present should pass", func(t *testing.T) {
		epochStartProvider.nodesConfig = &nodesCoordinator.NodesCoordinatorRegistry{
			EpochsConfig: map[string]*nodesCoordinator.EpochValidators{
				"9": {},
			},
		}
		errCheck := epochStartProvider.checkNodesConfigForEpoch(9)
		require.NoError(t, errCheck)
	})
}

func TestGetLastBootstrapData(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.initializeFromLocalStorage()

	round := int64(10)

	roundNum := bootstrapStorage.RoundNum{
		Num: round,
	}
	roundBytes, _ := json.Marshal(&roundNum)
	nodesCoordinatorConfigKey := []byte("key")

	nodesConfigRegistry := nodesCoordinator.NodesCoordinatorRegistry{
		CurrentEpoch: 10,
	}
	bootstrapData := bootstrapStorage.BootstrapData{
		NodesCoordinatorConfigKey: nodesCoordinatorConfigKey,
	}

	storer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) (b []byte, err error) {
			switch {
			case bytes.Equal([]byte(common.HighestRoundFromBootStorage), key):
				return roundBytes, nil
			case bytes.Equal([]byte(strconv.FormatInt(round, 10)), key):

				bootstrapDataBytes, _ := json.Marshal(bootstrapData)
				return bootstrapDataBytes, nil
			default:
				return nil, nil
			}
		},
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			nodesConfigRegistryBytes, _ := json.Marshal(nodesConfigRegistry)
			return nodesConfigRegistryBytes, nil
		},
	}

	bootData, nodesRegistry, err := epochStartProvider.getLastBootstrapData(storer)
	assert.Nil(t, err)
	assert.Equal(t, &bootstrapData, bootData)
	assert.Equal(t, &nodesConfigRegistry, nodesRegistry)
}

func TestCheckIfShuffledOut_ValidatorIsInWaitingList(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.initializeFromLocalStorage()
	epochStartProvider.baseData.lastEpoch = 0

	publicKey := []byte("pubKey")
	nodesConfig := &nodesCoordinator.NodesCoordinatorRegistry{
		CurrentEpoch: 1,
		EpochsConfig: map[string]*nodesCoordinator.EpochValidators{
			"0": {
				WaitingValidators: map[string][]*nodesCoordinator.SerializableValidator{
					"0": {{PubKey: publicKey, Chances: 0, Index: 0}},
				},
			},
		},
	}

	shardId, result := epochStartProvider.checkIfShuffledOut(publicKey, nodesConfig)
	assert.False(t, result)
	assert.Equal(t, shardId, epochStartProvider.baseData.shardId)
}

func TestCheckIfShuffledOut_ValidatorIsInEligibleList(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.initializeFromLocalStorage()
	epochStartProvider.baseData.lastEpoch = 0

	publicKey := []byte("pubKey")
	nodesConfig := &nodesCoordinator.NodesCoordinatorRegistry{
		CurrentEpoch: 1,
		EpochsConfig: map[string]*nodesCoordinator.EpochValidators{
			"0": {
				EligibleValidators: map[string][]*nodesCoordinator.SerializableValidator{
					"0": {{PubKey: publicKey, Chances: 0, Index: 0}},
				},
			},
		},
	}

	shardId, result := epochStartProvider.checkIfShuffledOut(publicKey, nodesConfig)
	assert.False(t, result)
	assert.Equal(t, shardId, epochStartProvider.baseData.shardId)
}

func TestCheckIfShuffledOut_ValidatorIsShuffledToEligibleList(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.initializeFromLocalStorage()
	epochStartProvider.baseData.lastEpoch = 0
	epochStartProvider.baseData.shardId = 1

	publicKey := []byte("pubKey")
	nodesConfig := &nodesCoordinator.NodesCoordinatorRegistry{
		CurrentEpoch: 1,
		EpochsConfig: map[string]*nodesCoordinator.EpochValidators{
			"0": {
				EligibleValidators: map[string][]*nodesCoordinator.SerializableValidator{
					"0": {{PubKey: publicKey, Chances: 0, Index: 0}},
				},
			},
		},
	}

	shardId, result := epochStartProvider.checkIfShuffledOut(publicKey, nodesConfig)
	assert.True(t, result)
	assert.NotEqual(t, shardId, epochStartProvider.baseData.shardId)
}

func TestCheckIfShuffledOut_ValidatorNotInEligibleOrWaiting(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	epochStartProvider, _ := NewEpochStartBootstrap(args)
	epochStartProvider.initializeFromLocalStorage()
	epochStartProvider.baseData.lastEpoch = 0

	publicKey := []byte("pubKey")
	nodesConfig := &nodesCoordinator.NodesCoordinatorRegistry{
		CurrentEpoch: 1,
		EpochsConfig: map[string]*nodesCoordinator.EpochValidators{
			"0": {
				EligibleValidators: map[string][]*nodesCoordinator.SerializableValidator{},
				WaitingValidators:  map[string][]*nodesCoordinator.SerializableValidator{},
			},
		},
	}

	shardId, result := epochStartProvider.checkIfShuffledOut(publicKey, nodesConfig)
	assert.False(t, result)
	assert.Equal(t, epochStartProvider.baseData.shardId, shardId)
}

func TestStartFromSavedEpoch_ShuffledOutStaleEpochStartJoinsFromNetwork(t *testing.T) {
	epochStartRound := uint64(1000)
	roundsPerEpoch := int64(200)

	createProvider := func(shuffledOut bool, currentRound int64, storageUnitOpened *bool) *epochStartBootstrap {
		coreComp, cryptoComp := createComponentsForEpochStart()
		coreComp.ChainParametersHandlerField = &chainParameters.ChainParametersHandlerStub{
			CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
				return config.ChainParametersByEpochConfig{RoundsPerEpoch: roundsPerEpoch}
			},
			ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
				return config.ChainParametersByEpochConfig{RoundsPerEpoch: roundsPerEpoch}, nil
			},
		}
		args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
		args.RoundHandler = &mock.RoundHandlerStub{
			IndexCalled: func() int64 {
				return currentRound
			},
		}
		args.LatestStorageDataProvider = &mock.LatestStorageDataProviderStub{
			GetCalled: func() (storage.LatestDataFromStorage, error) {
				return storage.LatestDataFromStorage{Epoch: 10, ShardID: core.MetachainShardId, LastRound: int64(epochStartRound) + 5, EpochStartRound: epochStartRound}, nil
			},
		}
		args.StorageUnitOpener = &storageStubs.UnitOpenerStub{
			GetMostRecentStorageUnitCalled: func(config config.DBConfig) (storage.Storer, error) {
				*storageUnitOpened = true
				return nil, errors.New("stop the storage path here")
			},
		}
		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, err)
		epochStartProvider.shuffledOut = shuffledOut

		return epochStartProvider
	}

	t.Run("shuffled out with stale epoch start skips storage and continues from the network", func(t *testing.T) {
		storageUnitOpened := false
		staleRound := int64(epochStartRound) + roundsPerEpoch + 10
		epochStartProvider := createProvider(true, staleRound, &storageUnitOpened)

		params, shouldContinue, err := epochStartProvider.startFromSavedEpoch()
		require.Nil(t, err)
		require.True(t, shouldContinue)
		require.Equal(t, Parameters{}, params)
		require.False(t, storageUnitOpened, "prepareEpochFromStorage must not run on a stale epoch start")
	})

	t.Run("shuffled out with current epoch start keeps the storage path", func(t *testing.T) {
		storageUnitOpened := false
		freshRound := int64(epochStartRound) + roundsPerEpoch/2
		epochStartProvider := createProvider(true, freshRound, &storageUnitOpened)

		_, shouldContinue, err := epochStartProvider.startFromSavedEpoch()
		require.NotNil(t, err)
		require.False(t, shouldContinue, "shuffled out storage attempt must not fall through to the network")
		require.True(t, storageUnitOpened, "prepareEpochFromStorage should have been attempted")
	})

	t.Run("not shuffled out still attempts the storage path and can fall through", func(t *testing.T) {
		storageUnitOpened := false
		staleRound := int64(epochStartRound) + roundsPerEpoch + 10
		epochStartProvider := createProvider(false, staleRound, &storageUnitOpened)

		_, shouldContinue, err := epochStartProvider.startFromSavedEpoch()
		require.Nil(t, err)
		require.True(t, shouldContinue)
		require.True(t, storageUnitOpened, "prepareEpochFromStorage should have been attempted")
	})
}
