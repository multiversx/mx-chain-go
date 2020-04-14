package factory

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/epochStart/shardchain"
	"github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

func TestGetCacherFromConfig(t *testing.T) {
	t.Parallel()

	cfg := config.CacheConfig{
		Size:        100,
		Shards:      2,
		Type:        "lru",
		SizeInBytes: 128,
	}

	storageCacheConfig := GetCacherFromConfig(cfg)
	assert.Equal(t, storageUnit.CacheConfig{
		Size:        cfg.Size,
		SizeInBytes: cfg.SizeInBytes,
		Type:        storageUnit.CacheType(cfg.Type),
		Shards:      cfg.Shards,
	}, storageCacheConfig)
}

func TestGetDBFromConfig(t *testing.T) {
	t.Parallel()

	cfg := config.DBConfig{
		Type:              "lru",
		MaxBatchSize:      10,
		BatchDelaySeconds: 2,
		MaxOpenFiles:      20,
	}

	storageDBConfig := GetDBFromConfig(cfg)
	assert.Equal(t, storageUnit.DBConfig{
		Type:              storageUnit.DBType(cfg.Type),
		MaxBatchSize:      cfg.MaxBatchSize,
		BatchDelaySeconds: cfg.BatchDelaySeconds,
		MaxOpenFiles:      cfg.MaxOpenFiles,
	}, storageDBConfig)
}

func TestGetBloomFromConfig(t *testing.T) {
	t.Parallel()

	cfg := config.BloomFilterConfig{
		Size:     100,
		HashFunc: []string{"hashFunc"},
	}

	storageBloomConfig := GetBloomFromConfig(cfg)
	assert.Equal(t, storageUnit.BloomConfig{
		HashFunc: []storageUnit.HasherType{storageUnit.HasherType(cfg.HashFunc[0])},
		Size:     cfg.Size,
	}, storageBloomConfig)
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

	round, err := loadEpochStartRound(shardID, key, storer)
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

	round, err := loadEpochStartRound(shardID, key, storer)
	assert.NoError(t, err)
	assert.Equal(t, startRound, round)
}

func TestConvertShardIDToUint32(t *testing.T) {
	t.Parallel()

	shardID, err := convertShardIDToUint32("metachain")
	assert.NoError(t, err)
	assert.Equal(t, core.MetachainShardId, shardID)

	id := uint32(0)
	shardIDStr := fmt.Sprintf("%d", id)
	shardID, err = convertShardIDToUint32(shardIDStr)
	assert.NoError(t, err)
	assert.Equal(t, id, shardID)

	shardID, err = convertShardIDToUint32("wrongID")
	assert.Error(t, err)
	assert.Equal(t, uint32(0), shardID)
}
