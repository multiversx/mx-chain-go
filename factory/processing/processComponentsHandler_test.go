package processing_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	processComp "github.com/multiversx/mx-chain-go/factory/processing"
	"github.com/stretchr/testify/require"
)

func TestNewManagedProcessComponents(t *testing.T) {
	t.Parallel()

	t.Run("nil factory should error", func(t *testing.T) {
		t.Parallel()

		managedProcessComponents, err := processComp.NewManagedProcessComponents(nil)
		require.Equal(t, errorsMx.ErrNilProcessComponentsFactory, err)
		require.Nil(t, managedProcessComponents)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		processComponentsFactory, _ := processComp.NewProcessComponentsFactory(createMockProcessComponentsFactoryArgs())
		managedProcessComponents, err := processComp.NewManagedProcessComponents(processComponentsFactory)
		require.NoError(t, err)
		require.NotNil(t, managedProcessComponents)
	})
}

func TestManagedProcessComponents_Create(t *testing.T) {
	t.Parallel()

	t.Run("invalid params should error", func(t *testing.T) {
		t.Parallel()

		args := createMockProcessComponentsFactoryArgs()
		args.Config.PublicKeyPeerId.Type = "invalid"
		processComponentsFactory, _ := processComp.NewProcessComponentsFactory(args)
		managedProcessComponents, _ := processComp.NewManagedProcessComponents(processComponentsFactory)
		require.NotNil(t, managedProcessComponents)

		err := managedProcessComponents.Create()
		require.Error(t, err)
	})
	t.Run("should work with getters", func(t *testing.T) {
		t.Parallel()

		processComponentsFactory, _ := processComp.NewProcessComponentsFactory(createMockProcessComponentsFactoryArgs())
		managedProcessComponents, _ := processComp.NewManagedProcessComponents(processComponentsFactory)
		require.NotNil(t, managedProcessComponents)

		require.True(t, check.IfNil(managedProcessComponents.NodesCoordinator()))
		require.True(t, check.IfNil(managedProcessComponents.InterceptorsContainer()))
		require.True(t, check.IfNil(managedProcessComponents.ResolversContainer()))
		require.True(t, check.IfNil(managedProcessComponents.RequestersFinder()))
		require.True(t, check.IfNil(managedProcessComponents.RoundHandler()))
		require.True(t, check.IfNil(managedProcessComponents.ForkDetector()))
		require.True(t, check.IfNil(managedProcessComponents.BlockProcessor()))
		require.True(t, check.IfNil(managedProcessComponents.EpochStartTrigger()))
		require.True(t, check.IfNil(managedProcessComponents.EpochStartNotifier()))
		require.True(t, check.IfNil(managedProcessComponents.BlackListHandler()))
		require.True(t, check.IfNil(managedProcessComponents.BootStorer()))
		require.True(t, check.IfNil(managedProcessComponents.HeaderSigVerifier()))
		require.True(t, check.IfNil(managedProcessComponents.ValidatorsStatistics()))
		require.True(t, check.IfNil(managedProcessComponents.ValidatorsProvider()))
		require.True(t, check.IfNil(managedProcessComponents.BlockTracker()))
		require.True(t, check.IfNil(managedProcessComponents.PendingMiniBlocksHandler()))
		require.True(t, check.IfNil(managedProcessComponents.RequestHandler()))
		require.True(t, check.IfNil(managedProcessComponents.TxLogsProcessor()))
		require.True(t, check.IfNil(managedProcessComponents.HeaderConstructionValidator()))
		require.True(t, check.IfNil(managedProcessComponents.HeaderIntegrityVerifier()))
		require.True(t, check.IfNil(managedProcessComponents.CurrentEpochProvider()))
		require.True(t, check.IfNil(managedProcessComponents.NodeRedundancyHandler()))
		require.True(t, check.IfNil(managedProcessComponents.WhiteListHandler()))
		require.True(t, check.IfNil(managedProcessComponents.WhiteListerVerifiedTxs()))
		require.True(t, check.IfNil(managedProcessComponents.RequestedItemsHandler()))
		require.True(t, check.IfNil(managedProcessComponents.ImportStartHandler()))
		require.True(t, check.IfNil(managedProcessComponents.HistoryRepository()))
		require.True(t, check.IfNil(managedProcessComponents.TransactionSimulatorProcessor()))
		require.True(t, check.IfNil(managedProcessComponents.FallbackHeaderValidator()))
		require.True(t, check.IfNil(managedProcessComponents.PeerShardMapper()))
		require.True(t, check.IfNil(managedProcessComponents.ShardCoordinator()))
		require.True(t, check.IfNil(managedProcessComponents.TxsSenderHandler()))
		require.True(t, check.IfNil(managedProcessComponents.HardforkTrigger()))
		require.True(t, check.IfNil(managedProcessComponents.ProcessedMiniBlocksTracker()))
		require.True(t, check.IfNil(managedProcessComponents.AccountsParser()))
		require.True(t, check.IfNil(managedProcessComponents.ScheduledTxsExecutionHandler()))
		require.True(t, check.IfNil(managedProcessComponents.ESDTDataStorageHandlerForAPI()))
		require.True(t, check.IfNil(managedProcessComponents.ReceiptsRepository()))
		require.True(t, check.IfNil(managedProcessComponents.FullArchivePeerShardMapper()))

		err := managedProcessComponents.Create()
		require.NoError(t, err)
		require.False(t, check.IfNil(managedProcessComponents.NodesCoordinator()))
		require.False(t, check.IfNil(managedProcessComponents.InterceptorsContainer()))
		require.False(t, check.IfNil(managedProcessComponents.ResolversContainer()))
		require.False(t, check.IfNil(managedProcessComponents.RequestersFinder()))
		require.False(t, check.IfNil(managedProcessComponents.RoundHandler()))
		require.False(t, check.IfNil(managedProcessComponents.ForkDetector()))
		require.False(t, check.IfNil(managedProcessComponents.BlockProcessor()))
		require.False(t, check.IfNil(managedProcessComponents.EpochStartTrigger()))
		require.False(t, check.IfNil(managedProcessComponents.EpochStartNotifier()))
		require.False(t, check.IfNil(managedProcessComponents.BlackListHandler()))
		require.False(t, check.IfNil(managedProcessComponents.BootStorer()))
		require.False(t, check.IfNil(managedProcessComponents.HeaderSigVerifier()))
		require.False(t, check.IfNil(managedProcessComponents.ValidatorsStatistics()))
		require.False(t, check.IfNil(managedProcessComponents.ValidatorsProvider()))
		require.False(t, check.IfNil(managedProcessComponents.BlockTracker()))
		require.False(t, check.IfNil(managedProcessComponents.PendingMiniBlocksHandler()))
		require.False(t, check.IfNil(managedProcessComponents.RequestHandler()))
		require.False(t, check.IfNil(managedProcessComponents.TxLogsProcessor()))
		require.False(t, check.IfNil(managedProcessComponents.HeaderConstructionValidator()))
		require.False(t, check.IfNil(managedProcessComponents.HeaderIntegrityVerifier()))
		require.False(t, check.IfNil(managedProcessComponents.CurrentEpochProvider()))
		require.False(t, check.IfNil(managedProcessComponents.NodeRedundancyHandler()))
		require.False(t, check.IfNil(managedProcessComponents.WhiteListHandler()))
		require.False(t, check.IfNil(managedProcessComponents.WhiteListerVerifiedTxs()))
		require.False(t, check.IfNil(managedProcessComponents.RequestedItemsHandler()))
		require.False(t, check.IfNil(managedProcessComponents.ImportStartHandler()))
		require.False(t, check.IfNil(managedProcessComponents.HistoryRepository()))
		require.False(t, check.IfNil(managedProcessComponents.TransactionSimulatorProcessor()))
		require.False(t, check.IfNil(managedProcessComponents.FallbackHeaderValidator()))
		require.False(t, check.IfNil(managedProcessComponents.PeerShardMapper()))
		require.False(t, check.IfNil(managedProcessComponents.ShardCoordinator()))
		require.False(t, check.IfNil(managedProcessComponents.TxsSenderHandler()))
		require.False(t, check.IfNil(managedProcessComponents.HardforkTrigger()))
		require.False(t, check.IfNil(managedProcessComponents.ProcessedMiniBlocksTracker()))
		require.False(t, check.IfNil(managedProcessComponents.AccountsParser()))
		require.False(t, check.IfNil(managedProcessComponents.ScheduledTxsExecutionHandler()))
		require.False(t, check.IfNil(managedProcessComponents.ESDTDataStorageHandlerForAPI()))
		require.False(t, check.IfNil(managedProcessComponents.ReceiptsRepository()))
		require.False(t, check.IfNil(managedProcessComponents.FullArchivePeerShardMapper()))

		require.Equal(t, factory.ProcessComponentsName, managedProcessComponents.String())
	})
}

