package runType_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/runType"
)

func createCoreComponents() (factory.RunTypeCoreComponentsHandler, error) {
	rccf := runType.NewRunTypeCoreComponentsFactory()
	return runType.NewManagedRunTypeCoreComponents(rccf)
}

func TestNewManagedRunTypeCoreComponents(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		managedRunTypeCoreComponents, err := runType.NewManagedRunTypeCoreComponents(nil)
		require.ErrorIs(t, err, errors.ErrNilRunTypeCoreComponentsFactory)
		require.True(t, managedRunTypeCoreComponents.IsInterfaceNil())
	})
	t.Run("should work", func(t *testing.T) {
		rccf := runType.NewRunTypeCoreComponentsFactory()
		managedRunTypeCoreComponents, err := runType.NewManagedRunTypeCoreComponents(rccf)
		require.NoError(t, err)
		require.False(t, managedRunTypeCoreComponents.IsInterfaceNil())
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
		require.Nil(t, managedRunTypeCoreComponents.EnableEpochsFactoryCreator())

		err = managedRunTypeCoreComponents.Create()
		require.NoError(t, err)

		require.NotNil(t, managedRunTypeCoreComponents.GenesisNodesSetupFactoryCreator())
		require.NotNil(t, managedRunTypeCoreComponents.RatingsDataFactoryCreator())
		require.NotNil(t, managedRunTypeCoreComponents.EnableEpochsFactoryCreator())

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
