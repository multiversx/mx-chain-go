package txcache

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateCache_ReturnsDisabledWhenBadConfig(t *testing.T) {
	cache := CreateCache(CacheConfig{})
	_, isDisabled := cache.(*DisabledCache)
	require.True(t, isDisabled)
}

func TestCreateCache_ReturnsFailsafe(t *testing.T) {
	config := CacheConfig{
		Name:                       "test",
		NumChunksHint:              16,
		NumBytesPerSenderThreshold: math.MaxUint32,
		CountPerSenderThreshold:    math.MaxUint32,
		MinGasPriceNanoErd:         100,
	}

	cache := CreateCache(config)
	_, isFailsafe := cache.(*txCacheFailsafeDecorator)
	require.True(t, isFailsafe)
}
