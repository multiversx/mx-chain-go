package bootstrap

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/epochStart/mock"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/directoryhandler"
	storageFactory "github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/latestData"
	storageMock "github.com/multiversx/mx-chain-go/storage/mock"
	"github.com/multiversx/mx-chain-go/testscommon/chainParameters"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
)

func TestRecoveryBootstrapOnlyUsesNetworkWithoutLocalStorage(t *testing.T) {
	parent := t.TempDir()
	missing := filepath.Join(parent, "missing")
	provider := &epochStartBootstrap{latestStorageDataProvider: &mock.LatestStorageDataProviderStub{
		GetParentDirectoryCalled: func() string { return missing },
	}}
	require.True(t, provider.hasNoLocalStorage())

	provider.latestStorageDataProvider = &mock.LatestStorageDataProviderStub{
		GetParentDirectoryCalled: func() string { return parent },
	}
	require.True(t, provider.hasNoLocalStorage())
	require.NoError(t, os.Mkdir(filepath.Join(parent, "Epoch_1"), 0o700))
	require.False(t, provider.hasNoLocalStorage())
}

func TestRecoveryBootstrapDataUsesExactRound(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	approvedHash := bytes.Repeat([]byte{1}, 32)
	args.GeneralConfig.HardforkRoundExclusions = []config.HardforkRoundExclusionConfig{{StartRound: 101, EndRound: 199}}
	args.GeneralConfig.HardforkRecoveryCheckpoint = config.HardforkRecoveryCheckpointConfig{
		Enabled: true,
		Round:   100,
		Headers: []config.HardforkRecoveryHeaderConfig{
			{ShardID: 0, Hash: hex.EncodeToString(approvedHash)},
			{ShardID: core.MetachainShardId, Hash: hex.EncodeToString(approvedHash)},
		},
	}
	provider, err := NewEpochStartBootstrap(args)
	require.NoError(t, err)
	provider.baseData.shardId = 0
	provider.baseData.lastEpoch = 8

	selected := bootstrapStorage.BootstrapData{
		LastHeader:                bootstrapStorage.BootstrapHeaderInfo{ShardId: 0, Epoch: 7, Hash: approvedHash},
		NodesCoordinatorConfigKey: []byte("registry"),
	}
	roundBytes, _ := json.Marshal(&bootstrapStorage.RoundNum{Num: 105})
	selectedBytes, _ := json.Marshal(selected)
	registryBytes, _ := json.Marshal(&nodesCoordinator.NodesCoordinatorRegistry{})
	storer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			if bytes.Equal(key, []byte(common.HighestRoundFromBootStorage)) {
				return roundBytes, nil
			}
			require.Equal(t, []byte("100"), key)
			return selectedBytes, nil
		},
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			require.Equal(t, []byte(common.NodesCoordinatorRegistryKeyPrefix+"registry"), key)
			return registryBytes, nil
		},
	}
	data, _, err := provider.getLastBootstrapData(storer)
	require.NoError(t, err)
	require.Equal(t, selected.LastHeader, data.LastHeader)
	require.Equal(t, uint32(8), provider.baseData.lastEpoch)
	require.Equal(t, int64(100), provider.baseData.lastRound)
}

func TestRecoveryBootstrapDataBeforeCheckpointUsesSavedRound(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	setRecoveryCheckpointConfig(&args.GeneralConfig)
	provider, err := NewEpochStartBootstrap(args)
	require.NoError(t, err)
	provider.baseData.shardId = 0

	selected := bootstrapStorage.BootstrapData{
		LastHeader:                bootstrapStorage.BootstrapHeaderInfo{ShardId: 0, Epoch: 7},
		NodesCoordinatorConfigKey: []byte("registry"),
	}
	roundBytes, err := json.Marshal(&bootstrapStorage.RoundNum{Num: 99})
	require.NoError(t, err)
	selectedBytes, err := json.Marshal(selected)
	require.NoError(t, err)
	registryBytes, err := json.Marshal(&nodesCoordinator.NodesCoordinatorRegistry{})
	require.NoError(t, err)
	storer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			if bytes.Equal(key, []byte(common.HighestRoundFromBootStorage)) {
				return roundBytes, nil
			}
			require.Equal(t, []byte("99"), key)
			return selectedBytes, nil
		},
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			require.Equal(t, []byte(common.NodesCoordinatorRegistryKeyPrefix+"registry"), key)
			return registryBytes, nil
		},
	}
	data, _, err := provider.getLastBootstrapData(storer)
	require.NoError(t, err)
	require.Equal(t, selected.LastHeader, data.LastHeader)
	require.Equal(t, int64(99), provider.baseData.lastRound)
}

