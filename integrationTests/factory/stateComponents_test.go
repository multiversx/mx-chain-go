package factory

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/require"
)

func TestStateComponents_Create_Close_ShouldWork(t *testing.T) {
	generalConfig, _ := loadMainConfig(configPath)
	ratingsConfig, _ := loadRatingsConfig(ratingsPath)
	economicsConfig, _ := loadEconomicsConfig(economicsPath)
	preferencesConfig, _ := loadPreferencesConfig(prefsPath)
	systemScConfig, _ := loadSystemSmartContractsConfig(systemSCConfigPath)
	p2pConfig, _ := core.LoadP2PConfig(p2pPath)

	nrBefore := runtime.NumGoroutine()

	coreComponents, err := createCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
	require.Nil(t, err)
	require.NotNil(t, coreComponents)

	crytoComponents, err := createCryptoComponents(*generalConfig, *systemScConfig, coreComponents)
	require.Nil(t, err)
	require.NotNil(t, crytoComponents)

	networkComponents, err := createNetworkComponents(*generalConfig, *p2pConfig, *ratingsConfig, coreComponents)
	require.Nil(t, err)
	require.NotNil(t, networkComponents)

	bootstrapComponents, err := createBootstrapComponents(
		*generalConfig, preferencesConfig.Preferences, coreComponents, crytoComponents, networkComponents,
	)
	require.Nil(t, err)
	require.NotNil(t, bootstrapComponents)
	time.Sleep(5 * time.Second)

	stateComponents, err := createStateComponents(*generalConfig, coreComponents, bootstrapComponents)
	require.Nil(t, err)
	require.NotNil(t, stateComponents)
	time.Sleep(2 * time.Second)

	err = stateComponents.Close()
	require.Nil(t, err)
	err = bootstrapComponents.Close()
	require.Nil(t, err)
	err = networkComponents.Close()
	require.Nil(t, err)
	err = crytoComponents.Close()
	require.Nil(t, err)
	err = coreComponents.Close()
	require.Nil(t, err)

	time.Sleep(2 * time.Second)
	nrAfter := runtime.NumGoroutine()

	if nrBefore != nrAfter {
		printStack()
	}

	require.Equal(t, nrBefore, nrAfter)
}

func createStateComponents(
	genConfig config.Config,
	coreComponents factory.CoreComponentsHolder,
	bootstrapComponents factory.BootstrapComponentsHolder,
) (factory.StateComponentsHandler, error) {
	nbShards := uint32(3)
	selfShardID := uint32(0)
	shardCoordinator, err := sharding.NewMultiShardCoordinator(nbShards, selfShardID)
	if err != nil {
		return nil, err
	}

	triesComponents, trieStorageManagers := bootstrapComponents.EpochStartBootstrapper().GetTriesComponents()
	stateArgs := factory.StateComponentsFactoryArgs{
		Config:              genConfig,
		ShardCoordinator:    shardCoordinator,
		Core:                coreComponents,
		TriesContainer:      triesComponents,
		TrieStorageManagers: trieStorageManagers,
	}

	stateComponentsFactory, err := factory.NewStateComponentsFactory(stateArgs)
	if err != nil {
		return nil, fmt.Errorf("NewStateComponentsFactory failed: %w", err)
	}

	managedStateComponents, err := factory.NewManagedStateComponents(stateComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedStateComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedStateComponents, nil
}
