package block_test

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createSovChainBlockProcessorArgs() blproc.ArgShardProcessor {
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

	coreComp, dataComp, bootstrapComp, statusComp := createComponentHolderMocks()
	coreComp.Hash = &hashingMocks.HasherMock{}

	arguments := CreateMockArguments(coreComp, dataComp, bootstrapComp, statusComp)
	arguments.BlockTracker, _ = track.NewSovereignChainShardBlockTrack(sbt)
	arguments.RequestHandler, _ = requestHandlers.NewSovereignResolverRequestHandler(rrh)

	return arguments
}

// CreateSovereignChainShardTrackerMockArguments -
func CreateSovereignChainShardTrackerMockArguments() track.ArgShardTracker {
	argsHeaderValidator := blproc.ArgsHeaderValidator{
		Hasher:      &hashingMocks.HasherMock{},
		Marshalizer: &marshallerMock.MarshalizerMock{},
	}
	headerValidator, _ := blproc.NewHeaderValidator(argsHeaderValidator)

	arguments := track.ArgShardTracker{
		ArgBaseTracker: track.ArgBaseTracker{
			Hasher:           &hashingMocks.HasherMock{},
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

		scbp, err := blproc.NewSovereignChainBlockProcessor(blproc.ArgsSovereignChainBlockProcessor{
			ShardProcessor:               nil,
			ValidatorStatisticsProcessor: &mock.ValidatorStatisticsProcessorStub{},
			OutgoingOperationsFormatter:  &sovereign.OutgoingOperationsFormatterStub{},
			OutGoingOperationsPool:       &sovereign.OutGoingOperationsPoolStub{},
		})

		assert.Nil(t, scbp)
		assert.ErrorIs(t, err, process.ErrNilBlockProcessor)
	})

	t.Run("should error when validator statistics is nil", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		sp, _ := blproc.NewShardProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(blproc.ArgsSovereignChainBlockProcessor{
			ShardProcessor:               sp,
			ValidatorStatisticsProcessor: nil,
			OutgoingOperationsFormatter:  &sovereign.OutgoingOperationsFormatterStub{},
			OutGoingOperationsPool:       &sovereign.OutGoingOperationsPoolStub{},
		})

		assert.Nil(t, scbp)
		assert.ErrorIs(t, err, process.ErrNilValidatorStatistics)
	})

	t.Run("should error when outgoing operations formatter is nil", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		sp, _ := blproc.NewShardProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(blproc.ArgsSovereignChainBlockProcessor{
			ShardProcessor:               sp,
			ValidatorStatisticsProcessor: &mock.ValidatorStatisticsProcessorStub{},
			OutgoingOperationsFormatter:  nil,
			OutGoingOperationsPool:       &sovereign.OutGoingOperationsPoolStub{},
		})

		assert.Nil(t, scbp)
		assert.ErrorIs(t, err, errors.ErrNilOutgoingOperationsFormatter)
	})

	t.Run("should error when outgoing operation pool is nil", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		sp, _ := blproc.NewShardProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(blproc.ArgsSovereignChainBlockProcessor{
			ShardProcessor:               sp,
			ValidatorStatisticsProcessor: &mock.ValidatorStatisticsProcessorStub{},
			OutgoingOperationsFormatter:  &sovereign.OutgoingOperationsFormatterStub{},
			OutGoingOperationsPool:       nil,
		})

		require.Nil(t, scbp)
		require.Equal(t, errors.ErrNilOutGoingOperationsPool, err)
	})

	t.Run("should error when type assertion to extendedShardHeaderTrackHandler fails", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		sp, _ := blproc.NewShardProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(blproc.ArgsSovereignChainBlockProcessor{
			ShardProcessor:               sp,
			ValidatorStatisticsProcessor: &mock.ValidatorStatisticsProcessorStub{},
			OutgoingOperationsFormatter:  &sovereign.OutgoingOperationsFormatterStub{},
			OutGoingOperationsPool:       &sovereign.OutGoingOperationsPoolStub{},
		})

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
		scbp, err := blproc.NewSovereignChainBlockProcessor(blproc.ArgsSovereignChainBlockProcessor{
			ShardProcessor:               sp,
			ValidatorStatisticsProcessor: &mock.ValidatorStatisticsProcessorStub{},
			OutgoingOperationsFormatter:  &sovereign.OutgoingOperationsFormatterStub{},
			OutGoingOperationsPool:       &sovereign.OutGoingOperationsPoolStub{},
		})

		assert.Nil(t, scbp)
		assert.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arguments := createSovChainBlockProcessorArgs()
		sp, _ := blproc.NewShardProcessor(arguments)
		scbp, err := blproc.NewSovereignChainBlockProcessor(blproc.ArgsSovereignChainBlockProcessor{
			ShardProcessor:               sp,
			ValidatorStatisticsProcessor: &mock.ValidatorStatisticsProcessorStub{},
			OutgoingOperationsFormatter:  &sovereign.OutgoingOperationsFormatterStub{},
			OutGoingOperationsPool:       &sovereign.OutGoingOperationsPoolStub{},
		})

		assert.NotNil(t, scbp)
		assert.Nil(t, err)
	})
}