func TestPrepareEpochFromStorage_RecoveryKeepsSnapshotStorageEpoch(t *testing.T) {
	for _, tc := range []struct {
		name         string
		highestRound int64
		lastRound    int64
		headerEpoch  uint32
	}{
		{name: "snapshot at checkpoint", highestRound: 100, headerEpoch: 7},
		{name: "snapshot before checkpoint", highestRound: 99, headerEpoch: 7},
		{name: "committed tip after exclusion", highestRound: 200, lastRound: 100, headerEpoch: 7},
		{name: "rewind within same epoch", highestRound: 105, lastRound: 99, headerEpoch: 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			coreComp, cryptoComp := createComponentsForEpochStart()
			args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
			setRecoveryCheckpointConfig(&args.GeneralConfig)
			approvedHash, err := hex.DecodeString(args.GeneralConfig.HardforkRecoveryCheckpoint.Headers[0].Hash)
			require.NoError(t, err)
			bootstrapData := bootstrapStorage.BootstrapData{
				LastHeader:             bootstrapStorage.BootstrapHeaderInfo{ShardId: 0, Epoch: tc.headerEpoch, Nonce: 10, Hash: approvedHash},
				HighestFinalBlockNonce: 10, LastRound: tc.lastRound,
				NodesCoordinatorConfigKey: []byte("registry"), EpochStartTriggerConfigKey: []byte("trigger"),
			}
			roundBytes, err := json.Marshal(&bootstrapStorage.RoundNum{Num: tc.highestRound})
			require.NoError(t, err)
			bootstrapBytes, err := json.Marshal(&bootstrapData)
			require.NoError(t, err)
			registryBytes, err := json.Marshal(&nodesCoordinator.NodesCoordinatorRegistry{})
			require.NoError(t, err)
			triggerBytes, err := json.Marshal(&block.ShardTriggerRegistry{EpochStartRound: 100})
			require.NoError(t, err)
			metaBytes, err := json.Marshal(&block.MetaBlock{Epoch: 8})
			require.NoError(t, err)
			selectedRound := tc.highestRound
			if selectedRound >= 100 && selectedRound <= 199 {
				selectedRound = 100
			}
			storer := &storageStubs.StorerStub{
				GetCalled: func(key []byte) ([]byte, error) {
					switch string(key) {
					case common.HighestRoundFromBootStorage:
						return roundBytes, nil
					case common.TriggerRegistryKeyPrefix + "trigger":
						return triggerBytes, nil
					default:
						require.Equal(t, strconv.FormatInt(selectedRound, 10), string(key))
						return bootstrapBytes, nil
					}
				},
				SearchFirstCalled: func(key []byte) ([]byte, error) {
					if string(key) == common.NodesCoordinatorRegistryKeyPrefix+"registry" {
						return registryBytes, nil
					}
					require.Equal(t, core.EpochStartIdentifier(8), string(key))
					return metaBytes, nil
				},
			}
			args.LatestStorageDataProvider, err = latestData.NewLatestDataProvider(latestData.ArgsLatestDataProvider{
				GeneralConfig: args.GeneralConfig, ParentDir: "db", DefaultEpochString: "Epoch", DefaultShardString: "Shard",
				DirectoryReader: &storageMock.DirectoryReaderStub{
					ListDirectoriesAsStringCalled: func(path string) ([]string, error) {
						if path == "db" {
							return []string{"Epoch_8"}, nil
						}
						return []string{"Shard_0"}, nil
					},
				},
				BootstrapDataProvider: &storageMock.BootStrapDataProviderStub{
					LoadForPathCalled: func(_ storage.PersisterFactory, path string) (*bootstrapStorage.BootstrapData, storage.Storer, error) {
						require.Contains(t, path, filepath.Join("Epoch_8", "Shard_0"))
						return &bootstrapData, storer, nil
					},
				},
			})
			require.NoError(t, err)
			args.StorageUnitOpener = &storageStubs.UnitOpenerStub{
				GetMostRecentStorageUnitCalled: func(_ config.DBConfig) (storage.Storer, error) { return storer, nil },
			}
			provider, err := NewEpochStartBootstrap(args)
			require.NoError(t, err)
			_, err = args.LatestStorageDataProvider.Get()
			require.NoError(t, err)
			provider.initializeFromLocalStorage()
			require.True(t, provider.baseData.storageExists)
			params, err := provider.prepareEpochFromStorage()
			require.NoError(t, err)
			require.Equal(t, uint32(8), params.Epoch)
			require.Equal(t, selectedRound, provider.baseData.lastRound)
			require.Equal(t, tc.headerEpoch, bootstrapData.LastHeader.Epoch)
			require.Equal(t, approvedHash, bootstrapData.LastHeader.Hash)
		})
	}
}

