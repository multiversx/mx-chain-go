package core_test

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/config"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	coreComp "github.com/multiversx/mx-chain-go/factory/core"
	"github.com/multiversx/mx-chain-go/testscommon"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

func TestManagedCoreComponents(t *testing.T) {
	t.Parallel()

	t.Run("nil factory should error", func(t *testing.T) {
		t.Parallel()

		managedCoreComponents, err := coreComp.NewManagedCoreComponents(nil)
		require.Equal(t, errorsMx.ErrNilCoreComponentsFactory, err)
		require.Nil(t, managedCoreComponents)
	})
	t.Run("invalid args should error", func(t *testing.T) {
		t.Parallel()

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
		require.Nil(t, managedCoreComponents.InternalMarshalizer())
	})
	t.Run("should work with getters", func(t *testing.T) {
		t.Parallel()

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
		require.Nil(t, managedCoreComponents.PathHandler())
		require.Equal(t, "", managedCoreComponents.ChainID())
		require.Nil(t, managedCoreComponents.AddressPubKeyConverter())
		require.Nil(t, managedCoreComponents.EnableRoundsHandler())
		require.Nil(t, managedCoreComponents.WasmVMChangeLocker())
		require.Nil(t, managedCoreComponents.ProcessStatusHandler())
		require.True(t, len(managedCoreComponents.HardforkTriggerPubKey()) == 0)
		require.Nil(t, managedCoreComponents.TxSignHasher())
		require.Zero(t, managedCoreComponents.MinTransactionVersion())
		require.Nil(t, managedCoreComponents.TxVersionChecker())
		require.Zero(t, managedCoreComponents.EncodedAddressLen())
		require.Nil(t, managedCoreComponents.AlarmScheduler())
		require.Nil(t, managedCoreComponents.SyncTimer())
		require.Equal(t, time.Time{}, managedCoreComponents.GenesisTime())
		require.Nil(t, managedCoreComponents.Watchdog())
		require.Nil(t, managedCoreComponents.EconomicsData())
		require.Nil(t, managedCoreComponents.APIEconomicsData())
		require.Nil(t, managedCoreComponents.RatingsData())
		require.Nil(t, managedCoreComponents.Rater())
		require.Nil(t, managedCoreComponents.GenesisNodesSetup())
		require.Nil(t, managedCoreComponents.RoundHandler())
		require.Nil(t, managedCoreComponents.NodesShuffler())
		require.Nil(t, managedCoreComponents.EpochNotifier())
		require.Nil(t, managedCoreComponents.EpochStartNotifierWithConfirm())
		require.Nil(t, managedCoreComponents.ChanStopNodeProcess())
		require.Nil(t, managedCoreComponents.NodeTypeProvider())
		require.Nil(t, managedCoreComponents.EnableEpochsHandler())

		err = managedCoreComponents.Create()
		require.NoError(t, err)
		require.NotNil(t, managedCoreComponents.Hasher())
		require.NotNil(t, managedCoreComponents.InternalMarshalizer())
		require.NotNil(t, managedCoreComponents.VmMarshalizer())
		require.NotNil(t, managedCoreComponents.TxMarshalizer())
		require.NotNil(t, managedCoreComponents.Uint64ByteSliceConverter())
		require.NotNil(t, managedCoreComponents.AddressPubKeyConverter())
		require.NotNil(t, managedCoreComponents.ValidatorPubKeyConverter())
		require.NotNil(t, managedCoreComponents.PathHandler())
		require.NotEqual(t, "", managedCoreComponents.ChainID())
		require.NotNil(t, managedCoreComponents.AddressPubKeyConverter())
		require.NotNil(t, managedCoreComponents.EnableRoundsHandler())
		require.NotNil(t, managedCoreComponents.WasmVMChangeLocker())
		require.NotNil(t, managedCoreComponents.ProcessStatusHandler())
		expectedBytes, _ := managedCoreComponents.ValidatorPubKeyConverter().Decode(componentsMock.DummyPk)
		require.Equal(t, expectedBytes, managedCoreComponents.HardforkTriggerPubKey())
		require.NotNil(t, managedCoreComponents.TxSignHasher())
		require.NotZero(t, managedCoreComponents.MinTransactionVersion())
		require.NotNil(t, managedCoreComponents.TxVersionChecker())
		require.NotZero(t, managedCoreComponents.EncodedAddressLen())
		require.NotNil(t, managedCoreComponents.AlarmScheduler())
		require.NotNil(t, managedCoreComponents.SyncTimer())
		require.NotNil(t, managedCoreComponents.GenesisTime())
		require.NotNil(t, managedCoreComponents.Watchdog())
		require.NotNil(t, managedCoreComponents.EconomicsData())
		require.NotNil(t, managedCoreComponents.APIEconomicsData())
		require.NotNil(t, managedCoreComponents.RatingsData())
		require.NotNil(t, managedCoreComponents.Rater())
		require.NotNil(t, managedCoreComponents.GenesisNodesSetup())
		require.NotNil(t, managedCoreComponents.RoundHandler())
		require.NotNil(t, managedCoreComponents.NodesShuffler())
		require.NotNil(t, managedCoreComponents.EpochNotifier())
		require.NotNil(t, managedCoreComponents.EpochStartNotifierWithConfirm())
		require.NotNil(t, managedCoreComponents.ChanStopNodeProcess())
		require.NotNil(t, managedCoreComponents.NodeTypeProvider())
		require.NotNil(t, managedCoreComponents.EnableEpochsHandler())
		require.Nil(t, managedCoreComponents.SetInternalMarshalizer(&testscommon.MarshalizerStub{}))

		require.Equal(t, factory.CoreComponentsName, managedCoreComponents.String())
	})
}

func TestManagedCoreComponents_CheckSubcomponents(t *testing.T) {
	t.Parallel()

	coreArgs := componentsMock.GetCoreArgs()
	coreComponentsFactory, _ := coreComp.NewCoreComponentsFactory(coreArgs)
	managedCoreComponents, err := coreComp.NewManagedCoreComponents(coreComponentsFactory)
	require.NoError(t, err)
	require.Equal(t, errorsMx.ErrNilCoreComponents, managedCoreComponents.CheckSubcomponents())

	err = managedCoreComponents.Create()
	require.NoError(t, err)
	require.Nil(t, managedCoreComponents.CheckSubcomponents())
}

func TestManagedCoreComponents_Close(t *testing.T) {
	t.Parallel()

	coreComponentsFactory, _ := coreComp.NewCoreComponentsFactory(componentsMock.GetCoreArgs())
	managedCoreComponents, _ := coreComp.NewManagedCoreComponents(coreComponentsFactory)
	err := managedCoreComponents.Close()
	require.NoError(t, err)
	err = managedCoreComponents.Create()
	require.NoError(t, err)
	err = managedCoreComponents.Close()
	require.NoError(t, err)
}

func TestManagedCoreComponents_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	managedCoreComponents, _ := coreComp.NewManagedCoreComponents(nil)
	require.True(t, managedCoreComponents.IsInterfaceNil())

	coreArgs := componentsMock.GetCoreArgs()
	coreComponentsFactory, _ := coreComp.NewCoreComponentsFactory(coreArgs)
	managedCoreComponents, _ = coreComp.NewManagedCoreComponents(coreComponentsFactory)
	require.False(t, managedCoreComponents.IsInterfaceNil())
}
