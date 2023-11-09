package factory

import (
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/stretchr/testify/assert"
)

func TestGetCacherFromConfig(t *testing.T) {
	t.Parallel()

	cfg := config.CacheConfig{
		Capacity:    100,
		Shards:      2,
		Type:        "lru",
		SizeInBytes: 128,
	}

	storageCacheConfig := GetCacherFromConfig(cfg)
	assert.Equal(t, storageunit.CacheConfig{
		Capacity:    cfg.Capacity,
		SizeInBytes: cfg.SizeInBytes,
		Type:        storageunit.CacheType(cfg.Type),
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
	assert.Equal(t, storageunit.DBConfig{
		Type:              storageunit.DBType(cfg.Type),
		MaxBatchSize:      cfg.MaxBatchSize,
		BatchDelaySeconds: cfg.BatchDelaySeconds,
		MaxOpenFiles:      cfg.MaxOpenFiles,
	}, storageDBConfig)
}