func TestRecoverySnapshotRestartWithRealStorage(t *testing.T) {
	for _, round := range []int64{99, 100, 200} {
		t.Run(strconv.FormatInt(round, 10), func(t *testing.T) {
			parentDir := t.TempDir()
			coreComp, cryptoComp := createComponentsForEpochStart()
			coreComp.IntMarsh = &marshal.GogoProtoMarshalizer{}
			args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
			setRecoveryCheckpointConfig(&args.GeneralConfig)
			dbConfig := config.DBConfig{
				FilePath: "BootstrapData", Type: "LvlDBSerial", BatchDelaySeconds: 1, MaxBatchSize: 10, MaxOpenFiles: 10,
			}
			args.GeneralConfig.BootstrapStorage.DB = dbConfig
			hash, err := hex.DecodeString(args.GeneralConfig.HardforkRecoveryCheckpoint.Headers[0].Hash)
			require.NoError(t, err)
			snapshot := bootstrapStorage.BootstrapData{
				LastHeader:             bootstrapStorage.BootstrapHeaderInfo{ShardId: 0, Epoch: 7, Nonce: 10, Hash: hash},
				HighestFinalBlockNonce: 10, LastRound: 0,
				NodesCoordinatorConfigKey: []byte("registry"), EpochStartTriggerConfigKey: []byte("trigger"),
			}
			marshaller := coreComp.InternalMarshalizer()
			bootstrapBytes, err := marshaller.Marshal(&snapshot)
			require.NoError(t, err)
			roundBytes, err := marshaller.Marshal(&bootstrapStorage.RoundNum{Num: round})
			require.NoError(t, err)
			triggerBytes, err := marshaller.Marshal(&block.ShardTriggerRegistryV3{
				EpochStartRound: 100, EpochStartShardHeader: &block.HeaderV3{Epoch: 8, Round: 100},
			})
			require.NoError(t, err)
			metaBytes, err := marshaller.Marshal(&block.MetaBlock{Epoch: 8})
			require.NoError(t, err)
			registryBytes, err := json.Marshal(&nodesCoordinator.NodesCoordinatorRegistry{})
			require.NoError(t, err)
			persisterFactory, err := storageFactory.NewPersisterFactory(dbConfig)
			require.NoError(t, err)
			persister, err := persisterFactory.Create(filepath.Join(parentDir, "Epoch_8", "Shard_0", dbConfig.FilePath))
			require.NoError(t, err)
			for key, value := range map[string][]byte{
				common.HighestRoundFromBootStorage:                    roundBytes,
				strconv.FormatInt(round, 10):                          bootstrapBytes,
				common.TriggerRegistryKeyPrefix + "trigger":           triggerBytes,
				common.NodesCoordinatorRegistryKeyPrefix + "registry": registryBytes,
				core.EpochStartIdentifier(8):                          metaBytes,
			} {
				require.NoError(t, persister.Put([]byte(key), value))
			}
			require.NoError(t, persister.Close())
			bootstrapProvider, err := storageFactory.NewBootstrapDataProvider(marshaller)
			require.NoError(t, err)
			loaded, opened, err := bootstrapProvider.LoadForPath(persisterFactory, filepath.Join(parentDir, "Epoch_8", "Shard_0", dbConfig.FilePath))
			require.NoError(t, err)
			require.Equal(t, snapshot, *loaded)
			storedTrigger, err := opened.Get([]byte(common.TriggerRegistryKeyPrefix + "trigger"))
			require.NoError(t, err)
			_, err = epochStart.UnmarshalShardTrigger(marshaller, storedTrigger)
			require.NoError(t, err)
			require.NoError(t, opened.Close())
			args.LatestStorageDataProvider, err = latestData.NewLatestDataProvider(latestData.ArgsLatestDataProvider{
				GeneralConfig: args.GeneralConfig, BootstrapDataProvider: bootstrapProvider,
				DirectoryReader: directoryhandler.NewDirectoryReader(), ParentDir: parentDir,
				DefaultEpochString: "Epoch", DefaultShardString: "Shard",
			})
			require.NoError(t, err)
			openerArgs := storageFactory.ArgsNewOpenStorageUnits{
				BootstrapDataProvider: bootstrapProvider, LatestStorageDataProvider: args.LatestStorageDataProvider,
				DefaultEpochString: "Epoch", DefaultShardString: "Shard",
			}
			ordinaryOpener, err := storageFactory.NewStorageUnitOpenHandler(openerArgs)
			require.NoError(t, err)
			_, err = ordinaryOpener.GetMostRecentStorageUnit(dbConfig)
			require.ErrorIs(t, err, storage.ErrBootstrapDataNotFoundInStorage)
			openerArgs.RecoveryCheckpointEnabled = true
			args.StorageUnitOpener, err = storageFactory.NewStorageUnitOpenHandler(openerArgs)
			require.NoError(t, err)
			_, err = args.LatestStorageDataProvider.Get()
			require.NoError(t, err)
			provider, err := NewEpochStartBootstrap(args)
			require.NoError(t, err)
			provider.initializeFromLocalStorage()
			require.True(t, provider.baseData.storageExists)
			require.Equal(t, uint32(8), provider.baseData.lastEpoch)
			highestRound, err := provider.getHighestStoredRound()
			require.NoError(t, err)
			require.Equal(t, round, highestRound)
			params, err := provider.prepareEpochFromStorage()
			require.NoError(t, err)
			require.Equal(t, uint32(8), params.Epoch)
			require.Equal(t, uint32(0), params.SelfShardId)
			require.Equal(t, round, provider.baseData.lastRound)
			if round == int64(args.GeneralConfig.HardforkRecoveryCheckpoint.Round) {
				restarted, createErr := NewEpochStartBootstrap(args)
				require.NoError(t, createErr)
				params, err = restarted.Bootstrap()
				require.NoError(t, err)
				require.Equal(t, uint32(8), params.Epoch)
			}
		})
	}
}

