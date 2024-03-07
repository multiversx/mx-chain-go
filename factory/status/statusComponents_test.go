package status_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/mock"
	statusComp "github.com/multiversx/mx-chain-go/factory/status"
	testsMocks "github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/genesisMocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/require"
)

func createMockStatusComponentsFactoryArgs() statusComp.StatusComponentsFactoryArgs {
	return statusComp.StatusComponentsFactoryArgs{
		Config: testscommon.GetGeneralConfig(),
		ExternalConfig: config.ExternalConfig{
			ElasticSearchConnector: config.ElasticSearchConfig{
				Enabled:        false,
				URL:            "url",
				Username:       "user",
				Password:       "pass",
				EnabledIndexes: []string{"transactions", "blocks"},
			},
			HostDriversConfig: []config.HostDriversConfig{
				{
					MarshallerType: "json",
				},
			},
			EventNotifierConnector: config.EventNotifierConfig{
				MarshallerType: "json",
			},
		},
		EconomicsConfig:    config.EconomicsConfig{},
		ShardCoordinator:   &testscommon.ShardsCoordinatorMock{},
		NodesCoordinator:   &shardingMocks.NodesCoordinatorMock{},
		EpochStartNotifier: &mock.EpochStartNotifierStub{},
		CoreComponents: &mock.CoreComponentsMock{
			NodesConfig: &genesisMocks.NodesSetupStub{
				GetRoundDurationCalled: func() uint64 {
					return 1000
				},
			},
			EpochChangeNotifier: &epochNotifier.EpochNotifierStub{},
		},
		StatusCoreComponents: &factory.StatusCoreComponentsStub{
			AppStatusHandlerField:  &statusHandler.AppStatusHandlerStub{},
			NetworkStatisticsField: &testscommon.NetworkStatisticsProviderStub{},
		},
		NetworkComponents: &testsMocks.NetworkComponentsStub{},
		StateComponents:   &mock.StateComponentsHolderStub{},
		CryptoComponents: &mock.CryptoComponentsMock{
			ManagedPeersHolderField: &testscommon.ManagedPeersHolderStub{},
		},
		IsInImportMode: false,
	}
}

func TestNewStatusComponentsFactory(t *testing.T) {
	// no t.Parallel for these tests as they create real components

	t.Run("nil CoreComponents should error", func(t *testing.T) {
		args := createMockStatusComponentsFactoryArgs()
		args.CoreComponents = nil
		scf, err := statusComp.NewStatusComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errorsMx.ErrNilCoreComponentsHolder, err)
	})
	t.Run("CoreComponents with nil GenesisNodesSetup should error", func(t *testing.T) {
		args := createMockStatusComponentsFactoryArgs()
		args.CoreComponents = &mock.CoreComponentsMock{
			NodesConfig: nil,
		}
		scf, err := statusComp.NewStatusComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errorsMx.ErrNilGenesisNodesSetupHandler, err)
	})
	t.Run("nil NetworkComponents should error", func(t *testing.T) {
		args := createMockStatusComponentsFactoryArgs()
		args.NetworkComponents = nil
		scf, err := statusComp.NewStatusComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errorsMx.ErrNilNetworkComponentsHolder, err)
	})
	t.Run("nil ShardCoordinator should error", func(t *testing.T) {
		args := createMockStatusComponentsFactoryArgs()
		args.ShardCoordinator = nil
		scf, err := statusComp.NewStatusComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errorsMx.ErrNilShardCoordinator, err)
	})
	t.Run("nil NodesCoordinator should error", func(t *testing.T) {
		args := createMockStatusComponentsFactoryArgs()
		args.NodesCoordinator = nil
		scf, err := statusComp.NewStatusComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errorsMx.ErrNilNodesCoordinator, err)
	})
	t.Run("nil EpochStartNotifier should error", func(t *testing.T) {
		args := createMockStatusComponentsFactoryArgs()
		args.EpochStartNotifier = nil
		scf, err := statusComp.NewStatusComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errorsMx.ErrNilEpochStartNotifier, err)
	})
	t.Run("nil StatusCoreComponents should error", func(t *testing.T) {
		args := createMockStatusComponentsFactoryArgs()
		args.StatusCoreComponents = nil
		scf, err := statusComp.NewStatusComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errorsMx.ErrNilStatusCoreComponents, err)
	})
	t.Run("nil CryptoComponents should error", func(t *testing.T) {
		args := createMockStatusComponentsFactoryArgs()
		args.CryptoComponents = nil
		scf, err := statusComp.NewStatusComponentsFactory(args)
		require.Nil(t, scf)
		require.Equal(t, errorsMx.ErrNilCryptoComponents, err)
	})
	t.Run("should work", func(t *testing.T) {
		scf, err := statusComp.NewStatusComponentsFactory(createMockStatusComponentsFactoryArgs())
		require.NotNil(t, scf)
		require.NoError(t, err)
	})
}

