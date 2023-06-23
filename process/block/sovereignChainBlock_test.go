package block_test

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/process"
	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

// CreateSovereignChainShardTrackerMockArguments -
func CreateSovereignChainShardTrackerMockArguments() track.ArgShardTracker {
	argsHeaderValidator := blproc.ArgsHeaderValidator{
		Hasher:      &testscommon.HasherStub{},
		Marshalizer: &marshallerMock.MarshalizerMock{},
	}
	headerValidator, _ := blproc.NewHeaderValidator(argsHeaderValidator)

	arguments := track.ArgShardTracker{
		ArgBaseTracker: track.ArgBaseTracker{
			Hasher:           &testscommon.HasherStub{},
			HeaderValidator:  headerValidator,
			Marshalizer:      &marshallerMock.MarshalizerStub{},
			RequestHandler:   &testscommon.ExtendedShardHeaderRequestHandlerStub{},
			RoundHandler:     &testscommon.RoundHandlerMock{},
			ShardCoordinator: &testscommon.ShardsCoordinatorMock{},
			Store:            &storage.ChainStorerStub{},
			StartHeaders:     createGenesisBlocks(&testscommon.ShardsCoordinatorMock{NoShards: 1}),
			PoolsHolder:      dataRetrieverMock.NewPoolsHolderMock(),
			WhitelistHandler: &testscommon.WhiteListHandlerStub{},
			FeeHandler:       &economicsmocks.EconomicsHandlerStub{},
		},
	}

	return arguments
}

func TestSovereignBlockProcessor_NewSovereignChainBlockProcessorShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("should error when shard processor is nil", func(t *testing.T) {
		t.Parallel()

		scbp, err := blproc.NewSovereignChainBlockProcessor(nil, &mock.ValidatorStatisticsProcessorStub{})

		assert.Nil(t, scbp)
		assert.ErrorIs(t, err, process.ErrNilBlockProcessor)
	})

	t.Run("should error when validator statistics is nil", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		sp, _ := blproc.NewShardProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(sp, nil)

		assert.Nil(t, scbp)
		assert.ErrorIs(t, err, process.ErrNilValidatorStatistics)
	})

	t.Run("should error when type assertion to extendedShardHeaderTrackHandler fails", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		sp, _ := blproc.NewShardProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(sp, &mock.ValidatorStatisticsProcessorStub{})

		assert.Nil(t, scbp)
		assert.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})

	t.Run("should error when type assertion to extendedShardHeaderRequestHandler fails", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.BlockTracker, _ = track.NewSovereignChainShardBlockTrack(sbt)
		sp, _ := blproc.NewShardProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(sp, &mock.ValidatorStatisticsProcessorStub{})

		assert.Nil(t, scbp)
		assert.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		rrh, _ := requestHandlers.NewResolverRequestHandler(
			&dataRetrieverMock.RequestersFinderStub{},
			&mock.RequestedItemsHandlerStub{},
			&testscommon.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.BlockTracker, _ = track.NewSovereignChainShardBlockTrack(sbt)
		arguments.RequestHandler, _ = requestHandlers.NewSovereignResolverRequestHandler(rrh)
		sp, _ := blproc.NewShardProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(sp, &mock.ValidatorStatisticsProcessorStub{})

		assert.NotNil(t, scbp)
		assert.Nil(t, err)
	})
}

//TODO: More unit tests should be added. Created PR https://multiversxlabs.atlassian.net/browse/MX-14149
