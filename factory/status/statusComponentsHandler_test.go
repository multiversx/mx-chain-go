package status_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/factory/mock"
	statusComp "github.com/ElrondNetwork/elrond-go/factory/status"
	componentsMock "github.com/ElrondNetwork/elrond-go/testscommon/components"
	"github.com/ElrondNetwork/elrond-go/testscommon/factory"
	"github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedStatusComponents --------------------
func TestManagedStatusComponents_CreateWithInvalidArgsShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	statusArgs, _ := componentsMock.GetStatusComponentsFactoryArgsAndProcessComponents(shardCoordinator)
	statusCoreComponents := &factory.StatusCoreComponentsStub{
		AppStatusHandlerField: &statusHandler.AppStatusHandlerStub{},
	}
	statusArgs.StatusCoreComponents = statusCoreComponents

	statusComponentsFactory, _ := statusComp.NewStatusComponentsFactory(statusArgs)
	managedStatusComponents, err := statusComp.NewManagedStatusComponents(statusComponentsFactory)
	require.NoError(t, err)

	statusCoreComponents.AppStatusHandlerField = nil
	err = managedStatusComponents.Create()
	require.Error(t, err)
}

func TestManagedStatusComponents_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	statusArgs, _ := componentsMock.GetStatusComponentsFactoryArgsAndProcessComponents(shardCoordinator)
	statusComponentsFactory, _ := statusComp.NewStatusComponentsFactory(statusArgs)
	managedStatusComponents, err := statusComp.NewManagedStatusComponents(statusComponentsFactory)
	require.NoError(t, err)
	require.Nil(t, managedStatusComponents.OutportHandler())
	require.Nil(t, managedStatusComponents.SoftwareVersionChecker())

	err = managedStatusComponents.Create()
	require.NoError(t, err)
	require.NotNil(t, managedStatusComponents.OutportHandler())
	require.NotNil(t, managedStatusComponents.SoftwareVersionChecker())
}

func TestManagedStatusComponents_Close(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	statusArgs, _ := componentsMock.GetStatusComponentsFactoryArgsAndProcessComponents(shardCoordinator)
	statusComponentsFactory, _ := statusComp.NewStatusComponentsFactory(statusArgs)
	managedStatusComponents, _ := statusComp.NewManagedStatusComponents(statusComponentsFactory)
	err := managedStatusComponents.Create()
	require.NoError(t, err)

	err = managedStatusComponents.Close()
	require.NoError(t, err)
}

func TestManagedStatusComponents_CheckSubcomponents(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	statusArgs, _ := componentsMock.GetStatusComponentsFactoryArgsAndProcessComponents(shardCoordinator)
	statusComponentsFactory, _ := statusComp.NewStatusComponentsFactory(statusArgs)
	managedStatusComponents, _ := statusComp.NewManagedStatusComponents(statusComponentsFactory)
	err := managedStatusComponents.Create()
	require.NoError(t, err)

	err = managedStatusComponents.CheckSubcomponents()
	require.NoError(t, err)
}