func TestStatusComponentsFactory_Create(t *testing.T) {
	// no t.Parallel for these tests as they create real components

	t.Run("NewSoftwareVersionFactory fails should return error", func(t *testing.T) {
		args := createMockStatusComponentsFactoryArgs()
		args.StatusCoreComponents = &factory.StatusCoreComponentsStub{
			AppStatusHandlerField: nil, // make NewSoftwareVersionFactory fail
		}
		scf, _ := statusComp.NewStatusComponentsFactory(args)
		require.NotNil(t, scf)

		sc, err := scf.Create()
		require.Error(t, err)
		require.Nil(t, sc)
	})
	t.Run("softwareVersionCheckerFactory.Create fails should return error", func(t *testing.T) {
		args := createMockStatusComponentsFactoryArgs()
		args.Config.SoftwareVersionConfig.PollingIntervalInMinutes = 0
		scf, _ := statusComp.NewStatusComponentsFactory(args)
		require.NotNil(t, scf)

		sc, err := scf.Create()
		require.Error(t, err)
		require.Nil(t, sc)
	})
	t.Run("invalid round duration should error", func(t *testing.T) {
		args := createMockStatusComponentsFactoryArgs()
		args.CoreComponents = &mock.CoreComponentsMock{
			NodesConfig: &genesisMocks.NodesSetupStub{
				GetRoundDurationCalled: func() uint64 {
					return 0
				},
			},
		}
		scf, _ := statusComp.NewStatusComponentsFactory(args)
		require.NotNil(t, scf)

		sc, err := scf.Create()
		require.Equal(t, errorsMx.ErrInvalidRoundDuration, err)
		require.Nil(t, sc)
	})
	t.Run("makeWebSocketDriverArgs fails due to invalid marshaller type should error", func(t *testing.T) {
		args := createMockStatusComponentsFactoryArgs()
		args.ExternalConfig.HostDriversConfig[0].Enabled = true
		args.ExternalConfig.HostDriversConfig[0].MarshallerType = "invalid type"
		scf, _ := statusComp.NewStatusComponentsFactory(args)
		require.NotNil(t, scf)

		sc, err := scf.Create()
		require.Error(t, err)
		require.Nil(t, sc)
	})
	t.Run("should work", func(t *testing.T) {
		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		shardCoordinator.SelfIDCalled = func() uint32 {
			return core.MetachainShardId // coverage
		}
		args, _ := componentsMock.GetStatusComponentsFactoryArgsAndProcessComponents(shardCoordinator)
		args.ExternalConfig.HostDriversConfig[0].Enabled = true // coverage
		scf, err := statusComp.NewStatusComponentsFactory(args)
		require.Nil(t, err)

		sc, err := scf.Create()
		require.NoError(t, err)
		require.NotNil(t, sc)

		require.NoError(t, sc.Close())
	})
}

func TestStatusComponentsFactory_epochStartEventHandler(t *testing.T) {
	// no t.Parallel for these tests as they create real components

	args := createMockStatusComponentsFactoryArgs()
	args.NodesCoordinator = &shardingMocks.NodesCoordinatorStub{
		GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
			return make(map[uint32][][]byte), errors.New("fail for coverage")
		},
	}
	scf, _ := statusComp.NewStatusComponentsFactory(args)
	require.NotNil(t, scf)

	sc, _ := scf.Create()
	require.NotNil(t, sc)

	handler := sc.EpochStartEventHandler()
	require.NotNil(t, handler)
	handler.EpochStartAction(&testscommon.HeaderHandlerStub{})
}

func TestStatusComponentsFactory_IsInterfaceNil(t *testing.T) {
	// no t.Parallel for these tests as they create real components

	args := createMockStatusComponentsFactoryArgs()
	args.CoreComponents = nil
	scf, _ := statusComp.NewStatusComponentsFactory(args)
	require.True(t, scf.IsInterfaceNil())

	scf, _ = statusComp.NewStatusComponentsFactory(createMockStatusComponentsFactoryArgs())
	require.False(t, scf.IsInterfaceNil())
}

func TestStatusComponents_Close(t *testing.T) {
	// no t.Parallel for these tests as they create real components

	scf, _ := statusComp.NewStatusComponentsFactory(createMockStatusComponentsFactoryArgs())
	cc, err := scf.Create()
	require.Nil(t, err)

	err = cc.Close()
	require.NoError(t, err)
}

func TestMakeHostDriversArgs(t *testing.T) {
	// no t.Parallel for these tests as they create real components

	args := createMockStatusComponentsFactoryArgs()
	args.ExternalConfig.HostDriversConfig = []config.HostDriversConfig{
		{
			Enabled:            false,
			URL:                "localhost",
			RetryDurationInSec: 1,
			MarshallerType:     "json",
			Mode:               data.ModeClient,
		},
		{
			Enabled:            true,
			URL:                "localhost",
			RetryDurationInSec: 1,
			MarshallerType:     "json",
			Mode:               data.ModeClient,
		},
	}
	scf, _ := statusComp.NewStatusComponentsFactory(args)
	res, err := scf.MakeHostDriversArgs()
	require.Nil(t, err)
	require.Equal(t, 1, len(res))
}