func TestGetHighestStoredRoundUsesBootstrapIndex(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	provider, err := NewEpochStartBootstrap(args)
	require.NoError(t, err)
	provider.baseData.lastRound = 99

	roundBytes, err := json.Marshal(&bootstrapStorage.RoundNum{Num: 100})
	require.NoError(t, err)
	provider.storageOpenerHandler = &storageStubs.UnitOpenerStub{
		GetMostRecentStorageUnitCalled: func(_ config.DBConfig) (storage.Storer, error) {
			return &storageStubs.StorerStub{GetCalled: func(key []byte) ([]byte, error) {
				require.Equal(t, []byte(common.HighestRoundFromBootStorage), key)
				return roundBytes, nil
			}}, nil
		},
	}

	highestRound, err := provider.getHighestStoredRound()
	require.NoError(t, err)
	require.Equal(t, int64(100), highestRound)
	require.Equal(t, int64(99), provider.baseData.lastRound)
}

func TestRecoveryEpochStartLookupDoesNotFallBack(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.GeneralConfig.HardforkRecoveryCheckpoint.Enabled = true
	args.GeneralConfig.HardforkRecoveryCheckpoint.Round = 100
	provider, err := NewEpochStartBootstrap(args)
	require.NoError(t, err)
	provider.baseData.lastEpoch = 7
	provider.baseData.lastRound = 100
	require.True(t, provider.isRecoveryCheckpointSelected())
	lookups := 0
	storer := &storageStubs.StorerStub{SearchFirstCalled: func(_ []byte) ([]byte, error) {
		lookups++
		return nil, storage.ErrKeyNotFound
	}}
	_, err = provider.getEpochStartMetaFromStorage(storer)
	require.ErrorIs(t, err, storage.ErrKeyNotFound)
	require.Equal(t, 1, lookups)
}