func TestManagedProcessComponents_CheckSubcomponents(t *testing.T) {
	t.Parallel()

	processComponentsFactory, _ := processComp.NewProcessComponentsFactory(createMockProcessComponentsFactoryArgs())
	managedProcessComponents, _ := processComp.NewManagedProcessComponents(processComponentsFactory)
	require.NotNil(t, managedProcessComponents)
	require.Equal(t, errorsMx.ErrNilProcessComponents, managedProcessComponents.CheckSubcomponents())

	err := managedProcessComponents.Create()
	require.NoError(t, err)

	require.Nil(t, managedProcessComponents.CheckSubcomponents())
}

func TestManagedProcessComponents_Close(t *testing.T) {
	t.Parallel()

	processComponentsFactory, _ := processComp.NewProcessComponentsFactory(createMockProcessComponentsFactoryArgs())
	managedProcessComponents, _ := processComp.NewManagedProcessComponents(processComponentsFactory)
	err := managedProcessComponents.Create()
	require.NoError(t, err)

	err = managedProcessComponents.Close()
	require.NoError(t, err)

	err = managedProcessComponents.Close()
	require.NoError(t, err)
}

func TestManagedProcessComponents_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	managedProcessComponents, _ := processComp.NewManagedProcessComponents(nil)
	require.True(t, managedProcessComponents.IsInterfaceNil())

	processComponentsFactory, _ := processComp.NewProcessComponentsFactory(createMockProcessComponentsFactoryArgs())
	managedProcessComponents, _ = processComp.NewManagedProcessComponents(processComponentsFactory)
	require.False(t, managedProcessComponents.IsInterfaceNil())
}
