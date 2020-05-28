package txcache

import (
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
)

// ConfigSourceMe holds cache configuration
type ConfigSourceMe struct {
	Name                       string
	NumChunksHint              uint32
	EvictionEnabled            bool
	NumBytesThreshold          uint32
	NumBytesPerSenderThreshold uint32
	CountThreshold             uint32
	CountPerSenderThreshold    uint32
	NumSendersToEvictInOneStep uint32
	MinGasPriceNanoErd         uint32
}

type senderConstraints struct {
	maxNumTxs   uint32
	maxNumBytes uint32
}

// TODO: perhaps add better constraints for "CountThreshold" and "NumBytesThreshold"?
func (config *ConfigSourceMe) verify() error {
	if len(config.Name) == 0 {
		return fmt.Errorf("%w: config.Name is invalid", errInvalidCacheConfig)
	}
	if config.NumChunksHint == 0 {
		return fmt.Errorf("%w: config.NumChunksHint is invalid", errInvalidCacheConfig)
	}
	if config.NumBytesPerSenderThreshold == 0 {
		return fmt.Errorf("%w: config.NumBytesPerSenderThreshold is invalid", errInvalidCacheConfig)
	}
	if config.CountPerSenderThreshold == 0 {
		return fmt.Errorf("%w: config.CountPerSenderThreshold is invalid", errInvalidCacheConfig)
	}
	if config.MinGasPriceNanoErd == 0 {
		return fmt.Errorf("%w: config.MinGasPriceNanoErd is invalid", errInvalidCacheConfig)
	}
	if config.EvictionEnabled {
		if config.NumBytesThreshold == 0 {
			return fmt.Errorf("%w: config.NumBytesThreshold is invalid", errInvalidCacheConfig)
		}

		if config.CountThreshold == 0 {
			return fmt.Errorf("%w: config.CountThreshold is invalid", errInvalidCacheConfig)
		}

		if config.NumSendersToEvictInOneStep == 0 {
			return fmt.Errorf("%w: config.NumSendersToEvictInOneStep is invalid", errInvalidCacheConfig)
		}
	}

	return nil
}

func (config *ConfigSourceMe) getSenderConstraints() senderConstraints {
	return senderConstraints{
		maxNumBytes: config.NumBytesPerSenderThreshold,
		maxNumTxs:   config.CountPerSenderThreshold,
	}
}

func (config *ConfigSourceMe) String() string {
	bytes, _ := json.Marshal(config)
	return string(bytes)
}

// ConfigDestinationMe holds cache configuration
type ConfigDestinationMe struct {
	Name                        string
	NumChunks                   uint32
	MaxNumItems                 uint32
	MaxNumBytes                 uint32
	NumItemsToPreemptivelyEvict uint32
}

func (config *ConfigDestinationMe) verify() error {
	if len(config.Name) == 0 {
		return fmt.Errorf("%w: config.Name is invalid", errInvalidCacheConfig)
	}
	if config.NumChunks == 0 {
		return fmt.Errorf("%w: config.NumChunks is invalid", errInvalidCacheConfig)
	}
	if config.MaxNumItems == 0 {
		return fmt.Errorf("%w: config.MaxNumItems is invalid", errInvalidCacheConfig)
	}
	if config.MaxNumBytes == 0 {
		return fmt.Errorf("%w: config.MaxNumBytes is invalid", errInvalidCacheConfig)
	}
	if config.NumItemsToPreemptivelyEvict == 0 {
		return fmt.Errorf("%w: config.NumItemsToPreemptivelyEvict is invalid", errInvalidCacheConfig)
	}

	return nil
}

func (config *ConfigDestinationMe) getChunkConfig() crossTxChunkConfig {
	numChunks := core.MaxUint32(config.NumChunks, 1)

	return crossTxChunkConfig{
		maxNumItems:                 config.MaxNumItems / numChunks,
		maxNumBytes:                 config.MaxNumBytes / numChunks,
		numItemsToPreemptivelyEvict: config.NumItemsToPreemptivelyEvict / numChunks,
	}
}

func (config *ConfigDestinationMe) String() string {
	bytes, _ := json.Marshal(config)
	return string(bytes)
}

type crossTxChunkConfig struct {
	maxNumItems                 uint32
	maxNumBytes                 uint32
	numItemsToPreemptivelyEvict uint32
}

func (config *crossTxChunkConfig) String() string {
	return fmt.Sprintf("maxNumItems: %d, maxNumBytes: %d, numItemsToPreemptivelyEvict: %d", config.maxNumItems, config.maxNumBytes, config.numItemsToPreemptivelyEvict)
}
