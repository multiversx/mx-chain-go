package txcache

import (
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon/txcachemocks"
	"github.com/multiversx/mx-chain-storage-go/common"
	"github.com/stretchr/testify/assert"
)

func TestNewTxCache(t *testing.T) {
	t.Parallel()

	t.Run("nil parameter should error", func(t *testing.T) {
		t.Parallel()

		cfg := ConfigSourceMe{
			Name:                          "test",
			NumChunks:                     1,
			NumBytesThreshold:             1000,
			NumBytesPerSenderThreshold:    100,
			CountThreshold:                10,
			CountPerSenderThreshold:       100,
			NumSendersToPreemptivelyEvict: 1,
		}

		cache, err := NewTxCache(cfg, nil)
		assert.Nil(t, cache)
		assert.Equal(t, common.ErrNilTxGasHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cfg := ConfigSourceMe{
			Name:                          "test",
			NumChunks:                     1,
			NumBytesThreshold:             1000,
			NumBytesPerSenderThreshold:    100,
			CountThreshold:                10,
			CountPerSenderThreshold:       100,
			NumSendersToPreemptivelyEvict: 1,
		}

		cache, err := NewTxCache(cfg, &txcachemocks.TxGasHandlerMock{
			GasProcessingDivisor: 1,
			MinimumGasPrice:      1,
			MinimumGasMove:       1,
		})
		assert.NotNil(t, cache)
		assert.Nil(t, err)
	})
}

func TestNewDisabledCache(t *testing.T) {
	t.Parallel()

	cache := NewDisabledCache()
	assert.NotNil(t, cache)
}

func TestNewCrossTxCache(t *testing.T) {
	t.Parallel()

	t.Run("invalid config should error", func(t *testing.T) {
		t.Parallel()

		cfg := ConfigDestinationMe{
			Name:                        "",
			NumChunks:                   1,
			MaxNumItems:                 100,
			MaxNumBytes:                 1000,
			NumItemsToPreemptivelyEvict: 1,
		}

		cache, err := NewCrossTxCache(cfg)
		assert.Nil(t, cache)
		assert.ErrorIs(t, err, common.ErrInvalidConfig)
		assert.True(t, strings.Contains(err.Error(), "config.Name is invalid"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cfg := ConfigDestinationMe{
			Name:                        "test",
			NumChunks:                   1,
			MaxNumItems:                 100,
			MaxNumBytes:                 1000,
			NumItemsToPreemptivelyEvict: 1,
		}

		cache, err := NewCrossTxCache(cfg)
		assert.NotNil(t, cache)
		assert.Nil(t, err)
	})
}
