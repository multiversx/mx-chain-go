package storageBootstrap

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/process/sync"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	testsDataRetriever "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/outport"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestNewShardBootstrapFactory(t *testing.T) {
	sbf, err := NewShardBootstrapFactory()
	require.NotNil(t, sbf)
	require.Nil(t, err)
}

func TestShardBootstrapFactory_CreateShardBootstrapFactory(t *testing.T) {
	sbf, err := NewShardBootstrapFactory()
	require.NotNil(t, sbf)
	require.Nil(t, err)

	bootStrapper, err := sbf.CreateShardBootstrapFactory(getDefaultArgs())
	require.NotNil(t, bootStrapper)
	require.Nil(t, err)
}

func TestShardBootstrapFactory_IsInterfaceNil(t *testing.T) {
	sbf, _ := NewShardBootstrapFactory()

	require.False(t, sbf.IsInterfaceNil())
}

func getDefaultArgs() sync.ArgShardBootstrapper {
	bootStorer := genericMocks.NewStorerMock()
	argBaseBoostrapper := sync.ArgBaseBootstrapper{
		PoolsHolder: testsDataRetriever.NewPoolsHolderMock(),
		Store: &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return bootStorer, nil
			},
		},
		ChainHandler:                 &testscommon.ChainHandlerStub{},
		RoundHandler:                 &mock.RoundHandlerMock{},
		BlockProcessor:               &testscommon.BlockProcessorStub{},
		WaitTime:                     time.Second,
		Hasher:                       &hashingMocks.HasherMock{},
		Marshalizer:                  &testscommon.ProtoMarshalizerMock{},
		ForkDetector:                 &mock.ForkDetectorStub{},
		RequestHandler:               &testscommon.RequestHandlerStub{},
		ShardCoordinator:             &testscommon.ShardsCoordinatorMock{},
		Accounts:                     &stateMock.AccountsStub{},
		BlackListHandler:             &testscommon.TimeCacheStub{},
		NetworkWatcher:               &mock.NetworkConnectionWatcherStub{},
		BootStorer:                   &mock.BoostrapStorerMock{},
		StorageBootstrapper:          &mock.StorageBootstrapperMock{},
		EpochHandler:                 &mock.EpochStartTriggerStub{},
		MiniblocksProvider:           &mock.MiniBlocksProviderStub{},
		Uint64Converter:              &mock.Uint64ByteSliceConverterMock{},
		AppStatusHandler:             &statusHandlerMock.AppStatusHandlerStub{},
		OutportHandler:               &outport.OutportStub{},
		AccountsDBSyncer:             &mock.AccountsDBSyncerStub{},
		CurrentEpochProvider:         &testscommon.CurrentEpochProviderStub{},
		HistoryRepo:                  &dblookupext.HistoryRepositoryStub{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessWaitTime:              time.Second,
		RepopulateTokensSupplies:     false,
	}

	return sync.ArgShardBootstrapper{
		ArgBaseBootstrapper: argBaseBoostrapper,
	}
}
