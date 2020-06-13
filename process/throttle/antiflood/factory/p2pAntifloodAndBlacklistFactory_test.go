package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/p2p/mock"
	"github.com/ElrondNetwork/elrond-go/process/throttle/antiflood/disabled"
	"github.com/stretchr/testify/assert"
)

const currentPid = core.PeerID("current pid")

func TestNewP2PAntiFloodAndBlackList_NilStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	af, bl, err := NewP2PAntiFloodAndBlackList(cfg, nil, currentPid)
	assert.Nil(t, af)
	assert.Nil(t, bl)
	assert.Equal(t, p2p.ErrNilStatusHandler, err)
}

func TestNewP2PAntiFloodAndBlackList_ShouldWorkAndReturnDisabledImplementations(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Antiflood: config.AntifloodConfig{
			Enabled: false,
		},
	}
	ash := &mock.AppStatusHandlerMock{}
	af, bl, err := NewP2PAntiFloodAndBlackList(cfg, ash, currentPid)
	assert.NotNil(t, af)
	assert.NotNil(t, bl)
	assert.Nil(t, err)

	_, ok1 := af.(*disabled.AntiFlood)
	_, ok2 := bl.(*disabled.PeerBlacklistHandler)
	assert.True(t, ok1)
	assert.True(t, ok2)
}

func TestNewP2PAntiFloodAndBlackList_ShouldWorkAndReturnOkImplementations(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Antiflood: config.AntifloodConfig{
			Enabled: true,
			Cache: config.CacheConfig{
				Type:     "LRU",
				Capacity: 10,
				Shards:   2,
			},
			FastReacting: createFloodPreventerConfig(),
			SlowReacting: createFloodPreventerConfig(),
			OutOfSpecs:   createFloodPreventerConfig(),
			Topic: config.TopicAntifloodConfig{
				DefaultMaxMessagesPerSec: 10,
			},
		},
	}

	ash := &mock.AppStatusHandlerMock{}
	af, bl, err := NewP2PAntiFloodAndBlackList(cfg, ash, currentPid)
	assert.Nil(t, err)
	assert.NotNil(t, af)
	assert.NotNil(t, bl)
}

func createFloodPreventerConfig() config.FloodPreventerConfig {
	return config.FloodPreventerConfig{
		IntervalInSeconds: 1,
		PeerMaxInput: config.AntifloodLimitsConfig{
			BaseMessagesPerInterval: 10,
			TotalSizePerInterval:    10,
		},
		BlackList: config.BlackListConfig{
			ThresholdNumMessagesPerInterval: 10,
			ThresholdSizePerInterval:        10,
			NumFloodingRounds:               10,
			PeerBanDurationInSeconds:        10,
		},
	}
}
