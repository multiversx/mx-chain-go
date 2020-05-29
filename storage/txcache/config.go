package txcache

import (
	"encoding/json"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ConfigSourceMe holds cache configuration
type ConfigSourceMe struct {
	Name                          string
	NumChunks                     uint32
	EvictionEnabled               bool
	NumBytesThreshold             uint32
	NumBytesPerSenderThreshold    uint32
	CountThreshold                uint32
	CountPerSenderThreshold       uint32
	NumSendersToPreemptivelyEvict uint32
	MinGasPriceNanoErd            uint32
}

type senderConstraints struct {
	maxNumTxs   uint32
	maxNumBytes uint32
}

// TODO: Upon further analysis and brainstorming, add some sensible minimum accepted values for the appropriate fields.
func (config *ConfigSourceMe) verify() error {
	if len(config.Name) == 0 {
		return fmt.Errorf("%w: config.Name is invalid", storage.ErrInvalidConfig)
	}
	if config.NumChunks == 0 {
		return fmt.Errorf("%w: config.NumChunks is invalid", storage.ErrInvalidConfig)
	}
	if config.NumBytesPerSenderThreshold == 0 {
		return fmt.Errorf("%w: config.NumBytesPerSenderThreshold is invalid", storage.ErrInvalidConfig)
	}
	if config.CountPerSenderThreshold == 0 {
		return fmt.Errorf("%w: config.CountPerSenderThreshold is invalid", storage.ErrInvalidConfig)
	}
	if config.MinGasPriceNanoErd == 0 {
		return fmt.Errorf("%w: config.MinGasPriceNanoErd is invalid", storage.ErrInvalidConfig)
	}
	if config.EvictionEnabled {
		if config.NumBytesThreshold == 0 {
			return fmt.Errorf("%w: config.NumBytesThreshold is invalid", storage.ErrInvalidConfig)
		}

		if config.CountThreshold == 0 {
			return fmt.Errorf("%w: config.CountThreshold is invalid", storage.ErrInvalidConfig)
		}

		if config.NumSendersToPreemptivelyEvict == 0 {
			return fmt.Errorf("%w: config.NumSendersToPreemptivelyEvict is invalid", storage.ErrInvalidConfig)
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

// String returns a readable representation of the object
func (config *ConfigSourceMe) String() string {
	bytes, err := json.Marshal(config)
	if err != nil {
		log.Error("ConfigSourceMe.String()", "err", err)
	}

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

// TODO: Upon further analysis and brainstorming, add some sensible minimum accepted values for the appropriate fields.
func (config *ConfigDestinationMe) verify() error {
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

func (config *ConfigDestinationMe) getChunkConfig() crossTxChunkConfig {
	numChunks := core.MaxUint32(config.NumChunks, 1)

	return crossTxChunkConfig{
		maxNumItems:                 config.MaxNumItems / numChunks,
		maxNumBytes:                 config.MaxNumBytes / numChunks,
		numItemsToPreemptivelyEvict: config.NumItemsToPreemptivelyEvict / numChunks,
	}
}

// String returns a readable representation of the object
func (config *ConfigDestinationMe) String() string {
	bytes, err := json.Marshal(config)
	if err != nil {
		log.Error("ConfigDestinationMe.String()", "err", err)
	}

	return string(bytes)
}

type crossTxChunkConfig struct {
	maxNumItems                 uint32
	maxNumBytes                 uint32
	numItemsToPreemptivelyEvict uint32
}

func (config *crossTxChunkConfig) String() string {
	return fmt.Sprintf(
		"maxNumItems: %d, maxNumBytes: %d, numItemsToPreemptivelyEvict: %d",
		config.maxNumItems,
		config.maxNumBytes,
		config.numItemsToPreemptivelyEvict,
	)
}
