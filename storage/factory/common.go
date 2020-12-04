package factory

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

const allFiles = -1

// GetCacherFromConfig will return the cache config needed for storage unit from a config came from the toml file
func GetCacherFromConfig(cfg config.CacheConfig) storageUnit.CacheConfig {
	return storageUnit.CacheConfig{
		Name:                 cfg.Name,
		Capacity:             cfg.Capacity,
		SizePerSender:        cfg.SizePerSender,
		SizeInBytes:          cfg.SizeInBytes,
		SizeInBytesPerSender: cfg.SizeInBytesPerSender,
		Type:                 storageUnit.CacheType(cfg.Type),
		Shards:               cfg.Shards,
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
