package latestData

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/mock"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLatestDataProvider_ShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		ldp, err := NewLatestDataProvider(getLatestDataProviderArgs())
		require.NotNil(t, ldp)
		require.NoError(t, err)
	})
	t.Run("nil DirectoryReader should error", func(t *testing.T) {
		t.Parallel()

		args := getLatestDataProviderArgs()
		args.DirectoryReader = nil
		ldp, err := NewLatestDataProvider(args)
		require.Nil(t, ldp)
		require.Equal(t, storage.ErrNilDirectoryReader, err)
	})
	t.Run("nil BootstrapDataProvider should error", func(t *testing.T) {
		t.Parallel()

		args := getLatestDataProviderArgs()
		args.BootstrapDataProvider = nil
		ldp, err := NewLatestDataProvider(args)
		require.Nil(t, ldp)
		require.Equal(t, storage.ErrNilBootstrapDataProvider, err)
	})
}

func TestLatestDataProvider_GetParentDirectory(t *testing.T) {
	t.Parallel()

	args := getLatestDataProviderArgs()
	ldp, _ := NewLatestDataProvider(args)
	require.Equal(t, args.ParentDir, ldp.GetParentDirectory())
}

func TestGetShardsFromDirectory(t *testing.T) {
	t.Parallel()

	path := "testPath"
	shards := []string{"0", "1"}
	lastDirectories := []string{"WrongShard", "Shard", fmt.Sprintf("Shard_%s", shards[0]), fmt.Sprintf("Shard_%s", shards[1])}
	args := getLatestDataProviderArgs()
	args.DirectoryReader = &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(directoryPath string) ([]string, error) {
			if directoryPath == path {
				return lastDirectories, nil
			}
			return nil, nil
		},
	}
	ldp, _ := NewLatestDataProvider(args)

	result, err := ldp.GetShardsFromDirectory(path)
	assert.NoError(t, err)
	assert.Equal(t, shards, result)
}

func TestLatestDataProvider_GetCannotGetListDirectoriesShouldErr(t *testing.T) {
	t.Parallel()

	localErr := errors.New("localErr")
	args := getLatestDataProviderArgs()
	args.DirectoryReader = &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(directoryPath string) ([]string, error) {
			return nil, localErr
		},
	}
	ldp, _ := NewLatestDataProvider(args)

	_, err := ldp.Get()
	assert.Equal(t, localErr, err)
}

func TestLatestDataProvider_Get(t *testing.T) {
	t.Parallel()

	shardID := uint32(0)
	defaultPath := "db"
	args := getLatestDataProviderArgs()
	args.ParentDir = defaultPath
	lastEpoch := uint32(1)
	args.DirectoryReader = &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(directoryPath string) ([]string, error) {
			if directoryPath == defaultPath {
				return []string{fmt.Sprintf("Epoch_%d", lastEpoch)}, nil
			}
			return []string{fmt.Sprintf("Shard_%d", shardID)}, nil
		},
	}

	startRound, lastRound := uint64(5), int64(10)
	state := &block.ShardTriggerRegistry{
		EpochStartRound: startRound,
	}
	storer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			stateBytes, _ := json.Marshal(state)
			return stateBytes, nil
		},
	}

	args.BootstrapDataProvider = &mock.BootStrapDataProviderStub{
		LoadForPathCalled: func(persisterFactory storage.PersisterFactory, path string) (*bootstrapStorage.BootstrapData, storage.Storer, error) {
			bootstrapData := &bootstrapStorage.BootstrapData{
				LastRound:  lastRound,
				LastHeader: bootstrapStorage.BootstrapHeaderInfo{Epoch: lastEpoch},
			}

			return bootstrapData, storer, nil
		},
	}

	ldp, _ := NewLatestDataProvider(args)

	expectedRes := storage.LatestDataFromStorage{
		Epoch:           lastEpoch,
		ShardID:         shardID,
		LastRound:       lastRound,
		EpochStartRound: startRound,
	}
	result, err := ldp.Get()
	assert.NoError(t, err)
	assert.Equal(t, expectedRes, result)
}

func getLatestDataProviderArgs() ArgsLatestDataProvider {
	return ArgsLatestDataProvider{
		GeneralConfig:         config.Config{},
		BootstrapDataProvider: &mock.BootStrapDataProviderStub{},
		DirectoryReader:       &mock.DirectoryReaderStub{},
		ParentDir:             "db",
		DefaultEpochString:    "Epoch",
		DefaultShardString:    "Shard",
	}
}

