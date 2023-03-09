package heartbeat_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	errErd "github.com/multiversx/mx-chain-go/errors"
	bootstrapComp "github.com/multiversx/mx-chain-go/factory/bootstrap"
	heartbeatComp "github.com/multiversx/mx-chain-go/factory/heartbeat"
	"github.com/multiversx/mx-chain-go/factory/mock"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/assert"
)

func createMockHeartbeatV2ComponentsFactoryArgs() heartbeatComp.ArgHeartbeatV2ComponentsFactory {
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	bootStrapArgs := componentsMock.GetBootStrapFactoryArgs()
	bootstrapComponentsFactory, _ := bootstrapComp.NewBootstrapComponentsFactory(bootStrapArgs)
	bootstrapC, _ := bootstrapComp.NewTestManagedBootstrapComponents(bootstrapComponentsFactory)
	_ = bootstrapC.Create()

	_ = bootstrapC.SetShardCoordinator(shardCoordinator)

	statusCoreC := componentsMock.GetStatusCoreComponents()
	coreC := componentsMock.GetCoreComponents()
	cryptoC := componentsMock.GetCryptoComponents(coreC)
	networkC := componentsMock.GetNetworkComponents(cryptoC)
	dataC := componentsMock.GetDataComponents(coreC, shardCoordinator)
	stateC := componentsMock.GetStateComponents(coreC, shardCoordinator)
	processC := componentsMock.GetProcessComponents(shardCoordinator, coreC, networkC, dataC, cryptoC, stateC)
	return heartbeatComp.ArgHeartbeatV2ComponentsFactory{
		Config: config.Config{
			HeartbeatV2: config.HeartbeatV2Config{
				PeerAuthenticationTimeBetweenSendsInSec:          1,
				PeerAuthenticationTimeBetweenSendsWhenErrorInSec: 1,
				PeerAuthenticationTimeThresholdBetweenSends:      0.1,
				HeartbeatTimeBetweenSendsInSec:                   1,
				HeartbeatTimeBetweenSendsDuringBootstrapInSec:    1,
				HeartbeatTimeBetweenSendsWhenErrorInSec:          1,
				HeartbeatTimeThresholdBetweenSends:               0.1,
				HeartbeatExpiryTimespanInSec:                     30,
				MinPeersThreshold:                                0.8,
				DelayBetweenPeerAuthenticationRequestsInSec:      10,
				PeerAuthenticationMaxTimeoutForRequestsInSec:     60,
				PeerShardTimeBetweenSendsInSec:                   5,
				PeerShardTimeThresholdBetweenSends:               0.1,
				MaxMissingKeysInRequest:                          100,
				MaxDurationPeerUnresponsiveInSec:                 10,
				HideInactiveValidatorIntervalInSec:               60,
				HardforkTimeBetweenSendsInSec:                    5,
				TimeBetweenConnectionsMetricsUpdateInSec:         10,
				TimeToReadDirectConnectionsInSec:                 15,
				PeerAuthenticationTimeBetweenChecksInSec:         6,
				HeartbeatPool: config.CacheConfig{
					Type:     "LRU",
					Capacity: 1000,
					Shards:   1,
				},
			},
			Hardfork: config.HardforkConfig{
				PublicKeyToListenFrom: componentsMock.DummyPk,
			},
		},
		Prefs: config.Preferences{
			Preferences: config.PreferencesConfig{
				NodeDisplayName: "node",
				Identity:        "identity",
			},
		},
		BaseVersion:        "test-base",
		AppVersion:           "test",
		BootstrapComponents:  bootstrapC,
		CoreComponents:       coreC,
		DataComponents:       dataC,
		NetworkComponents:    networkC,
		CryptoComponents:     cryptoC,
		ProcessComponents:    processC,
		StatusCoreComponents: statusCoreC,
	}
}

func Test_heartbeatV2Components_Create(t *testing.T) {
	t.Parallel()

	t.Run("invalid config should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.Config.HeartbeatV2.HeartbeatExpiryTimespanInSec = args.Config.HeartbeatV2.PeerAuthenticationTimeBetweenSendsInSec
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.False(t, check.IfNil(hcf))
		assert.Nil(t, err)

		hc, err := hcf.Create()
		assert.Nil(t, hc)
		assert.True(t, errors.Is(err, errErd.ErrInvalidHeartbeatV2Config))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.False(t, check.IfNil(hcf))
		assert.Nil(t, err)

		hc, err := hcf.Create()
		assert.NotNil(t, hc)
		assert.Nil(t, err)

		err = hc.Close()
		assert.Nil(t, err)
	})
}
