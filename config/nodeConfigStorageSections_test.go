package config_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
)

// guards the section boundaries of the shipped config: a section inserted inside another silently drops keys
func TestNodeConfigStorageSections(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	err := core.LoadTomlFile(&cfg, "../cmd/node/config/config.toml")
	require.Nil(t, err)

	require.Equal(t, uint64(3), cfg.StoragePruning.NumActivePersisters)
	require.Equal(t, uint32(3), cfg.StoragePruning.AssumedPeersNumActivePersisters)
	require.Equal(t, uint32(10), cfg.StoragePruning.FullArchiveNumActivePersisters)
	require.Equal(t, uint32(256), cfg.StorageEngine.SharedCacheSizeMB)
	require.Equal(t, uint32(4096), cfg.StorageEngine.SharedFileCacheSize)
}
