package factory_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedStatusComponents --------------------
func TestManagedStatusComponents_CreateWithInvalidArgs_ShouldErr(t *testing.T) {
	statusArgs, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	coreComponents := getDefaultCoreComponents()
	statusArgs.CoreComponents = coreComponents

	statusComponentsFactory, _ := factory.NewStatusComponentsFactory(statusArgs)
	managedStatusComponents, err := factory.NewManagedStatusComponents(statusComponentsFactory)
	require.NoError(t, err)

	coreComponents.StatusHdl = nil
	err = managedStatusComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedStatusComponents.StatusHandler())
}

func TestManagedStatusComponents_Create_ShouldWork(t *testing.T) {
	statusArgs, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	statusComponentsFactory, _ := factory.NewStatusComponentsFactory(statusArgs)
	managedStatusComponents, err := factory.NewManagedStatusComponents(statusComponentsFactory)
	require.NoError(t, err)
	require.Nil(t, managedStatusComponents.StatusHandler())
	require.Nil(t, managedStatusComponents.ElasticIndexer())
	require.Nil(t, managedStatusComponents.SoftwareVersionChecker())
	require.Nil(t, managedStatusComponents.TpsBenchmark())

	err = managedStatusComponents.Create()
	require.NoError(t, err)
	require.NotNil(t, managedStatusComponents.StatusHandler())
	require.NotNil(t, managedStatusComponents.ElasticIndexer())
	require.NotNil(t, managedStatusComponents.SoftwareVersionChecker())
	require.NotNil(t, managedStatusComponents.TpsBenchmark())
}

func TestManagedStatusComponents_Close(t *testing.T) {
	statusArgs, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	statusComponentsFactory, _ := factory.NewStatusComponentsFactory(statusArgs)
	managedStatusComponents, _ := factory.NewManagedStatusComponents(statusComponentsFactory)
	err := managedStatusComponents.Create()
	require.NoError(t, err)

	err = managedStatusComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedStatusComponents.StatusHandler())
}

func TestManagedStatusComponents_CheckSubcomponents(t *testing.T) {
	statusArgs, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	statusComponentsFactory, _ := factory.NewStatusComponentsFactory(statusArgs)
	managedStatusComponents, _ := factory.NewManagedStatusComponents(statusComponentsFactory)
	err := managedStatusComponents.Create()
	require.NoError(t, err)

	err = managedStatusComponents.CheckSubcomponents()
	require.NoError(t, err)
}

func TestManagedStatusComponents_StartPolling(t *testing.T) {
	statusArgs, _ := getStatusComponentsFactoryArgsAndProcessComponents()
	statusArgs.Config.GeneralSettings.StatusPollingIntervalSec = 10
	statusComponentsFactory, _ := factory.NewStatusComponentsFactory(statusArgs)
	managedStatusComponents, _ := factory.NewManagedStatusComponents(statusComponentsFactory)
	err := managedStatusComponents.Create()
	require.NoError(t, err)

	err = managedStatusComponents.StartPolling()
	require.NoError(t, err)
	time.Sleep(3 * time.Second)
}