func TestRecoveryEpochStartLookupAllowsPostExclusionFallback(t *testing.T) {
	coreComp, cryptoComp := createComponentsForEpochStart()
	args := createMockEpochStartBootstrapArgs(coreComp, cryptoComp)
	args.GeneralConfig.HardforkRecoveryCheckpoint.Enabled = true
	args.GeneralConfig.HardforkRecoveryCheckpoint.Round = 100
	provider, err := NewEpochStartBootstrap(args)
	require.NoError(t, err)
	provider.baseData.lastEpoch = 7
	provider.baseData.lastRound = 200
	require.False(t, provider.isRecoveryCheckpointSelected())

	meta := &block.MetaBlock{Epoch: 6}
	metaBytes, err := json.Marshal(meta)
	require.NoError(t, err)
	lookups := 0
	storer := &storageStubs.StorerStub{SearchFirstCalled: func(_ []byte) ([]byte, error) {
		lookups++
		if lookups == 1 {
			return nil, storage.ErrKeyNotFound
		}
		return metaBytes, nil
	}}
	got, err := provider.getEpochStartMetaFromStorage(storer)
	require.NoError(t, err)
	require.Equal(t, meta, got)
	require.Equal(t, 2, lookups)
	require.Equal(t, uint32(6), provider.baseData.lastEpoch)
}

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
	// only the node type distinguishes "found in my own shard" from "not found"
	assert.Equal(t, core.NodeTypeValidator, epochStartProvider.nodeType)
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
	assert.Equal(t, core.NodeTypeValidator, epochStartProvider.nodeType)
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
	assert.Equal(t, core.NodeTypeValidator, epochStartProvider.nodeType)
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
	assert.Equal(t, core.NodeTypeObserver, epochStartProvider.nodeType)
}

