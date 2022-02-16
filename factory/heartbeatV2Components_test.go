package factory_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/config"
	elrondErrors "github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/stretchr/testify/assert"
)

func createMockHeartbeatV2ComponentsFactoryArgs() factory.ArgHeartbeatV2ComponentsFactory {
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	coreC := getCoreComponents()
	networkC := getNetworkComponents()
	dataC := getDataComponents(coreC, shardCoordinator)
	cryptoC := getCryptoComponents(coreC)

	return factory.ArgHeartbeatV2ComponentsFactory{
		Config: config.Config{
			HeartbeatV2: config.HeartbeatV2Config{
				PeerAuthenticationTimeBetweenSendsInSec:          1,
				PeerAuthenticationTimeBetweenSendsWhenErrorInSec: 1,
				PeerAuthenticationThresholdBetweenSends:          0.1,
				HeartbeatTimeBetweenSendsInSec:                   1,
				HeartbeatTimeBetweenSendsWhenErrorInSec:          1,
				HeartbeatThresholdBetweenSends:                   0.1,
				HeartbeatExpiryTimespanInSec:                     30,
				PeerAuthenticationPool: config.PeerAuthenticationPoolConfig{
					DefaultSpanInSec: 30,
					CacheExpiryInSec: 30,
				},
				HeartbeatPool: config.CacheConfig{
					Type:     "LRU",
					Capacity: 1000,
					Shards:   1,
				},
			},
		},
		Prefs: config.Preferences{
			Preferences: config.PreferencesConfig{
				NodeDisplayName: "node",
				Identity:        "identity",
			},
		},
		AppVersion: "test",
		RedundancyHandler: &mock.RedundancyHandlerStub{
			ObserverPrivateKeyCalled: func() crypto.PrivateKey {
				return &mock.PrivateKeyStub{
					GeneratePublicHandler: func() crypto.PublicKey {
						return &mock.PublicKeyMock{}
					},
				}
			},
		},
		CoreComponents:    coreC,
		DataComponents:    dataC,
		NetworkComponents: networkC,
		CryptoComponents:  cryptoC,
	}
}

func TestNewHeartbeatV2ComponentsFactory(t *testing.T) {
	t.Parallel()

	t.Run("nil core components should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.CoreComponents = nil
		hcf, err := factory.NewHeartbeatV2ComponentsFactory(args)
		assert.True(t, check.IfNil(hcf))
		assert.Equal(t, elrondErrors.ErrNilCoreComponentsHolder, err)
	})
	t.Run("nil data components should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.DataComponents = nil
		hcf, err := factory.NewHeartbeatV2ComponentsFactory(args)
		assert.True(t, check.IfNil(hcf))
		assert.Equal(t, elrondErrors.ErrNilDataComponentsHolder, err)
	})
	t.Run("nil network components should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.NetworkComponents = nil
		hcf, err := factory.NewHeartbeatV2ComponentsFactory(args)
		assert.True(t, check.IfNil(hcf))
		assert.Equal(t, elrondErrors.ErrNilNetworkComponentsHolder, err)
	})
	t.Run("nil crypto components should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.CryptoComponents = nil
		hcf, err := factory.NewHeartbeatV2ComponentsFactory(args)
		assert.True(t, check.IfNil(hcf))
		assert.Equal(t, elrondErrors.ErrNilCryptoComponentsHolder, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		hcf, err := factory.NewHeartbeatV2ComponentsFactory(args)
		assert.False(t, check.IfNil(hcf))
		assert.Nil(t, err)
	})
}

func Test_heartbeatV2ComponentsFactory_Create(t *testing.T) {
	t.Parallel()

	t.Run("new sender returns error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.Config.HeartbeatV2.HeartbeatTimeBetweenSendsInSec = 0
		hcf, err := factory.NewHeartbeatV2ComponentsFactory(args)
		assert.False(t, check.IfNil(hcf))
		assert.Nil(t, err)

		hc, err := hcf.Create()
		assert.Nil(t, hc)
		assert.True(t, errors.Is(err, heartbeat.ErrInvalidTimeDuration))
		assert.True(t, strings.Contains(err.Error(), "timeBetweenSends"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		hcf, err := factory.NewHeartbeatV2ComponentsFactory(args)
		assert.False(t, check.IfNil(hcf))
		assert.Nil(t, err)

		hc, err := hcf.Create()
		assert.NotNil(t, hc)
		assert.Nil(t, err)
	})
}

func Test_heartbeatV2Components_Close(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	args := createMockHeartbeatV2ComponentsFactoryArgs()
	hcf, err := factory.NewHeartbeatV2ComponentsFactory(args)
	assert.False(t, check.IfNil(hcf))
	assert.Nil(t, err)

	hc, err := hcf.Create()
	assert.NotNil(t, hc)
	assert.Nil(t, err)
}
