package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/stretchr/testify/assert"
)

func createMockHeartbeatV2ComponentsFactoryArgs() factory.ArgHeartbeatV2ComponentsFactory {
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	bootStrapArgs := getBootStrapArgs()
	bootstrapComponentsFactory, _ := factory.NewBootstrapComponentsFactory(bootStrapArgs)
	bootstrapC, _ := factory.NewManagedBootstrapComponents(bootstrapComponentsFactory)
	_ = bootstrapC.Create()
	factory.SetShardCoordinator(shardCoordinator, bootstrapC)

	coreC := getCoreComponents()
	networkC := getNetworkComponents()
	dataC := getDataComponents(coreC, shardCoordinator)
	cryptoC := getCryptoComponents(coreC)
	stateC := getStateComponents(coreC, shardCoordinator)
	processC := getProcessComponents(shardCoordinator, coreC, networkC, dataC, cryptoC, stateC)
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
				MinPeersThreshold:                                0.8,
				DelayBetweenRequestsInSec:                        10,
				MaxTimeoutInSec:                                  60,
				PeerShardTimeBetweenSendsInSec:                   5,
				PeerShardThresholdBetweenSends:                   0.1,
				MaxMissingKeysInRequest:                          100,
				MaxDurationPeerUnresponsiveInSec:                 10,
				HideInactiveValidatorIntervalInSec:               60,
				HardforkTimeBetweenSendsInSec:                    5,
				TimeBetweenConnectionsMetricsUpdateInSec:         10,
				HeartbeatPool: config.CacheConfig{
					Type:     "LRU",
					Capacity: 1000,
					Shards:   1,
				},
			},
			Hardfork: config.HardforkConfig{
				PublicKeyToListenFrom: dummyPk,
			},
		},
		Prefs: config.Preferences{
			Preferences: config.PreferencesConfig{
				NodeDisplayName: "node",
				Identity:        "identity",
			},
		},
		AppVersion:              "test",
		BoostrapComponents:      bootstrapC,
		CoreComponents:          coreC,
		DataComponents:          dataC,
		NetworkComponents:       networkC,
		CryptoComponents:        cryptoC,
		ProcessComponents:       processC,
		HeartbeatV1DisableEpoch: 1,
	}
}

func Test_heartbeatV2Components_Create_ShouldWork(t *testing.T) {
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

	err = hc.Close()
	assert.Nil(t, err)
}
