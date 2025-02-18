package block_test

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	errMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	storageStub "github.com/multiversx/mx-chain-go/testscommon/storage"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	sovereignCore "github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/stretchr/testify/require"
)

func createSovChainBaseBlockProcessorArgs() blproc.ArgShardProcessor {
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

func createSovChainBlockProcessorArgs() blproc.ArgsSovereignChainBlockProcessor {
	baseArgs := createSovChainBaseBlockProcessorArgs()
	sp, _ := blproc.NewShardProcessor(baseArgs)
	return blproc.ArgsSovereignChainBlockProcessor{
		ShardProcessor:                  sp,
		ValidatorStatisticsProcessor:    &testscommon.ValidatorStatisticsProcessorStub{},
		OutgoingOperationsFormatter:     &sovereign.OutgoingOperationsFormatterMock{},
		OutGoingOperationsPool:          &sovereign.OutGoingOperationsPoolMock{},
		OperationsHasher:                &testscommon.HasherStub{},
		EpochStartDataCreator:           &mock.EpochStartDataCreatorStub{},
		EpochRewardsCreator:             &testscommon.RewardsCreatorStub{},
		ValidatorInfoCreator:            &testscommon.EpochValidatorInfoCreatorStub{},
		EpochSystemSCProcessor:          &testscommon.EpochStartSystemSCStub{},
		EpochEconomics:                  &mock.EpochEconomicsStub{},
		SCToProtocol:                    &mock.SCToProtocolStub{},
		MainChainNotarizationStartRound: 11,
	}
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
			Store:            &storageStub.ChainStorerStub{},
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

		args := createSovChainBlockProcessorArgs()
		args.ShardProcessor = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(args)

		require.Nil(t, scbp)
		require.ErrorIs(t, err, process.ErrNilBlockProcessor)
	})

	t.Run("should error when validator statistics is nil", func(t *testing.T) {
		t.Parallel()

		args := createSovChainBlockProcessorArgs()
		args.ValidatorStatisticsProcessor = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(args)

		require.Nil(t, scbp)
		require.ErrorIs(t, err, process.ErrNilValidatorStatistics)
	})

	t.Run("should error when outgoing operations formatter is nil", func(t *testing.T) {
		t.Parallel()

		args := createSovChainBlockProcessorArgs()
		args.OutgoingOperationsFormatter = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(args)

		require.Nil(t, scbp)
		require.ErrorIs(t, err, errMx.ErrNilOutgoingOperationsFormatter)
	})

	t.Run("should error when outgoing operation pool is nil", func(t *testing.T) {
		t.Parallel()

		args := createSovChainBlockProcessorArgs()
		args.OutGoingOperationsPool = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(args)

		require.Nil(t, scbp)
		require.Equal(t, errMx.ErrNilOutGoingOperationsPool, err)
	})

	t.Run("should error when operations hasher is nil", func(t *testing.T) {
		t.Parallel()

		args := createSovChainBlockProcessorArgs()
		args.OperationsHasher = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(args)

		require.Nil(t, scbp)
		require.Equal(t, errMx.ErrNilOperationsHasher, err)
	})

	t.Run("should error when epoch start data creator is nil", func(t *testing.T) {
		t.Parallel()

		args := createSovChainBlockProcessorArgs()
		args.EpochStartDataCreator = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(args)

		require.Nil(t, scbp)
		require.Equal(t, process.ErrNilEpochStartDataCreator, err)
	})

	t.Run("should error when epoch rewards creator is nil", func(t *testing.T) {
		t.Parallel()

		args := createSovChainBlockProcessorArgs()
		args.EpochRewardsCreator = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(args)

		require.Nil(t, scbp)
		require.Equal(t, process.ErrNilRewardsCreator, err)
	})

	t.Run("should error when epoch validator infor creator is nil", func(t *testing.T) {
		t.Parallel()

		args := createSovChainBlockProcessorArgs()
		args.ValidatorInfoCreator = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(args)

		require.Nil(t, scbp)
		require.Equal(t, process.ErrNilEpochStartValidatorInfoCreator, err)
	})

	t.Run("should error when epoch start system sc processor is nil", func(t *testing.T) {
		t.Parallel()

		args := createSovChainBlockProcessorArgs()
		args.EpochSystemSCProcessor = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(args)

		require.Nil(t, scbp)
		require.Equal(t, process.ErrNilEpochStartSystemSCProcessor, err)
	})

	t.Run("should error when type assertion to extendedShardHeaderTrackHandler fails", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		sp, _ := blproc.NewShardProcessor(arguments)
		args := createSovChainBlockProcessorArgs()
		args.ShardProcessor = sp
		scbp, err := blproc.NewSovereignChainBlockProcessor(args)

		require.Nil(t, scbp)
		require.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})

	t.Run("should error when type assertion to extendedShardHeaderRequestHandler fails", func(t *testing.T) {
		t.Parallel()

		shardArguments := CreateSovereignChainShardTrackerMockArguments()
		sbt, _ := track.NewShardBlockTrack(shardArguments)

		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.BlockTracker, _ = track.NewSovereignChainShardBlockTrack(sbt)
		sp, _ := blproc.NewShardProcessor(arguments)
		args := createSovChainBlockProcessorArgs()
		args.ShardProcessor = sp
		scbp, err := blproc.NewSovereignChainBlockProcessor(args)

		require.Nil(t, scbp)
		require.ErrorIs(t, err, process.ErrWrongTypeAssertion)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arguments := createSovChainBlockProcessorArgs()
		scbp, err := blproc.NewSovereignChainBlockProcessor(arguments)

		require.NotNil(t, scbp)
		require.Nil(t, err)
	})
}

