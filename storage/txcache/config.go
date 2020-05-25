package txcache

import "fmt"

// CacheConfig holds cache configuration
type CacheConfig struct {
	Name                       string
	NumChunksHint              uint32
	EvictionEnabled            bool
	NumBytesThreshold          uint32
	CountThreshold             uint32
	NumSendersToEvictInOneStep uint32
	LargeNumOfTxsForASender    uint32
	NumTxsToEvictFromASender   uint32
	MinGasPriceNanoErd         uint32
}

type senderConstraints struct {
	maxNumTxs   uint32
	maxNumBytes uint32
}

func (config *CacheConfig) verify() error {
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

func (config *CacheConfig) getSenderConstraints() senderConstraints {
	return senderConstraints{
		maxNumBytes: config.NumBytesPerSenderThreshold,
		maxNumTxs:   config.CountPerSenderThreshold,
	}
}