func TestSovereignChainBlockProcessor_createAndSetOutGoingMiniBlock(t *testing.T) {
	arguments := createSovChainBlockProcessorArgs()

	expectedLogs := []*data.LogData{
		{
			TxHash: "txHash1",
		},
	}
	arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
		GetAllCurrentLogsCalled: func() []*data.LogData {
			return expectedLogs
		},
	}
	bridgeOp1 := []byte("bridgeOp@123@rcv1@token1@val1")
	bridgeOp2 := []byte("bridgeOp@124@rcv2@token2@val2")

	hasher := arguments.CoreComponents.Hasher()
	bridgeOp1Hash := hasher.Compute(string(bridgeOp1))
	bridgeOp2Hash := hasher.Compute(string(bridgeOp2))
	bridgeOpsHash := hasher.Compute(string(append(bridgeOp1Hash, bridgeOp2Hash...)))

	outgoingOperationsFormatter := &sovereign.OutgoingOperationsFormatterStub{
		CreateOutgoingTxDataCalled: func(logs []*data.LogData) [][]byte {
			require.Equal(t, expectedLogs, logs)
			return [][]byte{bridgeOp1, bridgeOp2}
		},
	}

	poolAddCt := 0
	outGoingOperationsPool := &sovereign.OutGoingOperationsPoolStub{
		AddCalled: func(hash []byte, data []byte) {
			defer func() {
				poolAddCt++
			}()

			switch poolAddCt {
			case 0:
				require.Equal(t, bridgeOp1Hash, hash)
				require.Equal(t, bridgeOp1, data)
			case 1:
				require.Equal(t, bridgeOp2Hash, hash)
				require.Equal(t, bridgeOp2, data)
			default:
				require.Fail(t, "should not add in pool any other operation")
			}
		},
	}

	sp, _ := blproc.NewShardProcessor(arguments)
	scbp, _ := blproc.NewSovereignChainBlockProcessor(blproc.ArgsSovereignChainBlockProcessor{
		ShardProcessor:               sp,
		ValidatorStatisticsProcessor: &mock.ValidatorStatisticsProcessorStub{},
		OutgoingOperationsFormatter:  outgoingOperationsFormatter,
		OutGoingOperationsPool:       outGoingOperationsPool,
	})

	sovChainHdr := &block.SovereignChainHeader{}
	blockBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				ReceiverShardID: core.SovereignChainShardId,
				SenderShardID:   core.MainChainShardId,
			},
		},
	}

	err := scbp.CreateAndSetOutGoingMiniBlock(sovChainHdr, blockBody)
	require.Nil(t, err)

	expectedOutGoingMb := &block.MiniBlock{
		TxHashes:        [][]byte{bridgeOp1Hash, bridgeOp2Hash},
		ReceiverShardID: core.MainChainShardId,
		SenderShardID:   arguments.BootstrapComponents.ShardCoordinator().SelfId(),
	}
	expectedBlockBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				ReceiverShardID: core.SovereignChainShardId,
				SenderShardID:   core.MainChainShardId,
			},
			expectedOutGoingMb,
		},
	}
	require.Equal(t, expectedBlockBody, blockBody)

	expectedOutGoingMbHash, err := core.CalculateHash(arguments.CoreComponents.InternalMarshalizer(), hasher, expectedOutGoingMb)
	require.Nil(t, err)

	expectedSovChainHeader := &block.SovereignChainHeader{
		OutGoingMiniBlockHeader: &block.OutGoingMiniBlockHeader{
			Hash:                   expectedOutGoingMbHash,
			OutGoingOperationsHash: bridgeOpsHash,
		},
	}
	require.Equal(t, expectedSovChainHeader, sovChainHdr)

}

//TODO: More unit tests should be added. Created PR https://multiversxlabs.atlassian.net/browse/MX-14149