func TestSovereignChainBlockProcessor_createAndSetOutGoingMiniBlock(t *testing.T) {
	arguments := createSovChainBaseBlockProcessorArgs()

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

	outgoingOpsHasher := &mock.HasherStub{}
	bridgeOp1Hash := outgoingOpsHasher.Compute(string(bridgeOp1))
	bridgeOp2Hash := outgoingOpsHasher.Compute(string(bridgeOp2))
	bridgeOpsHash := outgoingOpsHasher.Compute(string(append(bridgeOp1Hash, bridgeOp2Hash...)))

	outgoingOperationsFormatter := &sovereign.OutgoingOperationsFormatterMock{
		CreateOutgoingTxDataCalled: func(logs []*data.LogData) ([][]byte, error) {
			require.Equal(t, expectedLogs, logs)
			return [][]byte{bridgeOp1, bridgeOp2}, nil
		},
	}

	poolAddCt := 0
	outGoingOperationsPool := &sovereign.OutGoingOperationsPoolMock{
		AddCalled: func(data *sovereignCore.BridgeOutGoingData) {
			defer func() {
				poolAddCt++
			}()

			switch poolAddCt {
			case 0:
				require.Equal(t, &sovereignCore.BridgeOutGoingData{
					Hash: bridgeOpsHash,
					OutGoingOperations: []*sovereignCore.OutGoingOperation{
						{
							Hash: bridgeOp1Hash,
							Data: bridgeOp1,
						},
						{
							Hash: bridgeOp2Hash,
							Data: bridgeOp2,
						},
					},
				}, data)
			default:
				require.Fail(t, "should not add in pool any other operation")
			}
		},
	}

	sp, _ := blproc.NewShardProcessor(arguments)
	scbp, _ := blproc.NewSovereignChainBlockProcessor(blproc.ArgsSovereignChainBlockProcessor{
		ShardProcessor:               sp,
		ValidatorStatisticsProcessor: &testscommon.ValidatorStatisticsProcessorStub{},
		OutgoingOperationsFormatter:  outgoingOperationsFormatter,
		OutGoingOperationsPool:       outGoingOperationsPool,
		OperationsHasher:             outgoingOpsHasher,
		EpochStartDataCreator:        &mock.EpochStartDataCreatorStub{},
		EpochRewardsCreator:          &testscommon.RewardsCreatorStub{},
		ValidatorInfoCreator:         &testscommon.EpochValidatorInfoCreatorStub{},
		EpochSystemSCProcessor:       &testscommon.EpochStartSystemSCStub{},
		EpochEconomics:               &mock.EpochEconomicsStub{},
		SCToProtocol:                 &mock.SCToProtocolStub{},
	})

	sovChainHdr := &block.SovereignChainHeader{}
	processedMb := &block.MiniBlock{
		ReceiverShardID: core.SovereignChainShardId,
		SenderShardID:   core.MainChainShardId,
	}
	blockBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{processedMb},
	}

	err := scbp.CreateAndSetOutGoingMiniBlock(sovChainHdr, blockBody)
	require.Nil(t, err)
	require.Equal(t, 1, poolAddCt)

	expectedOutGoingMb := &block.MiniBlock{
		TxHashes:        [][]byte{bridgeOp1Hash, bridgeOp2Hash},
		ReceiverShardID: core.MainChainShardId,
		SenderShardID:   arguments.BootstrapComponents.ShardCoordinator().SelfId(),
	}
	expectedBlockBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{processedMb, expectedOutGoingMb},
	}
	require.Equal(t, expectedBlockBody, blockBody)

	expectedOutGoingMbHash, err := core.CalculateHash(arguments.CoreComponents.InternalMarshalizer(), arguments.CoreComponents.Hasher(), expectedOutGoingMb)
	require.Nil(t, err)

	expectedSovChainHeader := &block.SovereignChainHeader{
		OutGoingMiniBlockHeader: &block.OutGoingMiniBlockHeader{
			Hash:                   expectedOutGoingMbHash,
			OutGoingOperationsHash: bridgeOpsHash,
		},
	}
	require.Equal(t, expectedSovChainHeader, sovChainHdr)
}

