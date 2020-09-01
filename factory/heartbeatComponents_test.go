package factory_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/stretchr/testify/require"
)

// ------------ Test HeartbeatComponents --------------------
func TestHeartbeatComponents_Close_ShouldWork(t *testing.T) {
	t.Parallel()

	heartbeatArgs := getDefaultHeartbeatComponents()
	hcf, _ := factory.NewHeartbeatComponentsFactory(heartbeatArgs)
	cc, _ := hcf.Create()

	err := cc.Close()
	require.NoError(t, err)
}

func getDefaultHeartbeatComponents() factory.HeartbeatComponentsFactoryArgs {
	coreComponents := getCoreComponents()
	networkComponents := getNetworkComponents()
	dataComponents := getDataComponents(coreComponents)
	cryptoComponents := getCryptoComponents(coreComponents)
	stateComponents := getStateComponents(coreComponents)
	processComponents := getProcessComponents(
		coreComponents,
		networkComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
	)

	return factory.HeartbeatComponentsFactoryArgs{
		Config: config.Config{
			Heartbeat: config.HeartbeatConfig{
				MinTimeToWaitBetweenBroadcastsInSec: 20,
				MaxTimeToWaitBetweenBroadcastsInSec: 25,
				HeartbeatRefreshIntervalInSec:       60,
				HideInactiveValidatorIntervalInSec:  3600,
				DurationToConsiderUnresponsiveInSec: 60,
				HeartbeatStorage: config.StorageConfig{
					Cache: config.CacheConfig{
						Capacity: 10000,
						Type:     "LRU",
						Shards:   1,
					},
					DB: config.DBConfig{
						FilePath:          "HeartbeatStorage",
						Type:              "MemoryDB",
						BatchDelaySeconds: 30,
						MaxBatchSize:      6,
						MaxOpenFiles:      10,
					},
				},
			},
			ValidatorStatistics: config.ValidatorStatisticsConfig{
				CacheRefreshIntervalInSec: uint32(100),
			},
		},
		Prefs:             config.Preferences{},
		AppVersion:        "test",
		GenesisTime:       time.Time{},
		HardforkTrigger:   &mock.HardforkTriggerStub{},
		CoreComponents:    coreComponents,
		DataComponents:    dataComponents,
		NetworkComponents: networkComponents,
		CryptoComponents:  cryptoComponents,
		ProcessComponents: processComponents,
	}
}
