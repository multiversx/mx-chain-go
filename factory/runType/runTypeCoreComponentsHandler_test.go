package runType_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/runType"
)

func TestNewManagedRunTypeCoreComponents(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		managedRunTypeCoreComponents, err := createCoreComponents()
		require.NoError(t, err)
		require.NotNil(t, managedRunTypeCoreComponents)
	})
}

func TestManagedRunTypeCoreComponents_Create(t *testing.T) {
	t.Parallel()

	t.Run("should work with getters", func(t *testing.T) {
		t.Parallel()

		managedRunTypeCoreComponents, err := createCoreComponents()
		require.NoError(t, err)

		require.Nil(t, managedRunTypeCoreComponents.GenesisNodesSetupFactoryCreator())
		require.Nil(t, managedRunTypeCoreComponents.RatingsDataFactoryCreator())

		err = managedRunTypeCoreComponents.Create()
		require.NoError(t, err)

		require.NotNil(t, managedRunTypeCoreComponents.GenesisNodesSetupFactoryCreator())
		require.NotNil(t, managedRunTypeCoreComponents.RatingsDataFactoryCreator())

		require.Equal(t, factory.RunTypeCoreComponentsName, managedRunTypeCoreComponents.String())
		require.NoError(t, managedRunTypeCoreComponents.Close())
	})
}

func TestManagedRunTypeCoreComponents_Close(t *testing.T) {
	t.Parallel()

	managedRunTypeCoreComponents, _ := createCoreComponents()
	require.NoError(t, managedRunTypeCoreComponents.Close())

	err := managedRunTypeCoreComponents.Create()
	require.NoError(t, err)

	require.NoError(t, managedRunTypeCoreComponents.Close())
	require.Nil(t, managedRunTypeCoreComponents.GenesisNodesSetupFactoryCreator())
	require.Nil(t, managedRunTypeCoreComponents.RatingsDataFactoryCreator())
}

func TestManagedRunTypeCoreComponents_CheckSubcomponents(t *testing.T) {
	t.Parallel()

	managedRunTypeCoreComponents, _ := createCoreComponents()
	err := managedRunTypeCoreComponents.CheckSubcomponents()
	require.Equal(t, errors.ErrNilRunTypeCoreComponents, err)

	err = managedRunTypeCoreComponents.Create()
	require.NoError(t, err)

	//TODO check for nil each subcomponent - MX-15371
	err = managedRunTypeCoreComponents.CheckSubcomponents()
	require.NoError(t, err)

	require.NoError(t, managedRunTypeCoreComponents.Close())
}

func TestManagedRunTypeCoreComponents_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var managedRunTypeCoreComponents factory.RunTypeCoreComponentsHandler
	managedRunTypeCoreComponents, _ = runType.NewManagedRunTypeCoreComponents(nil)
	require.True(t, managedRunTypeCoreComponents.IsInterfaceNil())

	managedRunTypeCoreComponents, _ = createCoreComponents()
	require.False(t, managedRunTypeCoreComponents.IsInterfaceNil())
}

func createCoreComponents() (factory.RunTypeCoreComponentsHandler, error) {
	rccf := runType.NewRunTypeCoreComponentsFactory()
	return runType.NewManagedRunTypeCoreComponents(rccf)
}
