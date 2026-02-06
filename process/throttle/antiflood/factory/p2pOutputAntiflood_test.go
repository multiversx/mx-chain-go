package factory

import (
	"context"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process/throttle/antiflood/disabled"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewP2POutputAntiFlood_ShouldWorkAndReturnDisabledImplementations(t *testing.T) {
	t.Parallel()

	antifloodConfigHandler := &testscommon.AntifloodConfigsHandlerStub{
		IsEnabledCalled: func() bool {
			return false
		},
	}

	ctx := context.Background()
	af, err := NewP2POutputAntiFlood(ctx, antifloodConfigHandler)
	assert.NotNil(t, af)
	assert.Nil(t, err)

	_, ok := af.(*disabled.AntiFlood)
	assert.True(t, ok)
}

func TestNewP2POutputAntiFlood_BadCacheConfigShouldErr(t *testing.T) {
	t.Parallel()

	antifloodConfigHandler := &testscommon.AntifloodConfigsHandlerStub{
		IsEnabledCalled: func() bool {
			return true
		},
		GetCurrentConfigCalled: func() config.AntifloodConfigByRound {
			return config.AntifloodConfigByRound{
				Cache: config.CacheConfig{
					Type:     "unknown type",
					Capacity: 10,
					Shards:   2,
				},
				PeerMaxOutput: config.FloodPreventerConfig{
					PeerMaxInput: config.AntifloodLimitsConfig{
						BaseMessagesPerInterval: 10,
						TotalSizePerInterval:    10,
					},
				},
			}
		},
	}

	ctx := context.Background()
	af, err := NewP2POutputAntiFlood(ctx, antifloodConfigHandler)
	assert.NotNil(t, err)
	assert.True(t, check.IfNil(af))
}

func TestNewP2POutputAntiFlood_ShouldWorkAndReturnOkImplementations(t *testing.T) {
	t.Parallel()

	antifloodConfigHandler := &testscommon.AntifloodConfigsHandlerStub{
		IsEnabledCalled: func() bool {
			return true
		},
		GetCurrentConfigCalled: func() config.AntifloodConfigByRound {
			return config.AntifloodConfigByRound{
				Cache: config.CacheConfig{
					Type:     "LRU",
					Capacity: 10,
					Shards:   2,
				},
				PeerMaxOutput: config.FloodPreventerConfig{
					PeerMaxInput: config.AntifloodLimitsConfig{
						BaseMessagesPerInterval: 10,
						TotalSizePerInterval:    10,
					},
				},
			}
		},
	}

	ctx := context.Background()
	af, err := NewP2POutputAntiFlood(ctx, antifloodConfigHandler)
	assert.Nil(t, err)
	assert.NotNil(t, af)
}
