package factory_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-storage-go/common"
)

func TestNewPebbleResources(t *testing.T) {
	t.Parallel()

	t.Run("sharing disabled should return nil", func(t *testing.T) {
		t.Parallel()

		resources, err := factory.NewPebbleResources(config.StorageEngineConfig{})
		require.Nil(t, err)
		require.Nil(t, resources)
	})
	t.Run("invalid file cache size should error", func(t *testing.T) {
		t.Parallel()

		resources, err := factory.NewPebbleResources(config.StorageEngineConfig{SharedCacheSizeMB: 1})
		require.ErrorIs(t, err, common.ErrInvalidConfig)
		require.Nil(t, resources)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		resources, err := factory.NewPebbleResources(config.StorageEngineConfig{SharedCacheSizeMB: 1, SharedFileCacheSize: 10})
		require.Nil(t, err)
		require.NotNil(t, resources)
		resources.Close()
	})
}
