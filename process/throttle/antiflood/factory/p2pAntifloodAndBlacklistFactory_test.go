package factory

import (
	"context"
	"testing"
	"time"

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

	ctx := context.Background()
	cfg := config.Config{}
	components, err := NewP2PAntiFloodComponents(ctx, cfg, nil, currentPid)
	assert.Nil(t, components)
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
	ctx := context.Background()
	components, err := NewP2PAntiFloodComponents(ctx, cfg, ash, currentPid)
	assert.NotNil(t, components)
	assert.Nil(t, err)

	_, ok1 := components.AntiFloodHandler.(*disabled.AntiFlood)
	_, ok2 := components.BlacklistHandler.(*disabled.PeerBlacklistCacher)
	_, ok3 := components.PubKeysCacher.(*disabled.TimeCache)
	assert.True(t, ok1)
	assert.True(t, ok2)
	assert.True(t, ok3)
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

	ash := mock.NewAppStatusHandlerMock()
	ctx := context.Background()
	components, err := NewP2PAntiFloodComponents(ctx, cfg, ash, currentPid)
	assert.Nil(t, err)
	assert.NotNil(t, components.AntiFloodHandler)
	assert.NotNil(t, components.BlacklistHandler)
	assert.NotNil(t, components.PubKeysCacher)

	// we need this time sleep as to allow the code coverage tool to deterministically compute the code coverage
	//on the go routines that are automatically launched
	time.Sleep(time.Second * 2)
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
