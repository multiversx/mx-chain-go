package immunitycache

import (
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const numChunksLowerBound = 1
const numChunksUpperBound = 128
const maxNumItemsLowerBound = 4
const maxNumBytesLowerBound = maxNumItemsLowerBound * 1
const maxNumBytesUpperBound = 1_073_741_824 // one GB
const numItemsToPreemptivelyEvictLowerBound = 1

// CacheConfig holds cache configuration
type CacheConfig struct {
	Name                        string
	NumChunks                   uint32
	MaxNumItems                 uint32
	MaxNumBytes                 uint32
	NumItemsToPreemptivelyEvict uint32
}

// Verify verifies the validity of the configuration
func (config *CacheConfig) Verify() error {
	if len(config.Name) == 0 {
		return fmt.Errorf("%w: config.Name is invalid", storage.ErrInvalidConfig)
	}
	if config.NumChunks < numChunksLowerBound || config.NumChunks > numChunksUpperBound {
		return fmt.Errorf("%w: config.NumChunks is invalid", storage.ErrInvalidConfig)
	}
	if config.MaxNumItems < maxNumItemsLowerBound {
		return fmt.Errorf("%w: config.MaxNumItems is invalid", storage.ErrInvalidConfig)
	}
	if config.MaxNumBytes < maxNumBytesLowerBound || config.MaxNumBytes > maxNumBytesUpperBound {
		return fmt.Errorf("%w: config.MaxNumBytes is invalid", storage.ErrInvalidConfig)
	}
	if config.NumItemsToPreemptivelyEvict < numItemsToPreemptivelyEvictLowerBound {
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
