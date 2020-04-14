package factory

import (
	"encoding/json"
	"strconv"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/epochStart/metachain"
	"github.com/ElrondNetwork/elrond-go/epochStart/shardchain"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

const allFiles = -1

// GetCacherFromConfig will return the cache config needed for storage unit from a config came from the toml file
func GetCacherFromConfig(cfg config.CacheConfig) storageUnit.CacheConfig {
	return storageUnit.CacheConfig{
		Size:        cfg.Size,
		SizeInBytes: cfg.SizeInBytes,
		Type:        storageUnit.CacheType(cfg.Type),
		Shards:      cfg.Shards,
	}
}

// GetDBFromConfig will return the db config needed for storage unit from a config came from the toml file
func GetDBFromConfig(cfg config.DBConfig) storageUnit.DBConfig {
	return storageUnit.DBConfig{
		Type:              storageUnit.DBType(cfg.Type),
		MaxBatchSize:      cfg.MaxBatchSize,
		BatchDelaySeconds: cfg.BatchDelaySeconds,
		MaxOpenFiles:      cfg.MaxOpenFiles,
	}
}

// GetBloomFromConfig will return the bloom config needed for storage unit from a config came from the toml file
func GetBloomFromConfig(cfg config.BloomFilterConfig) storageUnit.BloomConfig {
	var hashFuncs []storageUnit.HasherType
	if cfg.HashFunc != nil {
		hashFuncs = make([]storageUnit.HasherType, len(cfg.HashFunc))
		for idx, hf := range cfg.HashFunc {
			hashFuncs[idx] = storageUnit.HasherType(hf)
		}
	}

	return storageUnit.BloomConfig{
		Size:     cfg.Size,
		HashFunc: hashFuncs,
	}
}

// loadEpochStartRound will return the epoch start round from the bootstrap unit
func loadEpochStartRound(
	shardID uint32,
	key []byte,
	storer storage.Storer,
) (uint64, error) {
	trigInternalKey := append([]byte(core.TriggerRegistryKeyPrefix), key...)
	data, err := storer.Get(trigInternalKey)
	if err != nil {
		return 0, err
	}

	if shardID == core.MetachainShardId {
		state := &metachain.TriggerRegistry{}
		err = json.Unmarshal(data, state)
		if err != nil {
			return 0, err
		}

		return state.CurrEpochStartRound, nil
	}

	state := &shardchain.TriggerRegistry{}
	err = json.Unmarshal(data, state)
	if err != nil {
		return 0, err
	}

	return state.EpochStartRound, nil
}

func convertShardIDToUint32(shardIDStr string) (uint32, error) {
	if shardIDStr == "metachain" {
		return core.MetachainShardId, nil
	}

	shardID, err := strconv.ParseInt(shardIDStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint32(shardID), nil
}
