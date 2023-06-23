package storageBootstrap

import (
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestNewShardStorageBootstrapperFactory(t *testing.T) {
	sbf, err := NewShardStorageBootstrapperFactory()
	require.NotNil(t, sbf)
	require.Nil(t, err)
}

func TestShardStorageBootstrapperFactory_CreateShardStorageBootstrapper(t *testing.T) {
	sbf, _ := NewShardStorageBootstrapperFactory()

	bootStrapper, err := sbf.CreateShardStorageBootstrapper(getDefaultArgShardBootstrapper())
	require.NotNil(t, bootStrapper)
	require.Nil(t, err)
}

func TestShardStorageBootstrapperFactory_IsInterfaceNil(t *testing.T) {
	sbf, _ := NewShardStorageBootstrapperFactory()

	require.False(t, sbf.IsInterfaceNil())
}

func getDefaultArgShardBootstrapper() ArgsShardStorageBootstrapper {
	bootStorer := genericMocks.NewStorerMock()
	argBaseBoostrapper := ArgsShardStorageBootstrapper{
		ArgsBaseStorageBootstrapper{
			BootStorer:     &mock.BoostrapStorerMock{},
			ForkDetector:   &mock.ForkDetectorStub{},
			BlockProcessor: &testscommon.BlockProcessorStub{},
			ChainHandler:   &testscommon.ChainHandlerStub{},
			Marshalizer:    &testscommon.ProtoMarshalizerMock{},
			Store: &storageStubs.ChainStorerStub{
				GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
					return bootStorer, nil
				},
			},
			Uint64Converter:              &mock.Uint64ByteSliceConverterMock{},
			BootstrapRoundIndex:          0,
			ShardCoordinator:             &testscommon.ShardsCoordinatorMock{},
			NodesCoordinator:             &shardingMocks.NodesCoordinatorMock{},
			EpochStartTrigger:            &testscommon.EpochStartTriggerStub{},
			BlockTracker:                 &mock.BlockTrackerStub{},
			ChainID:                      "2",
			ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
			MiniblocksProvider:           &mock.MiniBlocksProviderStub{},
			EpochNotifier:                &epochNotifier.EpochNotifierStub{},
			ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
			AppStatusHandler:             statusHandler.NewAppStatusHandlerMock(),
		},
	}

	return argBaseBoostrapper
}