func TestStartFromSavedEpoch_ShuffledOutStaleEpochStartJoinsFromNetwork(t *testing.T) {
	epochStartRound := uint64(1000)
	roundsPerEpoch := int64(200)

	const storedEpoch = uint32(10)
	publicKey := []byte("pubKey")

	// lets getShardIDForLatestEpoch run: eligible in shard 0 while stored as metachain
	createStorer := func(shuffledOut bool) storage.Storer {
		round := int64(10)
		roundBytes, _ := json.Marshal(&bootstrapStorage.RoundNum{Num: round})
		bootstrapData := bootstrapStorage.BootstrapData{NodesCoordinatorConfigKey: []byte("key")}
		bootstrapDataBytes, _ := json.Marshal(bootstrapData)
		nodesCoordinatorKey := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), bootstrapData.NodesCoordinatorConfigKey...)

		eligibleShard := "0"
		if !shuffledOut {
			eligibleShard = fmt.Sprint(core.MetachainShardId)
		}
		registryBytes, _ := json.Marshal(&nodesCoordinator.NodesCoordinatorRegistry{
			EpochsConfig: map[string]*nodesCoordinator.EpochValidators{
				fmt.Sprint(storedEpoch): {
					EligibleValidators: map[string][]*nodesCoordinator.SerializableValidator{
						eligibleShard: {{PubKey: publicKey, Chances: 0, Index: 0}},
					},
				},
			},
		})
		metaBytes, _ := json.Marshal(&block.MetaBlock{Nonce: 1, Epoch: storedEpoch})

		return &storageStubs.StorerStub{
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
				case bytes.Equal([]byte(core.EpochStartIdentifier(storedEpoch)), key):
					return metaBytes, nil
				default:
					return nil, errors.New("unexpected key")
				}
			},
		}
	}

	createProvider := func(shuffledOut bool, currentRound int64, numStorageOpens *int) *epochStartBootstrap {
		coreComp, cryptoComp := createComponentsForEpochStart()
		coreComp.ChainParametersHandlerField = &chainParameters.ChainParametersHandlerStub{
			CurrentChainParametersCalled: func() config.ChainParametersByEpochConfig {
				return config.ChainParametersByEpochConfig{RoundsPerEpoch: roundsPerEpoch}
			},
			ChainParametersForEpochCalled: func(epoch uint32) (config.ChainParametersByEpochConfig, error) {
				return config.ChainParametersByEpochConfig{RoundsPerEpoch: roundsPerEpoch}, nil
			},
		}
		cryptoComp.PubKey = &cryptoMocks.PublicKeyStub{
			ToByteArrayStub: func() ([]byte, error) {
				return publicKey, nil
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
				return storage.LatestDataFromStorage{Epoch: storedEpoch, ShardID: core.MetachainShardId, LastRound: int64(epochStartRound) + 5, EpochStartRound: epochStartRound}, nil
			},
		}
		storer := createStorer(shuffledOut)
		args.StorageUnitOpener = &storageStubs.UnitOpenerStub{
			GetMostRecentStorageUnitCalled: func(config config.DBConfig) (storage.Storer, error) {
				*numStorageOpens++
				return storer, nil
			},
		}
		epochStartProvider, err := NewEpochStartBootstrap(args)
		require.Nil(t, err)

		return epochStartProvider
	}

	// the shuffle check itself reads the bootstrap storage, so a second open means prepareEpochFromStorage ran
	t.Run("shuffled out with stale epoch start skips storage and continues from the network", func(t *testing.T) {
		numStorageOpens := 0
		staleRound := int64(epochStartRound) + roundsPerEpoch + 10
		epochStartProvider := createProvider(true, staleRound, &numStorageOpens)

		params, shouldContinue, err := epochStartProvider.startFromSavedEpoch()
		require.Nil(t, err)
		require.True(t, shouldContinue)
		require.Equal(t, Parameters{}, params)
		require.True(t, epochStartProvider.shuffledOut, "the shuffle must be detected from storage, not injected")
		require.Equal(t, 1, numStorageOpens, "prepareEpochFromStorage must not run on a stale epoch start")
	})

	t.Run("shuffled out with current epoch start keeps the storage path", func(t *testing.T) {
		numStorageOpens := 0
		freshRound := int64(epochStartRound) + roundsPerEpoch/2
		epochStartProvider := createProvider(true, freshRound, &numStorageOpens)

		_, shouldContinue, err := epochStartProvider.startFromSavedEpoch()
		require.NotNil(t, err)
		require.True(t, epochStartProvider.shuffledOut)
		require.False(t, shouldContinue, "shuffled out storage attempt must not fall through to the network")
		require.Equal(t, 2, numStorageOpens, "prepareEpochFromStorage should have been attempted")
	})

	// a stale epoch start must not divert a node that was not shuffled out
	t.Run("not shuffled out takes the storage path even with a stale epoch start", func(t *testing.T) {
		numStorageOpens := 0
		staleRound := int64(epochStartRound) + roundsPerEpoch + 10
		epochStartProvider := createProvider(false, staleRound, &numStorageOpens)

		params, shouldContinue, err := epochStartProvider.startFromSavedEpoch()
		require.Nil(t, err)
		require.False(t, shouldContinue)
		require.False(t, epochStartProvider.shuffledOut)
		require.Equal(t, storedEpoch, params.Epoch)
		require.Equal(t, 2, numStorageOpens, "prepareEpochFromStorage should have been attempted")
	})
}
