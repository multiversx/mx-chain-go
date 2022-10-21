package processing_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory/processing"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	componentsMock "github.com/ElrondNetwork/elrond-go/testscommon/components"
	"github.com/stretchr/testify/assert"
)

func TestNewSovereignChainManagedProcessComponents_ShouldErrNilManagedProcessComponents(t *testing.T) {
	t.Parallel()

	scmpc, err := processing.NewSovereignChainManagedProcessComponents(nil, nil)

	assert.Nil(t, scmpc)
	assert.Equal(t, errors.ErrNilManagedProcessComponents, err)
}

func TestNewSovereignChainManagedProcessComponents_ShouldErrNilProcessComponentsFactory(t *testing.T) {
	t.Parallel()

	shardCoordinator := testscommon.NewMultiShardsCoordinatorMock(2)
	processArgs := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
	pcf, _ := processing.NewProcessComponentsFactory(processArgs)
	mpc, _ := processing.NewManagedProcessComponents(pcf)

	scmpc, err := processing.NewSovereignChainManagedProcessComponents(mpc, nil)

	assert.Nil(t, scmpc)
	assert.Equal(t, errors.ErrNilProcessComponentsFactory, err)
}

func TestNewSovereignChainManagedProcessComponents_ShouldWork(t *testing.T) {
	t.Parallel()

	shardCoordinator := testscommon.NewMultiShardsCoordinatorMock(2)
	processArgs := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
	pcf, _ := processing.NewProcessComponentsFactory(processArgs)
	mpc, _ := processing.NewManagedProcessComponents(pcf)
	scpcf, _ := processing.NewSovereignChainProcessComponentsFactory(pcf)

	scmpc, err := processing.NewSovereignChainManagedProcessComponents(mpc, scpcf)

	assert.NotNil(t, scmpc)
	assert.Nil(t, err)
}

func TestSovereignChainManagedProcessComponents_CreateWithInvalidArgsShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := testscommon.NewMultiShardsCoordinatorMock(2)
	processArgs := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
	_ = processArgs.CoreData.SetInternalMarshalizer(nil)
	pcf, _ := processing.NewProcessComponentsFactory(processArgs)
	mpc, _ := processing.NewManagedProcessComponents(pcf)
	scpcf, _ := processing.NewSovereignChainProcessComponentsFactory(pcf)
	scmpc, _ := processing.NewSovereignChainManagedProcessComponents(mpc, scpcf)

	err := scmpc.Create()

	assert.Error(t, err)
	assert.Nil(t, scmpc.NodesCoordinator())
}

