package components

import (
	"testing"

	disabledStatistics "github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/statusHandler/persister"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks"
	"github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/guardianMocks"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"

	"github.com/stretchr/testify/require"
)

func createArgsDataComponentsHolder() ArgsDataComponentsHolder {
	generalCfg := testscommon.GetGeneralConfig()
	persistentStatusHandler, _ := persister.NewPersistentStatus(&testscommon.MarshallerStub{}, &mock.Uint64ByteSliceConverterMock{})
	coreComp := createCoreComponents()
	coreComp.EconomicsDataCalled = func() process.EconomicsDataHandler {
		return &economicsmocks.EconomicsHandlerMock{
			MinGasPriceCalled: func() uint64 {
				return 1000000
			},
		}
	}
	cryptoComp := createCryptoComponents()

	return ArgsDataComponentsHolder{
		Configs: config.Configs{
			GeneralConfig:     &generalCfg,
			ImportDbConfig:    &config.ImportDbConfig{},
			PreferencesConfig: &config.Preferences{},
			FlagsConfig:       &config.ContextFlagsConfig{},
		},
		CoreComponents: coreComp,
		StatusCoreComponents: &factory.StatusCoreComponentsStub{
			AppStatusHandlerField:        &statusHandler.AppStatusHandlerStub{},
			StateStatsHandlerField:       disabledStatistics.NewStateStatistics(),
			PersistentStatusHandlerField: persistentStatusHandler,
		},
		BootstrapComponents: &mainFactoryMocks.BootstrapComponentsStub{
			ShCoordinator:              testscommon.NewMultiShardsCoordinatorMock(2),
			BootstrapParams:            &bootstrapMocks.BootstrapParamsHandlerMock{},
			HdrIntegrityVerifier:       &mock.HeaderIntegrityVerifierStub{},
			GuardedAccountHandlerField: &guardianMocks.GuardedAccountHandlerStub{},
			VersionedHdrFactory:        &testscommon.VersionedHeaderFactoryStub{},
		},
		CryptoComponents:  cryptoComp,
		RunTypeComponents: components.GetRunTypeComponents(coreComp, cryptoComp),
	}
}

func TestCreateDataComponents(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		comp, err := CreateDataComponents(createArgsDataComponentsHolder())
		require.NoError(t, err)
		require.NotNil(t, comp)

		require.Nil(t, comp.Create())
		require.Nil(t, comp.Close())
	})
	t.Run("nil core components should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsDataComponentsHolder()
		args.CoreComponents = nil
		comp, err := CreateDataComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("nil status components should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsDataComponentsHolder()
		args.StatusCoreComponents = nil
		comp, err := CreateDataComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("nil bootstrap components should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsDataComponentsHolder()
		args.BootstrapComponents = nil
		comp, err := CreateDataComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("nil crypto components should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsDataComponentsHolder()
		args.CryptoComponents = nil
		comp, err := CreateDataComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("nil runtype components should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsDataComponentsHolder()
		args.RunTypeComponents = nil
		comp, err := CreateDataComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
}

func TestDataComponentsHolder_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var comp *dataComponentsHolder
	require.True(t, comp.IsInterfaceNil())

	comp, _ = CreateDataComponents(createArgsDataComponentsHolder())
	require.False(t, comp.IsInterfaceNil())
	require.Nil(t, comp.Close())
}

func TestDataComponentsHolder_Getters(t *testing.T) {
	t.Parallel()

	comp, err := CreateDataComponents(createArgsDataComponentsHolder())
	require.NoError(t, err)

	require.NotNil(t, comp.Blockchain())
	require.Nil(t, comp.SetBlockchain(nil))
	require.Nil(t, comp.Blockchain())
	require.NotNil(t, comp.StorageService())
	require.NotNil(t, comp.Datapool())
	require.NotNil(t, comp.MiniBlocksProvider())
	require.Nil(t, comp.CheckSubcomponents())
	require.Empty(t, comp.String())
	require.Nil(t, comp.Close())
}

func TestDataComponentsHolder_Clone(t *testing.T) {
	t.Parallel()

	comp, err := CreateDataComponents(createArgsDataComponentsHolder())
	require.NoError(t, err)

	compClone := comp.Clone()
	require.Equal(t, comp, compClone)
	require.False(t, comp == compClone) // pointer testing
	require.Nil(t, comp.Close())
}
