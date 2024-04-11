package components

import (
	"testing"

	"github.com/stretchr/testify/require"

	disabledStatistics "github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/config"
	mocks "github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/guardianMocks"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
)

func createArgsDataComponentsHolder() ArgsDataComponentsHolder {
	generalCfg := testscommon.GetGeneralConfig()
	return ArgsDataComponentsHolder{
		Configs: config.Configs{
			GeneralConfig:     &generalCfg,
			ImportDbConfig:    &config.ImportDbConfig{},
			PreferencesConfig: &config.Preferences{},
			FlagsConfig:       &config.ContextFlagsConfig{},
		},
		CoreComponents: &factory.CoreComponentsHolderMock{},
		StatusCoreComponents: &factory.StatusCoreComponentsStub{
			AppStatusHandlerField:  &statusHandler.AppStatusHandlerStub{},
			StateStatsHandlerField: disabledStatistics.NewStateStatistics(),
		},
		BootstrapComponents: &mainFactoryMocks.BootstrapComponentsStub{
			ShCoordinator:              testscommon.NewMultiShardsCoordinatorMock(2),
			BootstrapParams:            &bootstrapMocks.BootstrapParamsHandlerMock{},
			HdrIntegrityVerifier:       &mocks.HeaderIntegrityVerifierStub{},
			GuardedAccountHandlerField: &guardianMocks.GuardedAccountHandlerStub{},
			VersionedHdrFactory:        &testscommon.VersionedHeaderFactoryStub{},
		},
		CryptoComponents: &mocks.CryptoComponentsStub{
			PubKey:                  &mocks.PublicKeyMock{},
			BlockSig:                &cryptoMocks.SingleSignerStub{},
			BlKeyGen:                &cryptoMocks.KeyGenStub{},
			TxSig:                   &cryptoMocks.SingleSignerStub{},
			TxKeyGen:                &cryptoMocks.KeyGenStub{},
			ManagedPeersHolderField: &testscommon.ManagedPeersHolderStub{},
		},
		RunTypeComponents: mainFactoryMocks.NewRunTypeComponentsStub(),
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
	//t.Run("NewMiniBlockProvider failure should error", func(t *testing.T) {
	//	t.Parallel()
	//
	//	args := createArgsDataComponentsHolder()
	//	args = &dataRetriever.PoolsHolderStub{
	//		MiniBlocksCalled: func() chainStorage.Cacher {
	//			return nil
	//		},
	//	}
	//	comp, err := CreateDataComponents(args)
	//	require.Error(t, err)
	//	require.Nil(t, comp)
	//})
	//t.Run("GetStorer failure should error", func(t *testing.T) {
	//	t.Parallel()
	//
	//	args := createArgsDataComponentsHolder()
	//	args.StorageService = &storage.ChainStorerStub{
	//		GetStorerCalled: func(unitType retriever.UnitType) (chainStorage.Storer, error) {
	//			return nil, expectedErr
	//		},
	//	}
	//	comp, err := CreateDataComponents(args)
	//	require.Equal(t, expectedErr, err)
	//	require.Nil(t, comp)
	//})
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
