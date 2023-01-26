package factory

import (
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
)

// GetCacherFromConfig will return the cache config needed for storage unit from a config came from the toml file
func GetCacherFromConfig(cfg config.CacheConfig) storageunit.CacheConfig {
	return storageunit.CacheConfig{
		Name:                 cfg.Name,
		Capacity:             cfg.Capacity,
		SizePerSender:        cfg.SizePerSender,
		SizeInBytes:          cfg.SizeInBytes,
		SizeInBytesPerSender: cfg.SizeInBytesPerSender,
		Type:                 storageunit.CacheType(cfg.Type),
		Shards:               cfg.Shards,
	}
}

// GetDBFromConfig will return the db config needed for storage unit from a config came from the toml file
func GetDBFromConfig(cfg config.DBConfig) storageunit.DBConfig {
	return storageunit.DBConfig{
		Type:              storageunit.DBType(cfg.Type),
		MaxBatchSize:      cfg.MaxBatchSize,
		BatchDelaySeconds: cfg.BatchDelaySeconds,
		MaxOpenFiles:      cfg.MaxOpenFiles,
	}
}