func TestSovereignChainManagedProcessComponents_CreateShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	coreComponents := componentsMock.GetCoreComponents()
	shardCoordinator := testscommon.NewMultiShardsCoordinatorMock(1)
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
	networkComponents := componentsMock.GetNetworkComponents()
	cryptoComponents := componentsMock.GetCryptoComponents(coreComponents)
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
	pcf, _ := processing.NewProcessComponentsFactory(processArgs)
	mpc, _ := processing.NewManagedProcessComponents(pcf)
	scpcf, _ := processing.NewSovereignChainProcessComponentsFactory(pcf)
	scmpc, _ := processing.NewSovereignChainManagedProcessComponents(mpc, scpcf)

	assert.True(t, check.IfNil(scmpc.NodesCoordinator()))
	assert.True(t, check.IfNil(scmpc.InterceptorsContainer()))
	assert.True(t, check.IfNil(scmpc.ResolversFinder()))
	assert.True(t, check.IfNil(scmpc.RoundHandler()))
	assert.True(t, check.IfNil(scmpc.ForkDetector()))
	assert.True(t, check.IfNil(scmpc.BlockProcessor()))
	assert.True(t, check.IfNil(scmpc.EpochStartTrigger()))
	assert.True(t, check.IfNil(scmpc.EpochStartNotifier()))
	assert.True(t, check.IfNil(scmpc.BlackListHandler()))
	assert.True(t, check.IfNil(scmpc.BootStorer()))
	assert.True(t, check.IfNil(scmpc.HeaderSigVerifier()))
	assert.True(t, check.IfNil(scmpc.ValidatorsStatistics()))
	assert.True(t, check.IfNil(scmpc.ValidatorsProvider()))
	assert.True(t, check.IfNil(scmpc.BlockTracker()))
	assert.True(t, check.IfNil(scmpc.PendingMiniBlocksHandler()))
	assert.True(t, check.IfNil(scmpc.RequestHandler()))
	assert.True(t, check.IfNil(scmpc.TxLogsProcessor()))
	assert.True(t, check.IfNil(scmpc.HeaderConstructionValidator()))
	assert.True(t, check.IfNil(scmpc.HeaderIntegrityVerifier()))
	assert.True(t, check.IfNil(scmpc.CurrentEpochProvider()))
	assert.True(t, check.IfNil(scmpc.NodeRedundancyHandler()))
	assert.True(t, check.IfNil(scmpc.WhiteListHandler()))
	assert.True(t, check.IfNil(scmpc.WhiteListerVerifiedTxs()))
	assert.True(t, check.IfNil(scmpc.RequestedItemsHandler()))
	assert.True(t, check.IfNil(scmpc.ImportStartHandler()))
	assert.True(t, check.IfNil(scmpc.HistoryRepository()))
	assert.True(t, check.IfNil(scmpc.TransactionSimulatorProcessor()))
	assert.True(t, check.IfNil(scmpc.FallbackHeaderValidator()))
	assert.True(t, check.IfNil(scmpc.PeerShardMapper()))
	assert.True(t, check.IfNil(scmpc.ShardCoordinator()))
	assert.True(t, check.IfNil(scmpc.TxsSenderHandler()))
	assert.True(t, check.IfNil(scmpc.HardforkTrigger()))
	assert.True(t, check.IfNil(scmpc.ProcessedMiniBlocksTracker()))

	err := scmpc.Create()

	assert.NoError(t, err)
	assert.False(t, check.IfNil(scmpc.NodesCoordinator()))
	assert.False(t, check.IfNil(scmpc.InterceptorsContainer()))
	assert.False(t, check.IfNil(scmpc.ResolversFinder()))
	assert.False(t, check.IfNil(scmpc.RoundHandler()))
	assert.False(t, check.IfNil(scmpc.ForkDetector()))
	assert.False(t, check.IfNil(scmpc.BlockProcessor()))
	assert.False(t, check.IfNil(scmpc.EpochStartTrigger()))
	assert.False(t, check.IfNil(scmpc.EpochStartNotifier()))
	assert.False(t, check.IfNil(scmpc.BlackListHandler()))
	assert.False(t, check.IfNil(scmpc.BootStorer()))
	assert.False(t, check.IfNil(scmpc.HeaderSigVerifier()))
	assert.False(t, check.IfNil(scmpc.ValidatorsStatistics()))
	assert.False(t, check.IfNil(scmpc.ValidatorsProvider()))
	assert.False(t, check.IfNil(scmpc.BlockTracker()))
	assert.False(t, check.IfNil(scmpc.PendingMiniBlocksHandler()))
	assert.False(t, check.IfNil(scmpc.RequestHandler()))
	assert.False(t, check.IfNil(scmpc.TxLogsProcessor()))
	assert.False(t, check.IfNil(scmpc.HeaderConstructionValidator()))
	assert.False(t, check.IfNil(scmpc.HeaderIntegrityVerifier()))
	assert.False(t, check.IfNil(scmpc.CurrentEpochProvider()))
	assert.False(t, check.IfNil(scmpc.NodeRedundancyHandler()))
	assert.False(t, check.IfNil(scmpc.WhiteListHandler()))
	assert.False(t, check.IfNil(scmpc.WhiteListerVerifiedTxs()))
	assert.False(t, check.IfNil(scmpc.RequestedItemsHandler()))
	assert.False(t, check.IfNil(scmpc.ImportStartHandler()))
	assert.False(t, check.IfNil(scmpc.HistoryRepository()))
	assert.False(t, check.IfNil(scmpc.TransactionSimulatorProcessor()))
	assert.False(t, check.IfNil(scmpc.FallbackHeaderValidator()))
	assert.False(t, check.IfNil(scmpc.PeerShardMapper()))
	assert.False(t, check.IfNil(scmpc.ShardCoordinator()))
	assert.False(t, check.IfNil(scmpc.TxsSenderHandler()))
	assert.False(t, check.IfNil(scmpc.HardforkTrigger()))
	assert.False(t, check.IfNil(scmpc.ProcessedMiniBlocksTracker()))

	nodeSkBytes, err := cryptoComponents.PrivateKey().ToByteArray()
	assert.Nil(t, err)
	observerSkBytes, err := scmpc.NodeRedundancyHandler().ObserverPrivateKey().ToByteArray()
	assert.Nil(t, err)
	assert.NotEqual(t, nodeSkBytes, observerSkBytes)
}
