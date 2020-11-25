package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedCoreComponents --------------------
func TestManagedCoreComponents_CreateWithInvalidArgs_ShouldErr(t *testing.T) {
	t.Parallel()

	coreArgs := getCoreArgs()
	coreArgs.Config.Marshalizer = config.MarshalizerConfig{
		Type:           "invalid_marshalizer_type",
		SizeCheckDelta: 0,
	}
	coreComponentsFactory, _ := factory.NewCoreComponentsFactory(coreArgs)
	managedCoreComponents, err := factory.NewManagedCoreComponents(coreComponentsFactory)
	require.NoError(t, err)
	err = managedCoreComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedCoreComponents.StatusHandler())
}

func TestManagedCoreComponents_Create_ShouldWork(t *testing.T) {
	t.Parallel()

	coreArgs := getCoreArgs()
	coreComponentsFactory, _ := factory.NewCoreComponentsFactory(coreArgs)
	managedCoreComponents, err := factory.NewManagedCoreComponents(coreComponentsFactory)
	require.NoError(t, err)
	require.Nil(t, managedCoreComponents.Hasher())
	require.Nil(t, managedCoreComponents.InternalMarshalizer())
	require.Nil(t, managedCoreComponents.VmMarshalizer())
	require.Nil(t, managedCoreComponents.TxMarshalizer())
	require.Nil(t, managedCoreComponents.Uint64ByteSliceConverter())
	require.Nil(t, managedCoreComponents.AddressPubKeyConverter())
	require.Nil(t, managedCoreComponents.ValidatorPubKeyConverter())
	require.Nil(t, managedCoreComponents.StatusHandler())
	require.Nil(t, managedCoreComponents.PathHandler())
	require.Equal(t, "", managedCoreComponents.ChainID())
	require.Nil(t, managedCoreComponents.AddressPubKeyConverter())

	err = managedCoreComponents.Create()
	require.NoError(t, err)
	require.NotNil(t, managedCoreComponents.Hasher())
	require.NotNil(t, managedCoreComponents.InternalMarshalizer())
	require.NotNil(t, managedCoreComponents.VmMarshalizer())
	require.NotNil(t, managedCoreComponents.TxMarshalizer())
	require.NotNil(t, managedCoreComponents.Uint64ByteSliceConverter())
	require.NotNil(t, managedCoreComponents.AddressPubKeyConverter())
	require.NotNil(t, managedCoreComponents.ValidatorPubKeyConverter())
	require.NotNil(t, managedCoreComponents.StatusHandler())
	require.NotNil(t, managedCoreComponents.PathHandler())
	require.NotEqual(t, "", managedCoreComponents.ChainID())
	require.NotNil(t, managedCoreComponents.AddressPubKeyConverter())
}

func TestManagedCoreComponents_Close(t *testing.T) {
	t.Parallel()

	coreArgs := getCoreArgs()
	coreComponentsFactory, _ := factory.NewCoreComponentsFactory(coreArgs)
	managedCoreComponents, _ := factory.NewManagedCoreComponents(coreComponentsFactory)
	err := managedCoreComponents.Create()
	require.NoError(t, err)

	err = managedCoreComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedCoreComponents.StatusHandler())
}
