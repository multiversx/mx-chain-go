package components

import (
	"testing"

	disabledStatistics "github.com/multiversx/mx-chain-go/common/statistics/disabled"
	mockFactory "github.com/multiversx/mx-chain-go/factory/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/require"
)

func createArgsStateComponents() ArgsStateComponents {
	return ArgsStateComponents{
		Config: testscommon.GetGeneralConfig(),
		CoreComponents: &mockFactory.CoreComponentsMock{
			IntMarsh:                     &testscommon.MarshallerStub{},
			Hash:                         &testscommon.HasherStub{},
			PathHdl:                      &testscommon.PathManagerStub{},
			ProcessStatusHandlerInternal: &testscommon.ProcessStatusHandlerStub{},
			EnableEpochsHandlerField:     &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
			AddrPubKeyConv:               &testscommon.PubkeyConverterStub{},
		},
		StatusCore: &factory.StatusCoreComponentsStub{
			AppStatusHandlerField:  &statusHandler.AppStatusHandlerStub{},
			StateStatsHandlerField: disabledStatistics.NewStateStatistics(),
		},
		StoreService: genericMocks.NewChainStorerMock(0),
		ChainHandler: &testscommon.ChainHandlerStub{},
	}
}

func TestCreateStateComponents(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		comp, err := CreateStateComponents(createArgsStateComponents())
		require.NoError(t, err)
		require.NotNil(t, comp)

		require.Nil(t, comp.Create())
		require.Nil(t, comp.Close())
	})
	t.Run("NewStateComponentsFactory failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsStateComponents()
		args.CoreComponents = nil
		comp, err := CreateStateComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
	t.Run("stateComp.Create failure should error", func(t *testing.T) {
		t.Parallel()

		args := createArgsStateComponents()
		coreMock, ok := args.CoreComponents.(*mockFactory.CoreComponentsMock)
		require.True(t, ok)
		coreMock.EnableEpochsHandlerField = nil
		comp, err := CreateStateComponents(args)
		require.Error(t, err)
		require.Nil(t, comp)
	})
}

func TestStateComponentsHolder_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var comp *stateComponentsHolder
	require.True(t, comp.IsInterfaceNil())

	comp, _ = CreateStateComponents(createArgsStateComponents())
	require.False(t, comp.IsInterfaceNil())
	require.Nil(t, comp.Close())
}

func TestStateComponentsHolder_Getters(t *testing.T) {
	t.Parallel()

	comp, err := CreateStateComponents(createArgsStateComponents())
	require.NoError(t, err)

	require.NotNil(t, comp.PeerAccounts())
	require.NotNil(t, comp.AccountsAdapter())
	require.NotNil(t, comp.AccountsAdapterAPI())
	require.NotNil(t, comp.AccountsRepository())
	require.NotNil(t, comp.TriesContainer())
	require.NotNil(t, comp.TrieStorageManagers())
	require.NotNil(t, comp.MissingTrieNodesNotifier())
	require.Nil(t, comp.CheckSubcomponents())
	require.Empty(t, comp.String())

	require.Nil(t, comp.Close())
}
