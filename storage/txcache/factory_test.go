package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateCache_ReturnsDisabledWhenBadConfig(t *testing.T) {
	cache := CreateCache(CacheConfig{})
	_, isDisabled := cache.(*DisabledCache)
	require.True(t, isDisabled)
}