func TestLatestDataProvider_RecoveryStorageEpoch(t *testing.T) {
	for _, tc := range []struct {
		name       string
		enabled    bool
		lastRound  int64
		finalNonce uint64
		wantEpoch  uint32
		wantError  bool
	}{
		{name: "snapshot", enabled: true, finalNonce: 10, wantEpoch: 8},
		{name: "committed header", enabled: true, lastRound: 99, finalNonce: 9, wantEpoch: 8},
		{name: "disabled keeps header epoch", lastRound: 99, finalNonce: 9, wantEpoch: 7},
		{name: "disabled keeps snapshot selection unchanged", finalNonce: 10, wantError: true},
		{name: "unfinished record is not a snapshot", enabled: true, finalNonce: 9, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := getLatestDataProviderArgs()
			args.GeneralConfig.HardforkRecoveryCheckpoint.Enabled = tc.enabled
			args.DirectoryReader = &mock.DirectoryReaderStub{
				ListDirectoriesAsStringCalled: func(path string) ([]string, error) {
					if path == args.ParentDir {
						return []string{"Epoch_8"}, nil
					}
					return []string{"Shard_0"}, nil
				},
			}
			trigger, err := json.Marshal(&block.ShardTriggerRegistry{EpochStartRound: 100})
			require.NoError(t, err)
			args.BootstrapDataProvider = &mock.BootStrapDataProviderStub{
				LoadForPathCalled: func(_ storage.PersisterFactory, path string) (*bootstrapStorage.BootstrapData, storage.Storer, error) {
					require.Contains(t, path, filepath.Join("Epoch_8", "Shard_0"))
					return &bootstrapStorage.BootstrapData{
						LastHeader: bootstrapStorage.BootstrapHeaderInfo{Epoch: 7, Nonce: 10, Hash: []byte("anchor")},
						LastRound:  tc.lastRound, HighestFinalBlockNonce: tc.finalNonce,
					}, &storageStubs.StorerStub{GetCalled: func(_ []byte) ([]byte, error) { return trigger, nil }}, nil
				},
			}
			provider, err := NewLatestDataProvider(args)
			require.NoError(t, err)
			data, err := provider.Get()
			if tc.wantError {
				require.ErrorIs(t, err, storage.ErrBootstrapDataNotFoundInStorage)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.wantEpoch, data.Epoch)
			require.Equal(t, tc.lastRound, data.LastRound)
			_, storageEpoch, err := provider.GetParentDirAndLastEpoch()
			require.NoError(t, err)
			require.Equal(t, uint32(8), storageEpoch)
		})
	}
}

func TestLoadEpochStartRoundShard(t *testing.T) {
	t.Parallel()

	key := []byte("123")
	shardID := uint32(0)
	startRound := uint64(100)
	state := &block.ShardTriggerRegistry{
		EpochStartRound: startRound,
	}
	storer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			stateBytes, _ := json.Marshal(state)
			return stateBytes, nil
		},
	}

	args := getLatestDataProviderArgs()
	ldp, _ := NewLatestDataProvider(args)

	round, err := ldp.loadEpochStartRound(shardID, key, storer)
	assert.NoError(t, err)
	assert.Equal(t, startRound, round)
}

func TestLoadEpochStartRoundMetachain(t *testing.T) {
	t.Parallel()

	key := []byte("123")
	shardID := core.MetachainShardId
	startRound := uint64(1000)
	state := &block.MetaTriggerRegistry{
		CurrEpochStartRound: startRound,
	}
	marshaller := &marshal.GogoProtoMarshalizer{}
	storer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			stateBytes, _ := marshaller.Marshal(state)
			return stateBytes, nil
		},
	}

	args := getLatestDataProviderArgs()
	ldp, _ := NewLatestDataProvider(args)

	round, err := ldp.loadEpochStartRound(shardID, key, storer)
	assert.NoError(t, err)
	assert.Equal(t, startRound, round)
}

func TestLatestDataProvider_ShouldWork(t *testing.T) {
	t.Parallel()

	workingDir := "testDir"
	defaultDbPath := "default"
	chainID := "chainID"
	lastEpoch := uint32(2)
	lastDirectories := []string{"WrongEpoch_10", "Epoch_1", fmt.Sprintf("Epoch_%d", lastEpoch), "Shard_1"}

	args := getLatestDataProviderArgs()
	args.ParentDir = filepath.Join(workingDir, defaultDbPath, chainID)
	args.DirectoryReader = &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(directoryPath string) ([]string, error) {
			return lastDirectories, nil
		},
	}

	bd := &bootstrapStorage.BootstrapData{
		LastHeader:             bootstrapStorage.BootstrapHeaderInfo{},
		HighestFinalBlockNonce: 1,
		LastRound:              1,
	}

	storerStub := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			return json.Marshal(&block.ShardTriggerRegistry{
				EpochStartRound: 1,
			})
		},
	}

	args.BootstrapDataProvider = &mock.BootStrapDataProviderStub{
		LoadForPathCalled: func(persisterFactory storage.PersisterFactory, path string) (
			*bootstrapStorage.BootstrapData, storage.Storer, error) {
			return bd, storerStub, nil
		}}

	ldp, _ := NewLatestDataProvider(args)

	parentDir, epoch, err := ldp.GetParentDirAndLastEpoch()
	assert.NoError(t, err)
	assert.Equal(t, workingDir+"/"+defaultDbPath+"/"+chainID, parentDir)
	assert.Equal(t, lastEpoch, epoch)
}

