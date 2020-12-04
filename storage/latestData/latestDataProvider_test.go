package latestData

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/epochStart/shardchain"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLatestDataProvider_ShouldWork(t *testing.T) {
	t.Parallel()

	ldp, err := NewLatestDataProvider(getLatestDataProviderArgs())
	require.False(t, check.IfNil(ldp))
	require.NoError(t, err)
}

func TestGetParentDirAndLastEpoch_ShouldWork(t *testing.T) {
	t.Parallel()

	workingDir := "testDir"
	defaultDbPath := "default"
	chainID := "chainID"
	lastEpoch := uint32(2)
	lastDirectories := []string{"WrongEpoch_10", "Epoch_1", fmt.Sprintf("Epoch_%d", lastEpoch)}

	args := getLatestDataProviderArgs()
	args.ChainID = chainID
	args.DefaultDBPath = defaultDbPath
	args.WorkingDir = workingDir
	args.DirectoryReader = &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(directoryPath string) ([]string, error) {
			return lastDirectories, nil
		},
	}

	ldp, _ := NewLatestDataProvider(args)

	parentDir, epoch, err := ldp.GetParentDirAndLastEpoch()
	assert.NoError(t, err)
	assert.Equal(t, workingDir+"/"+defaultDbPath+"/"+chainID, parentDir)
	assert.Equal(t, lastEpoch, epoch)
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
	args.DefaultDBPath = defaultPath
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
	state := &shardchain.TriggerRegistry{
		EpochStartRound: startRound,
	}
	storer := &mock.StorerStub{
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
		Marshalizer:           &mock.MarshalizerMock{},
		Hasher:                &mock.HasherMock{},
		BootstrapDataProvider: &mock.BootStrapDataProviderStub{},
		DirectoryReader:       &mock.DirectoryReaderStub{},
		WorkingDir:            "",
		ChainID:               "",
		DefaultDBPath:         "db",
		DefaultEpochString:    "Epoch",
		DefaultShardString:    "Shard",
	}
}

func TestLoadEpochStartRoundShard(t *testing.T) {
	t.Parallel()

	key := []byte("123")
	shardID := uint32(0)
	startRound := uint64(100)
	state := &shardchain.TriggerRegistry{
		EpochStartRound: startRound,
	}
	storer := &mock.StorerStub{
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
	state := &metachain.TriggerRegistry{
		CurrEpochStartRound: startRound,
	}
	storer := &mock.StorerStub{
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
