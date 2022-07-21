package core_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	coreComp "github.com/ElrondNetwork/elrond-go/factory/core"
	componentsMock "github.com/ElrondNetwork/elrond-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

// ------------ Test ManagedCoreComponents --------------------
func TestManagedCoreComponents_CreateWithInvalidArgsShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreArgs := componentsMock.GetCoreArgs()
	coreArgs.Config.Marshalizer = config.MarshalizerConfig{
		Type:           "invalid_marshalizer_type",
		SizeCheckDelta: 0,
	}
	coreComponentsFactory, _ := coreComp.NewCoreComponentsFactory(coreArgs)
	managedCoreComponents, err := coreComp.NewManagedCoreComponents(coreComponentsFactory)
	require.NoError(t, err)
	err = managedCoreComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedCoreComponents.StatusHandler())
}

func TestManagedCoreComponents_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreArgs := componentsMock.GetCoreArgs()
	coreComponentsFactory, _ := coreComp.NewCoreComponentsFactory(coreArgs)
	managedCoreComponents, err := coreComp.NewManagedCoreComponents(coreComponentsFactory)
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
	require.Nil(t, managedCoreComponents.RoundNotifier())
	require.Nil(t, managedCoreComponents.ArwenChangeLocker())
	require.Nil(t, managedCoreComponents.ProcessStatusHandler())
	require.True(t, len(managedCoreComponents.HardforkTriggerPubKey()) == 0)

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
	require.NotNil(t, managedCoreComponents.RoundNotifier())
	require.NotNil(t, managedCoreComponents.ArwenChangeLocker())
	require.NotNil(t, managedCoreComponents.ProcessStatusHandler())
	expectedBytes, _ := managedCoreComponents.ValidatorPubKeyConverter().Decode(componentsMock.DummyPk)
	require.Equal(t, expectedBytes, managedCoreComponents.HardforkTriggerPubKey())
}

func TestManagedCoreComponents_Close(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreArgs := componentsMock.GetCoreArgs()
	coreComponentsFactory, _ := coreComp.NewCoreComponentsFactory(coreArgs)
	managedCoreComponents, _ := coreComp.NewManagedCoreComponents(coreComponentsFactory)
	err := managedCoreComponents.Create()
	require.NoError(t, err)

	err = managedCoreComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedCoreComponents.StatusHandler())
}
