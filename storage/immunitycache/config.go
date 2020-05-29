package immunitycache

import (
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// CacheConfig holds cache configuration
type CacheConfig struct {
	Name                        string
	NumChunks                   uint32
	MaxNumItems                 uint32
	MaxNumBytes                 uint32
	NumItemsToPreemptivelyEvict uint32
}

func (config *CacheConfig) verify() error {
	if len(config.Name) == 0 {
		return fmt.Errorf("%w: config.Name is invalid", storage.ErrInvalidConfig)
	}
	if config.NumChunks == 0 {
		return fmt.Errorf("%w: config.NumChunks is invalid", storage.ErrInvalidConfig)
	}
	if config.MaxNumItems == 0 {
		return fmt.Errorf("%w: config.MaxNumItems is invalid", storage.ErrInvalidConfig)
	}
	if config.MaxNumBytes == 0 {
		return fmt.Errorf("%w: config.MaxNumBytes is invalid", storage.ErrInvalidConfig)
	}
	if config.NumItemsToPreemptivelyEvict == 0 {
		return fmt.Errorf("%w: config.NumItemsToPreemptivelyEvict is invalid", storage.ErrInvalidConfig)
	}

	return nil
}

func (config *CacheConfig) getChunkConfig() immunityChunkConfig {
	numChunks := core.MaxUint32(config.NumChunks, 1)

	return immunityChunkConfig{
		cacheName:                   config.Name,
		maxNumItems:                 config.MaxNumItems / numChunks,
		maxNumBytes:                 config.MaxNumBytes / numChunks,
		numItemsToPreemptivelyEvict: config.NumItemsToPreemptivelyEvict / numChunks,
	}
}

// String returns a readable representation of the object
func (config *CacheConfig) String() string {
	bytes, err := json.Marshal(config)
	if err != nil {
		log.Error("CacheConfig.String()", "err", err)
	}

	return string(bytes)
}

type immunityChunkConfig struct {
	cacheName                   string
	maxNumItems                 uint32
	maxNumBytes                 uint32
	numItemsToPreemptivelyEvict uint32
}

// String returns a readable representation of the object
func (config *immunityChunkConfig) String() string {
	return fmt.Sprintf(
		"maxNumItems: %d, maxNumBytes: %d, numItemsToPreemptivelyEvict: %d",
		config.maxNumItems,
		config.maxNumBytes,
		config.numItemsToPreemptivelyEvict,
	)
}
