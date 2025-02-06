package block_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/factory/shard"
	shardData "github.com/multiversx/mx-chain-go/process/factory/shard/data"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/track"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	commonMock "github.com/multiversx/mx-chain-go/testscommon/common"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"
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
		require.ErrorIs(t, err, errors.ErrNilOutgoingOperationsFormatter)
	})

	t.Run("should error when outgoing operation pool is nil", func(t *testing.T) {
		t.Parallel()

		args := createSovChainBlockProcessorArgs()
		args.OutGoingOperationsPool = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(args)

		require.Nil(t, scbp)
		require.Equal(t, errors.ErrNilOutGoingOperationsPool, err)
	})

	t.Run("should error when operations hasher is nil", func(t *testing.T) {
		t.Parallel()

		args := createSovChainBlockProcessorArgs()
		args.OperationsHasher = nil
		scbp, err := blproc.NewSovereignChainBlockProcessor(args)

		require.Nil(t, scbp)
		require.Equal(t, errors.ErrNilOperationsHasher, err)
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

func createTxCoordinator(
	store dataRetriever.StorageService,
	marshaller marshal.Marshalizer,
	Hasher hashing.Hasher,
	dataPool dataRetriever.PoolsHolder,
) process.TransactionCoordinator {
	args := shardData.ArgPreProcessorsContainerFactory{
		ShardCoordinator:             mock.NewMultiShardsCoordinatorMock(3),
		Store:                        store,
		Marshaller:                   marshaller,
		Hasher:                       Hasher,
		DataPool:                     dataPool,
		PubkeyConverter:              createMockPubkeyConverter(),
		Accounts:                     initAccountsMock(),
		RequestHandler:               &testscommon.RequestHandlerStub{},
		TxProcessor:                  &testscommon.TxProcessorMock{},
		ScProcessor:                  &testscommon.SCProcessorMock{},
		ScResultProcessor:            &testscommon.SmartContractResultsProcessorMock{},
		RewardsTxProcessor:           &testscommon.RewardTxProcessorMock{},
		EconomicsFee:                 &economicsmocks.EconomicsHandlerStub{},
		GasHandler:                   &testscommon.GasHandlerStub{},
		BlockTracker:                 &mock.BlockTrackerMock{},
		BlockSizeComputation:         &testscommon.BlockSizeComputationStub{},
		BalanceComputation:           &testscommon.BalanceComputationStub{},
		EnableEpochsHandler:          &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
		TxExecutionOrderHandler:      &commonMock.TxExecutionOrderHandlerStub{},
		RunTypeComponents:            processMocks.NewRunTypeComponentsStub(),
	}
	factory, _ := shard.NewPreProcessorsContainerFactory(args)
	container, _ := factory.Create()

	argsTransactionCoordinator := createMockTransactionCoordinatorArguments(initAccountsMock(), dataPool, container)
	tc, _ := coordinator.NewTransactionCoordinator(argsTransactionCoordinator)
	return tc
}

func TestSovereignChainBlockProcessor_RestoreBlockIntoPoolsShouldWork(t *testing.T) {
	t.Parallel()

	txHash := []byte("tx hash 1")

	datapool := dataRetrieverMock.NewPoolsHolderMock()
	marshalizerMock := &mock.MarshalizerMock{}
	hasherMock := &mock.HasherStub{}

	body := &block.Body{}
	tx := &transaction.Transaction{
		Nonce: 1,
		Value: big.NewInt(0),
	}
	buffTx, _ := marshalizerMock.Marshal(tx)

	store := &storageStubs.ChainStorerStub{
		GetAllCalled: func(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
			m := make(map[string][]byte)
			m[string(txHash)] = buffTx
			return m, nil
		},
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = datapool
	dataComponents.Storage = store
	coreComponents.Hash = hasherMock
	coreComponents.IntMarsh = marshalizerMock
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.TxCoordinator = createTxCoordinator(store, marshalizerMock, hasherMock, datapool)

	shardArguments := CreateSovereignChainShardTrackerMockArguments()
	sbt, _ := track.NewShardBlockTrack(shardArguments)
	arguments.BlockTracker, _ = track.NewSovereignChainShardBlockTrack(sbt)
	rrh, _ := requestHandlers.NewResolverRequestHandler(
		&dataRetrieverMock.RequestersFinderStub{},
		&mock.RequestedItemsHandlerStub{},
		&testscommon.WhiteListHandlerStub{},
		1,
		0,
		time.Second,
	)
	arguments.RequestHandler, _ = requestHandlers.NewSovereignResolverRequestHandler(rrh)

	sp, _ := blproc.NewShardProcessor(arguments)

	argsSovProc := createSovChainBlockProcessorArgs()
	argsSovProc.ShardProcessor = sp
	sovBlockProc, _ := blproc.NewSovereignChainBlockProcessor(argsSovProc)

	txHashes := make([][]byte, 0)
	txHashes = append(txHashes, txHash)
	miniblock := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   1,
		TxHashes:        txHashes,
	}
	body.MiniBlocks = append(body.MiniBlocks, &miniblock)

	miniblockHash := []byte("mini block hash 1")
	hasherMock.ComputeCalled = func(s string) []byte {
		return miniblockHash
	}

	extendedHeaderHash := []byte("meta block hash 1")
	extendedHeader := &block.ShardHeaderExtended{
		Header: &block.HeaderV2{
			ScheduledRootHash: []byte("root hash"),
		},
		IncomingMiniBlocks: nil,
		IncomingEvents:     nil,
	}
	datapool.Headers().AddHeader(extendedHeaderHash, extendedHeader)

	store.GetStorerCalled = func(unitType dataRetriever.UnitType) (storage.Storer, error) {
		return &storageStubs.StorerStub{
			RemoveCalled: func(key []byte) error {
				return nil
			},
			GetCalled: func(key []byte) ([]byte, error) {
				return marshalizerMock.Marshal(extendedHeader)
			},
		}, nil
	}

	miniBlockHeader := block.MiniBlockHeader{
		Hash:            miniblockHash,
		SenderShardID:   miniblock.SenderShardID,
		ReceiverShardID: miniblock.ReceiverShardID,
	}

	sovHdr := &block.SovereignChainHeader{
		Header: &block.Header{
			MiniBlockHeaders: []block.MiniBlockHeader{miniBlockHeader},
		},
		ExtendedShardHeaderHashes: [][]byte{extendedHeaderHash},
	}

	err := sovBlockProc.RestoreBlockIntoPools(sovHdr, body)
	assert.Nil(t, err)

	miniblockFromPool, _ := datapool.MiniBlocks().Get(miniblockHash)
	txFromPool, _ := datapool.Transactions().SearchFirstData(txHash)
	assert.Nil(t, err)
	assert.Equal(t, &miniblock, miniblockFromPool)
	assert.Equal(t, tx, txFromPool)
}

//TODO: More unit tests should be added. Created PR https://multiversxlabs.atlassian.net/browse/MX-14149