func TestSovereignChainBlockProcessor_RestoreBlockIntoPoolsInvalidHeaderType(t *testing.T) {
	t.Parallel()

	sovBlockProcArgs := createSovChainBlockProcessorArgs()
	sovBlockProc, _ := blproc.NewSovereignChainBlockProcessor(sovBlockProcArgs)

	err := sovBlockProc.RestoreBlockIntoPools(&block.HeaderV2{}, &block.Body{})
	require.ErrorIs(t, err, errMx.ErrWrongTypeAssertion)
}

func TestSovereignChainBlockProcessor_RestoreBlockIntoPoolsShouldWorkWhenHeaderNotCommitted(t *testing.T) {
	t.Parallel()

	store := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return nil, errors.New("header not fully processed yet")
		},
	}
	expectedBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				SenderShardID:   core.MainChainShardId,
				ReceiverShardID: core.SovereignChainShardId,
				TxHashes:        [][]byte{[]byte("txHash1")},
			},
		},
	}

	dataPool := dataRetrieverMock.NewPoolsHolderMock()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = dataPool
	dataComponents.Storage = store
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.RequestHandler = &testscommon.ExtendedShardHeaderRequestHandlerStub{}

	wasRestoreBlockDataFromStorageCalled := false
	arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
		RestoreBlockDataFromStorageCalled: func(body *block.Body) (int, error) {
			wasRestoreBlockDataFromStorageCalled = true
			require.Equal(t, body, expectedBody)
			return 0, nil
		},
	}

	wasLastSelfNotarizedHeaderRemoved := false
	wasLastCrossNotarizedHeaderRemoved := false
	arguments.BlockTracker = &testscommon.ExtendedShardHeaderTrackerStub{
		RemoveLastSelfNotarizedHeadersCalled: func() {
			wasLastSelfNotarizedHeaderRemoved = true
		},
		RemoveLastCrossNotarizedHeadersCalled: func() {
			wasLastCrossNotarizedHeaderRemoved = true
		},
	}

	sp, _ := blproc.NewShardProcessor(arguments)
	argsSovProc := createSovChainBlockProcessorArgs()
	argsSovProc.ShardProcessor = sp
	sovBlockProc, _ := blproc.NewSovereignChainBlockProcessor(argsSovProc)

	extendedHeaderHash := []byte("extendedHdrHash")
	sovHdr := createSovHeaderForRestoreBlocksTest(expectedBody, extendedHeaderHash)

	retrievedHdr, err := dataPool.Headers().GetHeaderByHash(extendedHeaderHash)
	require.NotNil(t, err)
	require.Nil(t, retrievedHdr)

	err = sovBlockProc.RestoreBlockIntoPools(sovHdr, expectedBody)
	require.Nil(t, err)
	require.True(t, wasRestoreBlockDataFromStorageCalled)
	require.True(t, wasLastSelfNotarizedHeaderRemoved)
	require.True(t, wasLastCrossNotarizedHeaderRemoved)

	retrievedHdr, err = dataPool.Headers().GetHeaderByHash(extendedHeaderHash)
	require.NotNil(t, err)
	require.Nil(t, retrievedHdr)

}

func TestSovereignChainBlockProcessor_RestoreBlockIntoPoolsShouldWork(t *testing.T) {
	t.Parallel()

	testRestoreBlockIntoPools(t, true)
	testRestoreBlockIntoPools(t, false)
}

