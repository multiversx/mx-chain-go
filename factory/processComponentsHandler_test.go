package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
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

	processComponentsFactory, err := factory.NewProcessComponentsFactory(processArgs)
	require.Nil(t, err)
	managedProcessComponents, err := factory.NewManagedProcessComponents(processComponentsFactory)
	require.NoError(t, err)
	require.Nil(t, managedProcessComponents.NodesCoordinator())
	require.Nil(t, managedProcessComponents.InterceptorsContainer())
	require.Nil(t, managedProcessComponents.ResolversFinder())
	require.Nil(t, managedProcessComponents.Rounder())
	require.Nil(t, managedProcessComponents.ForkDetector())
	require.Nil(t, managedProcessComponents.BlockProcessor())
	require.Nil(t, managedProcessComponents.EpochStartTrigger())
	require.Nil(t, managedProcessComponents.EpochStartNotifier())
	require.Nil(t, managedProcessComponents.BlackListHandler())
	require.Nil(t, managedProcessComponents.BootStorer())
	require.Nil(t, managedProcessComponents.HeaderSigVerifier())
	require.Nil(t, managedProcessComponents.ValidatorsStatistics())
	require.Nil(t, managedProcessComponents.ValidatorsProvider())
	require.Nil(t, managedProcessComponents.BlockTracker())
	require.Nil(t, managedProcessComponents.PendingMiniBlocksHandler())
	require.Nil(t, managedProcessComponents.RequestHandler())
	require.Nil(t, managedProcessComponents.TxLogsProcessor())
	require.Nil(t, managedProcessComponents.HeaderConstructionValidator())
	require.Nil(t, managedProcessComponents.HeaderIntegrityVerifier())

	err = managedProcessComponents.Create()
	require.NoError(t, err)
	require.NotNil(t, managedProcessComponents.NodesCoordinator())
	require.NotNil(t, managedProcessComponents.InterceptorsContainer())
	require.NotNil(t, managedProcessComponents.ResolversFinder())
	require.NotNil(t, managedProcessComponents.Rounder())
	require.NotNil(t, managedProcessComponents.ForkDetector())
	require.NotNil(t, managedProcessComponents.BlockProcessor())
	require.NotNil(t, managedProcessComponents.EpochStartTrigger())
	require.NotNil(t, managedProcessComponents.EpochStartNotifier())
	require.NotNil(t, managedProcessComponents.BlackListHandler())
	require.NotNil(t, managedProcessComponents.BootStorer())
	require.NotNil(t, managedProcessComponents.HeaderSigVerifier())
	require.NotNil(t, managedProcessComponents.ValidatorsStatistics())
	require.NotNil(t, managedProcessComponents.ValidatorsProvider())
	require.NotNil(t, managedProcessComponents.BlockTracker())
	require.NotNil(t, managedProcessComponents.PendingMiniBlocksHandler())
	require.NotNil(t, managedProcessComponents.RequestHandler())
	require.NotNil(t, managedProcessComponents.TxLogsProcessor())
	require.NotNil(t, managedProcessComponents.HeaderConstructionValidator())
	require.NotNil(t, managedProcessComponents.HeaderIntegrityVerifier())
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
