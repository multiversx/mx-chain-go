package factory

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	mainFactory "github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/require"
)

func TestDataComponents_Create_Close_ShouldWork(t *testing.T) {
	generalConfig, _ := loadMainConfig(configPath)
	ratingsConfig, _ := loadRatingsConfig(ratingsPath)
	economicsConfig, _ := loadEconomicsConfig(economicsPath)

	time.Sleep(2 * time.Second)
	nrBefore := runtime.NumGoroutine()

	coreComponents, err := createCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
	require.Nil(t, err)
	require.NotNil(t, coreComponents)

	dataComponents, err := createDataComponents(*generalConfig, *economicsConfig, coreComponents)
	require.Nil(t, err)
	require.NotNil(t, dataComponents)
	time.Sleep(2 * time.Second)

	err = dataComponents.Close()
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

func createDataComponents(genConfig config.Config, econConfig config.EconomicsConfig, coreComponents mainFactory.CoreComponentsHolder) (mainFactory.DataComponentsHandler, error) {
	epochStartNotifier := notifier.NewEpochStartSubscriptionHandler()

	currentEpoch := uint32(0)

	economicsData, err := economics.NewEconomicsData(&econConfig)
	if err != nil {
		return nil, err
	}

	nbShards := uint32(3)
	selfShardID := uint32(0)
	shardCoordinator, err := sharding.NewMultiShardCoordinator(nbShards, selfShardID)
	if err != nil {
		return nil, err
	}

	dataArgs := mainFactory.DataComponentsFactoryArgs{
		Config:             genConfig,
		EconomicsData:      economicsData,
		ShardCoordinator:   shardCoordinator,
		Core:               coreComponents,
		EpochStartNotifier: epochStartNotifier,
		CurrentEpoch:       currentEpoch,
	}

	dataComponentsFactory, err := mainFactory.NewDataComponentsFactory(dataArgs)
	if err != nil {
		return nil, fmt.Errorf("NewDataComponentsFactory failed: %w", err)
	}
	managedDataComponents, err := mainFactory.NewManagedDataComponents(dataComponentsFactory)
	if err != nil {
		return nil, err
	}
	err = managedDataComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedDataComponents, nil
}
