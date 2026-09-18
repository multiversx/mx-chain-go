package factory

import (
	"github.com/multiversx/mx-chain-storage-go/pebbledb"

	"github.com/multiversx/mx-chain-go/config"
)

const megabyte = 1 << 20

// NewPebbleResources creates the caches shared by the pebble persisters, or nil when sharing is disabled
func NewPebbleResources(cfg config.StorageEngineConfig) (*pebbledb.SharedResources, error) {
	if cfg.SharedCacheSizeMB == 0 {
		return nil, nil
	}

	return pebbledb.NewSharedResources(int64(cfg.SharedCacheSizeMB)*megabyte, int(cfg.SharedFileCacheSize))
}
