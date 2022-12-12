package processing_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	processComp "github.com/ElrondNetwork/elrond-go/factory/processing"
	componentsMock "github.com/ElrondNetwork/elrond-go/testscommon/components"
	"github.com/stretchr/testify/require"
)

// ------------ Test TestManagedProcessComponents --------------------
func TestManagedProcessComponents_CreateWithInvalidArgsShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	processArgs := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
	_ = processArgs.CoreData.SetInternalMarshalizer(nil)
	processComponentsFactory, _ := processComp.NewProcessComponentsFactory(processArgs)
	managedProcessComponents, err := processComp.NewManagedProcessComponents(processComponentsFactory)
	require.NoError(t, err)
	err = managedProcessComponents.Create()
	require.Error(t, err)
	require.Nil(t, managedProcessComponents.NodesCoordinator())
}

func TestManagedProcessComponents_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
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
	dataComponents := componentsMock.GetDataComponents(coreComponents, shardCoordinator)
	cryptoComponents := componentsMock.GetCryptoComponents(coreComponents)
	networkComponents := componentsMock.GetNetworkComponents(cryptoComponents)
	stateComponents := componentsMock.GetStateComponents(coreComponents, shardCoordinator)
	processArgs := componentsMock.GetProcessArgs(
		shardCoordinator,
		coreComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
		networkComponents,
	)

	componentsMock.SetShardCoordinator(t, processArgs.BootstrapComponents, shardCoordinator)

	processComponentsFactory, err := processComp.NewProcessComponentsFactory(processArgs)
	require.Nil(t, err)
	managedProcessComponents, err := processComp.NewManagedProcessComponents(processComponentsFactory)
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

	nodeSkBytes, err := cryptoComponents.PrivateKey().ToByteArray()
	require.Nil(t, err)
	observerSkBytes, err := managedProcessComponents.NodeRedundancyHandler().ObserverPrivateKey().ToByteArray()
	require.Nil(t, err)
	require.NotEqual(t, nodeSkBytes, observerSkBytes)
}

func TestManagedProcessComponents_Close(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	processArgs := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
	processComponentsFactory, _ := processComp.NewProcessComponentsFactory(processArgs)
	managedProcessComponents, _ := processComp.NewManagedProcessComponents(processComponentsFactory)
	err := managedProcessComponents.Create()
	require.NoError(t, err)

	err = managedProcessComponents.Close()
	require.NoError(t, err)
	require.Nil(t, managedProcessComponents.NodesCoordinator())
}