func testRestoreBlockIntoPools(t *testing.T, withExtendedHeader bool) {
	store := &storageStubs.ChainStorerStub{}
	marshaller := &mock.MarshalizerMock{}
	expectedBody := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			{
				SenderShardID:   core.MainChainShardId,
				ReceiverShardID: core.SovereignChainShardId,
				TxHashes:        [][]byte{[]byte("txHash1")},
			},
		},
	}

	extendedHdrNonce := uint64(4)
	extendedHdrNonceToBytes := []byte("toByteData")
	dataPool := dataRetrieverMock.NewPoolsHolderMock()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.UInt64ByteSliceConv = &testscommon.Uint64ByteSliceConverterStub{
		ToByteSliceCalled: func(u uint64) []byte {
			require.Equal(t, extendedHdrNonce, u)
			return extendedHdrNonceToBytes
		},
	}
	dataComponents.DataPool = dataPool
	dataComponents.Storage = store
	coreComponents.IntMarsh = marshaller
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.RequestHandler = &testscommon.ExtendedShardHeaderRequestHandlerStub{}

	wasRestoreBlockDataFromStorageCalled := false
	arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
		RestoreBlockDataFromStorageCalled: func(body *block.Body) (int, error) {
			wasRestoreBlockDataFromStorageCalled = true
			require.Equal(t, body, expectedBody)
			return 0, nil
		},
	}

	wasLastSelfNotarizedHeaderRemoved := false
	wasLastCrossNotarizedHeaderRemoved := false
	arguments.BlockTracker = &testscommon.ExtendedShardHeaderTrackerStub{
		RemoveLastSelfNotarizedHeadersCalled: func() {
			wasLastSelfNotarizedHeaderRemoved = true
		},
		RemoveLastCrossNotarizedHeadersCalled: func() {
			wasLastCrossNotarizedHeaderRemoved = true
		},
	}

	sp, _ := blproc.NewShardProcessor(arguments)
	argsSovProc := createSovChainBlockProcessorArgs()
	argsSovProc.ShardProcessor = sp
	sovBlockProc, _ := blproc.NewSovereignChainBlockProcessor(argsSovProc)

	extendedHeaderHash := []byte("extendedHdrHash")
	extendedHeader := &block.ShardHeaderExtended{
		Header: &block.HeaderV2{
			ScheduledRootHash: []byte("rootHash"),
			Header: &block.Header{
				Nonce: extendedHdrNonce,
			},
		},
	}

	wasExtendedHdrStorerRemoved := false
	storerExtendedHdr := &storageStubs.StorerStub{
		RemoveCalled: func(key []byte) error {
			wasExtendedHdrStorerRemoved = true
			require.Equal(t, extendedHeaderHash, key)
			return nil
		},
		GetCalled: func(key []byte) ([]byte, error) {
			return marshaller.Marshal(extendedHeader)
		},
	}

	wasNonceHashExtendedHdrStorerRemoved := false
	storerNonceHashExtendedHdr := &storageStubs.StorerStub{
		RemoveCalled: func(key []byte) error {
			wasNonceHashExtendedHdrStorerRemoved = true
			require.Equal(t, extendedHdrNonceToBytes, key)
			return nil
		},
	}

	store.GetStorerCalled = func(unitType dataRetriever.UnitType) (storage.Storer, error) {
		switch unitType {
		case dataRetriever.ExtendedShardHeadersUnit:
			return storerExtendedHdr, nil
		case dataRetriever.ExtendedShardHeadersNonceHashDataUnit:
			return storerNonceHashExtendedHdr, nil
		}

		return nil, errors.New("invalid storer type")
	}

	sovHdr := createSovHeaderForRestoreBlocksTest(expectedBody, extendedHeaderHash)
	if !withExtendedHeader {
		sovHdr.ExtendedShardHeaderHashes = nil
	}

	retrievedHdr, err := dataPool.Headers().GetHeaderByHash(extendedHeaderHash)
	require.NotNil(t, err)
	require.Nil(t, retrievedHdr)

	err = sovBlockProc.RestoreBlockIntoPools(sovHdr, expectedBody)
	require.Nil(t, err)
	require.True(t, wasRestoreBlockDataFromStorageCalled)
	require.True(t, wasLastSelfNotarizedHeaderRemoved)

	if withExtendedHeader {
		require.True(t, wasLastCrossNotarizedHeaderRemoved)
		require.True(t, wasExtendedHdrStorerRemoved)
		require.True(t, wasNonceHashExtendedHdrStorerRemoved)

		retrievedHdr, err = dataPool.Headers().GetHeaderByHash(extendedHeaderHash)
		require.Nil(t, err)
		require.Equal(t, extendedHeader, retrievedHdr)
	} else {
		require.False(t, wasLastCrossNotarizedHeaderRemoved)
		require.False(t, wasExtendedHdrStorerRemoved)
		require.False(t, wasNonceHashExtendedHdrStorerRemoved)

		retrievedHdr, err = dataPool.Headers().GetHeaderByHash(extendedHeaderHash)
		require.NotNil(t, err)
		require.Nil(t, retrievedHdr)
	}
}

func createSovHeaderForRestoreBlocksTest(body *block.Body, extendedHeaderHash []byte) *block.SovereignChainHeader {
	miniBlockHeader := block.MiniBlockHeader{
		SenderShardID:   body.MiniBlocks[0].SenderShardID,
		ReceiverShardID: body.MiniBlocks[0].ReceiverShardID,
	}

	return &block.SovereignChainHeader{
		Header: &block.Header{
			MiniBlockHeaders: []block.MiniBlockHeader{miniBlockHeader},
		},
		ExtendedShardHeaderHashes: [][]byte{extendedHeaderHash},
	}
}

//TODO: More unit tests should be added. Created PR https://multiversxlabs.atlassian.net/browse/MX-14149
