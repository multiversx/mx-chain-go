package processing_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	processComp "github.com/multiversx/mx-chain-go/factory/processing"
	"github.com/multiversx/mx-chain-go/process/mock"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/stretchr/testify/assert"
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

func TestManagedProcessComponents_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testManagedProcessComponentsCreateShouldWork(t, common.MetachainShardId, common.ChainRunTypeRegular)
	testManagedProcessComponentsCreateShouldWork(t, 0, common.ChainRunTypeRegular)
	testManagedProcessComponentsCreateShouldWork(t, 0, common.ChainRunTypeSovereign)
}

func testManagedProcessComponentsCreateShouldWork(t *testing.T, shardID uint32, chainType common.ChainRunType) {

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	shardCoordinator.CurrentShard = shardID

	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if core.IsSmartContractOnMetachain(address[len(address)-1:], address) {
			return core.MetachainShardId
		}
		return 0
	}

	args := createMockProcessComponentsFactoryArgs()
	componentsMock.SetShardCoordinator(t, args.BootstrapComponents, shardCoordinator)
	args.ChainRunType = chainType
	processComponentsFactory, _ := processComp.NewProcessComponentsFactory(args)
	processComponentsFactory.SetChainRunType(chainType)
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
	require.True(t, check.IfNil(managedProcessComponents.APITransactionEvaluator()))
	require.True(t, check.IfNil(managedProcessComponents.FallbackHeaderValidator()))
	require.True(t, check.IfNil(managedProcessComponents.PeerShardMapper()))
	require.True(t, check.IfNil(managedProcessComponents.ShardCoordinator()))
	require.True(t, check.IfNil(managedProcessComponents.TxsSenderHandler()))
	require.True(t, check.IfNil(managedProcessComponents.HardforkTrigger()))
	require.True(t, check.IfNil(managedProcessComponents.ProcessedMiniBlocksTracker()))

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
	require.False(t, check.IfNil(managedProcessComponents.APITransactionEvaluator()))
	require.False(t, check.IfNil(managedProcessComponents.FallbackHeaderValidator()))
	require.False(t, check.IfNil(managedProcessComponents.PeerShardMapper()))
	require.False(t, check.IfNil(managedProcessComponents.ShardCoordinator()))
	require.False(t, check.IfNil(managedProcessComponents.TxsSenderHandler()))
	require.False(t, check.IfNil(managedProcessComponents.HardforkTrigger()))
	require.False(t, check.IfNil(managedProcessComponents.ProcessedMiniBlocksTracker()))

	switch chainType {
	case common.ChainRunTypeRegular:
		switch shardID {
		case common.MetachainShardId:
			assert.Equal(t, "*sync.metaForkDetector", fmt.Sprintf("%T", managedProcessComponents.ForkDetector()))
			assert.Equal(t, "*track.metaBlockTrack", fmt.Sprintf("%T", managedProcessComponents.BlockTracker()))
		case 0:
			assert.Equal(t, "*sync.shardForkDetector", fmt.Sprintf("%T", managedProcessComponents.ForkDetector()))
			assert.Equal(t, "*track.shardBlockTrack", fmt.Sprintf("%T", managedProcessComponents.BlockTracker()))
		}
	case common.ChainRunTypeSovereign:
		assert.Equal(t, "*sync.sovereignChainShardForkDetector", fmt.Sprintf("%T", managedProcessComponents.ForkDetector()))
		assert.Equal(t, "*track.sovereignChainShardBlockTrack", fmt.Sprintf("%T", managedProcessComponents.BlockTracker()))
	}
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
