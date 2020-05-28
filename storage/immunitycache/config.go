package immunitycache

import (
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
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
		return fmt.Errorf("%w: config.Name is invalid", errInvalidConfig)
	}
	if config.NumChunks == 0 {
		return fmt.Errorf("%w: config.NumChunks is invalid", errInvalidConfig)
	}
	if config.MaxNumItems == 0 {
		return fmt.Errorf("%w: config.MaxNumItems is invalid", errInvalidConfig)
	}
	if config.MaxNumBytes == 0 {
		return fmt.Errorf("%w: config.MaxNumBytes is invalid", errInvalidConfig)
	}
	if config.NumItemsToPreemptivelyEvict == 0 {
		return fmt.Errorf("%w: config.NumItemsToPreemptivelyEvict is invalid", errInvalidConfig)
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

func (config *CacheConfig) String() string {
	bytes, _ := json.Marshal(config)
	return string(bytes)
}

type immunityChunkConfig struct {
	cacheName                   string
	maxNumItems                 uint32
	maxNumBytes                 uint32
	numItemsToPreemptivelyEvict uint32
}

func (config *immunityChunkConfig) String() string {
	return fmt.Sprintf("maxNumItems: %d, maxNumBytes: %d, numItemsToPreemptivelyEvict: %d", config.maxNumItems, config.maxNumBytes, config.numItemsToPreemptivelyEvict)
}