func TestFullHistoryGetShardsFromDirectory(t *testing.T) {
	t.Parallel()

	path := "testPath"
	shards := []string{"0", "1"}
	lastDirectories := []string{"WrongShard", "Shard", fmt.Sprintf("Shard_%s", shards[0]), fmt.Sprintf("Shard_%s", shards[1])}
	args := getLatestDataProviderArgs()
	args.DirectoryReader = &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(directoryPath string) ([]string, error) {
			if directoryPath == path {
				return lastDirectories, nil
			}
			return nil, nil
		},
	}
	ldp, _ := NewLatestDataProvider(args)

	result, err := ldp.GetShardsFromDirectory(path)
	assert.NoError(t, err)
	assert.Equal(t, shards, result)
}

func TestFullHistoryLatestDataProvider_GetCannotGetListDirectoriesShouldErr(t *testing.T) {
	t.Parallel()

	localErr := errors.New("localErr")
	args := getLatestDataProviderArgs()
	args.DirectoryReader = &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(directoryPath string) ([]string, error) {
			return nil, localErr
		},
	}
	ldp, _ := NewLatestDataProvider(args)

	_, err := ldp.Get()
	assert.Equal(t, localErr, err)
}

func TestFullHistoryLatestDataProvider_Get(t *testing.T) {
	t.Parallel()

	shardID := uint32(0)
	defaultPath := "db"
	args := getLatestDataProviderArgs()
	args.ParentDir = defaultPath
	lastEpoch := uint32(1)
	args.DirectoryReader = &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(directoryPath string) ([]string, error) {
			if directoryPath == defaultPath {
				return []string{fmt.Sprintf("Epoch_%d", lastEpoch)}, nil
			}
			return []string{fmt.Sprintf("Shard_%d", shardID)}, nil
		},
	}

	startRound, lastRound := uint64(5), int64(10)
	state := &block.ShardTriggerRegistry{
		EpochStartRound:       startRound,
		EpochStartShardHeader: &block.Header{},
	}
	marshaller := &marshal.GogoProtoMarshalizer{}
	storer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			stateBytes, _ := marshaller.Marshal(state)
			return stateBytes, nil
		},
	}

	args.BootstrapDataProvider = &mock.BootStrapDataProviderStub{
		LoadForPathCalled: func(persisterFactory storage.PersisterFactory, path string) (*bootstrapStorage.BootstrapData, storage.Storer, error) {
			bootstrapData := &bootstrapStorage.BootstrapData{
				LastRound:  lastRound,
				LastHeader: bootstrapStorage.BootstrapHeaderInfo{Epoch: lastEpoch},
			}

			return bootstrapData, storer, nil
		},
	}

	ldp, _ := NewLatestDataProvider(args)

	expectedRes := storage.LatestDataFromStorage{
		Epoch:           lastEpoch,
		ShardID:         shardID,
		LastRound:       lastRound,
		EpochStartRound: startRound,
	}
	result, err := ldp.Get()
	assert.NoError(t, err)
	assert.Equal(t, expectedRes, result)
}

func TestFullHistoryLoadEpochStartRoundShard(t *testing.T) {
	t.Parallel()

	key := []byte("123")
	shardID := uint32(0)
	startRound := uint64(100)
	state := &block.ShardTriggerRegistry{
		EpochStartRound:       startRound,
		EpochStartShardHeader: &block.Header{},
	}
	marshaller := &marshal.GogoProtoMarshalizer{}
	storer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			stateBytes, _ := marshaller.Marshal(state)
			return stateBytes, nil
		},
	}

	args := getLatestDataProviderArgs()
	ldp, _ := NewLatestDataProvider(args)

	round, err := ldp.loadEpochStartRound(shardID, key, storer)
	assert.NoError(t, err)
	assert.Equal(t, startRound, round)
}

func TestFullHistoryLoadEpochStartRoundMetachain(t *testing.T) {
	t.Parallel()

	key := []byte("123")
	shardID := core.MetachainShardId
	startRound := uint64(1000)
	state := &block.MetaTriggerRegistry{
		CurrEpochStartRound: startRound,
	}

	marshaller := &marshal.GogoProtoMarshalizer{}
	storer := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			stateBytes, _ := marshaller.Marshal(state)
			return stateBytes, nil
		},
	}

	args := getLatestDataProviderArgs()
	ldp, _ := NewLatestDataProvider(args)

	round, err := ldp.loadEpochStartRound(shardID, key, storer)
	assert.NoError(t, err)
	assert.Equal(t, startRound, round)
}

func TestLatestDataProvider_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var ldp *latestDataProvider
	require.True(t, ldp.IsInterfaceNil())

	ldp, _ = NewLatestDataProvider(getLatestDataProviderArgs())
	require.False(t, ldp.IsInterfaceNil())
}
