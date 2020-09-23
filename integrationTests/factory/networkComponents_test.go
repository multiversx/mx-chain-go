package factory

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	mainFactory "github.com/ElrondNetwork/elrond-go/factory"
	"github.com/stretchr/testify/require"
)

// ------------ Test NetworkComponents --------------------
func TestNetworkComponents_Create_ShouldWork(t *testing.T) {
	t.Skip()

	generalConfig, _ := loadMainConfig(configPath)
	ratingsConfig, _ := loadRatingsConfig(ratingsPath)
	economicsConfig, _ := loadEconomicsConfig(economicsPath)
	p2pConfig, _ := core.LoadP2PConfig(p2pPath)

	ccf, _ := createCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)

	networkComponents, err := createNetworkComponents(*generalConfig, *p2pConfig, *ratingsConfig, ccf)

	require.Nil(t, err)
	require.NotNil(t, networkComponents)
}

func TestNetworkComponents_Create_Close_ShouldWork(t *testing.T) {
	//	t.Skip()

	_ = logger.SetLogLevel("*:DEBUG")

	generalConfig, _ := loadMainConfig(configPath)
	ratingsConfig, _ := loadRatingsConfig(ratingsPath)
	economicsConfig, _ := loadEconomicsConfig(economicsPath)
	p2pConfig, _ := core.LoadP2PConfig(p2pPath)

	ccf, _ := createCoreComponents(*generalConfig, *ratingsConfig, *economicsConfig)
	time.Sleep(2 * time.Second)

	printStack()

	nrBefore := runtime.NumGoroutine()
	networkComponents, _ := createNetworkComponents(*generalConfig, *p2pConfig, *ratingsConfig, ccf)
	time.Sleep(2 * time.Second)

	_ = networkComponents.NetworkMessenger().Bootstrap()
	time.Sleep(3 * time.Second)

	err := networkComponents.Close()
	require.Nil(t, err)
	time.Sleep(5 * time.Second)
	nrAfter := runtime.NumGoroutine()

	if nrBefore != nrAfter {
		printStack()
	}

	require.Equal(t, nrBefore, nrAfter)

}

func createNetworkComponents(
	config config.Config,
	p2pConfig config.P2PConfig,
	ratingsConfig config.RatingsConfig,
	managedCoreComponents mainFactory.CoreComponentsHandler) (mainFactory.NetworkComponentsHandler, error) {
	networkComponentsFactoryArgs := mainFactory.NetworkComponentsFactoryArgs{
		P2pConfig:     p2pConfig,
		MainConfig:    config,
		RatingsConfig: ratingsConfig,
		StatusHandler: managedCoreComponents.StatusHandler(),
		Marshalizer:   managedCoreComponents.InternalMarshalizer(),
		Syncer:        managedCoreComponents.SyncTimer(),
	}

	networkComponentsFactory, err := mainFactory.NewNetworkComponentsFactory(networkComponentsFactoryArgs)
	if err != nil {
		return nil, fmt.Errorf("NewNetworkComponentsFactory failed: %w", err)
	}

	managedNetworkComponents, err := mainFactory.NewManagedNetworkComponents(networkComponentsFactory)
	if err != nil {
		return nil, err
	}
	err = managedNetworkComponents.Create()
	if err != nil {
		return nil, err
	}

	return managedNetworkComponents, nil
}
