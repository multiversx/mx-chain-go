package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/stretchr/testify/require"
)

// ------------ Test TestManagedProcessComponents --------------------
func TestManagedProcessComponents_CreateWithInvalidArgs_ShouldErr(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	processArgs := getProcessComponentsArgs(shardCoordinator)
	_ = processArgs.CoreData.SetInternalMarshalizer(nil)
	processComponentsFactory, _ := factory.NewProcessComponentsFactory(processArgs)
	managedProcessComponents, err := factory.NewManagedProcessComponents(processComponentsFactory)
	require.NoError(t, err)
	err = managedProcessComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedProcessComponents.NodesCoordinator())
}

func TestManagedProcessComponents_Create_ShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents := getCoreComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	shardCoordinator.SelfIDCalled = func() uint32 {
		return core.MetachainShardId
	}
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if core.IsSmartContractOnMetachain(address[len(address)-1:], address) {
			return core.MetachainShardId
		}

		return 0
	}

	shardCoordinator.CurrentShard = core.MetachainShardId
	dataComponents := getDataComponents(coreComponents, shardCoordinator)
	networkComponents := getNetworkComponents()
	cryptoComponents := getCryptoComponents(coreComponents)
	stateComponents := getStateComponents(coreComponents, shardCoordinator)
	processArgs := getProcessArgs(
		shardCoordinator,
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
		networkComponents,
	)

	factory.SetShardCoordinator(shardCoordinator, processArgs.BootstrapComponents)

	processComponentsFactory, err := factory.NewProcessComponentsFactory(processArgs)
	require.Nil(t, err)
	managedProcessComponents, err := factory.NewManagedProcessComponents(processComponentsFactory)
	require.NoError(t, err)
	require.True(t, check.IfNil(managedProcessComponents.NodesCoordinator()))
	require.True(t, check.IfNil(managedProcessComponents.InterceptorsContainer()))
	require.True(t, check.IfNil(managedProcessComponents.ResolversFinder()))
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
	require.True(t, check.IfNilReflect(managedProcessComponents.ArwenChangeLocker()))
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

	err = managedProcessComponents.Create()
	require.NoError(t, err)
	require.False(t, check.IfNil(managedProcessComponents.NodesCoordinator()))
	require.False(t, check.IfNil(managedProcessComponents.InterceptorsContainer()))
	require.False(t, check.IfNil(managedProcessComponents.ResolversFinder()))
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
	require.False(t, check.IfNilReflect(managedProcessComponents.ArwenChangeLocker()))
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
}

func TestManagedProcessComponents_Close(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	processArgs := getProcessComponentsArgs(shardCoordinator)
	processComponentsFactory, _ := factory.NewProcessComponentsFactory(processArgs)
	managedProcessComponents, _ := factory.NewManagedProcessComponents(processComponentsFactory)
	err := managedProcessComponents.Create()
	require.NoError(t, err)

	err = managedProcessComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedProcessComponents.NodesCoordinator())
}
