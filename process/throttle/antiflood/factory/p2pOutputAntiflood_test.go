package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/disabled"
	"github.com/stretchr/testify/assert"
)

func TestNewP2POutputAntiFlood_ShouldWorkAndReturnDisabledImplementations(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Antiflood: config.AntifloodConfig{
			Enabled: false,
		},
	}
	af, fp, err := NewP2POutputAntiFlood(cfg)
	assert.NotNil(t, af)
	assert.NotNil(t, fp)
	assert.Nil(t, err)

	_, ok := af.(*disabled.AntiFlood)
	assert.True(t, ok)
}

func TestNewP2POutputAntiFlood_BadCacheConfigShouldErr(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Antiflood: config.AntifloodConfig{
			Enabled: true,
			Cache: config.CacheConfig{
				Type:     "unknown type",
				Capacity: 10,
				Shards:   2,
			},
			PeerMaxOutput: config.AntifloodLimitsConfig{
				BaseMessagesPerInterval: 10,
				TotalSizePerInterval:    10,
			},
		},
	}

	af, fp, err := NewP2POutputAntiFlood(cfg)
	assert.NotNil(t, err)
	assert.True(t, check.IfNil(fp))
	assert.True(t, check.IfNil(af))
}

func TestNewP2POutputAntiFlood_BadConfigShouldErr(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Antiflood: config.AntifloodConfig{
			Enabled: true,
			Cache: config.CacheConfig{
				Type:     "LRU",
				Capacity: 10,
				Shards:   2,
			},
			PeerMaxOutput: config.AntifloodLimitsConfig{
				BaseMessagesPerInterval: 0,
				TotalSizePerInterval:    10,
			},
		},
	}

	af, fp, err := NewP2POutputAntiFlood(cfg)
	assert.NotNil(t, err)
	assert.True(t, check.IfNil(af))
	assert.True(t, check.IfNil(fp))
}

func TestNewP2POutputAntiFlood_ShouldWorkAndReturnOkImplementations(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Antiflood: config.AntifloodConfig{
			Enabled: true,
			Cache: config.CacheConfig{
				Type:     "LRU",
				Capacity: 10,
				Shards:   2,
			},
			PeerMaxOutput: config.AntifloodLimitsConfig{
				BaseMessagesPerInterval: 10,
				TotalSizePerInterval:    10,
			},
		},
	}

	af, fp, err := NewP2POutputAntiFlood(cfg)
	assert.Nil(t, err)
	assert.NotNil(t, af)
	assert.NotNil(t, fp)
}
