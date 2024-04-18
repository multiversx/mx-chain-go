package components

import (
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	disabledStatistics "github.com/multiversx/mx-chain-go/common/statistics/disabled"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/consensus"
	factoryErrors "github.com/multiversx/mx-chain-go/factory"
	integrationTestsMock "github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/genesisMocks"
	"github.com/multiversx/mx-chain-go/testscommon/guardianMocks"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	updateMocks "github.com/multiversx/mx-chain-go/update/mock"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/require"
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
		CoreComponents: &factory.CoreComponentsHolderMock{
			ChainIDCalled: func() string {
				return "T"
			},
			GenesisNodesSetupCalled: func() sharding.GenesisNodesSetupHandler {
				return &genesisMocks.NodesSetupStub{}
			},
			InternalMarshalizerCalled: func() marshal.Marshalizer {
				return &testscommon.MarshallerStub{}
			},
			EpochNotifierCalled: func() process.EpochNotifier {
				return &epochNotifier.EpochNotifierStub{}
			},
			EconomicsDataCalled: func() process.EconomicsDataHandler {
				return &economicsmocks.EconomicsHandlerMock{
					MinGasPriceCalled: func() uint64 {
						return 1000000
					},
				}
			},
			EpochStartNotifierWithConfirmCalled: func() factoryErrors.EpochStartNotifierWithConfirm {
				return &updateMocks.EpochStartNotifierStub{}
			},
			RaterCalled: func() sharding.PeerAccountListAndRatingHandler {
				return &testscommon.RaterMock{}
			},
			NodesShufflerCalled: func() nodesCoordinator.NodesShuffler {
				return &shardingMocks.NodeShufflerMock{}
			},
			RoundHandlerCalled: func() consensus.RoundHandler {
				return &testscommon.RoundHandlerMock{}
			},
			HasherCalled: func() hashing.Hasher {
				return &testscommon.HasherStub{}
			},
			PathHandlerCalled: func() storage.PathManagerHandler {
				return &testscommon.PathManagerStub{}
			},
			TxMarshalizerCalled: func() marshal.Marshalizer {
				return &testscommon.MarshallerStub{}
			},
			AddressPubKeyConverterCalled: func() core.PubkeyConverter {
				return &testscommon.PubkeyConverterStub{}
			},
			Uint64ByteSliceConverterCalled: func() typeConverters.Uint64ByteSliceConverter {
				return &mock.Uint64ByteSliceConverterMock{}
			},
			TxSignHasherCalled: func() hashing.Hasher {
				return &testscommon.HasherStub{}
			},
			EnableEpochsHandlerCalled: func() common.EnableEpochsHandler {
				return &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
			},
			NodeTypeProviderCalled: func() core.NodeTypeProviderHandler {
				return &nodeTypeProviderMock.NodeTypeProviderStub{}
			},
		},
		StatusCoreComponents: &factory.StatusCoreComponentsStub{
			AppStatusHandlerField:  &statusHandler.AppStatusHandlerStub{},
			StateStatsHandlerField: disabledStatistics.NewStateStatistics(),
		},
		BootstrapComponents: &mainFactoryMocks.BootstrapComponentsStub{
			ShCoordinator:              testscommon.NewMultiShardsCoordinatorMock(2),
			BootstrapParams:            &bootstrapMocks.BootstrapParamsHandlerMock{},
			HdrIntegrityVerifier:       &mock.HeaderIntegrityVerifierStub{},
			GuardedAccountHandlerField: &guardianMocks.GuardedAccountHandlerStub{},
			VersionedHdrFactory:        &testscommon.VersionedHeaderFactoryStub{},
		},
		CryptoComponents: &integrationTestsMock.CryptoComponentsStub{
			PubKey:                  &integrationTestsMock.PublicKeyMock{},
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
