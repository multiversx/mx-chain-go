package factory_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedHeartbeatComponents --------------------
func TestManagedHeartbeatComponents_CreateWithInvalidArgs_ShouldErr(t *testing.T) {
	heartbeatArgs := getDefaultHeartbeatComponents()
	heartbeatArgs.Config.Heartbeat.MaxTimeToWaitBetweenBroadcastsInSec = 0
	heartbeatComponentsFactory, _ := factory.NewHeartbeatComponentsFactory(heartbeatArgs)
	managedHeartbeatComponents, err := factory.NewManagedHeartbeatComponents(heartbeatComponentsFactory)
	require.NoError(t, err)
	err = managedHeartbeatComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedHeartbeatComponents.Monitor())
}

func TestManagedHeartbeatComponents_Create_ShouldWork(t *testing.T) {
	heartbeatArgs := getDefaultHeartbeatComponents()
	heartbeatComponentsFactory, _ := factory.NewHeartbeatComponentsFactory(heartbeatArgs)
	managedHeartbeatComponents, err := factory.NewManagedHeartbeatComponents(heartbeatComponentsFactory)
	require.NoError(t, err)
	require.Nil(t, managedHeartbeatComponents.Monitor())
	require.Nil(t, managedHeartbeatComponents.MessageHandler())
	require.Nil(t, managedHeartbeatComponents.Sender())
	require.Nil(t, managedHeartbeatComponents.Storer())

	err = managedHeartbeatComponents.Create()
	require.NoError(t, err)
	require.NotNil(t, managedHeartbeatComponents.Monitor())
	require.NotNil(t, managedHeartbeatComponents.MessageHandler())
	require.NotNil(t, managedHeartbeatComponents.Sender())
	require.NotNil(t, managedHeartbeatComponents.Storer())
}

func TestManagedHeartbeatComponents_Close(t *testing.T) {
	heartbeatArgs := getDefaultHeartbeatComponents()
	heartbeatComponentsFactory, _ := factory.NewHeartbeatComponentsFactory(heartbeatArgs)
	managedHeartbeatComponents, _ := factory.NewManagedHeartbeatComponents(heartbeatComponentsFactory)
	err := managedHeartbeatComponents.Create()
	require.NoError(t, err)

	err = managedHeartbeatComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedHeartbeatComponents.Monitor())
}

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
