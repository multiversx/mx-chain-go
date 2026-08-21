package block_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-core-go/core/sharding"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters/uint64ByteSlice"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/process/aotSelection"
	headersCache "github.com/multiversx/mx-chain-go/process/asyncExecution/cache"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/executionManager"
	"github.com/multiversx/mx-chain-go/testscommon/pool"

	"github.com/multiversx/mx-chain-go/process/asyncExecution/executionTrack"
	"github.com/multiversx/mx-chain-go/process/estimator"
	"github.com/multiversx/mx-chain-go/process/missingData"
	"github.com/multiversx/mx-chain-go/testscommon/mbSelection"
	"github.com/multiversx/mx-chain-go/testscommon/processMocks"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/graceperiod"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/blockchain"
	proofscache "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/proofsCache"
	"github.com/multiversx/mx-chain-go/process"
	blproc "github.com/multiversx/mx-chain-go/process/block"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/block/headerForBlock"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/factory/containers"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/disabled"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	commonMocks "github.com/multiversx/mx-chain-go/testscommon/common"
	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
)

var expectedErr = errors.New("expected error")

const (
	busyIdentifier = "busy"
	idleIdentifier = "idle"
)

func haveTime() time.Duration {
	return 2000 * time.Millisecond
}

func createArgBaseProcessor(
	coreComponents *mock.CoreComponentsMock,
	dataComponents *mock.DataComponentsMock,
	bootstrapComponents *mock.BootstrapComponentsMock,
	statusComponents *mock.StatusComponentsMock,
) blproc.ArgBaseProcessor {
	nodesCoordinatorInstance := shardingMocks.NewNodesCoordinatorMock()
	argsHeaderValidator := blproc.ArgsHeaderValidator{
		Hasher:              &hashingMocks.HasherMock{},
		Marshalizer:         &mock.MarshalizerMock{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	headerValidator, _ := blproc.NewHeaderValidator(argsHeaderValidator)

	startHeaders := createGenesisBlocks(mock.NewOneShardCoordinatorMock())

	accountsDb := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accounts := &stateMock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return nil, nil
		},
		RecreateTrieIfNeededCalled: func(options common.RootHashHolder) error {
			return nil
		},
		CommitCalled: func() ([]byte, error) {
			return nil, nil
		},
	}
	accountsDb[state.UserAccountsState] = accounts

	statusCoreComponents := &factory.StatusCoreComponentsStub{
		AppStatusHandlerField: &statusHandlerMock.AppStatusHandlerStub{},
	}

	blockTracker := mock.NewBlockTrackerMock(bootstrapComponents.ShardCoordinator(), startHeaders)
	var headersForBlock blproc.HeadersForBlock = &testscommon.HeadersForBlockMock{}
	if !check.IfNil(coreComponents) && !check.IfNil(bootstrapComponents) && !check.IfNil(dataComponents) {
		headersForBlock, _ = headerForBlock.NewHeadersForBlock(headerForBlock.ArgHeadersForBlock{
			DataPool:            dataComponents.DataPool,
			RequestHandler:      &testscommon.RequestHandlerStub{},
			EnableEpochsHandler: coreComponents.EnableEpochsHandler(),
			ShardCoordinator:    bootstrapComponents.ShardCoordinator(),
			BlockTracker:        blockTracker,
			TxCoordinator:       &testscommon.TransactionCoordinatorMock{},
			RoundHandler:        coreComponents.RoundHandler(),
			ExtraDelayForRequestBlockInfoInMilliseconds: 100,
			GenesisNonce: 0,
		})
	}

	var blockDataRequester process.BlockDataRequester
	var inclusionEstimator process.InclusionEstimator
	var execManager process.ExecutionManager
	var mbSelectionSession blproc.MiniBlocksSelectionSession
	var execResultsVerifier blproc.ExecutionResultsVerifier
	var missingDataResolver blproc.MissingDataResolver
	if check.IfNil(dataComponents) || check.IfNil(dataComponents.Datapool()) || check.IfNil(coreComponents) || check.IfNil(bootstrapComponents) {
		inclusionEstimator = &processMocks.InclusionEstimatorMock{}
		mbSelectionSession = &mbSelection.MiniBlockSelectionSessionStub{}
		execResultsVerifier = &processMocks.ExecutionResultsVerifierMock{}
		missingDataResolver = &processMocks.MissingDataResolverMock{}
	} else {
		preprocContainer := containers.NewPreProcessorsContainer()
		blockDataRequesterArgs := coordinator.BlockDataRequestArgs{
			RequestHandler:      &testscommon.RequestHandlerStub{},
			MiniBlockPool:       dataComponents.Datapool().MiniBlocks(),
			PreProcessors:       preprocContainer,
			ShardCoordinator:    bootstrapComponents.ShardCoordinator(),
			EnableEpochsHandler: coreComponents.EnableEpochsHandler(),
		}
		// second instance for proposal missing data fetching to avoid interferences
		blockDataRequester, _ = coordinator.NewBlockDataRequester(blockDataRequesterArgs)

		mbSelectionSession, _ = blproc.NewMiniBlocksSelectionSession(
			bootstrapComponents.ShardCoordinator().SelfId(),
			coreComponents.InternalMarshalizer(),
			coreComponents.Hasher(),
		)

		blocksCache := headersCache.NewHeaderBodyCache(config.HeaderBodyCacheConfig{})
		executionResultsTracker, _ := executionTrack.NewExecutionResultsTracker(disabled.NewDisabledStateAccessesCollector())
		_ = executionResultsTracker.SetLastNotarizedResult(&block.ExecutionResult{})
		execManager, _ = executionManager.NewExecutionManager(executionManager.ArgsExecutionManager{
			BlocksCache:             blocksCache,
			ExecutionResultsTracker: executionResultsTracker,
			BlockChain:              dataComponents.BlockChain,
			Headers:                 dataComponents.DataPool.Headers(),
			PostProcessTransactions: dataComponents.DataPool.PostProcessTransactions(),
			ExecutedMiniBlocks:      dataComponents.DataPool.ExecutedMiniBlocks(),
			StorageService:          dataComponents.StorageService(),
			Marshaller:              coreComponents.InternalMarshalizer(),
			ShardCoordinator:        bootstrapComponents.ShardCoordinator(),
		})
		execResultsVerifier, _ = blproc.NewExecutionResultsVerifier(dataComponents.BlockChain, execManager)
		inclusionEstimator, _ = estimator.NewExecutionResultInclusionEstimator(
			config.ExecutionResultInclusionEstimatorConfig{
				SafetyMargin:       110,
				MaxResultsPerBlock: 20,
			},
			coreComponents.RoundHandler(),
			&testscommon.ExecResSizeComputationStub{},
		)

		missingDataArgs := missingData.ResolverArgs{
			HeadersPool:         dataComponents.DataPool.Headers(),
			ProofsPool:          dataComponents.DataPool.Proofs(),
			RequestHandler:      &testscommon.RequestHandlerStub{},
			BlockDataRequester:  blockDataRequester,
			EnableEpochsHandler: coreComponents.EnableEpochsHandler(),
		}
		missingDataResolver, _ = missingData.NewMissingDataResolver(missingDataArgs)
	}

	return blproc.ArgBaseProcessor{
		CoreComponents:       coreComponents,
		DataComponents:       dataComponents,
		BootstrapComponents:  bootstrapComponents,
		StatusComponents:     statusComponents,
		StatusCoreComponents: statusCoreComponents,
		Config:               config.Config{},
		AccountsDB:           accountsDb,
		AccountsProposal:     accounts,
		ForkDetector:         &mock.ForkDetectorMock{},
		NodesCoordinator:     nodesCoordinatorInstance,
		FeeHandler:           &mock.FeeAccumulatorStub{},
		RequestHandler:       &testscommon.RequestHandlerStub{},
		BlockChainHook:       &testscommon.BlockChainHookStub{},
		TxCoordinator:        &testscommon.TransactionCoordinatorMock{},
		EpochStartTrigger:    &mock.EpochStartTriggerStub{},
		HeaderValidator:      headerValidator,
		BootStorer: &mock.BoostrapStorerMock{
			PutCalled: func(round int64, bootData bootstrapStorage.BootstrapData) error {
				return nil
			},
		},
		BlockTracker:                       blockTracker,
		MiniBlockTracker:                   &testscommon.MiniBlockTrackerStub{},
		BlockSizeThrottler:                 &mock.BlockSizeThrottlerStub{},
		Version:                            "softwareVersion",
		HistoryRepository:                  &dblookupext.HistoryRepositoryStub{},
		GasHandler:                         &mock.GasHandlerMock{},
		ScheduledTxsExecutionHandler:       &testscommon.ScheduledTxsExecutionStub{},
		OutportDataProvider:                &outport.OutportDataProviderStub{},
		ScheduledMiniBlocksEnableEpoch:     2,
		ProcessedMiniBlocksTracker:         &testscommon.ProcessedMiniBlocksTrackerStub{},
		ReceiptsRepository:                 &testscommon.ReceiptsRepositoryStub{},
		BlockProcessingCutoffHandler:       &testscommon.BlockProcessingCutoffStub{},
		ManagedPeersHolder:                 &testscommon.ManagedPeersHolderStub{},
		SentSignaturesTracker:              &testscommon.SentSignatureTrackerStub{},
		StateAccessesCollector:             disabled.NewDisabledStateAccessesCollector(),
		HeadersForBlock:                    headersForBlock,
		MiniBlocksSelectionSession:         mbSelectionSession,
		ExecutionResultsVerifier:           execResultsVerifier,
		MissingDataResolver:                missingDataResolver,
		ExecutionResultsInclusionEstimator: inclusionEstimator,
		GasComputation: &testscommon.GasComputationMock{
			AddIncomingMiniBlocksCalled: func(miniBlocks []data.MiniBlockHeaderHandler, transactions map[string][]data.TransactionHandler) (int, int, error) {
				return len(miniBlocks), 0, nil
			},
			AddOutgoingTransactionsCalled: func(txHashes [][]byte, transactions []data.TransactionHandler, isProposer bool) ([][]byte, []data.MiniBlockHeaderHandler, error) {
				return txHashes, nil, nil
			},
		},
		ExecutionManager:        execManager,
		TxExecutionOrderHandler: &commonMocks.TxExecutionOrderHandlerStub{},
		AOTSelector:             aotSelection.NewDisabledAOTSelector(),
	}
}

func createTestBlockchain() *testscommon.ChainHandlerStub {
	return &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}
}

func generateTestCache() storage.Cacher {
	c, _ := storageunit.NewCache(storageunit.CacheConfig{Type: storageunit.LRUCache, Capacity: 1000, Shards: 1, SizeInBytes: 0})
	return c
}

func generateTestUnit() storage.Storer {
	storer, _ := storageunit.NewStorageUnit(
		generateTestCache(),
		database.NewMemDB(),
	)

	return storer
}

func initDataPool() *dataRetrieverMock.PoolsHolderStub {
	transactionsPool := testscommon.NewShardedDataCacheNotifierMock()
	unsignedTransactionsPool := testscommon.NewShardedDataCacheNotifierMock()
	rewardTransactionsPool := testscommon.NewShardedDataCacheNotifierMock()
	validatorsInfoPool := testscommon.NewShardedDataCacheNotifierMock()

	metablocksPool := cache.NewCacherStub()
	miniblocksPool := cache.NewCacherStub()
	headersPool := &mock.HeadersCacherStub{}
	proofsPool := proofscache.NewProofsPool(3, 100)
	executedMBs := cache.NewCacherStub()
	postProcessTxs := cache.NewCacherStub()
	directSentTxs := cache.NewCacherStub()

	sdp := &dataRetrieverMock.PoolsHolderStub{
		TransactionsCalled:         func() dataRetriever.ShardedDataCacherNotifier { return transactionsPool },
		UnsignedTransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier { return unsignedTransactionsPool },
		RewardTransactionsCalled:   func() dataRetriever.ShardedDataCacherNotifier { return rewardTransactionsPool },
		ValidatorsInfoCalled:       func() dataRetriever.ShardedDataCacherNotifier { return validatorsInfoPool },
		MetaBlocksCalled: func() storage.Cacher {
			return metablocksPool
		},
		MiniBlocksCalled: func() storage.Cacher {
			return miniblocksPool
		},
		HeadersCalled: func() dataRetriever.HeadersPool {
			return headersPool
		},
		ProofsCalled: func() dataRetriever.ProofsPool {
			return proofsPool
		},
		ExecutedMiniBlocksCalled: func() storage.Cacher {
			return executedMBs
		},
		PostProcessTransactionsCalled: func() storage.Cacher {
			return postProcessTxs
		},
		DirectSentTransactionsCalled: func() storage.Cacher {
			return directSentTxs
		},
	}

	return sdp
}

func initStore() *dataRetriever.ChainStorer {
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TransactionUnit, generateTestUnit())
	store.AddStorer(dataRetriever.UnsignedTransactionUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MiniBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.RewardTransactionUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MetaBlockUnit, generateTestUnit())
	store.AddStorer(dataRetriever.PeerChangesUnit, generateTestUnit())
	store.AddStorer(dataRetriever.BlockHeaderUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ShardHdrNonceHashDataUnit, generateTestUnit())
	store.AddStorer(dataRetriever.MetaHdrNonceHashDataUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ReceiptsUnit, generateTestUnit())
	store.AddStorer(dataRetriever.TrieEpochRootHashUnit, generateTestUnit())
	store.AddStorer(dataRetriever.ScheduledSCRsUnit, generateTestUnit())
	return store
}

func createDummyMetaBlock(destShardId uint32, senderShardId uint32, miniBlockHashes ...[]byte) *block.MetaBlock {
	metaBlock := &block.MetaBlock{
		ShardInfo: []block.ShardData{
			{
				ShardID:               senderShardId,
				ShardMiniBlockHeaders: make([]block.MiniBlockHeader, len(miniBlockHashes)),
			},
		},
	}

	for idx, mbHash := range miniBlockHashes {
		metaBlock.ShardInfo[0].ShardMiniBlockHeaders[idx].ReceiverShardID = destShardId
		metaBlock.ShardInfo[0].ShardMiniBlockHeaders[idx].SenderShardID = senderShardId
		metaBlock.ShardInfo[0].ShardMiniBlockHeaders[idx].Hash = mbHash
	}

	return metaBlock
}

func createDummyMiniBlock(
	txHash string,
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	destShardId uint32,
	senderShardId uint32) (*block.MiniBlock, []byte) {

	miniblock := &block.MiniBlock{
		TxHashes:        [][]byte{[]byte(txHash)},
		ReceiverShardID: destShardId,
		SenderShardID:   senderShardId,
	}

	buff, _ := marshalizer.Marshal(miniblock)
	hash := hasher.Compute(string(buff))

	return miniblock, hash
}

func isInTxHashes(searched []byte, list [][]byte) bool {
	for _, txHash := range list {
		if bytes.Equal(txHash, searched) {
			return true
		}
	}
	return false
}

type wrongBody struct {
}

func (wr *wrongBody) Clone() data.BodyHandler {
	wrCopy := *wr

	return &wrCopy
}

func (wr *wrongBody) IntegrityAndValidity() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (wr *wrongBody) IsInterfaceNil() bool {
	return wr == nil
}

func createComponentHolderMocks() (
	*mock.CoreComponentsMock,
	*mock.DataComponentsMock,
	*mock.BootstrapComponentsMock,
	*mock.StatusComponentsMock,
) {
	blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
	_ = blkc.SetGenesisHeader(&block.Header{Nonce: 0})

	gracePeriod, _ := graceperiod.NewEpochChangeGracePeriod([]config.EpochChangeGracePeriodByEpoch{{EnableEpoch: 0, GracePeriodInRounds: 1}})

	coreComponents := &mock.CoreComponentsMock{
		IntMarsh:                           &mock.MarshalizerMock{},
		Hash:                               &mock.HasherStub{},
		UInt64ByteSliceConv:                &mock.Uint64ByteSliceConverterMock{},
		StatusField:                        &statusHandlerMock.AppStatusHandlerStub{},
		RoundField:                         &mock.RoundHandlerMock{},
		ProcessStatusHandlerField:          &testscommon.ProcessStatusHandlerStub{},
		EpochNotifierField:                 &epochNotifier.EpochNotifierStub{},
		EnableEpochsHandlerField:           enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
		RoundNotifierField:                 &epochNotifier.RoundNotifierStub{},
		EnableRoundsHandlerField:           &testscommon.EnableRoundsHandlerStub{},
		EpochChangeGracePeriodHandlerField: gracePeriod,
		ProcessConfigsHandlerField:         testscommon.GetDefaultProcessConfigsHandler(),
		ClosingNodeStartedField:            &atomic.Bool{},
	}

	dataComponents := &mock.DataComponentsMock{
		Storage:    initStore(),
		DataPool:   initDataPool(),
		BlockChain: blkc,
	}

	boostrapComponents := &mock.BootstrapComponentsMock{
		Coordinator:          mock.NewOneShardCoordinatorMock(),
		HdrIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		VersionedHdrFactory: &testscommon.VersionedHeaderFactoryStub{
			CreateCalled: func(epoch uint32, _ uint64) data.HeaderHandler {
				return &block.Header{}
			},
		},
	}

	statusComponents := &mock.StatusComponentsMock{
		Outport: &outport.OutportStub{},
	}

	return coreComponents, dataComponents, boostrapComponents, statusComponents
}

func CreateMockArguments(
	coreComponents *mock.CoreComponentsMock,
	dataComponents *mock.DataComponentsMock,
	bootstrapComponents *mock.BootstrapComponentsMock,
	statusComponents *mock.StatusComponentsMock,
) blproc.ArgShardProcessor {
	return blproc.ArgShardProcessor{
		ArgBaseProcessor: createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents),
	}
}

func createMockTransactionCoordinatorArguments(
	accountAdapter state.AccountsAdapter,
	poolsHolder dataRetriever.PoolsHolder,
	preProcessorsContainer process.PreProcessorsContainer,
	preProcessorsContainerProposal process.PreProcessorsContainer,
) coordinator.ArgTransactionCoordinator {

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
	enableEpochsHandler := enableEpochsHandlerMock.NewEnableEpochsHandlerStub()
	enableRoundsHandler := &testscommon.EnableRoundsHandlerStub{}

	blockDataRequesterArgs := coordinator.BlockDataRequestArgs{
		RequestHandler:      &testscommon.RequestHandlerStub{},
		MiniBlockPool:       poolsHolder.MiniBlocks(),
		PreProcessors:       preProcessorsContainer,
		ShardCoordinator:    shardCoordinator,
		EnableEpochsHandler: enableEpochsHandler,
	}

	blockDataRequester, _ := coordinator.NewBlockDataRequester(blockDataRequesterArgs)

	blockDataRequesterArgsProposal := coordinator.BlockDataRequestArgs{
		RequestHandler:      &testscommon.RequestHandlerStub{},
		MiniBlockPool:       poolsHolder.MiniBlocks(),
		PreProcessors:       preProcessorsContainerProposal,
		ShardCoordinator:    shardCoordinator,
		EnableEpochsHandler: enableEpochsHandler,
	}
	blockDataRequesterProposal, _ := coordinator.NewBlockDataRequester(blockDataRequesterArgsProposal)

	argsTransactionCoordinator := coordinator.ArgTransactionCoordinator{
		Hasher:                &hashingMocks.HasherMock{},
		Marshalizer:           &mock.MarshalizerMock{},
		ShardCoordinator:      shardCoordinator,
		Accounts:              accountAdapter,
		DataPool:              poolsHolder,
		PreProcessors:         preProcessorsContainer,
		PreProcessorsProposal: preProcessorsContainerProposal,
		InterProcessors: &mock.InterimProcessorContainerMock{
			KeysCalled: func() []block.Type {
				return []block.Type{block.SmartContractResultBlock}
			},
		},
		GasHandler:                   &testscommon.GasHandlerStub{},
		FeeHandler:                   &mock.FeeAccumulatorStub{},
		BlockSizeComputation:         &testscommon.BlockSizeComputationStub{},
		BalanceComputation:           &testscommon.BalanceComputationStub{},
		EconomicsFee:                 &economicsmocks.EconomicsHandlerMock{},
		TxTypeHandler:                &testscommon.TxTypeHandlerMock{},
		TransactionsLogProcessor:     &mock.TxLogsProcessorStub{},
		EnableEpochsHandler:          enableEpochsHandler,
		EnableRoundsHandler:          enableRoundsHandler,
		ScheduledTxsExecutionHandler: &testscommon.ScheduledTxsExecutionStub{},
		DoubleTransactionsDetector:   &testscommon.DoubleTransactionsDetector{},
		ProcessedMiniBlocksTracker:   &testscommon.ProcessedMiniBlocksTrackerStub{},
		TxExecutionOrderHandler:      &commonMocks.TxExecutionOrderHandlerStub{},
		BlockDataRequester:           blockDataRequester,
		BlockDataRequesterProposal:   blockDataRequesterProposal,
		GasComputation: &testscommon.GasComputationMock{
			AddIncomingMiniBlocksCalled: func(miniBlocks []data.MiniBlockHeaderHandler, transactions map[string][]data.TransactionHandler) (int, int, error) {
				return len(miniBlocks), 0, nil
			},
			AddOutgoingTransactionsCalled: func(txHashes [][]byte, transactions []data.TransactionHandler, isProposer bool) ([][]byte, []data.MiniBlockHeaderHandler, error) {
				return txHashes, nil, nil
			},
		},
		AOTSelector: aotSelection.NewDisabledAOTSelector(),
	}

	return argsTransactionCoordinator
}

func TestCheckProcessorNilParameters(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

	tests := []struct {
		args        func() blproc.ArgBaseProcessor
		expectedErr error
	}{
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.AccountsDB[state.UserAccountsState] = nil
				return args
			},
			expectedErr: process.ErrNilAccountsAdapter,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				return createArgBaseProcessor(coreComponents, nil, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilDataComponentsHolder,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				return createArgBaseProcessor(nil, dataComponents, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilCoreComponentsHolder,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.BootstrapComponents = nil
				return args
			},
			expectedErr: process.ErrNilBootstrapComponentsHolder,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				return createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, nil)
			},
			expectedErr: process.ErrNilStatusComponentsHolder,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.ForkDetector = nil
				return args
			},
			expectedErr: process.ErrNilForkDetector,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.Hash = nil
				args := createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
				return args
			},
			expectedErr: process.ErrNilHasher,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.IntMarsh = nil
				args := createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
				return args
			},
			expectedErr: process.ErrNilMarshalizer,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				dataCompCopy := *dataComponents
				dataCompCopy.Storage = nil
				args := createArgBaseProcessor(coreComponents, &dataCompCopy, bootstrapComponents, statusComponents)
				return args
			},
			expectedErr: process.ErrNilStorage,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				dataCompCopy := *dataComponents
				dataCompCopy.DataPool = &dataRetrieverMock.PoolsHolderStub{
					TransactionsCalled: func() dataRetriever.ShardedDataCacherNotifier {
						return &testscommon.ShardedDataCacheNotifierMock{}
					},
					HeadersCalled: func() dataRetriever.HeadersPool {
						return &pool.HeadersPoolStub{}
					},
					ProofsCalled: func() dataRetriever.ProofsPool {
						return &dataRetrieverMock.ProofsPoolMock{}
					},
				}
				args := createArgBaseProcessor(coreComponents, &dataCompCopy, bootstrapComponents, statusComponents)
				return args
			},
			expectedErr: process.ErrNilDirectSentCache,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.BootstrapComponents = &mainFactoryMocks.BootstrapComponentsStub{ShCoordinator: nil}
				return args
			},
			expectedErr: process.ErrNilShardCoordinator,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.NodesCoordinator = nil
				return args
			},
			expectedErr: process.ErrNilNodesCoordinator,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.UInt64ByteSliceConv = nil
				return createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilUint64Converter,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.RequestHandler = nil
				return args
			},
			expectedErr: process.ErrNilRequestHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.EpochStartTrigger = nil
				return args
			},
			expectedErr: process.ErrNilEpochStartTrigger,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.RoundNotifierField = nil
				return createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilRoundNotifier,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.RoundField = nil
				return createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilRoundHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.BootStorer = nil
				return args
			},
			expectedErr: process.ErrNilStorage,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.BlockChainHook = nil
				return args
			},
			expectedErr: process.ErrNilBlockChainHook,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.TxCoordinator = nil
				return args
			},
			expectedErr: process.ErrNilTransactionCoordinator,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.HeaderValidator = nil
				return args
			},
			expectedErr: process.ErrNilHeaderValidator,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.BlockTracker = nil
				return args
			},
			expectedErr: process.ErrNilBlockTracker,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.FeeHandler = nil
				return args
			},
			expectedErr: process.ErrNilEconomicsFeeHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				dataComp := &mock.DataComponentsMock{
					Storage:    dataComponents.Storage,
					DataPool:   dataComponents.DataPool,
					BlockChain: nil,
				}
				return createArgBaseProcessor(coreComponents, dataComp, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilBlockChain,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.BlockSizeThrottler = nil
				return args
			},
			expectedErr: process.ErrNilBlockSizeThrottler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				statusCompCopy := *statusComponents
				statusCompCopy.Outport = nil
				return createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, &statusCompCopy)
			},
			expectedErr: process.ErrNilOutportHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.HistoryRepository = nil
				return args
			},
			expectedErr: process.ErrNilHistoryRepository,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				bootStrapCopy := *bootstrapComponents
				bootStrapCopy.HdrIntegrityVerifier = nil
				return createArgBaseProcessor(coreComponents, dataComponents, &bootStrapCopy, statusComponents)
			},
			expectedErr: process.ErrNilHeaderIntegrityVerifier,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.EnableRoundsHandlerField = nil
				args := createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
				return args
			},
			expectedErr: process.ErrNilEnableRoundsHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.StatusCoreComponents = nil
				return args
			},
			expectedErr: process.ErrNilStatusCoreComponentsHolder,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.StatusCoreComponents = &factory.StatusCoreComponentsStub{
					AppStatusHandlerField: nil,
				}
				return args
			},
			expectedErr: process.ErrNilAppStatusHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.EnableEpochsHandlerField = nil
				return createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilEnableEpochsHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.EpochNotifierField = nil
				return createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
			},
			expectedErr: process.ErrNilEpochNotifier,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.GasHandler = nil
				return args
			},
			expectedErr: process.ErrNilGasHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.ScheduledTxsExecutionHandler = nil
				return args
			},
			expectedErr: process.ErrNilScheduledTxsExecutionHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.ProcessedMiniBlocksTracker = nil
				return args
			},
			expectedErr: process.ErrNilProcessedMiniBlocksTracker,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.ReceiptsRepository = nil
				return args
			},
			expectedErr: process.ErrNilReceiptsRepository,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				bootstrapCopy := *bootstrapComponents
				bootstrapCopy.VersionedHdrFactory = nil
				return createArgBaseProcessor(coreComponents, dataComponents, &bootstrapCopy, statusComponents)
			},
			expectedErr: process.ErrNilVersionedHeaderFactory,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.ManagedPeersHolder = nil
				return args
			},
			expectedErr: process.ErrNilManagedPeersHolder,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.OutportDataProvider = nil
				return args
			},
			expectedErr: process.ErrNilOutportDataProvider,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.BlockProcessingCutoffHandler = nil
				return args
			},
			expectedErr: process.ErrNilBlockProcessingCutoffHandler,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.ManagedPeersHolder = nil
				return args
			},
			expectedErr: process.ErrNilManagedPeersHolder,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.SentSignaturesTracker = nil
				return args
			},
			expectedErr: process.ErrNilSentSignatureTracker,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.ExecutionResultsInclusionEstimator = nil
				return args
			},
			expectedErr: process.ErrNilExecutionResultsInclusionEstimator,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.ExecutionManager = nil
				return args
			},
			expectedErr: process.ErrNilExecutionManager,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.MiniBlocksSelectionSession = nil
				return args
			},
			expectedErr: process.ErrNilMiniBlocksSelectionSession,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.ExecutionResultsVerifier = nil
				return args
			},
			expectedErr: process.ErrNilExecutionResultsVerifier,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.MissingDataResolver = nil
				return args
			},
			expectedErr: process.ErrNilMissingDataResolver,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				args := createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
				args.GasComputation = nil
				return args
			},
			expectedErr: process.ErrNilGasComputation,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				return createArgBaseProcessor(coreComponents, dataComponents, bootstrapComponents, statusComponents)
			},
			expectedErr: nil,
		},
		{
			args: func() blproc.ArgBaseProcessor {
				coreCompCopy := *coreComponents
				coreCompCopy.ClosingNodeStartedField = nil
				args := createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
				return args
			},
			expectedErr: process.ErrNilClosingNodeStartedFlag,
		},
	}

	for _, test := range tests {
		err := blproc.CheckProcessorNilParameters(test.args())
		require.Equal(t, test.expectedErr, err)
	}

	coreCompCopy := *coreComponents
	coreCompCopy.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStubWithNoFlagsDefined()
	args := createArgBaseProcessor(&coreCompCopy, dataComponents, bootstrapComponents, statusComponents)
	err := blproc.CheckProcessorNilParameters(args)
	require.True(t, errors.Is(err, core.ErrInvalidEnableEpochsHandler))
}

func TestBlockProcessor_CheckBlockValidity(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.Hash = &hashingMocks.HasherMock{}
	blkc := createTestBlockchain()
	dataComponents.BlockChain = blkc
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	bp, _ := blproc.NewShardProcessor(arguments)

	body := &block.Body{}
	hdr := &block.Header{}
	hdr.Nonce = 1
	hdr.Round = 1
	hdr.TimeStamp = 0
	hdr.PrevHash = []byte("X")
	err := bp.CheckBlockValidity(hdr, body)
	assert.Equal(t, process.ErrBlockHashDoesNotMatch, err)

	hdr.PrevHash = []byte("")
	err = bp.CheckBlockValidity(hdr, body)
	assert.Nil(t, err)

	hdr.Nonce = 2
	err = bp.CheckBlockValidity(hdr, body)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)

	blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
		return &block.Header{Round: 1, Nonce: 1}
	}
	prevHash := []byte("X")
	blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
		return prevHash
	}
	hdr = &block.Header{}

	err = bp.CheckBlockValidity(hdr, body)
	assert.Equal(t, process.ErrLowerRoundInBlock, err)

	hdr.Round = 2
	hdr.Nonce = 1
	err = bp.CheckBlockValidity(hdr, body)
	assert.Equal(t, process.ErrWrongNonceInBlock, err)

	hdr.Nonce = 2
	hdr.PrevHash = []byte("XX")
	err = bp.CheckBlockValidity(hdr, body)
	assert.Equal(t, process.ErrBlockHashDoesNotMatch, err)

	hdr.PrevHash = blkc.GetCurrentBlockHeaderHash()
	hdr.PrevRandSeed = []byte("X")
	err = bp.CheckBlockValidity(hdr, body)
	assert.Equal(t, process.ErrRandSeedDoesNotMatch, err)

	hdr.PrevRandSeed = []byte("")
	err = bp.CheckBlockValidity(hdr, body)
	assert.Nil(t, err)
}

func TestBlockProcessor_CheckBlockValidityTimestamp(t *testing.T) {
	t.Parallel()

	t.Run("genesis+1 block with valid timestamp should pass", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.Hash = &hashingMocks.HasherMock{}
		coreComponents.RoundField = &testscommon.RoundHandlerMock{
			GetTimeStampForRoundCalled: func(round uint64) uint64 {
				return round * 6000 // 6s per round in ms
			},
		}

		blkc := createTestBlockchain()
		dataComponents.BlockChain = blkc
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		body := &block.Body{}
		hdr := &block.Header{}
		hdr.Nonce = 1
		hdr.Round = 1
		hdr.TimeStamp = 6 // 6 seconds = 6000ms
		hdr.PrevHash = []byte("")

		err := bp.CheckBlockValidity(hdr, body)
		assert.Nil(t, err)
	})

	t.Run("genesis+1 block with invalid timestamp should fail", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.Hash = &hashingMocks.HasherMock{}
		coreComponents.RoundField = &testscommon.RoundHandlerMock{
			GetTimeStampForRoundCalled: func(round uint64) uint64 {
				return round * 6000
			},
		}

		blkc := createTestBlockchain()
		dataComponents.BlockChain = blkc
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		body := &block.Body{}
		hdr := &block.Header{}
		hdr.Nonce = 1
		hdr.Round = 1
		hdr.TimeStamp = 999 // wrong timestamp
		hdr.PrevHash = []byte("")

		err := bp.CheckBlockValidity(hdr, body)
		assert.Equal(t, process.ErrInvalidTimestamp, err)
	})

	t.Run("non-genesis block with valid timestamp should pass", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.Hash = &hashingMocks.HasherMock{}
		coreComponents.RoundField = &testscommon.RoundHandlerMock{
			GetTimeStampForRoundCalled: func(round uint64) uint64 {
				return round * 6000
			},
		}

		blkc := createTestBlockchain()
		prevHash := []byte("X")
		blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
			return &block.Header{Round: 1, Nonce: 1}
		}
		blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
			return prevHash
		}
		dataComponents.BlockChain = blkc
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		body := &block.Body{}
		hdr := &block.Header{}
		hdr.Nonce = 2
		hdr.Round = 2
		hdr.TimeStamp = 12 // 12 seconds = 12000ms
		hdr.PrevHash = prevHash
		hdr.PrevRandSeed = []byte("")

		err := bp.CheckBlockValidity(hdr, body)
		assert.Nil(t, err)
	})

	t.Run("non-genesis block with invalid timestamp should fail", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.Hash = &hashingMocks.HasherMock{}
		coreComponents.RoundField = &testscommon.RoundHandlerMock{
			GetTimeStampForRoundCalled: func(round uint64) uint64 {
				return round * 6000
			},
		}

		blkc := createTestBlockchain()
		prevHash := []byte("X")
		blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
			return &block.Header{Round: 1, Nonce: 1}
		}
		blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
			return prevHash
		}
		dataComponents.BlockChain = blkc
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		body := &block.Body{}
		hdr := &block.Header{}
		hdr.Nonce = 2
		hdr.Round = 2
		hdr.TimeStamp = 999 // wrong timestamp
		hdr.PrevHash = prevHash
		hdr.PrevRandSeed = []byte("")

		err := bp.CheckBlockValidity(hdr, body)
		assert.Equal(t, process.ErrInvalidTimestamp, err)
	})

	t.Run("other checks still fail before timestamp check", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.Hash = &hashingMocks.HasherMock{}
		coreComponents.RoundField = &testscommon.RoundHandlerMock{
			GetTimeStampForRoundCalled: func(round uint64) uint64 {
				return round * 6000
			},
		}

		blkc := createTestBlockchain()
		blkc.GetCurrentBlockHeaderCalled = func() data.HeaderHandler {
			return &block.Header{Round: 1, Nonce: 1}
		}
		blkc.GetCurrentBlockHeaderHashCalled = func() []byte {
			return []byte("X")
		}
		dataComponents.BlockChain = blkc
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		body := &block.Body{}
		hdr := &block.Header{}
		hdr.Nonce = 2
		hdr.Round = 0 // invalid round
		hdr.TimeStamp = 999

		err := bp.CheckBlockValidity(hdr, body)
		assert.Equal(t, process.ErrLowerRoundInBlock, err)
	})
}

func TestVerifyStateRoot_ShouldWork(t *testing.T) {
	t.Parallel()
	rootHash := []byte("root hash to be tested")
	accounts := &stateMock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
	}

	arguments := CreateMockArguments(createComponentHolderMocks())
	arguments.AccountsDB[state.UserAccountsState] = accounts
	bp, _ := blproc.NewShardProcessor(arguments)

	assert.True(t, bp.VerifyStateRoot(rootHash))
}

func TestBaseProcessor_SetIndexOfFirstTxProcessed(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments(createComponentHolderMocks())
	processedMiniBlocksTracker := processedMb.NewProcessedMiniBlocksTracker()
	arguments.ProcessedMiniBlocksTracker = processedMiniBlocksTracker
	bp, _ := blproc.NewShardProcessor(arguments)

	metaHash := []byte("meta_hash")
	mbHash := []byte("mb_hash")
	miniBlockHeader := &block.MiniBlockHeader{
		Hash: mbHash,
	}

	processedMbInfo := &processedMb.ProcessedMiniBlockInfo{
		FullyProcessed:         false,
		IndexOfLastTxProcessed: 8,
	}
	processedMiniBlocksTracker.SetProcessedMiniBlockInfo(metaHash, mbHash, processedMbInfo)
	err := bp.SetIndexOfFirstTxProcessed(miniBlockHeader)
	assert.Nil(t, err)
	assert.Equal(t, int32(9), miniBlockHeader.GetIndexOfFirstTxProcessed())
}

func TestBaseProcessor_CheckHeaderBodyCorrelationIndexOfFirstTxProcessed(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherStub{}
	marshaller := &mock.MarshalizerMock{}

	t.Run("fresh incoming mb with non-zero IndexOfFirstTxProcessed should error", func(t *testing.T) {
		t.Parallel()

		hdr, body := createOneHeaderOneBody()
		hdr.MiniBlockHeaders[0].TxCount = 3
		body.MiniBlocks[0].TxHashes = [][]byte{[]byte("tx1"), []byte("tx2"), []byte("tx3")}
		mbBytes, _ := marshaller.Marshal(body.MiniBlocks[0])
		hdr.MiniBlockHeaders[0].Hash = hasher.Compute(string(mbBytes))
		_ = hdr.MiniBlockHeaders[0].SetIndexOfFirstTxProcessed(2)
		_ = hdr.MiniBlockHeaders[0].SetIndexOfLastTxProcessed(2)

		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.ProcessedMiniBlocksTracker = processedMb.NewProcessedMiniBlocksTracker()
		sp, _ := blproc.NewShardProcessor(arguments)

		err := sp.CheckHeaderBodyCorrelation(hdr, body)
		assert.Equal(t, process.ErrIndexOfFirstTxProcessedMismatch, err)
	})

	t.Run("fresh incoming mb with IndexOfFirstTxProcessed=0 should pass", func(t *testing.T) {
		t.Parallel()

		hdr, body := createOneHeaderOneBody()

		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.ProcessedMiniBlocksTracker = processedMb.NewProcessedMiniBlocksTracker()
		sp, _ := blproc.NewShardProcessor(arguments)

		err := sp.CheckHeaderBodyCorrelation(hdr, body)
		assert.Nil(t, err)
	})

	t.Run("partially processed mb with matching continuation should pass", func(t *testing.T) {
		t.Parallel()

		hdr, body := createOneHeaderOneBody()
		hdr.MiniBlockHeaders[0].TxCount = 5
		body.MiniBlocks[0].TxHashes = [][]byte{[]byte("tx1"), []byte("tx2"), []byte("tx3"), []byte("tx4"), []byte("tx5")}
		mbBytes, _ := marshaller.Marshal(body.MiniBlocks[0])
		mbHash := hasher.Compute(string(mbBytes))
		hdr.MiniBlockHeaders[0].Hash = mbHash
		// tracker says we already processed indices 0, 1, 2 so next first must be 3
		_ = hdr.MiniBlockHeaders[0].SetIndexOfFirstTxProcessed(3)
		_ = hdr.MiniBlockHeaders[0].SetIndexOfLastTxProcessed(4)

		arguments := CreateMockArguments(createComponentHolderMocks())
		tracker := processedMb.NewProcessedMiniBlocksTracker()
		tracker.SetProcessedMiniBlockInfo([]byte("meta_hash"), mbHash, &processedMb.ProcessedMiniBlockInfo{
			FullyProcessed:         false,
			IndexOfLastTxProcessed: 2,
		})
		arguments.ProcessedMiniBlocksTracker = tracker
		sp, _ := blproc.NewShardProcessor(arguments)

		err := sp.CheckHeaderBodyCorrelation(hdr, body)
		assert.Nil(t, err)
	})

	t.Run("partially processed mb with mismatched continuation should error", func(t *testing.T) {
		t.Parallel()

		hdr, body := createOneHeaderOneBody()
		hdr.MiniBlockHeaders[0].TxCount = 5
		body.MiniBlocks[0].TxHashes = [][]byte{[]byte("tx1"), []byte("tx2"), []byte("tx3"), []byte("tx4"), []byte("tx5")}
		mbBytes, _ := marshaller.Marshal(body.MiniBlocks[0])
		mbHash := hasher.Compute(string(mbBytes))
		hdr.MiniBlockHeaders[0].Hash = mbHash
		// tracker says next first must be 3, but proposer forged 0
		_ = hdr.MiniBlockHeaders[0].SetIndexOfFirstTxProcessed(0)
		_ = hdr.MiniBlockHeaders[0].SetIndexOfLastTxProcessed(4)

		arguments := CreateMockArguments(createComponentHolderMocks())
		tracker := processedMb.NewProcessedMiniBlocksTracker()
		tracker.SetProcessedMiniBlockInfo([]byte("meta_hash"), mbHash, &processedMb.ProcessedMiniBlockInfo{
			FullyProcessed:         false,
			IndexOfLastTxProcessed: 2,
		})
		arguments.ProcessedMiniBlocksTracker = tracker
		sp, _ := blproc.NewShardProcessor(arguments)

		err := sp.CheckHeaderBodyCorrelation(hdr, body)
		assert.Equal(t, process.ErrIndexOfFirstTxProcessedMismatch, err)
	})

	t.Run("intra shard mb should skip the tracker check", func(t *testing.T) {
		t.Parallel()

		hdr, body := createOneHeaderOneBody()
		hdr.MiniBlockHeaders[0].TxCount = 3
		body.MiniBlocks[0].TxHashes = [][]byte{[]byte("tx1"), []byte("tx2"), []byte("tx3")}
		body.MiniBlocks[0].SenderShardID = 0
		body.MiniBlocks[0].ReceiverShardID = 0
		hdr.MiniBlockHeaders[0].SenderShardID = 0
		hdr.MiniBlockHeaders[0].ReceiverShardID = 0
		mbBytes, _ := marshaller.Marshal(body.MiniBlocks[0])
		hdr.MiniBlockHeaders[0].Hash = hasher.Compute(string(mbBytes))
		_ = hdr.MiniBlockHeaders[0].SetIndexOfFirstTxProcessed(1)
		_ = hdr.MiniBlockHeaders[0].SetIndexOfLastTxProcessed(2)

		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.ProcessedMiniBlocksTracker = processedMb.NewProcessedMiniBlocksTracker()
		sp, _ := blproc.NewShardProcessor(arguments)

		err := sp.CheckHeaderBodyCorrelation(hdr, body)
		assert.Nil(t, err)
	})
}

func TestBaseProcessor_SetIndexOfLastTxProcessed(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments(createComponentHolderMocks())
	bp, _ := blproc.NewShardProcessor(arguments)

	mbHash := []byte("mb_hash")
	processedMiniBlocksDestMeInfo := make(map[string]*processedMb.ProcessedMiniBlockInfo)
	miniBlockHeader := &block.MiniBlockHeader{
		Hash:    mbHash,
		TxCount: 100,
	}

	err := bp.SetIndexOfLastTxProcessed(miniBlockHeader, processedMiniBlocksDestMeInfo)
	assert.Nil(t, err)
	assert.Equal(t, int32(99), miniBlockHeader.GetIndexOfLastTxProcessed())

	processedMbInfo := &processedMb.ProcessedMiniBlockInfo{
		FullyProcessed:         false,
		IndexOfLastTxProcessed: 8,
	}
	processedMiniBlocksDestMeInfo[string(mbHash)] = processedMbInfo

	err = bp.SetIndexOfLastTxProcessed(miniBlockHeader, processedMiniBlocksDestMeInfo)
	assert.Nil(t, err)
	assert.Equal(t, int32(8), miniBlockHeader.GetIndexOfLastTxProcessed())
}

func TestBaseProcessor_SetProcessingTypeAndConstructionStateForScheduledMb(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments(createComponentHolderMocks())
	bp, _ := blproc.NewShardProcessor(arguments)

	mbHash := []byte("mb_hash")
	processedMiniBlocksDestMeInfo := make(map[string]*processedMb.ProcessedMiniBlockInfo)
	miniBlockHeader := &block.MiniBlockHeader{
		Hash: mbHash,
	}

	processedMbInfo := &processedMb.ProcessedMiniBlockInfo{
		FullyProcessed: false,
	}

	miniBlockHeader.SenderShardID = 0
	err := bp.SetProcessingTypeAndConstructionStateForScheduledMb(miniBlockHeader, processedMiniBlocksDestMeInfo)
	assert.Nil(t, err)
	assert.Equal(t, int32(block.Proposed), miniBlockHeader.GetConstructionState())
	assert.Equal(t, int32(block.Scheduled), miniBlockHeader.GetProcessingType())

	miniBlockHeader.SenderShardID = 1

	err = bp.SetProcessingTypeAndConstructionStateForScheduledMb(miniBlockHeader, processedMiniBlocksDestMeInfo)
	assert.Nil(t, err)
	assert.Equal(t, int32(block.Final), miniBlockHeader.GetConstructionState())
	assert.Equal(t, int32(block.Scheduled), miniBlockHeader.GetProcessingType())

	processedMiniBlocksDestMeInfo[string(mbHash)] = processedMbInfo

	err = bp.SetProcessingTypeAndConstructionStateForScheduledMb(miniBlockHeader, processedMiniBlocksDestMeInfo)
	assert.Nil(t, err)
	assert.Equal(t, int32(block.PartialExecuted), miniBlockHeader.GetConstructionState())
	assert.Equal(t, int32(block.Scheduled), miniBlockHeader.GetProcessingType())
}

func TestBaseProcessor_SetProcessingTypeAndConstructionStateForNormalMb(t *testing.T) {
	t.Parallel()

	t.Run("set processing/construction for normal mini blocks not processed, should work", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		mbHash := []byte("mb_hash")
		processedMiniBlocksDestMeInfo := make(map[string]*processedMb.ProcessedMiniBlockInfo)
		miniBlockHeader := &block.MiniBlockHeader{
			Hash: mbHash,
		}

		processedMbInfo := &processedMb.ProcessedMiniBlockInfo{
			FullyProcessed: false,
		}

		err := bp.SetProcessingTypeAndConstructionStateForNormalMb(miniBlockHeader, processedMiniBlocksDestMeInfo)
		assert.Nil(t, err)
		assert.Equal(t, int32(block.Final), miniBlockHeader.GetConstructionState())
		assert.Equal(t, int32(block.Normal), miniBlockHeader.GetProcessingType())

		processedMiniBlocksDestMeInfo[string(mbHash)] = processedMbInfo

		err = bp.SetProcessingTypeAndConstructionStateForNormalMb(miniBlockHeader, processedMiniBlocksDestMeInfo)
		assert.Nil(t, err)
		assert.Equal(t, int32(block.PartialExecuted), miniBlockHeader.GetConstructionState())
		assert.Equal(t, int32(block.Normal), miniBlockHeader.GetProcessingType())
	})

	t.Run("set processing/construction for normal mini blocks already processed, should work", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			IsMiniBlockExecutedCalled: func(i []byte) bool {
				return true
			},
		}
		bp, _ := blproc.NewShardProcessor(arguments)

		mbHash := []byte("mb_hash")
		processedMiniBlocksDestMeInfo := make(map[string]*processedMb.ProcessedMiniBlockInfo)
		miniBlockHeader := &block.MiniBlockHeader{
			Hash: mbHash,
		}

		processedMbInfo := &processedMb.ProcessedMiniBlockInfo{
			FullyProcessed: false,
		}

		err := bp.SetProcessingTypeAndConstructionStateForNormalMb(miniBlockHeader, processedMiniBlocksDestMeInfo)
		assert.Nil(t, err)
		assert.Equal(t, int32(block.Final), miniBlockHeader.GetConstructionState())
		assert.Equal(t, int32(block.Processed), miniBlockHeader.GetProcessingType())

		processedMiniBlocksDestMeInfo[string(mbHash)] = processedMbInfo

		err = bp.SetProcessingTypeAndConstructionStateForNormalMb(miniBlockHeader, processedMiniBlocksDestMeInfo)
		assert.Nil(t, err)
		assert.Equal(t, int32(block.PartialExecuted), miniBlockHeader.GetConstructionState())
		assert.Equal(t, int32(block.Processed), miniBlockHeader.GetProcessingType())
	})
}

// ------- RevertState
func TestBaseProcessor_RevertStateRecreateTrieFailsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("err")
	arguments := CreateMockArguments(createComponentHolderMocks())
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		RecreateTrieCalled: func(rootHash common.RootHashHolder) error {
			return expectedErr
		},
	}

	bp, _ := blproc.NewShardProcessor(arguments)

	hdr := block.Header{Nonce: 37}
	err := bp.RevertStateToBlock(&hdr, hdr.RootHash)
	assert.Equal(t, expectedErr, err)
}

// removeHeadersBehindNonceFromPools
func TestBaseProcessor_RemoveHeadersBehindNonceFromPools(t *testing.T) {
	t.Parallel()

	removeFromDataPoolWasCalled := false
	dataPool := initDataPool()
	dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
		cs := &mock.HeadersCacherStub{}
		cs.RegisterHandlerCalled = func(i func(header data.HeaderHandler, key []byte)) {
		}
		cs.GetHeaderByHashCalled = func(hash []byte) (handler data.HeaderHandler, err error) {
			return nil, err
		}
		cs.RemoveHeaderByHashCalled = func(key []byte) {
		}
		cs.LenCalled = func() int {
			return 0
		}
		cs.MaxSizeCalled = func() int {
			return 1000
		}
		cs.NoncesCalled = func(shardId uint32) []uint64 {
			return []uint64{1, 2, 3}
		}
		cs.GetHeaderByNonceAndShardIdCalled = func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
			hdrs := make([]data.HeaderHandler, 0)
			hdrs = append(hdrs, &block.Header{Nonce: 2})
			return hdrs, nil, nil
		}

		return cs
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = dataPool
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
		RemoveBlockDataFromPoolCalled: func(body *block.Body) error {
			removeFromDataPoolWasCalled = true
			return nil
		},
	}
	bp, _ := blproc.NewShardProcessor(arguments)

	bp.RemoveHeadersBehindNonceFromPools(true, 0, 4)

	assert.True(t, removeFromDataPoolWasCalled)
}

// ------- ComputeNewNoncePrevHash

func TestBlockProcessor_computeHeaderHashMarshalizerFail1ShouldErr(t *testing.T) {
	t.Parallel()
	marshalizer := &mock.MarshalizerStub{}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.IntMarsh = marshalizer
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	bp, _ := blproc.NewShardProcessor(arguments)
	hdr, txBlock := createTestHdrTxBlockBody()
	expectedError := errors.New("marshalizer fail")
	marshalizer.MarshalCalled = func(obj interface{}) (bytes []byte, e error) {
		if hdr == obj {
			return nil, expectedError
		}

		if reflect.DeepEqual(txBlock, obj) {
			return []byte("txBlockBodyMarshalized"), nil
		}
		return nil, nil
	}
	_, err := bp.ComputeHeaderHash(hdr)
	assert.Equal(t, expectedError, err)
}

func TestBlockPorcessor_ComputeNewNoncePrevHashShouldWork(t *testing.T) {
	t.Parallel()
	marshalizer := &mock.MarshalizerStub{}
	hasher := &mock.HasherStub{}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.IntMarsh = marshalizer
	coreComponents.Hash = hasher
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	bp, _ := blproc.NewShardProcessor(arguments)
	hdr, txBlock := createTestHdrTxBlockBody()
	marshalizer.MarshalCalled = func(obj interface{}) (bytes []byte, e error) {
		if hdr == obj {
			return []byte("hdrHeaderMarshalized"), nil
		}
		if reflect.DeepEqual(txBlock, obj) {
			return []byte("txBlockBodyMarshalized"), nil
		}
		return nil, nil
	}
	hasher.ComputeCalled = func(s string) []byte {
		if s == "hdrHeaderMarshalized" {
			return []byte("hdr hash")
		}
		if s == "txBlockBodyMarshalized" {
			return []byte("tx block body hash")
		}
		return nil
	}
	_, err := bp.ComputeHeaderHash(hdr)
	assert.Nil(t, err)
}

func createShardProcessHeadersToSaveLastNotarized(
	highestNonce uint64,
	genesisHdr data.HeaderHandler,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) []data.HeaderHandler {
	rootHash := []byte("roothash")
	processedHdrs := make([]data.HeaderHandler, 0)

	headerMarsh, _ := marshalizer.Marshal(genesisHdr)
	headerHash := hasher.Compute(string(headerMarsh))

	for i := uint64(1); i <= highestNonce; i++ {
		hdr := &block.Header{
			Nonce:         i,
			Round:         i,
			Signature:     rootHash,
			RandSeed:      rootHash,
			PrevRandSeed:  rootHash,
			PubKeysBitmap: rootHash,
			RootHash:      rootHash,
			PrevHash:      headerHash}
		processedHdrs = append(processedHdrs, hdr)

		headerMarsh, _ = marshalizer.Marshal(hdr)
		headerHash = hasher.Compute(string(headerMarsh))
	}

	return processedHdrs
}

func createMetaProcessHeadersToSaveLastNoterized(
	highestNonce uint64,
	genesisHdr data.HeaderHandler,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) []data.HeaderHandler {
	rootHash := []byte("roothash")
	processedHdrs := make([]data.HeaderHandler, 0)

	headerMarsh, _ := marshalizer.Marshal(genesisHdr)
	headerHash := hasher.Compute(string(headerMarsh))

	for i := uint64(1); i <= highestNonce; i++ {
		hdr := &block.MetaBlock{
			Nonce:         i,
			Round:         i,
			Signature:     rootHash,
			RandSeed:      rootHash,
			PrevRandSeed:  rootHash,
			PubKeysBitmap: rootHash,
			RootHash:      rootHash,
			PrevHash:      headerHash}
		processedHdrs = append(processedHdrs, hdr)

		headerMarsh, _ = marshalizer.Marshal(hdr)
		headerHash = hasher.Compute(string(headerMarsh))
	}

	return processedHdrs
}

func TestBaseProcessor_SaveLastNotarizedInOneShardHdrsSliceForShardIsNil(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.Hash = &hashingMocks.HasherMock{}
	coreComponents.IntMarsh = &mock.MarshalizerMock{}
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)
	prHdrs := createShardProcessHeadersToSaveLastNotarized(10, &block.Header{}, &hashingMocks.HasherMock{}, &mock.MarshalizerMock{})

	err := sp.SaveLastNotarizedHeader(2, prHdrs)

	assert.Equal(t, process.ErrNotarizedHeadersSliceForShardIsNil, err)
}

func TestBaseProcessor_SaveLastNotarizedInMultiShardHdrsSliceForShardIsNil(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.Hash = &hashingMocks.HasherMock{}
	coreComponents.IntMarsh = &mock.MarshalizerMock{}
	bootstrapComponents.Coordinator = shardCoordinator
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	prHdrs := createShardProcessHeadersToSaveLastNotarized(10, &block.Header{}, &hashingMocks.HasherMock{}, &mock.MarshalizerMock{})

	err := sp.SaveLastNotarizedHeader(6, prHdrs)

	assert.Equal(t, process.ErrNotarizedHeadersSliceForShardIsNil, err)
}

func TestBaseProcessor_SaveLastNotarizedHdrShardGood(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.Hash = &hashingMocks.HasherMock{}
	coreComponents.IntMarsh = &mock.MarshalizerMock{}
	bootstrapComponents.Coordinator = shardCoordinator
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	sp, _ := blproc.NewShardProcessor(arguments)
	argsHeaderValidator := blproc.ArgsHeaderValidator{
		Hasher:              coreComponents.Hasher(),
		Marshalizer:         coreComponents.InternalMarshalizer(),
		EnableEpochsHandler: coreComponents.EnableEpochsHandler(),
	}
	headerValidator, _ := blproc.NewHeaderValidator(argsHeaderValidator)
	sp.SetHeaderValidator(headerValidator)

	genesisBlcks := createGenesisBlocks(shardCoordinator)

	highestNonce := uint64(10)
	shardId := uint32(0)
	prHdrs := createShardProcessHeadersToSaveLastNotarized(
		highestNonce,
		genesisBlcks[shardId],
		coreComponents.Hasher(),
		coreComponents.InternalMarshalizer())

	err := sp.SaveLastNotarizedHeader(shardId, prHdrs)
	assert.Nil(t, err)

	assert.Equal(t, highestNonce, sp.LastNotarizedHdrForShard(shardId).GetNonce())
}

func TestBaseProcessor_SaveLastNotarizedHdrMetaGood(t *testing.T) {
	t.Parallel()

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(5)
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.Hash = &hashingMocks.HasherMock{}
	coreComponents.IntMarsh = &mock.MarshalizerMock{}
	bootstrapComponents.Coordinator = shardCoordinator
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	argsHeaderValidator := blproc.ArgsHeaderValidator{
		Hasher:              coreComponents.Hasher(),
		Marshalizer:         coreComponents.InternalMarshalizer(),
		EnableEpochsHandler: coreComponents.EnableEpochsHandler(),
	}
	headerValidator, _ := blproc.NewHeaderValidator(argsHeaderValidator)
	sp.SetHeaderValidator(headerValidator)

	genesisBlcks := createGenesisBlocks(shardCoordinator)

	highestNonce := uint64(10)
	prHdrs := createMetaProcessHeadersToSaveLastNoterized(
		highestNonce,
		genesisBlcks[core.MetachainShardId],
		coreComponents.Hasher(),
		coreComponents.InternalMarshalizer())

	err := sp.SaveLastNotarizedHeader(core.MetachainShardId, prHdrs)
	assert.Nil(t, err)

	assert.Equal(t, highestNonce, sp.LastNotarizedHdrForShard(core.MetachainShardId).GetNonce())
}

func TestShardProcessor_ProcessBlockEpochDoesNotMatchShouldErr(t *testing.T) {
	t.Parallel()
	blockChain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Epoch: 2,
			}
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}
	coreComponents, dataComponents, bootstrapComponents, statusComponents := CreateCoreComponentsMultiShard()
	dataComponents.BlockChain = blockChain
	arguments := CreateMockArgumentsMultiShard(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)
	header := &block.Header{Round: 10, Nonce: 1}

	blk := &block.Body{}
	err := sp.ProcessBlock(header, blk, func() time.Duration { return time.Second })

	assert.True(t, errors.Is(err, process.ErrEpochDoesNotMatch))
}

func TestShardProcessor_ProcessBlockEpochDoesNotMatchShouldErr2(t *testing.T) {
	t.Parallel()

	randSeed := []byte("randseed")
	blockChain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Epoch:           1,
				RandSeed:        randSeed,
				AccumulatedFees: big.NewInt(0),
				DeveloperFees:   big.NewInt(0),
			}
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := CreateCoreComponentsMultiShard()
	dataComponents.BlockChain = blockChain
	arguments := CreateMockArgumentsMultiShard(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		EpochCalled: func() uint32 {
			return 1
		},
	}

	sp, _ := blproc.NewShardProcessor(arguments)
	header := &block.Header{Round: 10, Nonce: 1, Epoch: 5, RandSeed: randSeed, PrevRandSeed: randSeed}

	blk := &block.Body{}
	err := sp.ProcessBlock(header, blk, func() time.Duration { return time.Second })

	assert.True(t, errors.Is(err, process.ErrEpochDoesNotMatch))
}

func TestShardProcessor_ProcessBlockEpochDoesNotMatchShouldErr3(t *testing.T) {
	t.Parallel()

	randSeed := []byte("randseed")
	blockChain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Epoch:    3,
				RandSeed: randSeed,
			}
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := CreateCoreComponentsMultiShard()
	dataComponents.BlockChain = blockChain
	arguments := CreateMockArgumentsMultiShard(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		EpochCalled: func() uint32 {
			return 2
		},
		IsEpochStartCalled: func() bool {
			return true
		},
	}

	sp, _ := blproc.NewShardProcessor(arguments)
	header := &block.Header{Round: 10, Nonce: 1, Epoch: 5, RandSeed: randSeed, PrevRandSeed: randSeed}

	blk := &block.Body{}
	err := sp.ProcessBlock(header, blk, func() time.Duration { return time.Second })

	assert.True(t, errors.Is(err, process.ErrEpochDoesNotMatch))
}

func TestShardProcessor_ProcessBlockEpochDoesNotMatchShouldErrMetaHashDoesNotMatch(t *testing.T) {
	t.Parallel()

	randSeed := []byte("randseed")
	chain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Epoch:    2,
				RandSeed: randSeed,
			}
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}
	hasher := &mock.HasherStub{ComputeCalled: func(s string) []byte {
		return nil
	}}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := CreateCoreComponentsMultiShard()
	dataComponents.BlockChain = chain
	coreComponents.Hash = hasher
	arguments := CreateMockArgumentsMultiShard(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	epochStartTrigger := &mock.EpochStartTriggerStub{
		EpochCalled: func() uint32 {
			return 2
		},
		MetaEpochCalled: func() uint32 {
			return 3
		},
		IsEpochStartCalled: func() bool {
			return true
		},
		EpochFinalityAttestingRoundCalled: func() uint64 {
			return 100
		},
	}
	arguments.EpochStartTrigger = epochStartTrigger

	sp, _ := blproc.NewShardProcessor(arguments)
	rootHash, _ := arguments.AccountsDB[state.UserAccountsState].RootHash()
	epochStartHash := []byte("epochStartHash")
	header := &block.Header{
		Round:              10,
		Nonce:              1,
		Epoch:              3,
		RandSeed:           randSeed,
		PrevRandSeed:       randSeed,
		EpochStartMetaHash: epochStartHash,
		RootHash:           rootHash,
		AccumulatedFees:    big.NewInt(0),
		DeveloperFees:      big.NewInt(0),
	}

	blk := &block.Body{}
	err := sp.ProcessBlock(header, blk, func() time.Duration { return time.Second })
	assert.True(t, errors.Is(err, process.ErrEpochDoesNotMatch))

	epochStartTrigger.EpochStartMetaHdrHashCalled = func() []byte {
		return header.EpochStartMetaHash
	}
	err = sp.ProcessBlock(header, blk, func() time.Duration { return time.Second })
	assert.Nil(t, err)
}

func TestShardProcessor_ProcessBlockEpochDoesNotMatchShouldErrMetaHashDoesNotMatchForOldEpoch(t *testing.T) {
	t.Parallel()

	randSeed := []byte("randseed")
	chain := &testscommon.ChainHandlerStub{
		GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
			return &block.Header{
				Epoch:    2,
				RandSeed: randSeed,
			}
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{Nonce: 0}
		},
	}
	hasher := &mock.HasherStub{ComputeCalled: func(s string) []byte {
		return nil
	}}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := CreateCoreComponentsMultiShard()
	coreComponents.Hash = hasher
	dataComponents.BlockChain = chain
	arguments := CreateMockArgumentsMultiShard(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	arguments.EpochStartTrigger = &mock.EpochStartTriggerStub{
		EpochCalled: func() uint32 {
			return 5
		},
		MetaEpochCalled: func() uint32 {
			return 6
		},
		IsEpochStartCalled: func() bool {
			return true
		},
		EpochFinalityAttestingRoundCalled: func() uint64 {
			return 100
		},
	}

	sp, _ := blproc.NewShardProcessor(arguments)
	rootHash, _ := arguments.AccountsDB[state.UserAccountsState].RootHash()
	epochStartHash := []byte("epochStartHash")
	header := &block.Header{
		Round:              10,
		Nonce:              1,
		Epoch:              3,
		RandSeed:           randSeed,
		PrevRandSeed:       randSeed,
		EpochStartMetaHash: epochStartHash,
		RootHash:           rootHash,
		AccumulatedFees:    big.NewInt(0),
		DeveloperFees:      big.NewInt(0),
	}

	blk := &block.Body{}
	err := sp.ProcessBlock(header, blk, func() time.Duration { return time.Second })
	assert.True(t, errors.Is(err, process.ErrMissingHeader))

	metaHdr := &block.MetaBlock{}
	metaHdrData, _ := coreComponents.InternalMarshalizer().Marshal(metaHdr)
	_ = dataComponents.StorageService().Put(dataRetriever.MetaBlockUnit, header.EpochStartMetaHash, metaHdrData)

	err = sp.ProcessBlock(header, blk, func() time.Duration { return time.Second })
	assert.True(t, errors.Is(err, process.ErrEpochDoesNotMatch))

	metaHdr = &block.MetaBlock{Epoch: 3, EpochStart: block.EpochStart{
		LastFinalizedHeaders: []block.EpochStartShardData{{}},
		Economics:            block.Economics{},
	}}
	metaHdrData, _ = coreComponents.InternalMarshalizer().Marshal(metaHdr)
	_ = dataComponents.StorageService().Put(dataRetriever.MetaBlockUnit, header.EpochStartMetaHash, metaHdrData)

	err = sp.ProcessBlock(header, blk, func() time.Duration { return time.Second })
	assert.Nil(t, err)
}

func TestBlockProcessor_PruneStateOnRollbackPrunesPeerTrieIfAccPruneIsDisabled(t *testing.T) {
	t.Parallel()

	pruningCalled := 0
	peerAccDb := &stateMock.AccountsStub{
		PruneTrieCalled: func(rootHash []byte, identifier state.TriePruningIdentifier, _ state.PruningHandler) {
			pruningCalled++
		},
		CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
			pruningCalled++
		},
		IsPruningEnabledCalled: func() bool {
			return true
		},
	}

	arguments := CreateMockArguments(createComponentHolderMocks())
	arguments.AccountsDB[state.PeerAccountsState] = peerAccDb
	bp, _ := blproc.NewShardProcessor(arguments)

	prevHeader := &block.MetaBlock{
		RootHash:               []byte("prevRootHash"),
		ValidatorStatsRootHash: []byte("prevValidatorRootHash"),
	}
	currHeader := &block.MetaBlock{
		RootHash:               []byte("prevRootHash"),
		ValidatorStatsRootHash: []byte("currValidatorRootHash"),
	}

	bp.PruneStateOnRollback(currHeader, []byte("currHeaderHash"), prevHeader, []byte("prevHeaderHash"))
	assert.Equal(t, 2, pruningCalled)
}

func TestBlockProcessor_PruneStateOnRollbackPrunesPeerTrieIfSameRootHashButDifferentValidatorRootHash(t *testing.T) {
	t.Parallel()

	pruningCalled := 0
	peerAccDb := &stateMock.AccountsStub{
		PruneTrieCalled: func(rootHash []byte, identifier state.TriePruningIdentifier, _ state.PruningHandler) {
			pruningCalled++
		},
		CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
			pruningCalled++
		},
		IsPruningEnabledCalled: func() bool {
			return true
		},
	}

	accDb := &stateMock.AccountsStub{
		PruneTrieCalled: func(rootHash []byte, identifier state.TriePruningIdentifier, _ state.PruningHandler) {
			pruningCalled++
		},
		CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
			pruningCalled++
		},
		IsPruningEnabledCalled: func() bool {
			return true
		},
	}

	arguments := CreateMockArguments(createComponentHolderMocks())
	arguments.AccountsDB[state.PeerAccountsState] = peerAccDb
	arguments.AccountsDB[state.UserAccountsState] = accDb
	bp, _ := blproc.NewShardProcessor(arguments)

	prevHeader := &block.MetaBlock{
		RootHash:               []byte("prevRootHash"),
		ValidatorStatsRootHash: []byte("prevValidatorRootHash"),
	}
	currHeader := &block.MetaBlock{
		RootHash:               []byte("prevRootHash"),
		ValidatorStatsRootHash: []byte("currValidatorRootHash"),
	}

	bp.PruneStateOnRollback(currHeader, []byte("currHeaderHash"), prevHeader, []byte("prevHeaderHash"))
	assert.Equal(t, 2, pruningCalled)
}

func TestBlockProcessor_RequestHeadersIfMissingShouldWorkWhenSortedHeadersListIsEmpty(t *testing.T) {
	t.Parallel()

	var requestedNonces []uint64
	var mutRequestedNonces sync.Mutex

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	roundHandler := &mock.RoundHandlerMock{}
	coreComponents.RoundField = roundHandler
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	requestHandlerStub := &testscommon.RequestHandlerStub{
		RequestMetaHeaderByNonceCalled: func(nonce uint64) {
			mutRequestedNonces.Lock()
			requestedNonces = append(requestedNonces, nonce)
			mutRequestedNonces.Unlock()
		},
	}
	arguments.RequestHandler = requestHandlerStub
	sp, _ := blproc.NewShardProcessor(arguments)

	sortedHeaders := make([]data.HeaderHandler, 0)

	requestedNonces = make([]uint64, 0)
	roundHandler.RoundIndex = process.MaxHeaderRequestsAllowed + 5
	_ = sp.RequestHeadersIfMissing(sortedHeaders, core.MetachainShardId)
	time.Sleep(100 * time.Millisecond)
	mutRequestedNonces.Lock()
	sort.Slice(requestedNonces, func(i, j int) bool {
		return requestedNonces[i] < requestedNonces[j]
	})
	mutRequestedNonces.Unlock()
	expectedNonces := []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	assert.Equal(t, expectedNonces, requestedNonces)

	requestedNonces = make([]uint64, 0)
	roundHandler.RoundIndex = 5
	_ = sp.RequestHeadersIfMissing(sortedHeaders, core.MetachainShardId)
	time.Sleep(100 * time.Millisecond)
	mutRequestedNonces.Lock()
	sort.Slice(requestedNonces, func(i, j int) bool {
		return requestedNonces[i] < requestedNonces[j]
	})
	mutRequestedNonces.Unlock()
	expectedNonces = []uint64{1, 2, 3}
	assert.Equal(t, expectedNonces, requestedNonces)
}

func TestBlockProcessor_RequestHeadersIfMissingShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("without andromeda activated, should request only headers", func(t *testing.T) {
		t.Parallel()

		var requestedNonces []uint64
		var mutRequestedNonces sync.Mutex

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		coreComponents.EnableEpochsHandlerField = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag != common.AndromedaFlag
			},
		}

		roundHandler := &mock.RoundHandlerMock{}
		coreComponents.RoundField = roundHandler
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		requestProofCalls := 0
		requestHandlerStub := &testscommon.RequestHandlerStub{
			RequestMetaHeaderByNonceCalled: func(nonce uint64) {
				mutRequestedNonces.Lock()
				requestedNonces = append(requestedNonces, nonce)
				mutRequestedNonces.Unlock()
			},
			RequestEquivalentProofByNonceCalled: func(headerShard uint32, headerNonce uint64) {
				mutRequestedNonces.Lock()
				requestProofCalls++
				mutRequestedNonces.Unlock()
			},
		}
		arguments.RequestHandler = requestHandlerStub
		sp, _ := blproc.NewShardProcessor(arguments)

		sortedHeaders := make([]data.HeaderHandler, 0)

		crossNotarizedHeader := &block.MetaBlock{
			Nonce: 5,
			Round: 5,
		}
		arguments.BlockTracker.AddCrossNotarizedHeader(core.MetachainShardId, crossNotarizedHeader, []byte("hash"))

		hdr1 := &block.MetaBlock{
			Nonce: 1,
			Round: 1,
		}
		sortedHeaders = append(sortedHeaders, hdr1)

		hdr2 := &block.MetaBlock{
			Nonce: 8,
			Round: 8,
		}
		sortedHeaders = append(sortedHeaders, hdr2)

		hdr3 := &block.MetaBlock{
			Nonce: 10,
			Round: 10,
		}
		sortedHeaders = append(sortedHeaders, hdr3)

		requestedNonces = make([]uint64, 0)
		roundHandler.RoundIndex = 15
		_ = sp.RequestHeadersIfMissing(sortedHeaders, core.MetachainShardId)
		time.Sleep(100 * time.Millisecond)
		mutRequestedNonces.Lock()
		sort.Slice(requestedNonces, func(i, j int) bool {
			return requestedNonces[i] < requestedNonces[j]
		})
		mutRequestedNonces.Unlock()
		expectedNonces := []uint64{6, 7, 9, 11, 12, 13}
		assert.Equal(t, expectedNonces, requestedNonces)
		assert.Equal(t, 0, requestProofCalls)

		requestedNonces = make([]uint64, 0)
		roundHandler.RoundIndex = process.MaxHeaderRequestsAllowed + 10
		_ = sp.RequestHeadersIfMissing(sortedHeaders, core.MetachainShardId)
		time.Sleep(100 * time.Millisecond)
		mutRequestedNonces.Lock()
		sort.Slice(requestedNonces, func(i, j int) bool {
			return requestedNonces[i] < requestedNonces[j]
		})
		mutRequestedNonces.Unlock()
		expectedNonces = []uint64{6, 7, 9, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}
		assert.Equal(t, expectedNonces, requestedNonces)
		assert.Equal(t, 0, requestProofCalls)
	})

	t.Run("with andromeda activated, should request also proofs if needed", func(t *testing.T) {
		t.Parallel()

		var requestedNonces []uint64
		var mutRequestedNonces sync.Mutex

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		dataPool := initDataPool()
		dataPool.ProofsCalled = func() dataRetriever.ProofsPool {
			return &dataRetrieverMock.ProofsPoolMock{
				GetProofByNonceCalled: func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
					return nil, errors.New("err")
				},
			}
		}
		dataComponents.DataPool = dataPool

		coreComponents.EnableEpochsHandlerField = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}

		roundHandler := &mock.RoundHandlerMock{}
		coreComponents.RoundField = roundHandler
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		requestProofCalls := 0
		requestHandlerStub := &testscommon.RequestHandlerStub{
			RequestMetaHeaderByNonceCalled: func(nonce uint64) {
				mutRequestedNonces.Lock()
				requestedNonces = append(requestedNonces, nonce)
				mutRequestedNonces.Unlock()
			},
			RequestEquivalentProofByNonceCalled: func(headerShard uint32, headerNonce uint64) {
				mutRequestedNonces.Lock()
				requestProofCalls++
				mutRequestedNonces.Unlock()
			},
		}
		arguments.RequestHandler = requestHandlerStub
		sp, _ := blproc.NewShardProcessor(arguments)

		sortedHeaders := make([]data.HeaderHandler, 0)

		crossNotarizedHeader := &block.MetaBlock{
			Nonce: 5,
			Round: 5,
		}
		arguments.BlockTracker.AddCrossNotarizedHeader(core.MetachainShardId, crossNotarizedHeader, []byte("hash"))

		hdr1 := &block.MetaBlock{
			Nonce: 1,
			Round: 1,
		}
		sortedHeaders = append(sortedHeaders, hdr1)

		hdr2 := &block.MetaBlock{
			Nonce: 8,
			Round: 8,
		}
		sortedHeaders = append(sortedHeaders, hdr2)

		hdr3 := &block.MetaBlock{
			Nonce: 10,
			Round: 10,
		}
		sortedHeaders = append(sortedHeaders, hdr3)

		requestedNonces = make([]uint64, 0)
		roundHandler.RoundIndex = 15
		_ = sp.RequestHeadersIfMissing(sortedHeaders, core.MetachainShardId)
		time.Sleep(100 * time.Millisecond)
		mutRequestedNonces.Lock()
		sort.Slice(requestedNonces, func(i, j int) bool {
			return requestedNonces[i] < requestedNonces[j]
		})
		mutRequestedNonces.Unlock()
		expectedNonces := []uint64{6, 7, 9, 11, 12, 13}
		assert.Equal(t, expectedNonces, requestedNonces)
		assert.Equal(t, len(expectedNonces), requestProofCalls)

		requestProofCalls = 0
		requestedNonces = make([]uint64, 0)
		roundHandler.RoundIndex = process.MaxHeaderRequestsAllowed + 10
		_ = sp.RequestHeadersIfMissing(sortedHeaders, core.MetachainShardId)
		time.Sleep(100 * time.Millisecond)
		mutRequestedNonces.Lock()
		sort.Slice(requestedNonces, func(i, j int) bool {
			return requestedNonces[i] < requestedNonces[j]
		})
		mutRequestedNonces.Unlock()
		expectedNonces = []uint64{6, 7, 9, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}
		assert.Equal(t, expectedNonces, requestedNonces)
		assert.Equal(t, len(expectedNonces), requestProofCalls)
	})
}

func TestBlockProcessor_RequestHeadersIfMissingShouldAddHeaderIntoTrackerPool(t *testing.T) {
	t.Parallel()

	var mutRequestedNonces sync.Mutex
	var addedNonces map[uint64]struct{}

	sortedHeaders := make([]data.HeaderHandler, 0)

	crossNotarizedHeader := &block.MetaBlock{
		Nonce: 5,
		Round: 5,
	}

	hdr1 := &block.MetaBlock{
		Nonce: 1,
		Round: 1,
	}
	sortedHeaders = append(sortedHeaders, hdr1)

	hdr2 := &block.MetaBlock{
		Nonce: 8,
		Round: 8,
	}
	sortedHeaders = append(sortedHeaders, hdr2)

	hdr3 := &block.MetaBlock{
		Nonce: 10,
		Round: 10,
	}
	sortedHeaders = append(sortedHeaders, hdr3)

	expectedAddedNonces := []uint64{6, 7, 9}

	addedNonces = make(map[uint64]struct{})

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	roundHandler := &mock.RoundHandlerMock{}
	coreComponents.RoundField = roundHandler

	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	arguments.BlockTracker.AddCrossNotarizedHeader(core.MetachainShardId, crossNotarizedHeader, []byte("hash"))

	wg := &sync.WaitGroup{}
	wg.Add(len(expectedAddedNonces))

	requestHandlerStub := &testscommon.RequestHandlerStub{
		RequestMetaHeaderByNonceCalled: func(nonce uint64) {
			mutRequestedNonces.Lock()
			addedNonces[nonce] = struct{}{}
			mutRequestedNonces.Unlock()

			wg.Done()
		},
	}
	arguments.RequestHandler = requestHandlerStub

	roundHandler.RoundIndex = 12

	sp, _ := blproc.NewShardProcessor(arguments)

	_ = sp.RequestHeadersIfMissing(sortedHeaders, core.MetachainShardId)

	wg.Wait()

	// check if nonces were requested
	// requests are not necessarily in order
	mutRequestedNonces.Lock()
	for _, nonce := range expectedAddedNonces {
		_, ok := addedNonces[nonce]
		assert.True(t, ok)
	}
	mutRequestedNonces.Unlock()
}

func TestAddHeaderIntoTrackerPool_ShouldWork(t *testing.T) {
	t.Parallel()

	var wasCalled bool
	shardID := core.MetachainShardId
	nonce := uint64(1)
	poolsHolderStub := initDataPool()
	poolsHolderStub.HeadersCalled = func() dataRetriever.HeadersPool {
		return &mock.HeadersCacherStub{
			GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
				if hdrNonce == nonce && shardId == shardID {
					wasCalled = true
					return []data.HeaderHandler{&block.MetaBlock{Nonce: 1}}, [][]byte{[]byte("hash")}, nil
				}

				return nil, nil, errors.New("error")
			},
		}
	}

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = poolsHolderStub
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	wasCalled = false
	sp.AddHeaderIntoTrackerPool(nonce+1, shardID)
	assert.False(t, wasCalled)

	wasCalled = false
	sp.AddHeaderIntoTrackerPool(nonce, shardID)
	assert.True(t, wasCalled)
}

func TestBaseProcessor_commitTrieEpochRootHashIfNeededNilStorerShouldErr(t *testing.T) {
	t.Parallel()

	epoch := uint32(37)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	store := dataRetriever.NewChainStorer()
	dataComponents.Storage = store
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	sp, _ := blproc.NewShardProcessor(arguments)

	mb := &block.MetaBlock{Epoch: epoch}
	err := sp.CommitTrieEpochRootHashIfNeeded(mb, []byte("root"))
	require.NotNil(t, err)
	require.True(t, strings.Contains(err.Error(), dataRetriever.ErrStorerNotFound.Error()))
	require.True(t, strings.Contains(err.Error(), dataRetriever.TrieEpochRootHashUnit.String()))
}

func TestBaseProcessor_commitTrieEpochRootHashIfNeededDisabledStorerShouldNotErr(t *testing.T) {
	t.Parallel()

	epoch := uint32(37)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.Storage.AddStorer(dataRetriever.TrieEpochRootHashUnit, &storageunit.NilStorer{})
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	sp, _ := blproc.NewShardProcessor(arguments)

	mb := &block.MetaBlock{Epoch: epoch}
	err := sp.CommitTrieEpochRootHashIfNeeded(mb, []byte("root"))
	require.NoError(t, err)
}

func TestBaseProcessor_commitTrieEpochRootHashIfNeededCannotFindUserAccountStateShouldErr(t *testing.T) {
	t.Parallel()

	epoch := uint32(37)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB = map[state.AccountsDbIdentifier]state.AccountsAdapter{}

	sp, _ := blproc.NewShardProcessor(arguments)

	mb := &block.MetaBlock{Epoch: epoch}
	err := sp.CommitTrieEpochRootHashIfNeeded(mb, []byte("root"))
	require.True(t, errors.Is(err, process.ErrNilAccountsAdapter))
}

func TestBaseProcessor_commitTrieEpochRootHashIfNeededShouldWork(t *testing.T) {
	t.Parallel()

	epoch := uint32(37)
	rootHash := []byte("root-hash")

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	coreComponents.UInt64ByteSliceConv = uint64ByteSlice.NewBigEndianConverter()
	store := dataRetriever.NewChainStorer()
	store.AddStorer(dataRetriever.TrieEpochRootHashUnit,
		&storageStubs.StorerStub{
			PutCalled: func(key, data []byte) error {
				restoredEpoch, err := coreComponents.UInt64ByteSliceConv.ToUint64(key)
				require.NoError(t, err)
				require.Equal(t, epoch, uint32(restoredEpoch))
				return nil
			},
		},
	)
	dataComponents.Storage = store

	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.AccountsDB = map[state.AccountsDbIdentifier]state.AccountsAdapter{
		state.UserAccountsState: &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return rootHash, nil
			},
			GetAllLeavesCalled: func(channels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.TrieLeafParser) error {
				close(channels.LeavesChan)
				channels.ErrChan.Close()
				return nil
			},
		},
	}

	sp, _ := blproc.NewShardProcessor(arguments)

	mb := &block.MetaBlock{Epoch: epoch}
	err := sp.CommitTrieEpochRootHashIfNeeded(mb, []byte("root"))
	require.NoError(t, err)
}

func TestBaseProcessor_commitTrieEpochRootHashIfNeeded_GetAllLeaves(t *testing.T) {
	t.Parallel()

	epoch := uint32(37)
	rootHash := []byte("root-hash")

	t.Run("error on getting the leaves", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.UInt64ByteSliceConv = uint64ByteSlice.NewBigEndianConverter()
		store := dataRetriever.NewChainStorer()
		store.AddStorer(dataRetriever.TrieEpochRootHashUnit,
			&storageStubs.StorerStub{
				PutCalled: func(key, data []byte) error {
					return nil
				},
			},
		)
		dataComponents.Storage = store

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		arguments.AccountsDB = map[state.AccountsDbIdentifier]state.AccountsAdapter{
			state.UserAccountsState: &stateMock.AccountsStub{
				RootHashCalled: func() ([]byte, error) {
					return rootHash, nil
				},
				GetAllLeavesCalled: func(channels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, _ common.TrieLeafParser) error {
					close(channels.LeavesChan)
					channels.ErrChan.Close()
					return expectedErr
				},
			},
		}

		sp, _ := blproc.NewShardProcessor(arguments)

		mb := &block.MetaBlock{Epoch: epoch}
		err := sp.CommitTrieEpochRootHashIfNeeded(mb, []byte("root"))
		require.Equal(t, expectedErr, err)
	})

	t.Run("error on trie iterator chan", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.UInt64ByteSliceConv = uint64ByteSlice.NewBigEndianConverter()
		store := dataRetriever.NewChainStorer()
		store.AddStorer(dataRetriever.TrieEpochRootHashUnit,
			&storageStubs.StorerStub{
				PutCalled: func(key, data []byte) error {
					return nil
				},
			},
		)
		dataComponents.Storage = store

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		arguments.AccountsDB = map[state.AccountsDbIdentifier]state.AccountsAdapter{
			state.UserAccountsState: &stateMock.AccountsStub{
				RootHashCalled: func() ([]byte, error) {
					return rootHash, nil
				},
				GetAllLeavesCalled: func(channels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, trieLeafParser common.TrieLeafParser) error {
					channels.ErrChan.WriteInChanNonBlocking(expectedErr)
					close(channels.LeavesChan)
					return nil
				},
			},
		}

		sp, _ := blproc.NewShardProcessor(arguments)

		mb := &block.MetaBlock{Epoch: epoch}
		err := sp.CommitTrieEpochRootHashIfNeeded(mb, []byte("root"))
		require.Equal(t, expectedErr, err)
	})
}

func TestBaseProcessor_commitTrieEpochRootHashIfNeededShouldUseDataTrieIfNeededWork(t *testing.T) {
	t.Parallel()

	var processDataTrieTests = []struct {
		processDataTrie        bool
		calledWithUserRootHash bool
	}{
		{false, false},
		{true, true},
	}

	for _, tt := range processDataTrieTests {
		epoch := uint32(37)
		rootHash := []byte("userAcc-root-hash")

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.UInt64ByteSliceConv = uint64ByteSlice.NewBigEndianConverter()
		coreComponents.IntMarsh = &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				userAccount := obj.(state.UserAccountHandler)
				userAccount.SetRootHash(rootHash)
				return nil
			},
		}

		store := dataRetriever.NewChainStorer()
		store.AddStorer(dataRetriever.TrieEpochRootHashUnit,
			&storageStubs.StorerStub{
				PutCalled: func(key, data []byte) error {
					restoredEpoch, err := coreComponents.UInt64ByteSliceConv.ToUint64(key)
					require.NoError(t, err)
					require.Equal(t, epoch, uint32(restoredEpoch))
					return nil
				},
			},
		)
		dataComponents.Storage = store
		calledWithUserAccountRootHash := false
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.AccountsDB = map[state.AccountsDbIdentifier]state.AccountsAdapter{
			state.UserAccountsState: &stateMock.AccountsStub{
				GetAllLeavesCalled: func(channels *common.TrieIteratorChannels, ctx context.Context, rh []byte, _ common.TrieLeafParser) error {
					if bytes.Equal(rootHash, rh) {
						calledWithUserAccountRootHash = true
						close(channels.LeavesChan)
						channels.ErrChan.Close()
						return nil
					}

					go func() {
						channels.LeavesChan <- keyValStorage.NewKeyValStorage([]byte("address"), []byte("bytes"))
						close(channels.LeavesChan)
						channels.ErrChan.Close()
					}()

					return nil
				},
			},
		}

		arguments.Config.Debug.EpochStart.ProcessDataTrieOnCommitEpoch = tt.processDataTrie
		sp, _ := blproc.NewShardProcessor(arguments)

		mb := &block.MetaBlock{Epoch: epoch}
		err := sp.CommitTrieEpochRootHashIfNeeded(mb, []byte("root"))
		require.NoError(t, err)

		require.Equal(t, tt.calledWithUserRootHash, calledWithUserAccountRootHash)
	}
}

func TestBaseProcessor_updateState(t *testing.T) {
	t.Parallel()

	var pruneRootHash []byte
	var cancelPruneRootHash []byte

	poolMock := dataRetrieverMock.NewPoolsHolderMock()

	numHeaders := 5
	headers := make([]block.Header, numHeaders)
	for i := 0; i < numHeaders; i++ {
		headers[i] = block.Header{
			Nonce:    uint64(i),
			RootHash: []byte(strconv.Itoa(i)),
		}
	}

	hdrStore := &storageStubs.StorerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			if len(headers) != 0 {
				header := headers[0]
				headers = headers[1:]
				return json.Marshal(header)
			}

			return nil, nil
		},
	}

	storer := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return hdrStore, nil
		},
	}

	shardC := mock.NewMultiShardsCoordinatorMock(3)

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	dataComponents.DataPool = poolMock
	dataComponents.Storage = storer
	bootstrapComponents.Coordinator = shardC
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

	arguments.BlockTracker = &mock.BlockTrackerMock{}
	arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
		IsPruningEnabledCalled: func() bool {
			return true
		},
		PruneTrieCalled: func(rootHashParam []byte, identifier state.TriePruningIdentifier, _ state.PruningHandler) {
			pruneRootHash = rootHashParam
		},
		CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
			cancelPruneRootHash = rootHash
		},
	}
	sp, _ := blproc.NewShardProcessor(arguments)

	prevRootHash := []byte("rootHash")
	for i := range headers {
		sp.UpdateState(
			&headers[i],
			headers[i].RootHash,
			prevRootHash,
			arguments.AccountsDB[state.UserAccountsState],
		)

		assert.Equal(t, prevRootHash, pruneRootHash)
		assert.Equal(t, prevRootHash, cancelPruneRootHash)

		prevRootHash = headers[i].RootHash
	}

	assert.Equal(t, []byte(strconv.Itoa(len(headers)-2)), pruneRootHash)
	assert.Equal(t, []byte(strconv.Itoa(len(headers)-2)), cancelPruneRootHash)
}

func TestBaseProcessor_ProcessScheduledBlockShouldErrWhenProcessorBusy(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments(createComponentHolderMocks())
	processHandler := arguments.CoreComponents.ProcessStatusHandler()
	mockProcessHandler := processHandler.(*testscommon.ProcessStatusHandlerStub)
	mockProcessHandler.TrySetBusyCalled = func(reason string) bool {
		return false
	}
	setIdleCalled := false
	mockProcessHandler.SetIdleCalled = func() {
		setIdleCalled = true
	}

	bp, _ := blproc.NewShardProcessor(arguments)

	err := bp.ProcessScheduledBlock(
		&block.MetaBlock{}, &block.Body{}, haveTime,
	)
	require.Equal(t, process.ErrBlockProcessorBusy, err)
	require.False(t, setIdleCalled, "SetIdle should not be called when TrySetBusy fails")
}

func TestBaseProcessor_ProcessScheduledBlockShouldFail(t *testing.T) {
	t.Parallel()

	t.Run("execute all scheduled txs fail", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		processHandler := arguments.CoreComponents.ProcessStatusHandler()
		mockProcessHandler := processHandler.(*testscommon.ProcessStatusHandlerStub)
		busyIdleCalled := make([]string, 0)
		mockProcessHandler.SetIdleCalled = func() {
			busyIdleCalled = append(busyIdleCalled, idleIdentifier)
		}
		mockProcessHandler.TrySetBusyCalled = func(reason string) bool {
			busyIdleCalled = append(busyIdleCalled, busyIdentifier)
			return true
		}

		scheduledTxsExec := &testscommon.ScheduledTxsExecutionStub{
			ExecuteAllCalled: func(func() time.Duration) error {
				return expectedError
			},
		}

		arguments.ScheduledTxsExecutionHandler = scheduledTxsExec
		bp, _ := blproc.NewShardProcessor(arguments)

		err := bp.ProcessScheduledBlock(
			&block.MetaBlock{}, &block.Body{}, haveTime,
		)

		assert.Equal(t, expectedError, err)
		assert.Equal(t, []string{busyIdentifier, idleIdentifier}, busyIdleCalled)
	})
	t.Run("get root hash fail", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		processHandler := arguments.CoreComponents.ProcessStatusHandler()
		mockProcessHandler := processHandler.(*testscommon.ProcessStatusHandlerStub)
		busyIdleCalled := make([]string, 0)
		mockProcessHandler.SetIdleCalled = func() {
			busyIdleCalled = append(busyIdleCalled, idleIdentifier)
		}
		mockProcessHandler.TrySetBusyCalled = func(reason string) bool {
			busyIdleCalled = append(busyIdleCalled, busyIdentifier)
			return true
		}

		accounts := &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return nil, expectedError
			},
		}
		arguments.AccountsDB[state.UserAccountsState] = accounts

		bp, _ := blproc.NewShardProcessor(arguments)

		err := bp.ProcessScheduledBlock(
			&block.MetaBlock{}, &block.Body{}, haveTime,
		)

		assert.Equal(t, expectedError, err)
		assert.Equal(t, []string{busyIdentifier, idleIdentifier}, busyIdleCalled)
	})
}

func TestBaseProcessor_ProcessScheduledBlockShouldWork(t *testing.T) {
	t.Parallel()
	rootHash := []byte("root hash to be tested")
	accounts := &stateMock.AccountsStub{
		RootHashCalled: func() ([]byte, error) {
			return rootHash, nil
		},
	}

	initialGasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(11),
		DeveloperFees:   big.NewInt(12),
		GasProvided:     13,
		GasPenalized:    14,
		GasRefunded:     15,
	}

	finalGasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(101),
		DeveloperFees:   big.NewInt(103),
		GasProvided:     105,
		GasPenalized:    107,
		GasRefunded:     109,
	}

	feeHandler := createFeeHandlerMockForProcessScheduledBlock(initialGasAndFees, finalGasAndFees)
	gasHandler := createGasHandlerMockForProcessScheduledBlock(initialGasAndFees, finalGasAndFees)

	expectedGasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(90),
		DeveloperFees:   big.NewInt(91),
		GasProvided:     92,
		GasPenalized:    93,
		GasRefunded:     94,
	}

	wasCalledSetScheduledRootHash := false
	wasCalledSetScheduledGasAndFees := false
	scheduledTxsExec := &testscommon.ScheduledTxsExecutionStub{
		ExecuteAllCalled: func(func() time.Duration) error {
			return nil
		},
		SetScheduledRootHashCalled: func(hash []byte) {
			wasCalledSetScheduledRootHash = true
			require.Equal(t, rootHash, hash)
		},
		SetScheduledGasAndFeesCalled: func(gasAndFees scheduled.GasAndFees) {
			wasCalledSetScheduledGasAndFees = true
			require.Equal(t, expectedGasAndFees, gasAndFees)
		},
	}

	arguments := CreateMockArguments(createComponentHolderMocks())
	processHandler := arguments.CoreComponents.ProcessStatusHandler()
	mockProcessHandler := processHandler.(*testscommon.ProcessStatusHandlerStub)
	busyIdleCalled := make([]string, 0)
	mockProcessHandler.SetIdleCalled = func() {
		busyIdleCalled = append(busyIdleCalled, idleIdentifier)
	}
	mockProcessHandler.TrySetBusyCalled = func(reason string) bool {
		busyIdleCalled = append(busyIdleCalled, busyIdentifier)
		return true
	}

	arguments.AccountsDB[state.UserAccountsState] = accounts
	arguments.ScheduledTxsExecutionHandler = scheduledTxsExec
	arguments.FeeHandler = feeHandler
	arguments.GasHandler = gasHandler
	bp, _ := blproc.NewShardProcessor(arguments)

	err := bp.ProcessScheduledBlock(
		&block.MetaBlock{}, &block.Body{}, haveTime,
	)
	require.Nil(t, err)

	assert.True(t, wasCalledSetScheduledGasAndFees)
	assert.True(t, wasCalledSetScheduledRootHash)
	assert.Equal(t, []string{busyIdentifier, idleIdentifier}, busyIdleCalled) // the order is important
}

func TestBaseProcessor_CheckScheduledData(t *testing.T) {
	t.Parallel()

	scheduledGasAndFees := scheduled.GasAndFees{
		AccumulatedFees: big.NewInt(11),
		DeveloperFees:   big.NewInt(12),
		GasProvided:     13,
		GasPenalized:    14,
		GasRefunded:     15,
	}

	createProcessorAndHeader := func(t *testing.T) (interface {
		CheckScheduledData(data.HeaderHandler) error
	}, *block.HeaderV2) {
		t.Helper()
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ArgBaseProcessor.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return []byte("scheduled-root"), nil
			},
		}
		arguments.ArgBaseProcessor.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			GetScheduledGasAndFeesCalled: func() scheduled.GasAndFees {
				return scheduledGasAndFees
			},
		}
		processor, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)
		header := &block.HeaderV2{
			Header:                   &block.Header{},
			ScheduledRootHash:        []byte("scheduled-root"),
			ScheduledAccumulatedFees: big.NewInt(11),
			ScheduledDeveloperFees:   big.NewInt(12),
			ScheduledGasProvided:     13,
			ScheduledGasPenalized:    14,
			ScheduledGasRefunded:     15,
		}
		return processor, header
	}

	t.Run("should work when scheduled data matches", func(t *testing.T) {
		t.Parallel()

		processor, header := createProcessorAndHeader(t)
		err := processor.CheckScheduledData(header)

		require.NoError(t, err)
	})

	t.Run("should fail when scheduled accumulated fees mismatch", func(t *testing.T) {
		t.Parallel()

		processor, header := createProcessorAndHeader(t)
		header.ScheduledAccumulatedFees = big.NewInt(111)

		err := processor.CheckScheduledData(header)

		require.ErrorIs(t, err, process.ErrScheduledGasAndFeesDoesNotMatch)
	})

	t.Run("should fail when scheduled developer fees mismatch", func(t *testing.T) {
		t.Parallel()

		processor, header := createProcessorAndHeader(t)
		header.ScheduledDeveloperFees = big.NewInt(112)

		err := processor.CheckScheduledData(header)

		require.ErrorIs(t, err, process.ErrScheduledGasAndFeesDoesNotMatch)
	})

	t.Run("should fail when scheduled gas provided mismatch", func(t *testing.T) {
		t.Parallel()

		processor, header := createProcessorAndHeader(t)
		header.ScheduledGasProvided++

		err := processor.CheckScheduledData(header)

		require.ErrorIs(t, err, process.ErrScheduledGasAndFeesDoesNotMatch)
	})

	t.Run("should fail when scheduled gas penalized mismatch", func(t *testing.T) {
		t.Parallel()

		processor, header := createProcessorAndHeader(t)
		header.ScheduledGasPenalized++

		err := processor.CheckScheduledData(header)

		require.ErrorIs(t, err, process.ErrScheduledGasAndFeesDoesNotMatch)
	})

	t.Run("should fail when scheduled gas refunded mismatch", func(t *testing.T) {
		t.Parallel()

		processor, header := createProcessorAndHeader(t)
		header.ScheduledGasRefunded++

		err := processor.CheckScheduledData(header)

		require.ErrorIs(t, err, process.ErrScheduledGasAndFeesDoesNotMatch)
	})
}

// get initial fees on first getGasAndFees call and final fees on second call
func createFeeHandlerMockForProcessScheduledBlock(initial, final scheduled.GasAndFees) process.TransactionFeeHandler {
	runCount := 0
	return &mock.FeeAccumulatorStub{
		GetAccumulatedFeesCalled: func() *big.Int {
			if runCount%4 >= 2 {
				return final.AccumulatedFees
			}
			runCount++
			return initial.AccumulatedFees
		},
		GetDeveloperFeesCalled: func() *big.Int {
			if runCount%4 >= 2 {
				return final.DeveloperFees
			}
			runCount++
			return initial.DeveloperFees
		},
	}
}

// get initial gas consumed on first getGasAndFees call and final gas consumed on second call
func createGasHandlerMockForProcessScheduledBlock(initial, final scheduled.GasAndFees) process.GasHandler {
	runCount := 0
	return &mock.GasHandlerMock{
		TotalGasProvidedCalled: func() uint64 {
			return initial.GasProvided
		},
		TotalGasPenalizedCalled: func() uint64 {
			if runCount%4 >= 2 {
				return final.GasPenalized
			}
			runCount++
			return initial.GasPenalized
		},
		TotalGasRefundedCalled: func() uint64 {
			if runCount%4 >= 2 {
				return final.GasRefunded
			}
			runCount++
			return initial.GasRefunded
		},
		TotalGasProvidedWithScheduledCalled: func() uint64 {
			return final.GasProvided
		},
	}
}

func TestBaseProcessor_gasAndFeesDelta(t *testing.T) {
	zeroGasAndFees := process.GetZeroGasAndFees()

	t.Run("final accumulatedFees lower then initial accumulatedFees", func(t *testing.T) {
		t.Parallel()

		initialGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(100),
		}

		finalGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(10),
		}

		gasAndFees := blproc.GasAndFeesDelta(initialGasAndFees, finalGasAndFees)
		assert.Equal(t, zeroGasAndFees, gasAndFees)
	})
	t.Run("final devFees lower then initial devFees", func(t *testing.T) {
		t.Parallel()

		initialGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(10),
			DeveloperFees:   big.NewInt(100),
		}

		finalGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(100),
			DeveloperFees:   big.NewInt(10),
		}

		gasAndFees := blproc.GasAndFeesDelta(initialGasAndFees, finalGasAndFees)
		assert.Equal(t, zeroGasAndFees, gasAndFees)
	})
	t.Run("final gasProvided lower then initial gasProvided", func(t *testing.T) {
		t.Parallel()

		initialGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(11),
			DeveloperFees:   big.NewInt(12),
			GasProvided:     100,
		}

		finalGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(101),
			DeveloperFees:   big.NewInt(102),
			GasProvided:     10,
		}

		gasAndFees := blproc.GasAndFeesDelta(initialGasAndFees, finalGasAndFees)
		assert.Equal(t, zeroGasAndFees, gasAndFees)
	})
	t.Run("final gasPenalized lower then initial gasPenalized", func(t *testing.T) {
		t.Parallel()

		initialGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(11),
			DeveloperFees:   big.NewInt(12),
			GasProvided:     13,
			GasPenalized:    100,
		}

		finalGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(101),
			DeveloperFees:   big.NewInt(102),
			GasProvided:     103,
			GasPenalized:    10,
		}

		gasAndFees := blproc.GasAndFeesDelta(initialGasAndFees, finalGasAndFees)
		assert.Equal(t, zeroGasAndFees, gasAndFees)
	})
	t.Run("final gasRefunded lower then initial gasRefunded", func(t *testing.T) {
		t.Parallel()

		initialGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(11),
			DeveloperFees:   big.NewInt(12),
			GasProvided:     13,
			GasPenalized:    14,
			GasRefunded:     100,
		}

		finalGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(101),
			DeveloperFees:   big.NewInt(102),
			GasProvided:     103,
			GasPenalized:    104,
			GasRefunded:     10,
		}

		gasAndFees := blproc.GasAndFeesDelta(initialGasAndFees, finalGasAndFees)
		assert.Equal(t, zeroGasAndFees, gasAndFees)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		initialGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(11),
			DeveloperFees:   big.NewInt(12),
			GasProvided:     13,
			GasPenalized:    14,
			GasRefunded:     15,
		}

		finalGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(101),
			DeveloperFees:   big.NewInt(103),
			GasProvided:     105,
			GasPenalized:    107,
			GasRefunded:     109,
		}

		expectedGasAndFees := scheduled.GasAndFees{
			AccumulatedFees: big.NewInt(0).Sub(finalGasAndFees.AccumulatedFees, initialGasAndFees.AccumulatedFees),
			DeveloperFees:   big.NewInt(0).Sub(finalGasAndFees.DeveloperFees, initialGasAndFees.DeveloperFees),
			GasProvided:     finalGasAndFees.GasProvided - initialGasAndFees.GasProvided,
			GasPenalized:    finalGasAndFees.GasPenalized - initialGasAndFees.GasPenalized,
			GasRefunded:     finalGasAndFees.GasRefunded - initialGasAndFees.GasRefunded,
		}

		gasAndFees := blproc.GasAndFeesDelta(initialGasAndFees, finalGasAndFees)

		assert.Equal(t, expectedGasAndFees, gasAndFees)
	})

}

func TestBaseProcessor_getIndexOfFirstMiniBlockToBeExecuted(t *testing.T) {
	t.Parallel()

	t.Run("scheduledMiniBlocks flag not set", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		index, err := bp.GetIndexOfFirstMiniBlockToBeExecuted(&block.MetaBlock{})
		assert.Nil(t, err)
		assert.Equal(t, 0, index)
	})

	t.Run("scheduledMiniBlocks flag is set, empty block", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		index, err := bp.GetIndexOfFirstMiniBlockToBeExecuted(&block.MetaBlock{})
		assert.Nil(t, err)
		assert.Equal(t, 0, index)
	})

	t.Run("get first index for the miniBlockHeader which is not processed executionType", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			IsMiniBlockExecutedCalled: func(_ []byte) bool {
				return true
			},
		}
		bp, _ := blproc.NewShardProcessor(arguments)

		mbh1 := block.MiniBlockHeader{}
		mbhReserved1 := block.MiniBlockHeaderReserved{ExecutionType: block.Processed}
		mbh1.Reserved, _ = mbhReserved1.Marshal()

		mbh2 := block.MiniBlockHeader{}
		mbhReserved2 := block.MiniBlockHeaderReserved{ExecutionType: block.Normal}
		mbh2.Reserved, _ = mbhReserved2.Marshal()

		metaBlock := &block.MetaBlock{
			MiniBlockHeaders: []block.MiniBlockHeader{
				mbh1,
				mbh2,
			},
		}

		index, err := bp.GetIndexOfFirstMiniBlockToBeExecuted(metaBlock)
		assert.Nil(t, err)
		assert.Equal(t, 1, index)
	})

	t.Run("leading processed miniBlock not executed locally is rejected", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			IsMiniBlockExecutedCalled: func(_ []byte) bool {
				return false
			},
		}
		bp, _ := blproc.NewShardProcessor(arguments)

		mbh := block.MiniBlockHeader{}
		mbhReserved := block.MiniBlockHeaderReserved{ExecutionType: block.Processed}
		mbh.Reserved, _ = mbhReserved.Marshal()

		metaBlock := &block.MetaBlock{MiniBlockHeaders: []block.MiniBlockHeader{mbh}}

		index, err := bp.GetIndexOfFirstMiniBlockToBeExecuted(metaBlock)
		assert.Zero(t, index)
		assert.ErrorIs(t, err, process.ErrMiniBlockNotExecuted)
	})

	t.Run("processed miniBlock after a non-processed one is rejected", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			IsMiniBlockExecutedCalled: func(_ []byte) bool {
				return true
			},
		}
		bp, _ := blproc.NewShardProcessor(arguments)

		mbhNormal := block.MiniBlockHeader{}
		mbhNormalReserved := block.MiniBlockHeaderReserved{ExecutionType: block.Normal}
		mbhNormal.Reserved, _ = mbhNormalReserved.Marshal()

		mbhProcessed := block.MiniBlockHeader{}
		mbhProcessedReserved := block.MiniBlockHeaderReserved{ExecutionType: block.Processed}
		mbhProcessed.Reserved, _ = mbhProcessedReserved.Marshal()

		metaBlock := &block.MetaBlock{
			MiniBlockHeaders: []block.MiniBlockHeader{
				mbhNormal,
				mbhProcessed,
			},
		}

		index, err := bp.GetIndexOfFirstMiniBlockToBeExecuted(metaBlock)
		assert.Zero(t, index)
		assert.ErrorIs(t, err, process.ErrProcessedMiniBlockNotInLeadingPrefix)
	})
}

func TestBaseProcessor_getFinalMiniBlocks(t *testing.T) {
	t.Parallel()

	t.Run("scheduledMiniBlocks flag not set", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		body, _, err := bp.GetFinalMiniBlocks([]byte("hash"), &block.MetaBlock{}, &block.Body{})
		assert.Nil(t, err)
		assert.Equal(t, &block.Body{}, body)
	})

	t.Run("scheduledMiniBlocks flag is set, empty body", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		body, _, err := bp.GetFinalMiniBlocks([]byte("hash"), &block.MetaBlock{}, &block.Body{})
		assert.Nil(t, err)
		assert.Equal(t, &block.Body{}, body)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		mb1 := &block.MiniBlock{
			TxHashes: [][]byte{[]byte("txHash1")},
		}
		mb2 := &block.MiniBlock{
			TxHashes: [][]byte{[]byte("txHash2")},
		}
		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				mb1,
				mb2,
			},
		}

		mbh1 := block.MiniBlockHeader{}
		mbhReserved1 := block.MiniBlockHeaderReserved{State: block.Proposed}
		mbh1.Reserved, _ = mbhReserved1.Marshal()

		mbh2 := block.MiniBlockHeader{}
		mbhReserved2 := block.MiniBlockHeaderReserved{State: block.Final}
		mbh2.Reserved, _ = mbhReserved2.Marshal()

		metaBlock := &block.MetaBlock{
			MiniBlockHeaders: []block.MiniBlockHeader{
				mbh1,
				mbh2,
			},
		}

		expectedBody := &block.Body{MiniBlocks: block.MiniBlockSlice{mb2}}

		retBody, _, err := bp.GetFinalMiniBlocks([]byte("hash"), metaBlock, body)
		assert.Nil(t, err)
		assert.Equal(t, expectedBody, retBody)
	})
}

func TestBaseProcessor_getScheduledMiniBlocksFromMe(t *testing.T) {
	t.Parallel()

	t.Run("wrong body type", func(t *testing.T) {
		t.Parallel()

		retBody, err := blproc.GetScheduledMiniBlocksFromMe(&block.Header{}, &wrongBody{})
		assert.Equal(t, process.ErrWrongTypeAssertion, err)
		assert.Nil(t, retBody)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		mb1 := &block.MiniBlock{
			TxHashes: [][]byte{[]byte("txHash1")},
		}
		mb2 := &block.MiniBlock{
			TxHashes: [][]byte{[]byte("txHash2")},
		}
		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				mb1,
				mb2,
			},
		}

		mbh1 := block.MiniBlockHeader{
			SenderShardID: 1,
		}
		mbhReserved1 := block.MiniBlockHeaderReserved{ExecutionType: block.Normal}
		mbh1.Reserved, _ = mbhReserved1.Marshal()

		mbh2 := block.MiniBlockHeader{
			SenderShardID: 1,
		}
		mbhReserved2 := block.MiniBlockHeaderReserved{ExecutionType: block.Scheduled}
		mbh2.Reserved, _ = mbhReserved2.Marshal()

		header := &block.Header{
			ShardID: 1,
			MiniBlockHeaders: []block.MiniBlockHeader{
				mbh1,
				mbh2,
			},
		}

		retBody, err := blproc.GetScheduledMiniBlocksFromMe(header, body)
		assert.Nil(t, err)
		assert.Equal(t, block.MiniBlockSlice{mb2}, retBody)
	})
}

func TestBaseProcessor_checkScheduledMiniBlockValidity(t *testing.T) {
	t.Parallel()

	hash1 := []byte("Hash1")

	t.Run("scheduledMiniBlocks flag not set", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		err := bp.CheckScheduledMiniBlocksValidity(&block.MetaBlock{})
		assert.Nil(t, err)
	})

	t.Run("fail to calculate hash", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		coreComponents.IntMarsh = &marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			GetScheduledMiniBlocksCalled: func() block.MiniBlockSlice {
				return block.MiniBlockSlice{&block.MiniBlock{
					TxHashes: [][]byte{hash1},
				}}
			},
		}

		bp, _ := blproc.NewShardProcessor(arguments)

		header := &block.Header{
			MiniBlockHeaders: []block.MiniBlockHeader{
				{Hash: []byte("differentHash")},
			},
		}

		err := bp.CheckScheduledMiniBlocksValidity(header)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("scheduled miniblocks mismatch", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		coreComponents.Hash = &mock.HasherStub{
			ComputeCalled: func(s string) []byte {
				return hash1
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			GetScheduledMiniBlocksCalled: func() block.MiniBlockSlice {
				return block.MiniBlockSlice{&block.MiniBlock{
					TxHashes: [][]byte{hash1},
				}}
			},
		}

		bp, _ := blproc.NewShardProcessor(arguments)

		header := &block.Header{
			MiniBlockHeaders: []block.MiniBlockHeader{
				{Hash: []byte("differentHash")},
			},
		}

		err := bp.CheckScheduledMiniBlocksValidity(header)
		assert.Equal(t, process.ErrScheduledMiniBlocksMismatch, err)
	})

	t.Run("num header miniblocks lower than scheduled miniblocks, should fail", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			GetScheduledMiniBlocksCalled: func() block.MiniBlockSlice {
				return block.MiniBlockSlice{
					&block.MiniBlock{
						TxHashes: [][]byte{hash1},
					},
					&block.MiniBlock{
						TxHashes: [][]byte{[]byte("hash2")},
					},
				}
			},
		}

		bp, _ := blproc.NewShardProcessor(arguments)

		header := &block.Header{
			MiniBlockHeaders: []block.MiniBlockHeader{
				{Hash: hash1},
			},
		}

		err := bp.CheckScheduledMiniBlocksValidity(header)
		assert.Equal(t, process.ErrScheduledMiniBlocksMismatch, err)
	})

	t.Run("same hash, should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.Hash = &mock.HasherStub{
			ComputeCalled: func(s string) []byte {
				return hash1
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			GetScheduledMiniBlocksCalled: func() block.MiniBlockSlice {
				return block.MiniBlockSlice{
					&block.MiniBlock{
						TxHashes: [][]byte{hash1},
					},
				}
			},
		}

		bp, _ := blproc.NewShardProcessor(arguments)

		header := &block.Header{
			MiniBlockHeaders: []block.MiniBlockHeader{
				{Hash: hash1},
			},
		}

		err := bp.CheckScheduledMiniBlocksValidity(header)
		assert.Nil(t, err)
	})
}

func TestBaseProcessor_setMiniBlockHeaderReservedField(t *testing.T) {
	t.Parallel()

	miniBlockHash := []byte("miniBlockHash")

	t.Run("scheduledMiniBlocks flag not set", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		err := bp.SetMiniBlockHeaderReservedField(&block.MiniBlock{}, &block.MiniBlockHeader{Hash: []byte{}}, make(map[string]*processedMb.ProcessedMiniBlockInfo))
		assert.Nil(t, err)
	})

	t.Run("no scheduled miniBlock, miniBlock Not executed", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			IsScheduledTxCalled: func(hash []byte) bool {
				return false
			},
			IsMiniBlockExecutedCalled: func(hash []byte) bool {
				assert.Equal(t, miniBlockHash, hash)
				return false
			},
		}
		bp, _ := blproc.NewShardProcessor(arguments)

		mbHandler := &block.MiniBlockHeader{
			Hash: miniBlockHash,
		}

		err := bp.SetMiniBlockHeaderReservedField(&block.MiniBlock{}, mbHandler, make(map[string]*processedMb.ProcessedMiniBlockInfo))
		assert.Nil(t, err)
		assert.Equal(t, int32(block.Normal), mbHandler.GetProcessingType())
		assert.Equal(t, int32(block.Final), mbHandler.GetConstructionState())
	})

	t.Run("no scheduled miniBlock, miniBlock executed", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			IsScheduledTxCalled: func(hash []byte) bool {
				return false
			},
			IsMiniBlockExecutedCalled: func(hash []byte) bool {
				assert.Equal(t, miniBlockHash, hash)
				return true
			},
		}
		bp, _ := blproc.NewShardProcessor(arguments)

		mbHandler := &block.MiniBlockHeader{
			Hash: miniBlockHash,
		}

		err := bp.SetMiniBlockHeaderReservedField(&block.MiniBlock{}, mbHandler, make(map[string]*processedMb.ProcessedMiniBlockInfo))
		assert.Nil(t, err)
		assert.Equal(t, int32(block.Processed), mbHandler.GetProcessingType())
		assert.Equal(t, int32(block.Final), mbHandler.GetConstructionState())
	})

	t.Run("is scheduled miniBlock, different shardId", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		bootstrapComponents.Coordinator = &testscommon.ShardsCoordinatorMock{
			SelfIDCalled: func() uint32 {
				return 1
			},
		}

		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			IsScheduledTxCalled: func(hash []byte) bool {
				return true
			},
		}
		bp, _ := blproc.NewShardProcessor(arguments)

		mb := &block.MiniBlock{
			TxHashes: [][]byte{[]byte("hash")},
		}

		mbHandler := &block.MiniBlockHeader{
			Hash:          miniBlockHash,
			SenderShardID: 2,
		}

		err := bp.SetMiniBlockHeaderReservedField(mb, mbHandler, make(map[string]*processedMb.ProcessedMiniBlockInfo))
		assert.Nil(t, err)
		assert.Equal(t, int32(block.Scheduled), mbHandler.GetProcessingType())
		assert.Equal(t, int32(block.Final), mbHandler.GetConstructionState())
	})

	t.Run("is scheduled miniBlock, same shardId", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.ScheduledMiniBlocksFlag)
		shardId := uint32(1)
		bootstrapComponents.Coordinator = &testscommon.ShardsCoordinatorMock{
			SelfIDCalled: func() uint32 {
				return shardId
			},
		}

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ScheduledTxsExecutionHandler = &testscommon.ScheduledTxsExecutionStub{
			IsScheduledTxCalled: func(hash []byte) bool {
				return true
			},
		}
		bp, _ := blproc.NewShardProcessor(arguments)

		mb := &block.MiniBlock{
			TxHashes: [][]byte{[]byte("hash")},
		}

		mbHandler := &block.MiniBlockHeader{
			Hash:          miniBlockHash,
			SenderShardID: shardId,
		}

		err := bp.SetMiniBlockHeaderReservedField(mb, mbHandler, make(map[string]*processedMb.ProcessedMiniBlockInfo))
		assert.Nil(t, err)
		assert.Equal(t, int32(block.Scheduled), mbHandler.GetProcessingType())
		assert.Equal(t, int32(block.Proposed), mbHandler.GetConstructionState())
	})
}

func TestMetaProcessor_RestoreBlockBodyIntoPoolsShouldErrNilBlockBody(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
	dataComponents.Storage = initStore()
	arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	mp, _ := blproc.NewMetaProcessor(arguments)

	err := mp.RestoreBlockBodyIntoPools(nil)
	assert.Equal(t, err, process.ErrNilBlockBody)
}

func TestMetaProcessor_RestoreBlockBodyIntoPoolsShouldErrWhenRestoreBlockDataFromStorageFails(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("error")

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
	dataComponents.Storage = initStore()
	arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
		RestoreBlockDataFromStorageCalled: func(body *block.Body) (int, error) {
			return 0, expectedError
		},
	}
	mp, _ := blproc.NewMetaProcessor(arguments)

	err := mp.RestoreBlockBodyIntoPools(&block.Body{})
	assert.Equal(t, err, expectedError)
}

func TestMetaProcessor_RestoreBlockBodyIntoPoolsShouldWork(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createMockComponentHolders()
	dataComponents.Storage = initStore()
	arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
		RestoreBlockDataFromStorageCalled: func(body *block.Body) (int, error) {
			return 1, nil
		},
	}
	mp, _ := blproc.NewMetaProcessor(arguments)

	err := mp.RestoreBlockBodyIntoPools(&block.Body{})
	assert.Nil(t, err)
}

func TestBaseProcessor_getPruningHandler(t *testing.T) {
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.Config = config.Config{}
	arguments.StatusCoreComponents = &factory.StatusCoreComponentsStub{
		AppStatusHandlerField: &statusHandlerMock.AppStatusHandlerStub{},
	}
	bp, errCtor := blproc.NewShardProcessor(arguments)
	require.Nil(t, errCtor)

	bp.SetLastRestartNonce(1)
	ph := bp.GetPruningHandler(10)
	assert.False(t, ph.IsPruningEnabled())

	bp.SetLastRestartNonce(1)
	ph = bp.GetPruningHandler(11)
	assert.False(t, ph.IsPruningEnabled())

	bp.SetLastRestartNonce(1)
	ph = bp.GetPruningHandler(14)
	assert.True(t, ph.IsPruningEnabled())

	bp.SetLastRestartNonce(15)
	ph = bp.GetPruningHandler(14)
	assert.False(t, ph.IsPruningEnabled())

	bp.SetClosingNodeStarted(true)
	ph = bp.GetPruningHandler(14)
	assert.False(t, ph.IsPruningEnabled())
}

func TestBaseProcessor_getPruningHandlerSetsDefaulPruningDelay(t *testing.T) {
	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	arguments.Config = config.Config{}
	bp, _ := blproc.NewShardProcessor(arguments)

	bp.SetLastRestartNonce(0)
	ph := bp.GetPruningHandler(9)
	assert.False(t, ph.IsPruningEnabled())
}

func TestCheckConstructionStateProcessingTypeAndIndexesCorrectness(t *testing.T) {
	t.Parallel()

	const blockShard = uint32(1)
	const otherShard = uint32(2)

	makeMb := func(sender, receiver uint32, mbType block.Type, bodyScheduled bool, txCount int) *block.MiniBlock {
		mb := &block.MiniBlock{
			SenderShardID:   sender,
			ReceiverShardID: receiver,
			Type:            mbType,
			TxHashes:        make([][]byte, txCount),
		}
		for i := range mb.TxHashes {
			mb.TxHashes[i] = []byte{byte(i)}
		}
		if bodyScheduled {
			reserved, _ := (&block.MiniBlockReserved{ExecutionType: block.Scheduled}).Marshal()
			mb.Reserved = reserved
		}
		return mb
	}

	makeMbh := func(mb *block.MiniBlock, hdrPT block.ProcessingType, state block.MiniBlockState, lastIdx int32) *block.MiniBlockHeader {
		mbh := &block.MiniBlockHeader{
			SenderShardID:   mb.SenderShardID,
			ReceiverShardID: mb.ReceiverShardID,
			Type:            mb.Type,
			TxCount:         uint32(len(mb.TxHashes)),
		}
		_ = mbh.SetProcessingType(int32(hdrPT))
		_ = mbh.SetConstructionState(int32(state))
		_ = mbh.SetIndexOfLastTxProcessed(lastIdx)
		return mbh
	}

	t.Run("legal cells pass", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			name     string
			sender   uint32
			receiver uint32
			body     bool
			hdrPT    block.ProcessingType
			state    block.MiniBlockState
			txCount  int
			lastIdx  int32
		}{
			{"normal intra", blockShard, blockShard, false, block.Normal, block.Final, 3, 2},
			{"normal outgoing", blockShard, otherShard, false, block.Normal, block.Final, 3, 2},
			{"normal incoming", otherShard, blockShard, false, block.Normal, block.Final, 3, 2},
			{"normal incoming with scheduled body", otherShard, blockShard, true, block.Normal, block.Final, 3, 2},
			{"scheduled intra", blockShard, blockShard, true, block.Scheduled, block.Proposed, 3, 2},
			{"scheduled outgoing", blockShard, otherShard, true, block.Scheduled, block.Proposed, 3, 2},
			{"scheduled incoming with scheduled body", otherShard, blockShard, true, block.Scheduled, block.Final, 3, 2},
			{"scheduled incoming with normal body", otherShard, blockShard, false, block.Scheduled, block.Final, 3, 2},
			{"processed intra", blockShard, blockShard, true, block.Processed, block.Final, 3, 2},
			{"processed outgoing", blockShard, otherShard, true, block.Processed, block.Final, 3, 2},
			{"broadcast peer mb", blockShard, core.AllShardId, false, block.Normal, block.Final, 1, 0},
		}
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				mb := makeMb(tc.sender, tc.receiver, block.TxBlock, tc.body, tc.txCount)
				mbh := makeMbh(mb, tc.hdrPT, tc.state, tc.lastIdx)
				err := blproc.CheckConstructionStateProcessingTypeAndIndexesCorrectness(mbh, mb, blockShard)
				assert.NoError(t, err)
			})
		}
	})

	t.Run("scheduled plus partial executed allowed at sender", func(t *testing.T) {
		t.Parallel()
		mb := makeMb(blockShard, blockShard, block.TxBlock, true, 3)
		mbh := makeMbh(mb, block.Scheduled, block.PartialExecuted, 1)
		err := blproc.CheckConstructionStateProcessingTypeAndIndexesCorrectness(mbh, mb, blockShard)
		assert.NoError(t, err)
	})

	t.Run("scheduled plus partial executed allowed at incoming", func(t *testing.T) {
		t.Parallel()
		mb := makeMb(otherShard, blockShard, block.TxBlock, true, 3)
		mbh := makeMbh(mb, block.Scheduled, block.PartialExecuted, 1)
		err := blproc.CheckConstructionStateProcessingTypeAndIndexesCorrectness(mbh, mb, blockShard)
		assert.NoError(t, err)
	})

	t.Run("scheduled body required when header is scheduled", func(t *testing.T) {
		t.Parallel()
		mb := makeMb(blockShard, blockShard, block.TxBlock, false, 3)
		mbh := makeMbh(mb, block.Scheduled, block.Proposed, 2)
		err := blproc.CheckConstructionStateProcessingTypeAndIndexesCorrectness(mbh, mb, blockShard)
		assert.ErrorIs(t, err, process.ErrProcessingTypeBodyHeaderMismatch)
	})

	t.Run("sender shard normal header with scheduled body rejected", func(t *testing.T) {
		t.Parallel()
		mb := makeMb(blockShard, otherShard, block.TxBlock, true, 3)
		mbh := makeMbh(mb, block.Normal, block.Final, 2)
		err := blproc.CheckConstructionStateProcessingTypeAndIndexesCorrectness(mbh, mb, blockShard)
		assert.ErrorIs(t, err, process.ErrProcessingTypeBodyHeaderMismatch)
	})

	t.Run("processed must have sender equal block shard", func(t *testing.T) {
		t.Parallel()
		mb := makeMb(otherShard, blockShard, block.TxBlock, true, 3)
		mbh := makeMbh(mb, block.Processed, block.Final, 2)
		err := blproc.CheckConstructionStateProcessingTypeAndIndexesCorrectness(mbh, mb, blockShard)
		assert.ErrorIs(t, err, process.ErrInvalidMiniBlockShardRole)
	})

	t.Run("processed requires scheduled body", func(t *testing.T) {
		t.Parallel()
		mb := makeMb(blockShard, blockShard, block.TxBlock, false, 3)
		mbh := makeMbh(mb, block.Processed, block.Final, 2)
		err := blproc.CheckConstructionStateProcessingTypeAndIndexesCorrectness(mbh, mb, blockShard)
		assert.ErrorIs(t, err, process.ErrProcessingTypeBodyHeaderMismatch)
	})

	t.Run("incoming normal partial executed allowed", func(t *testing.T) {
		t.Parallel()
		mb := makeMb(otherShard, blockShard, block.TxBlock, false, 3)
		mbh := makeMbh(mb, block.Normal, block.PartialExecuted, 1)
		err := blproc.CheckConstructionStateProcessingTypeAndIndexesCorrectness(mbh, mb, blockShard)
		assert.NoError(t, err)
	})

	t.Run("sender shard normal partial executed rejected", func(t *testing.T) {
		t.Parallel()
		mb := makeMb(blockShard, otherShard, block.TxBlock, false, 3)
		mbh := makeMbh(mb, block.Normal, block.PartialExecuted, 1)
		err := blproc.CheckConstructionStateProcessingTypeAndIndexesCorrectness(mbh, mb, blockShard)
		assert.ErrorIs(t, err, process.ErrInvalidConstructionState)
	})

	t.Run("outgoing normal proposed with final index rejected", func(t *testing.T) {
		t.Parallel()
		mb := makeMb(blockShard, otherShard, block.TxBlock, false, 3)
		mbh := makeMbh(mb, block.Normal, block.Proposed, 2)
		err := blproc.CheckConstructionStateProcessingTypeAndIndexesCorrectness(mbh, mb, blockShard)
		assert.ErrorIs(t, err, process.ErrInvalidConstructionState)
	})

	t.Run("non TxBlock cannot be scheduled", func(t *testing.T) {
		t.Parallel()
		mb := makeMb(blockShard, blockShard, block.SmartContractResultBlock, true, 2)
		mbh := makeMbh(mb, block.Scheduled, block.Proposed, 1)
		err := blproc.CheckConstructionStateProcessingTypeAndIndexesCorrectness(mbh, mb, blockShard)
		assert.ErrorIs(t, err, process.ErrInvalidMiniBlockProcessingTypeForType)
	})

	t.Run("index inconsistency with partial executed", func(t *testing.T) {
		t.Parallel()
		mb := makeMb(blockShard, blockShard, block.TxBlock, true, 3)
		mbh := makeMbh(mb, block.Processed, block.PartialExecuted, 2)
		err := blproc.CheckConstructionStateProcessingTypeAndIndexesCorrectness(mbh, mb, blockShard)
		assert.ErrorIs(t, err, process.ErrInvalidConstructionState)
	})
}

func TestBaseProcessor_ConcurrentCallsNonceOfFirstCommittedBlock(t *testing.T) {
	t.Parallel()

	arguments := CreateMockArguments(createComponentHolderMocks())
	bp, _ := blproc.NewShardProcessor(arguments)

	numCalls := 1000
	wg := &sync.WaitGroup{}
	wg.Add(numCalls)

	mutValuesRead := sync.Mutex{}
	values := make(map[uint64]int)
	noValues := 0
	lastValRead := uint64(0)

	for i := 0; i < numCalls; i++ {
		go func(idx int) {
			time.Sleep(time.Millisecond * 10)

			switch idx % 2 {
			case 0:
				val := bp.NonceOfFirstCommittedBlock()

				mutValuesRead.Lock()
				if val.HasValue {
					values[val.Value]++
					lastValRead = val.Value
				} else {
					noValues++
				}
				mutValuesRead.Unlock()
			case 1:
				bp.SetNonceOfFirstCommittedBlock(uint64(idx))
			}

			wg.Done()
		}(i)
	}

	wg.Wait()

	mutValuesRead.Lock()
	defer mutValuesRead.Unlock()

	assert.True(t, len(values) <= 1) // we can have the situation when all reads are done before the first set
	assert.Equal(t, numCalls/2, values[lastValRead]+noValues)
}

func TestBaseProcessor_CheckSentSignaturesAtCommitTime(t *testing.T) {
	t.Parallel()

	t.Run("nodes coordinator errors, should return error", func(t *testing.T) {
		nodesCoordinatorInstance := shardingMocks.NewNodesCoordinatorMock()
		nodesCoordinatorInstance.ComputeValidatorsGroupCalled = func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validatorsGroup []nodesCoordinator.Validator, err error) {
			return nil, nil, expectedErr
		}

		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.SentSignaturesTracker = &testscommon.SentSignatureTrackerStub{
			ResetCountersForManagedBlockSignerCalled: func(signerPk []byte) {
				assert.Fail(t, "should have not called ResetCountersManagedBlockSigners")
			},
		}
		arguments.NodesCoordinator = nodesCoordinatorInstance
		bp, _ := blproc.NewShardProcessor(arguments)

		err := bp.CheckSentSignaturesAtCommitTime(&block.Header{
			RandSeed:     []byte("randSeed"),
			PrevRandSeed: []byte("prevRandSeed"),
		})
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work with bitmap", func(t *testing.T) {
		validator0, _ := nodesCoordinator.NewValidator([]byte("pk0"), 0, 0)
		validator1, _ := nodesCoordinator.NewValidator([]byte("pk1"), 1, 1)
		validator2, _ := nodesCoordinator.NewValidator([]byte("pk2"), 2, 2)

		nodesCoordinatorInstance := shardingMocks.NewNodesCoordinatorMock()
		nodesCoordinatorInstance.ComputeValidatorsGroupCalled = func(randomness []byte, round uint64, shardId uint32, epoch uint32) (leader nodesCoordinator.Validator, validatorsGroup []nodesCoordinator.Validator, err error) {
			return validator0, []nodesCoordinator.Validator{validator0, validator1, validator2}, nil
		}

		resetCountersCalled := make([][]byte, 0)
		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.SentSignaturesTracker = &testscommon.SentSignatureTrackerStub{
			ResetCountersForManagedBlockSignerCalled: func(signerPk []byte) {
				resetCountersCalled = append(resetCountersCalled, signerPk)
			},
		}
		arguments.NodesCoordinator = nodesCoordinatorInstance
		bp, _ := blproc.NewShardProcessor(arguments)

		err := bp.CheckSentSignaturesAtCommitTime(&block.Header{
			RandSeed:      []byte("randSeed"),
			PrevRandSeed:  []byte("prevRandSeed"),
			PubKeysBitmap: []byte{0b00000101},
		})
		assert.Nil(t, err)

		assert.Equal(t, [][]byte{validator0.PubKey(), validator2.PubKey()}, resetCountersCalled)
	})
}

func TestBaseProcessor_DisplayHeader(t *testing.T) {
	t.Parallel()

	t.Run("shard header with proof info", func(t *testing.T) {
		t.Parallel()

		header := &block.HeaderV2{
			Header: &block.Header{
				ChainID:         []byte("1"),
				Epoch:           2,
				Round:           3,
				TimeStamp:       4,
				Nonce:           5,
				PrevHash:        []byte("prevHash"),
				PrevRandSeed:    []byte("prevRandSeed"),
				RandSeed:        []byte("randSeed"),
				LeaderSignature: []byte("leaderSig"),
				RootHash:        []byte("rootHash"),
				ReceiptsHash:    []byte("receiptsHash"),
			},
			ScheduledRootHash:        []byte("schRootHash"),
			ScheduledAccumulatedFees: big.NewInt(6),
			ScheduledDeveloperFees:   big.NewInt(7),
			ScheduledGasProvided:     8,
			ScheduledGasPenalized:    9,
			ScheduledGasRefunded:     10,
		}
		proof := &block.HeaderProof{
			PubKeysBitmap:       []byte("bitmap"),
			AggregatedSignature: []byte("sig"),
			HeaderHash:          []byte("prevHash"),
			HeaderEpoch:         2,
			HeaderNonce:         4,
			HeaderShardId:       0,
			HeaderRound:         2,
			IsStartOfEpoch:      false,
		}

		lines := blproc.DisplayHeader(header, proof)
		require.Equal(t, 22, len(lines))
	})
	t.Run("shard header V3 with proof info", func(t *testing.T) {
		t.Parallel()

		header := &block.HeaderV3{
			ChainID:         []byte("1"),
			Epoch:           2,
			Round:           3,
			TimestampMs:     4,
			Nonce:           5,
			PrevHash:        []byte("prevHash"),
			PrevRandSeed:    []byte("prevRandSeed"),
			RandSeed:        []byte("randSeed"),
			LeaderSignature: []byte("leaderSig"),
			ReceiptsHash:    []byte("receiptsHash"),
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderHash: []byte("lastExecResult"),
				},
			},
			ExecutionResults: []*block.ExecutionResult{
				{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderHash: []byte("execResult0"),
					},
				},
				{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderHash: []byte("execResult1"),
					},
				},
			},
		}
		proof := &block.HeaderProof{
			PubKeysBitmap:       []byte("bitmap"),
			AggregatedSignature: []byte("sig"),
			HeaderHash:          []byte("prevHash"),
			HeaderEpoch:         2,
			HeaderNonce:         4,
			HeaderShardId:       0,
			HeaderRound:         2,
			IsStartOfEpoch:      false,
		}

		lines := blproc.DisplayHeader(header, proof)
		require.Equal(t, 23, len(lines))
	})
	t.Run("meta header with proof info", func(t *testing.T) {
		t.Parallel()

		header := &block.MetaBlock{
			Nonce:           5,
			Epoch:           2,
			Round:           3,
			TimeStamp:       4,
			LeaderSignature: []byte("leaderSig"),
			PrevHash:        []byte("prevHash"),
			PrevRandSeed:    []byte("prevRandSeed"),
			RandSeed:        []byte("randSeed"),
			RootHash:        []byte("rootHash"),
			ReceiptsHash:    []byte("receiptsHash"),
			EpochStart:      block.EpochStart{},
			ChainID:         []byte("1"),
		}
		proof := &block.HeaderProof{
			PubKeysBitmap:       []byte("bitmap"),
			AggregatedSignature: []byte("sig"),
			HeaderHash:          []byte("prevHash"),
			HeaderEpoch:         2,
			HeaderNonce:         4,
			HeaderShardId:       0,
			HeaderRound:         2,
			IsStartOfEpoch:      false,
		}

		lines := blproc.DisplayHeader(header, proof)
		require.Equal(t, 22, len(lines))
	})
}

func TestBaseProcessor_computeOwnShardStuckIfNeeded(t *testing.T) {
	t.Parallel()

	t.Run("header is not V3, should exit early without error", func(t *testing.T) {
		baseProcessor := blproc.CreateBaseProcessorWithMockedTracker(&mock.BlockTrackerMock{
			ComputeOwnShardStuckCalled: func(_ data.BaseExecutionResultHandler, _ uint64) {
				require.Fail(t, "should not be called")
			},
		})
		header := &block.Header{}

		err := baseProcessor.ComputeOwnShardStuckIfNeeded(header)
		assert.Nil(t, err)
	})

	t.Run("header is V3 but last executed results is nil", func(t *testing.T) {
		header := &block.HeaderV3{
			LastExecutionResult: nil,
		}
		baseProcessor := blproc.CreateBaseProcessorWithMockedTracker(&mock.BlockTrackerMock{
			ComputeOwnShardStuckCalled: func(_ data.BaseExecutionResultHandler, _ uint64) {
				require.Fail(t, "should not be called")
			},
		})

		err := baseProcessor.ComputeOwnShardStuckIfNeeded(header)
		assert.Equal(t, process.ErrNilLastExecutionResultHandler, err)
	})

	t.Run("header is metablock v3, last executed result is nil", func(t *testing.T) {
		header := &block.MetaBlockV3{
			LastExecutionResult: nil,
		}

		baseProcessor := blproc.CreateBaseProcessorWithMockedTracker(&mock.BlockTrackerMock{
			ComputeOwnShardStuckCalled: func(_ data.BaseExecutionResultHandler, _ uint64) {
				require.Fail(t, "should not be called")
			},
		})

		err := baseProcessor.ComputeOwnShardStuckIfNeeded(header)
		assert.Equal(t, process.ErrNilLastExecutionResultHandler, err)
	})

	t.Run("valid shard header v3 with valid last execution result", func(t *testing.T) {
		baseExecutionResults := &block.BaseExecutionResult{
			HeaderHash:  []byte("hash"),
			HeaderNonce: 100,
			HeaderRound: 200,
			RootHash:    []byte("rootHash"),
		}
		header := &block.HeaderV3{
			LastExecutionResult: &block.ExecutionResultInfo{
				NotarizedInRound: 201,
				ExecutionResult:  baseExecutionResults,
			},
		}
		called := false
		baseProcessor := blproc.CreateBaseProcessorWithMockedTracker(&mock.BlockTrackerMock{
			ComputeOwnShardStuckCalled: func(_ data.BaseExecutionResultHandler, _ uint64) {
				called = true
			},
		})

		err := baseProcessor.ComputeOwnShardStuckIfNeeded(header)
		assert.Nil(t, err)
		require.True(t, called)
	})
}

func TestBaseProcessor_updateGasConsumptionLimitsIfNeeded(t *testing.T) {
	t.Parallel()

	isOwnShardStuck := false
	bp := blproc.CreateBaseProcessorWithMockedTracker(&mock.BlockTrackerMock{
		IsOwnShardStuckCalled: func() bool {
			return isOwnShardStuck
		},
	})
	wasResetIncomingLimitCalled := false
	wasResetOutgoingLimitCalled := false
	wasZeroIncomingLimitCalled := false
	wasZeroOutgoingLimitCalled := false
	bp.SetGasComputation(&testscommon.GasComputationMock{
		ResetIncomingLimitCalled: func() {
			wasResetIncomingLimitCalled = true
		},
		ResetOutgoingLimitCalled: func() {
			wasResetOutgoingLimitCalled = true
		},
		ZeroIncomingLimitCalled: func() {
			wasZeroIncomingLimitCalled = true
		},
		ZeroOutgoingLimitCalled: func() {
			wasZeroOutgoingLimitCalled = true
		},
	})

	require.False(t, wasResetIncomingLimitCalled)
	require.False(t, wasResetOutgoingLimitCalled)
	require.False(t, wasZeroIncomingLimitCalled)
	require.False(t, wasZeroOutgoingLimitCalled)

	bp.UpdateGasConsumptionLimitsIfNeeded()
	require.True(t, wasResetIncomingLimitCalled)
	require.True(t, wasResetOutgoingLimitCalled)
	require.False(t, wasZeroIncomingLimitCalled)
	require.False(t, wasZeroOutgoingLimitCalled)

	// set the Reset.* variables to false again
	wasResetIncomingLimitCalled = false
	wasResetOutgoingLimitCalled = false

	// set the shard is stuck to true
	isOwnShardStuck = true

	bp.UpdateGasConsumptionLimitsIfNeeded()
	require.False(t, wasResetIncomingLimitCalled)
	require.False(t, wasResetOutgoingLimitCalled)
	require.True(t, wasZeroIncomingLimitCalled)
	require.True(t, wasZeroOutgoingLimitCalled)
}

func TestCheckHeaderBodyCorrelationProposal(t *testing.T) {
	t.Parallel()

	shardID := uint32(0)
	epoch := uint32(0)
	relayedV1V2DisableEpoch := uint32(5)
	createEnableEpochsHandlerStub := func() *enableEpochsHandlerMock.EnableEpochsHandlerStub {
		return &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.RelayedTransactionsV1V2DisableFlag && epoch >= relayedV1V2DisableEpoch
			},
		}
	}
	t.Run("different number of miniblock headers and miniblocks should error ", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		err := bp.CheckHeaderBodyCorrelationProposal(
			nil,
			&block.Body{MiniBlocks: []*block.MiniBlock{
				{SenderShardID: 0},
			}},
			shardID,
			epoch,
		)
		require.Equal(t, process.ErrHeaderBodyMismatch, err)
	})

	t.Run("nil miniblock should error", func(t *testing.T) {
		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		mbHeaders := make([]data.MiniBlockHeaderHandler, 1)
		mbHeaders[0] = &block.MiniBlockHeader{}

		err := bp.CheckHeaderBodyCorrelationProposal(
			mbHeaders,
			&block.Body{MiniBlocks: []*block.MiniBlock{
				nil,
			}},
			shardID,
			epoch,
		)
		require.Equal(t, process.ErrNilMiniBlock, err)
	})

	t.Run("nil miniblock header should error", func(t *testing.T) {
		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		mbHeaders := make([]data.MiniBlockHeaderHandler, 1)
		mbHeaders[0] = nil

		err := bp.CheckHeaderBodyCorrelationProposal(
			mbHeaders,
			&block.Body{MiniBlocks: []*block.MiniBlock{
				{},
			}},
			shardID,
			epoch,
		)
		require.Equal(t, process.ErrNilMiniBlockHeader, err)
	})
	t.Run("different hash mb header and miniblock", func(t *testing.T) {
		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		mbHeaders := make([]data.MiniBlockHeaderHandler, 1)
		mbHeaders[0] = &block.MiniBlockHeader{
			Hash: []byte("hash"),
		}

		err := bp.CheckHeaderBodyCorrelationProposal(
			mbHeaders,
			&block.Body{MiniBlocks: []*block.MiniBlock{
				{},
			}},
			shardID,
			epoch,
		)
		require.Equal(t, process.ErrHeaderBodyMismatch, err)
	})

	t.Run("different tx count mb header and miniblock", func(t *testing.T) {
		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		miniBlock := &block.MiniBlock{}

		mbHash, _ := core.CalculateHash(arguments.CoreComponents.InternalMarshalizer(), arguments.CoreComponents.Hasher(), miniBlock)

		mbHeaders := make([]data.MiniBlockHeaderHandler, 1)
		mbHeaders[0] = &block.MiniBlockHeader{
			Hash:    mbHash,
			TxCount: 1,
		}

		err := bp.CheckHeaderBodyCorrelationProposal(
			mbHeaders,
			&block.Body{MiniBlocks: []*block.MiniBlock{
				miniBlock,
			}},
			shardID,
			epoch,
		)
		require.Equal(t, process.ErrHeaderBodyMismatch, err)
	})

	t.Run("different receiver shard mb header and mini block", func(t *testing.T) {
		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		miniBlock := &block.MiniBlock{
			ReceiverShardID: 2,
		}

		mbHash, _ := core.CalculateHash(arguments.CoreComponents.InternalMarshalizer(), arguments.CoreComponents.Hasher(), miniBlock)

		mbHeaders := make([]data.MiniBlockHeaderHandler, 1)
		mbHeaders[0] = &block.MiniBlockHeader{
			Hash:            mbHash,
			ReceiverShardID: 1,
		}

		err := bp.CheckHeaderBodyCorrelationProposal(
			mbHeaders,
			&block.Body{MiniBlocks: []*block.MiniBlock{
				miniBlock,
			}},
			shardID,
			epoch,
		)
		require.ErrorIs(t, err, process.ErrHeaderBodyMismatch)
	})

	t.Run("different sender shard mb header and miniblock", func(t *testing.T) {
		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		miniBlock := &block.MiniBlock{
			SenderShardID:   0,
			ReceiverShardID: 2,
		}

		mbHash, _ := core.CalculateHash(arguments.CoreComponents.InternalMarshalizer(), arguments.CoreComponents.Hasher(), miniBlock)

		mbHeaders := make([]data.MiniBlockHeaderHandler, 1)
		mbHeaders[0] = &block.MiniBlockHeader{
			Hash:            mbHash,
			SenderShardID:   2,
			ReceiverShardID: 2,
		}

		err := bp.CheckHeaderBodyCorrelationProposal(
			mbHeaders,
			&block.Body{MiniBlocks: []*block.MiniBlock{
				miniBlock,
			}},
			shardID,
			epoch,
		)
		require.ErrorIs(t, err, process.ErrHeaderBodyMismatch)
	})

	t.Run("wrong construction state should error", func(t *testing.T) {
		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		miniBlock := &block.MiniBlock{
			ReceiverShardID: 2,
			SenderShardID:   0,
			TxHashes:        [][]byte{[]byte("tx1"), []byte("tx2")},
			Type:            block.TxBlock,
			Reserved:        nil,
		}

		mbHash, _ := core.CalculateHash(arguments.CoreComponents.InternalMarshalizer(), arguments.CoreComponents.Hasher(), miniBlock)

		mbHeaders := make([]data.MiniBlockHeaderHandler, 1)
		mbHeaders[0] = &block.MiniBlockHeader{
			Hash:            mbHash,
			SenderShardID:   0,
			ReceiverShardID: 2,
			TxCount:         2,
			Type:            block.TxBlock,
			Reserved:        nil,
		}
		_ = mbHeaders[0].SetConstructionState(int32(block.PartialExecuted))

		err := bp.CheckHeaderBodyCorrelationProposal(
			mbHeaders,
			&block.Body{MiniBlocks: []*block.MiniBlock{
				miniBlock,
			}},
			shardID,
			epoch,
		)
		require.Equal(t, process.ErrWrongMiniBlockConstructionState, err)
	})

	t.Run("wrong processing type should error", func(t *testing.T) {
		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		miniBlock := &block.MiniBlock{
			SenderShardID:   0,
			ReceiverShardID: 2,
			TxHashes:        [][]byte{[]byte("tx1"), []byte("tx2")},
			Type:            block.TxBlock,
			Reserved:        nil,
		}

		mbHash, _ := core.CalculateHash(arguments.CoreComponents.InternalMarshalizer(), arguments.CoreComponents.Hasher(), miniBlock)

		mbHeaders := make([]data.MiniBlockHeaderHandler, 1)
		mbHeaders[0] = &block.MiniBlockHeader{
			Hash:            mbHash,
			SenderShardID:   0,
			ReceiverShardID: 2,
			TxCount:         2,
			Type:            block.TxBlock,
			Reserved:        nil,
		}
		_ = mbHeaders[0].SetConstructionState(int32(block.Proposed))
		_ = mbHeaders[0].SetProcessingType(int32(block.Scheduled))

		err := bp.CheckHeaderBodyCorrelationProposal(
			mbHeaders,
			&block.Body{MiniBlocks: []*block.MiniBlock{
				miniBlock,
			}},
			shardID,
			epoch,
		)
		require.Equal(t, process.ErrWrongMiniBlockProcessingType, err)
	})

	t.Run("should work", func(t *testing.T) {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		bootstrapComponents.Coordinator, _ = sharding.NewMultiShardCoordinator(3, 0)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		miniBlock := &block.MiniBlock{
			SenderShardID:   2,
			ReceiverShardID: 0,
			TxHashes:        [][]byte{[]byte("tx1"), []byte("tx2")},
			Type:            block.TxBlock,
			Reserved:        nil,
		}

		mbHash, _ := core.CalculateHash(arguments.CoreComponents.InternalMarshalizer(), arguments.CoreComponents.Hasher(), miniBlock)

		mbHeaders := make([]data.MiniBlockHeaderHandler, 1)
		mbHeaders[0] = &block.MiniBlockHeader{
			Hash:            mbHash,
			SenderShardID:   2,
			ReceiverShardID: 0,
			TxCount:         2,
			Type:            block.TxBlock,
			Reserved:        nil,
		}
		_ = mbHeaders[0].SetConstructionState(int32(block.Proposed))
		_ = mbHeaders[0].SetProcessingType(int32(block.Normal))

		err := bp.CheckHeaderBodyCorrelationProposal(
			mbHeaders,
			&block.Body{MiniBlocks: []*block.MiniBlock{
				miniBlock,
			}},
			shardID,
			epoch,
		)
		require.NoError(t, err)
	})

	t.Run("duplicate tx hash across miniblocks with proposal should error only after relayed v1/v2 disable epoch", func(t *testing.T) {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.Hash = &hashingMocks.HasherMock{}
		coreComponents.EnableEpochsHandlerField = createEnableEpochsHandlerStub()
		bootstrapComponents.Coordinator, _ = sharding.NewMultiShardCoordinator(3, 0)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		miniBlock1 := &block.MiniBlock{
			SenderShardID:   1,
			ReceiverShardID: 0,
			TxHashes:        [][]byte{[]byte("tx1"), []byte("tx2")},
			Type:            block.TxBlock,
			Reserved:        nil,
		}
		miniBlock2 := &block.MiniBlock{
			SenderShardID:   2,
			ReceiverShardID: 0,
			TxHashes:        [][]byte{[]byte("tx1")},
			Type:            block.TxBlock,
			Reserved:        nil,
		}

		mbHash1, _ := core.CalculateHash(coreComponents.IntMarsh, coreComponents.Hash, miniBlock1)
		mbHash2, _ := core.CalculateHash(coreComponents.IntMarsh, coreComponents.Hash, miniBlock2)

		mbHeaders := make([]data.MiniBlockHeaderHandler, 2)
		mbHeaders[0] = &block.MiniBlockHeader{
			Hash:            mbHash1,
			SenderShardID:   1,
			ReceiverShardID: 0,
			TxCount:         2,
			Type:            block.TxBlock,
			Reserved:        nil,
		}
		_ = mbHeaders[0].SetConstructionState(int32(block.Proposed))
		_ = mbHeaders[0].SetProcessingType(int32(block.Normal))
		mbHeaders[1] = &block.MiniBlockHeader{
			Hash:            mbHash2,
			SenderShardID:   2,
			ReceiverShardID: 0,
			TxCount:         1,
			Type:            block.TxBlock,
			Reserved:        nil,
		}
		_ = mbHeaders[1].SetConstructionState(int32(block.Proposed))
		_ = mbHeaders[1].SetProcessingType(int32(block.Normal))

		body := &block.Body{MiniBlocks: []*block.MiniBlock{miniBlock1, miniBlock2}}

		err := bp.CheckHeaderBodyCorrelationProposal(mbHeaders, body, shardID, relayedV1V2DisableEpoch-1)
		require.NoError(t, err)

		err = bp.CheckHeaderBodyCorrelationProposal(mbHeaders, body, shardID, relayedV1V2DisableEpoch)
		require.Equal(t, process.ErrDuplicatedTransactionInBlockBody, err)
	})

	t.Run("duplicate tx hash within single miniblock with proposal should error only after relayed v1/v2 disable epoch", func(t *testing.T) {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.Hash = &hashingMocks.HasherMock{}
		coreComponents.EnableEpochsHandlerField = createEnableEpochsHandlerStub()
		bootstrapComponents.Coordinator, _ = sharding.NewMultiShardCoordinator(3, 0)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		miniBlock := &block.MiniBlock{
			SenderShardID:   1,
			ReceiverShardID: 0,
			TxHashes:        [][]byte{[]byte("tx1"), []byte("tx1")},
			Type:            block.TxBlock,
			Reserved:        nil,
		}

		mbHash, _ := core.CalculateHash(coreComponents.IntMarsh, coreComponents.Hash, miniBlock)

		mbHeaders := make([]data.MiniBlockHeaderHandler, 1)
		mbHeaders[0] = &block.MiniBlockHeader{
			Hash:            mbHash,
			SenderShardID:   1,
			ReceiverShardID: 0,
			TxCount:         2,
			Type:            block.TxBlock,
			Reserved:        nil,
		}
		_ = mbHeaders[0].SetConstructionState(int32(block.Proposed))
		_ = mbHeaders[0].SetProcessingType(int32(block.Normal))

		body := &block.Body{MiniBlocks: []*block.MiniBlock{miniBlock}}

		err := bp.CheckHeaderBodyCorrelationProposal(mbHeaders, body, shardID, relayedV1V2DisableEpoch-1)
		require.NoError(t, err)

		err = bp.CheckHeaderBodyCorrelationProposal(mbHeaders, body, shardID, relayedV1V2DisableEpoch)
		require.Equal(t, process.ErrDuplicatedTransactionInBlockBody, err)
	})

	t.Run("duplicate tx hash across miniblocks without proposal should error only after relayed v1/v2 disable epoch", func(t *testing.T) {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.Hash = &hashingMocks.HasherMock{}
		coreComponents.EnableEpochsHandlerField = createEnableEpochsHandlerStub()
		bootstrapComponents.Coordinator, _ = sharding.NewMultiShardCoordinator(3, 0)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		miniBlock1 := &block.MiniBlock{
			SenderShardID:   0,
			ReceiverShardID: 1,
			TxHashes:        [][]byte{[]byte("tx1"), []byte("tx2")},
			Type:            block.TxBlock,
			Reserved:        nil,
		}
		miniBlock2 := &block.MiniBlock{
			SenderShardID:   0,
			ReceiverShardID: 2,
			TxHashes:        [][]byte{[]byte("tx1")},
			Type:            block.TxBlock,
			Reserved:        nil,
		}

		mbHash1, _ := core.CalculateHash(coreComponents.IntMarsh, coreComponents.Hash, miniBlock1)
		mbHash2, _ := core.CalculateHash(coreComponents.IntMarsh, coreComponents.Hash, miniBlock2)

		hdr := &block.Header{
			ShardID: shardID,
			Epoch:   relayedV1V2DisableEpoch - 1,
			MiniBlockHeaders: []block.MiniBlockHeader{
				{
					Hash:            mbHash1,
					SenderShardID:   0,
					ReceiverShardID: 1,
					TxCount:         2,
					Type:            block.TxBlock,
					Reserved:        nil,
				},
				{
					Hash:            mbHash2,
					SenderShardID:   0,
					ReceiverShardID: 2,
					TxCount:         1,
					Type:            block.TxBlock,
					Reserved:        nil,
				},
			},
		}
		body := &block.Body{MiniBlocks: []*block.MiniBlock{miniBlock1, miniBlock2}}

		err := bp.CheckHeaderBodyCorrelation(hdr, body)
		require.NoError(t, err)

		hdr.Epoch = relayedV1V2DisableEpoch
		err = bp.CheckHeaderBodyCorrelation(hdr, body)
		require.Equal(t, process.ErrDuplicatedTransactionInBlockBody, err)
	})

	t.Run("duplicate tx hash within single miniblock without proposal should error only after relayed v1/v2 disable epoch", func(t *testing.T) {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.Hash = &hashingMocks.HasherMock{}
		coreComponents.EnableEpochsHandlerField = createEnableEpochsHandlerStub()
		bootstrapComponents.Coordinator, _ = sharding.NewMultiShardCoordinator(3, 0)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, _ := blproc.NewShardProcessor(arguments)

		miniBlock := &block.MiniBlock{
			SenderShardID:   0,
			ReceiverShardID: 1,
			TxHashes:        [][]byte{[]byte("tx1"), []byte("tx1")},
			Type:            block.TxBlock,
			Reserved:        nil,
		}

		mbHash, _ := core.CalculateHash(coreComponents.IntMarsh, coreComponents.Hash, miniBlock)

		hdr := &block.Header{
			ShardID: shardID,
			Epoch:   relayedV1V2DisableEpoch - 1,
			MiniBlockHeaders: []block.MiniBlockHeader{
				{
					Hash:            mbHash,
					SenderShardID:   0,
					ReceiverShardID: 1,
					TxCount:         2,
					Type:            block.TxBlock,
					Reserved:        nil,
				},
			},
		}
		body := &block.Body{MiniBlocks: []*block.MiniBlock{miniBlock}}

		err := bp.CheckHeaderBodyCorrelation(hdr, body)
		require.NoError(t, err)

		hdr.Epoch = relayedV1V2DisableEpoch
		err = bp.CheckHeaderBodyCorrelation(hdr, body)
		require.Equal(t, process.ErrDuplicatedTransactionInBlockBody, err)
	})
}

func createProposalMbHeader(
	t *testing.T,
	marshaller marshal.Marshalizer,
	hasher hashing.Hasher,
	mb *block.MiniBlock,
) *block.MiniBlockHeader {
	mbHash, err := core.CalculateHash(marshaller, hasher, mb)
	require.Nil(t, err)

	mbHeader := &block.MiniBlockHeader{
		Hash:            mbHash,
		SenderShardID:   mb.SenderShardID,
		ReceiverShardID: mb.ReceiverShardID,
		TxCount:         uint32(len(mb.TxHashes)),
		Type:            mb.Type,
	}
	require.Nil(t, mbHeader.SetConstructionState(int32(block.Proposed)))
	require.Nil(t, mbHeader.SetProcessingType(int32(block.Normal)))

	return mbHeader
}

func TestCheckHeaderBodyCorrelationProposal_BodyStructure(t *testing.T) {
	t.Parallel()

	selfShardID := uint32(0)
	epoch := uint32(0)
	scheduledReserved, _ := (&block.MiniBlockReserved{ExecutionType: block.Scheduled}).Marshal()
	processedReserved, _ := (&block.MiniBlockReserved{ExecutionType: block.Processed}).Marshal()

	newSelfMiniBlock := func(txHashes ...[]byte) *block.MiniBlock {
		return &block.MiniBlock{
			SenderShardID:   selfShardID,
			ReceiverShardID: selfShardID,
			TxHashes:        txHashes,
			Type:            block.TxBlock,
		}
	}
	newIncomingMiniBlock := func(txHashes ...[]byte) *block.MiniBlock {
		return &block.MiniBlock{
			SenderShardID:   1,
			ReceiverShardID: selfShardID,
			TxHashes:        txHashes,
			Type:            block.TxBlock,
		}
	}

	newMockArguments := func() blproc.ArgShardProcessor {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.Hash = &hashingMocks.HasherMock{}
		bootstrapComponents.Coordinator, _ = sharding.NewMultiShardCoordinator(3, 0)
		return CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	}

	checkBody := func(t *testing.T, blockShardID uint32, miniBlocks ...*block.MiniBlock) error {
		arguments := newMockArguments()
		bp, _ := blproc.NewShardProcessor(arguments)

		marshaller := arguments.CoreComponents.InternalMarshalizer()
		hasher := arguments.CoreComponents.Hasher()
		mbHeaders := make([]data.MiniBlockHeaderHandler, 0, len(miniBlocks))
		for _, mb := range miniBlocks {
			mbHeaders = append(mbHeaders, createProposalMbHeader(t, marshaller, hasher, mb))
		}

		return bp.CheckHeaderBodyCorrelationProposal(mbHeaders, &block.Body{MiniBlocks: miniBlocks}, blockShardID, epoch)
	}

	t.Run("self-sender with scheduled body marker should error", func(t *testing.T) {
		t.Parallel()

		mb := newSelfMiniBlock([]byte("tx1"), []byte("tx2"))
		mb.Reserved = scheduledReserved
		err := checkBody(t, selfShardID, mb)
		require.ErrorIs(t, err, process.ErrInvalidSelfSenderMiniBlock)
	})

	t.Run("self-sender with processed body marker should error", func(t *testing.T) {
		t.Parallel()

		mb := newSelfMiniBlock([]byte("tx1"))
		mb.Reserved = processedReserved
		err := checkBody(t, selfShardID, mb)
		require.ErrorIs(t, err, process.ErrInvalidSelfSenderMiniBlock)
	})

	t.Run("incoming after self-sender should error", func(t *testing.T) {
		t.Parallel()

		err := checkBody(t, selfShardID,
			newIncomingMiniBlock([]byte("tx1")),
			newSelfMiniBlock([]byte("tx2")),
			newIncomingMiniBlock([]byte("tx3")),
		)
		require.ErrorIs(t, err, process.ErrSelfSenderMiniBlockNotLast)
	})

	t.Run("two self-sender miniblocks should error", func(t *testing.T) {
		t.Parallel()

		err := checkBody(t, selfShardID,
			newSelfMiniBlock([]byte("tx1")),
			newSelfMiniBlock([]byte("tx2")),
		)
		require.ErrorIs(t, err, process.ErrMultipleSelfSenderMiniBlocks)
	})

	t.Run("self-sender smart contract result miniblock should error", func(t *testing.T) {
		t.Parallel()

		mb := newSelfMiniBlock([]byte("scr1"))
		mb.Type = block.SmartContractResultBlock
		err := checkBody(t, selfShardID, mb)
		require.ErrorIs(t, err, process.ErrInvalidSelfSenderMiniBlock)
	})

	t.Run("self-sender with foreign receiver should error", func(t *testing.T) {
		t.Parallel()

		mb := newSelfMiniBlock([]byte("tx1"))
		mb.ReceiverShardID = core.MetachainShardId
		err := checkBody(t, selfShardID, mb)
		require.ErrorIs(t, err, process.ErrInvalidSelfSenderMiniBlock)
	})

	t.Run("empty self-sender miniblock should error", func(t *testing.T) {
		t.Parallel()

		err := checkBody(t, selfShardID, newSelfMiniBlock())
		require.ErrorIs(t, err, process.ErrIndexIsOutOfBound)
	})

	t.Run("self-sender with partial indexes should error", func(t *testing.T) {
		t.Parallel()

		arguments := newMockArguments()
		bp, _ := blproc.NewShardProcessor(arguments)

		mb := newSelfMiniBlock([]byte("tx1"), []byte("tx2"), []byte("tx3"))
		mbHeader := createProposalMbHeader(t, arguments.CoreComponents.InternalMarshalizer(), arguments.CoreComponents.Hasher(), mb)
		require.Nil(t, mbHeader.SetIndexOfLastTxProcessed(1))

		err := bp.CheckHeaderBodyCorrelationProposal(
			[]data.MiniBlockHeaderHandler{mbHeader},
			&block.Body{MiniBlocks: []*block.MiniBlock{mb}},
			selfShardID,
			epoch,
		)
		require.ErrorIs(t, err, process.ErrInvalidSelfSenderIndexes)
	})

	t.Run("meta proposal with self-sender miniblock should error", func(t *testing.T) {
		t.Parallel()

		mb := &block.MiniBlock{
			SenderShardID:   core.MetachainShardId,
			ReceiverShardID: core.MetachainShardId,
			TxHashes:        [][]byte{[]byte("tx1")},
			Type:            block.TxBlock,
		}
		err := checkBody(t, core.MetachainShardId, mb)
		require.ErrorIs(t, err, process.ErrSelfSenderMiniBlockOnMeta)
	})

	t.Run("meta proposal with meta-sender rewards miniblock should error", func(t *testing.T) {
		t.Parallel()

		mb := &block.MiniBlock{
			SenderShardID:   core.MetachainShardId,
			ReceiverShardID: 0,
			TxHashes:        [][]byte{[]byte("rwd1")},
			Type:            block.RewardsBlock,
		}
		err := checkBody(t, core.MetachainShardId, mb)
		require.ErrorIs(t, err, process.ErrSelfSenderMiniBlockOnMeta)
	})

	t.Run("canonical body should work", func(t *testing.T) {
		t.Parallel()

		err := checkBody(t, selfShardID,
			newIncomingMiniBlock([]byte("tx1")),
			newIncomingMiniBlock([]byte("tx2"), []byte("tx3")),
			newSelfMiniBlock([]byte("tx4"), []byte("tx5")),
		)
		require.NoError(t, err)
	})

	t.Run("incoming-only and empty body should work", func(t *testing.T) {
		t.Parallel()

		err := checkBody(t, selfShardID, newIncomingMiniBlock([]byte("tx1")))
		require.NoError(t, err)

		err = checkBody(t, selfShardID)
		require.NoError(t, err)
	})

	t.Run("incoming with scheduled body marker should work", func(t *testing.T) {
		t.Parallel()

		mb := newIncomingMiniBlock([]byte("tx1"), []byte("tx2"))
		mb.Reserved = scheduledReserved
		err := checkBody(t, selfShardID, mb)
		require.NoError(t, err)
	})

	t.Run("incoming partial continuation follows the tracker rule", func(t *testing.T) {
		t.Parallel()

		arguments := newMockArguments()
		arguments.ProcessedMiniBlocksTracker = &testscommon.ProcessedMiniBlocksTrackerStub{
			GetProcessedMiniBlockInfoCalled: func(miniBlockHash []byte) (*processedMb.ProcessedMiniBlockInfo, []byte) {
				return &processedMb.ProcessedMiniBlockInfo{
					FullyProcessed:         false,
					IndexOfLastTxProcessed: 1,
				}, nil
			},
		}
		bp, _ := blproc.NewShardProcessor(arguments)

		mb := newIncomingMiniBlock([]byte("tx1"), []byte("tx2"), []byte("tx3"), []byte("tx4"))
		marshaller := arguments.CoreComponents.InternalMarshalizer()
		hasher := arguments.CoreComponents.Hasher()

		continuationHeader := createProposalMbHeader(t, marshaller, hasher, mb)
		require.Nil(t, continuationHeader.SetIndexOfFirstTxProcessed(2))
		err := bp.CheckHeaderBodyCorrelationProposal(
			[]data.MiniBlockHeaderHandler{continuationHeader},
			&block.Body{MiniBlocks: []*block.MiniBlock{mb}},
			selfShardID,
			epoch,
		)
		require.NoError(t, err)

		wrongStartHeader := createProposalMbHeader(t, marshaller, hasher, mb)
		err = bp.CheckHeaderBodyCorrelationProposal(
			[]data.MiniBlockHeaderHandler{wrongStartHeader},
			&block.Body{MiniBlocks: []*block.MiniBlock{mb}},
			selfShardID,
			epoch,
		)
		require.ErrorIs(t, err, process.ErrIndexOfFirstTxProcessedMismatch)
	})

	t.Run("legacy path still accepts shapes the proposal rules forbid", func(t *testing.T) {
		t.Parallel()

		arguments := newMockArguments()
		bp, _ := blproc.NewShardProcessor(arguments)

		marshaller := arguments.CoreComponents.InternalMarshalizer()
		hasher := arguments.CoreComponents.Hasher()

		mb1 := &block.MiniBlock{SenderShardID: 0, ReceiverShardID: 1, TxHashes: [][]byte{[]byte("tx1")}, Type: block.TxBlock}
		mb2 := &block.MiniBlock{SenderShardID: 0, ReceiverShardID: 2, TxHashes: [][]byte{[]byte("tx2")}, Type: block.TxBlock}
		mbHash1, _ := core.CalculateHash(marshaller, hasher, mb1)
		mbHash2, _ := core.CalculateHash(marshaller, hasher, mb2)

		hdr := &block.Header{
			ShardID: selfShardID,
			MiniBlockHeaders: []block.MiniBlockHeader{
				{Hash: mbHash1, SenderShardID: 0, ReceiverShardID: 1, TxCount: 1, Type: block.TxBlock},
				{Hash: mbHash2, SenderShardID: 0, ReceiverShardID: 2, TxCount: 1, Type: block.TxBlock},
			},
		}
		err := bp.CheckHeaderBodyCorrelation(hdr, &block.Body{MiniBlocks: []*block.MiniBlock{mb1, mb2}})
		require.NoError(t, err)

		mbSched := &block.MiniBlock{SenderShardID: 0, ReceiverShardID: 0, TxHashes: [][]byte{[]byte("tx3")}, Type: block.TxBlock, Reserved: scheduledReserved}
		mbSchedHash, _ := core.CalculateHash(marshaller, hasher, mbSched)
		hdrSched := &block.Header{
			ShardID: selfShardID,
			MiniBlockHeaders: []block.MiniBlockHeader{
				{Hash: mbSchedHash, SenderShardID: 0, ReceiverShardID: 0, TxCount: 1, Type: block.TxBlock},
			},
		}
		err = bp.CheckHeaderBodyCorrelation(hdrSched, &block.Body{MiniBlocks: []*block.MiniBlock{mbSched}})
		require.ErrorIs(t, err, process.ErrProcessingTypeBodyHeaderMismatch)
	})
}

func TestCheckProposalMiniBlocksConsistency(t *testing.T) {
	t.Parallel()

	selfShardID := uint32(0)
	scheduledReserved, _ := (&block.MiniBlockReserved{ExecutionType: block.Scheduled}).Marshal()

	newSelfMb := func(txHashes ...[]byte) *block.MiniBlock {
		return &block.MiniBlock{SenderShardID: selfShardID, ReceiverShardID: selfShardID, TxHashes: txHashes, Type: block.TxBlock}
	}
	newIncomingMb := func(txHashes ...[]byte) *block.MiniBlock {
		return &block.MiniBlock{SenderShardID: 1, ReceiverShardID: selfShardID, TxHashes: txHashes, Type: block.TxBlock}
	}
	headersFor := func(mbs ...*block.MiniBlock) []data.MiniBlockHeaderHandler {
		mbHeaders := make([]data.MiniBlockHeaderHandler, 0, len(mbs))
		for _, mb := range mbs {
			mbHeaders = append(mbHeaders, newProposalMbHeaderForMb(mb))
		}
		return mbHeaders
	}

	t.Run("count mismatch should error", func(t *testing.T) {
		t.Parallel()

		mb := newIncomingMb([]byte("tx1"))
		err := blproc.CheckProposalMiniBlocksConsistency(headersFor(mb, mb), block.MiniBlockSlice{mb}, selfShardID)
		require.Equal(t, process.ErrNumOfMiniBlocksAndMiniBlocksHeadersMismatch, err)
	})

	t.Run("nil miniblock should error", func(t *testing.T) {
		t.Parallel()

		err := blproc.CheckProposalMiniBlocksConsistency(headersFor(newIncomingMb([]byte("tx1"))), block.MiniBlockSlice{nil}, selfShardID)
		require.Equal(t, process.ErrNilMiniBlock, err)
	})

	t.Run("nil miniblock header should error", func(t *testing.T) {
		t.Parallel()

		err := blproc.CheckProposalMiniBlocksConsistency([]data.MiniBlockHeaderHandler{nil}, block.MiniBlockSlice{newIncomingMb([]byte("tx1"))}, selfShardID)
		require.Equal(t, process.ErrNilMiniBlockHeader, err)
	})

	t.Run("header body field mismatch should error", func(t *testing.T) {
		t.Parallel()

		mb := newIncomingMb([]byte("tx1"))
		otherMb := newIncomingMb([]byte("tx1"))
		otherMb.ReceiverShardID = 1
		err := blproc.CheckProposalMiniBlocksConsistency(headersFor(otherMb), block.MiniBlockSlice{mb}, selfShardID)
		require.ErrorIs(t, err, process.ErrHeaderBodyMismatch)
	})

	t.Run("scheduled-marked self-sender should error", func(t *testing.T) {
		t.Parallel()

		mb := newSelfMb([]byte("tx1"))
		mb.Reserved = scheduledReserved
		err := blproc.CheckProposalMiniBlocksConsistency(headersFor(mb), block.MiniBlockSlice{mb}, selfShardID)
		require.ErrorIs(t, err, process.ErrInvalidSelfSenderMiniBlock)
	})

	t.Run("two self-sender miniblocks should error", func(t *testing.T) {
		t.Parallel()

		mb1 := newSelfMb([]byte("tx1"))
		mb2 := newSelfMb([]byte("tx2"))
		err := blproc.CheckProposalMiniBlocksConsistency(headersFor(mb1, mb2), block.MiniBlockSlice{mb1, mb2}, selfShardID)
		require.ErrorIs(t, err, process.ErrMultipleSelfSenderMiniBlocks)
	})

	t.Run("incoming after self-sender should error", func(t *testing.T) {
		t.Parallel()

		selfMb := newSelfMb([]byte("tx1"))
		incomingMb := newIncomingMb([]byte("tx2"))
		err := blproc.CheckProposalMiniBlocksConsistency(headersFor(selfMb, incomingMb), block.MiniBlockSlice{selfMb, incomingMb}, selfShardID)
		require.ErrorIs(t, err, process.ErrSelfSenderMiniBlockNotLast)
	})

	t.Run("meta-sender miniblock on meta should error", func(t *testing.T) {
		t.Parallel()

		mb := &block.MiniBlock{
			SenderShardID:   core.MetachainShardId,
			ReceiverShardID: 0,
			TxHashes:        [][]byte{[]byte("rwd1")},
			Type:            block.RewardsBlock,
		}
		err := blproc.CheckProposalMiniBlocksConsistency(headersFor(mb), block.MiniBlockSlice{mb}, core.MetachainShardId)
		require.ErrorIs(t, err, process.ErrSelfSenderMiniBlockOnMeta)
	})

	t.Run("canonical, incoming-only and empty should work", func(t *testing.T) {
		t.Parallel()

		incomingMb := newIncomingMb([]byte("tx1"))
		selfMb := newSelfMb([]byte("tx2"), []byte("tx3"))
		err := blproc.CheckProposalMiniBlocksConsistency(headersFor(incomingMb, selfMb), block.MiniBlockSlice{incomingMb, selfMb}, selfShardID)
		require.NoError(t, err)

		err = blproc.CheckProposalMiniBlocksConsistency(headersFor(incomingMb), block.MiniBlockSlice{incomingMb}, selfShardID)
		require.NoError(t, err)

		err = blproc.CheckProposalMiniBlocksConsistency(nil, block.MiniBlockSlice{}, selfShardID)
		require.NoError(t, err)
	})
}

func TestCheckLegacyPredecessorReadyForV3(t *testing.T) {
	t.Parallel()

	prevHash := []byte("prev hash")

	buildProcessor := func(t *testing.T, currentHeader data.HeaderHandler, currentHash []byte) interface {
		CheckLegacyPredecessorReadyForV3(header data.HeaderHandler) error
	} {
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return currentHeader
			},
			GetCurrentBlockHeaderHashCalled: func() []byte {
				return currentHash
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)
		return bp
	}

	newLegacyHeader := func(mbHeaders ...block.MiniBlockHeader) *block.Header {
		return &block.Header{Nonce: 1, MiniBlockHeaders: mbHeaders}
	}
	newScheduledMbHeader := func(hash string) block.MiniBlockHeader {
		mbh := block.MiniBlockHeader{Hash: []byte(hash)}
		_ = mbh.SetProcessingType(int32(block.Scheduled))
		_ = mbh.SetConstructionState(int32(block.Proposed))
		return mbh
	}
	candidate := &block.HeaderV3{Nonce: 2, PrevHash: prevHash}

	t.Run("nil chain head should work", func(t *testing.T) {
		t.Parallel()

		bp := buildProcessor(t, nil, prevHash)
		require.NoError(t, bp.CheckLegacyPredecessorReadyForV3(candidate))
	})

	t.Run("v3 predecessor should work", func(t *testing.T) {
		t.Parallel()

		bp := buildProcessor(t, &block.HeaderV3{Nonce: 1}, prevHash)
		require.NoError(t, bp.CheckLegacyPredecessorReadyForV3(candidate))
	})

	t.Run("prev hash mismatch should skip the check", func(t *testing.T) {
		t.Parallel()

		bp := buildProcessor(t, newLegacyHeader(newScheduledMbHeader("leftover")), []byte("other hash"))
		require.NoError(t, bp.CheckLegacyPredecessorReadyForV3(candidate))
	})

	t.Run("clean legacy predecessor should work", func(t *testing.T) {
		t.Parallel()

		bp := buildProcessor(t, newLegacyHeader(block.MiniBlockHeader{Hash: []byte("final")}), prevHash)
		require.NoError(t, bp.CheckLegacyPredecessorReadyForV3(candidate))
	})

	t.Run("scheduled leftover should error", func(t *testing.T) {
		t.Parallel()

		bp := buildProcessor(t, newLegacyHeader(newScheduledMbHeader("leftover")), prevHash)
		err := bp.CheckLegacyPredecessorReadyForV3(candidate)
		require.ErrorIs(t, err, process.ErrLeftoverScheduledMiniBlocksOnTransition)
	})

	t.Run("partially executed leftover should error", func(t *testing.T) {
		t.Parallel()

		partialMbHeader := block.MiniBlockHeader{Hash: []byte("partial")}
		_ = partialMbHeader.SetConstructionState(int32(block.PartialExecuted))
		bp := buildProcessor(t, newLegacyHeader(partialMbHeader), prevHash)
		err := bp.CheckLegacyPredecessorReadyForV3(candidate)
		require.ErrorIs(t, err, process.ErrLeftoverScheduledMiniBlocksOnTransition)
	})
}

func TestBaseProcessor_GetFinalMiniBlocksFromExecutionResult(t *testing.T) {
	t.Parallel()

	t.Run("no execution results, should return empty body", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		bp, _ := blproc.NewShardProcessor(arguments)

		header := &block.HeaderV3{}

		body, _, err := bp.GetFinalMiniBlocksFromExecutionResults(header)
		require.Nil(t, err)
		require.Equal(t, &block.Body{}, body)
	})

	t.Run("should fail if miniblock not found in cache", func(t *testing.T) {
		t.Parallel()

		executedMBs := &cache.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return nil, false
			},
		}
		dataPool := initDataPool()
		dataPool.ExecutedMiniBlocksCalled = func() storage.Cacher {
			return executedMBs
		}

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.DataPool = dataPool

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		bp, _ := blproc.NewShardProcessor(arguments)

		executionResults := []*block.ExecutionResult{
			{
				MiniBlockHeaders: []block.MiniBlockHeader{
					{
						Hash:            []byte("mbHash1"),
						ReceiverShardID: 1,
						SenderShardID:   0,
					},
					{
						Hash:            []byte("mbHash2"),
						ReceiverShardID: 1,
						SenderShardID:   0,
					},
				},
			},
		}
		header := &block.HeaderV3{
			ExecutionResults: executionResults,
		}

		body, _, err := bp.GetFinalMiniBlocksFromExecutionResults(header)
		require.Equal(t, process.ErrMissingMiniBlock, err)
		require.Nil(t, body)
	})

	t.Run("should fail if miniblock not marshalled properly", func(t *testing.T) {
		t.Parallel()

		executedMBs := &cache.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				return []byte("invalid miniblock"), true
			},
		}
		dataPool := initDataPool()
		dataPool.ExecutedMiniBlocksCalled = func() storage.Cacher {
			return executedMBs
		}

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.DataPool = dataPool

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		bp, _ := blproc.NewShardProcessor(arguments)

		executionResults := []*block.ExecutionResult{
			{
				MiniBlockHeaders: []block.MiniBlockHeader{
					{
						Hash:            []byte("mbHash1"),
						ReceiverShardID: 1,
						SenderShardID:   0,
					},
				},
			},
		}
		header := &block.HeaderV3{
			ExecutionResults: executionResults,
		}

		body, _, err := bp.GetFinalMiniBlocksFromExecutionResults(header)
		require.Error(t, err) // unmarshall err
		require.Nil(t, body)
	})

	t.Run("should work for shard header", func(t *testing.T) {
		t.Parallel()

		marshalizer := &mock.MarshalizerMock{
			Fail: false,
		}

		mb1 := &block.MiniBlock{
			TxHashes:        [][]byte{[]byte("txHash1")},
			ReceiverShardID: 1,
			SenderShardID:   2,
		}

		executedMBs := &cache.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				marshalledMb, _ := marshalizer.Marshal(mb1)
				return marshalledMb, true
			},
		}
		dataPool := initDataPool()
		dataPool.ExecutedMiniBlocksCalled = func() storage.Cacher {
			return executedMBs
		}

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.DataPool = dataPool
		coreComponents.IntMarsh = marshalizer

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		bp, _ := blproc.NewShardProcessor(arguments)

		executionResults := []*block.ExecutionResult{
			{
				MiniBlockHeaders: []block.MiniBlockHeader{
					{
						Hash:            []byte("mbHash1"),
						ReceiverShardID: 1,
						SenderShardID:   0,
					},
				},
			},
		}
		header := &block.HeaderV3{
			ExecutionResults: executionResults,
		}

		body, _, err := bp.GetFinalMiniBlocksFromExecutionResults(header)
		require.Nil(t, err)
		require.Equal(t, &block.Body{
			MiniBlocks: []*block.MiniBlock{mb1},
		}, body)
	})

	t.Run("should work for meta block", func(t *testing.T) {
		t.Parallel()

		marshalizer := &mock.MarshalizerMock{
			Fail: false,
		}

		mb1 := &block.MiniBlock{
			TxHashes:        [][]byte{[]byte("txHash1")},
			ReceiverShardID: 1,
			SenderShardID:   2,
		}

		executedMBs := &cache.CacherStub{
			GetCalled: func(key []byte) (value interface{}, ok bool) {
				marshalledMb, _ := marshalizer.Marshal(mb1)
				return marshalledMb, true
			},
		}
		dataPool := initDataPool()
		dataPool.ExecutedMiniBlocksCalled = func() storage.Cacher {
			return executedMBs
		}

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.DataPool = dataPool
		coreComponents.IntMarsh = marshalizer

		arguments := createMockMetaArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		mp, _ := blproc.NewMetaProcessor(arguments)

		executionResults := []*block.MetaExecutionResult{
			{
				MiniBlockHeaders: []block.MiniBlockHeader{
					{
						Hash:            []byte("mbHash1"),
						ReceiverShardID: 1,
						SenderShardID:   0,
					},
				},
			},
		}
		header := &block.MetaBlockV3{
			ExecutionResults: executionResults,
		}

		body, _, err := mp.GetFinalMiniBlocksFromExecutionResults(header)
		require.Nil(t, err)
		require.Equal(t, &block.Body{
			MiniBlocks: []*block.MiniBlock{mb1},
		}, body)
	})
}

func TestBaseProcessor_GetFinalBlockNonce(t *testing.T) {
	t.Parallel()

	t.Run("should return fork detector final nonce, if current block not header v3", func(t *testing.T) {
		t.Parallel()

		finalHash := []byte("finalHash")
		finalNonce := uint64(10)

		header := &block.Header{
			Nonce: 11,
		}

		dataPool := initDataPool()
		dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
			return &pool.HeadersPoolStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					return nil, expectedError
				},
			}
		}

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.DataPool = dataPool

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ForkDetector = &mock.ForkDetectorMock{
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return finalNonce
			},
			GetHighestFinalBlockHashCalled: func() []byte {
				return finalHash
			},
		}

		bp, _ := blproc.NewShardProcessor(arguments)

		retNoncesToFinal := bp.GetFinalBlockNonce(header)
		require.Equal(t, finalNonce, retNoncesToFinal)
	})

	t.Run("should return fork detector final nonce, if failed to get header from pool", func(t *testing.T) {
		t.Parallel()

		finalHash := []byte("finalHash")
		finalNonce := uint64(10)

		header := &block.HeaderV3{
			Nonce: 11,
		}

		dataPool := initDataPool()
		dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
			return &pool.HeadersPoolStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					return nil, expectedError
				},
			}
		}

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.DataPool = dataPool

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ForkDetector = &mock.ForkDetectorMock{
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return finalNonce
			},
			GetHighestFinalBlockHashCalled: func() []byte {
				return finalHash
			},
		}

		bp, _ := blproc.NewShardProcessor(arguments)

		retNoncesToFinal := bp.GetFinalBlockNonce(header)
		require.Equal(t, finalNonce, retNoncesToFinal)
	})

	t.Run("should return last final nonce, if final block not header v3", func(t *testing.T) {
		t.Parallel()

		finalHash := []byte("finalHash")

		header := &block.HeaderV3{
			Nonce: 11,
		}

		finalNonce := uint64(10)

		finalHeader := &block.Header{
			Nonce: finalNonce,
		}

		dataPool := initDataPool()
		dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
			return &pool.HeadersPoolStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					if bytes.Equal(hash, finalHash) {
						return finalHeader, nil
					}

					return nil, expectedError
				},
			}
		}

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.DataPool = dataPool

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ForkDetector = &mock.ForkDetectorMock{
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return finalNonce
			},
			GetHighestFinalBlockHashCalled: func() []byte {
				return finalHash
			},
		}

		bp, _ := blproc.NewShardProcessor(arguments)

		retNoncesToFinal := bp.GetFinalBlockNonce(header)
		require.Equal(t, finalNonce, retNoncesToFinal)
	})

	t.Run("should return last executed final nonce, if header v3", func(t *testing.T) {
		t.Parallel()
		finalHash := []byte("finalHash")

		header := &block.HeaderV3{
			Nonce: 11,
		}

		finalNonce := uint64(10)
		finalExecResNonce := uint64(9)

		finalHeader := &block.HeaderV3{
			Nonce: finalNonce,
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					HeaderNonce: finalExecResNonce,
				},
			},
		}

		dataPool := initDataPool()
		dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
			return &pool.HeadersPoolStub{
				GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
					if bytes.Equal(hash, finalHash) {
						return finalHeader, nil
					}

					return nil, expectedError
				},
			}
		}

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.DataPool = dataPool

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.ForkDetector = &mock.ForkDetectorMock{
			GetHighestFinalBlockNonceCalled: func() uint64 {
				return finalNonce
			},
			GetHighestFinalBlockHashCalled: func() []byte {
				return finalHash
			},
		}

		bp, _ := blproc.NewShardProcessor(arguments)

		retNoncesToFinal := bp.GetFinalBlockNonce(header)
		require.Equal(t, finalExecResNonce, retNoncesToFinal)
	})
}

func TestBaseProcessor_RecreateTrieIfNeeded(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockRootHashCalled: func() []byte {
				return []byte("rootHash")
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		accounts := &stateMock.AccountsStub{
			RecreateTrieIfNeededCalled: func(options common.RootHashHolder) error {
				if bytes.Equal(options.GetRootHash(), []byte("rootHash")) {
					return nil
				}
				return expectedErr
			},
		}
		arguments.AccountsProposal = accounts
		bp, err := blproc.NewShardProcessor(arguments)

		require.NoError(t, err)

		err = bp.RecreateTrieIfNeeded()
		require.NoError(t, err)
	})

	t.Run("should return error from accounts proposal", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		accounts := &stateMock.AccountsStub{
			RecreateTrieIfNeededCalled: func(options common.RootHashHolder) error {
				return expectedErr
			},
		}
		arguments.AccountsProposal = accounts
		bp, err := blproc.NewShardProcessor(arguments)

		require.NoError(t, err)

		err = bp.RecreateTrieIfNeeded()
		require.Equal(t, expectedErr, err)
	})
}

func TestBaseProcessor_OnExecutedBlock(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{
					RootHash: []byte("genesisRootHash"),
				}
			},
		}

		dataPool := initDataPool()
		dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataCacheNotifierMock{
				OnExecutedBlockCalled: func(blockHeader data.HeaderHandler, rootHash []byte) error {
					if bytes.Equal(rootHash, []byte("rootHash")) || bytes.Equal(rootHash, []byte("genesisRootHash")) {
						return nil
					}
					return expectedErr
				},
			}
		}
		dataComponents.DataPool = dataPool

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		err = bp.OnExecutedBlock(&block.Header{}, []byte("rootHash"))
		require.NoError(t, err)
	})

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetGenesisHeaderCalled: func() data.HeaderHandler {
				return &block.Header{
					RootHash: []byte("genesisRootHash"),
				}
			},
		}

		dataPool := initDataPool()
		dataPool.TransactionsCalled = func() dataRetriever.ShardedDataCacherNotifier {
			return &testscommon.ShardedDataCacheNotifierMock{
				OnExecutedBlockCalled: func(blockHeader data.HeaderHandler, rootHash []byte) error {
					if bytes.Equal(rootHash, []byte("genesisRootHash")) {
						return nil
					}
					return expectedErr
				},
			}
		}
		dataComponents.DataPool = dataPool

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		err = bp.OnExecutedBlock(nil, []byte("hash"))
		require.Equal(t, expectedErr, err)
	})
}

func TestBaseProcessor_RequestProof(t *testing.T) {
	t.Parallel()

	t.Run("should not request if flag not enabled", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag != common.AndromedaFlag
			},
		}

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		requestCalled := false
		arguments.RequestHandler = &testscommon.RequestHandlerStub{
			RequestEquivalentProofByNonceCalled: func(headerShard uint32, headerNonce uint64) {
				requestCalled = true
			},
		}

		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		bp.RequestProofIfNeeded(10, 1, 2)

		require.False(t, requestCalled)
	})

	t.Run("should not request if proof already in pool", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataPool := initDataPool()
		dataPool.ProofsCalled = func() dataRetriever.ProofsPool {
			return &dataRetrieverMock.ProofsPoolMock{
				GetProofByNonceCalled: func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
					return &block.HeaderProof{}, nil
				},
			}
		}
		dataComponents.DataPool = dataPool

		coreComponents.EnableEpochsHandlerField = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		requestCalled := false
		arguments.RequestHandler = &testscommon.RequestHandlerStub{
			RequestEquivalentProofByNonceCalled: func(headerShard uint32, headerNonce uint64) {
				requestCalled = true
			},
		}

		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		bp.RequestProofIfNeeded(10, 1, 2)

		require.False(t, requestCalled)
	})

	t.Run("should request if flag enabled and proof not already in pool", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.EnableEpochsHandlerField = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsFlagEnabledInEpochCalled: func(flag core.EnableEpochFlag, epoch uint32) bool {
				return flag == common.AndromedaFlag
			},
		}

		dataPool := initDataPool()
		dataPool.ProofsCalled = func() dataRetriever.ProofsPool {
			return &dataRetrieverMock.ProofsPoolMock{
				GetProofByNonceCalled: func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
					return nil, errors.New("fetch err")
				},
			}
		}
		dataComponents.DataPool = dataPool

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		requestCalled := false
		arguments.RequestHandler = &testscommon.RequestHandlerStub{
			RequestEquivalentProofByNonceCalled: func(headerShard uint32, headerNonce uint64) {
				requestCalled = true
			},
		}

		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		bp.RequestProofIfNeeded(10, 1, 2)

		require.True(t, requestCalled)
	})
}

func TestBaseProcessor_RequestHeadersFromHeaderIfNeeded(t *testing.T) {
	t.Parallel()

	t.Run("header not already in pool, should request", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.RoundField = &testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 20
			},
		}
		coreComponents.ProcessConfigsHandlerField = &testscommon.ProcessConfigsHandlerStub{
			GetMaxRoundsWithoutNewBlockReceivedByRoundCalled: func(round uint64) uint32 {
				return 5
			},
		}

		headersPool := &mock.HeadersCacherStub{
			GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
				return make([]data.HeaderHandler, 0), [][]byte{}, errors.New("some err")
			},
		}
		dataPool := initDataPool()
		dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
			return headersPool
		}
		dataComponents.DataPool = dataPool

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		numCalls := 0
		arguments.RequestHandler = &testscommon.RequestHandlerStub{
			RequestShardHeaderByNonceCalled: func(shardID uint32, nonce uint64) {
				numCalls++
			},
		}

		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		header := &block.HeaderV3{
			Round:   10,
			ShardID: 2,
		}

		bp.RequestHeadersFromHeaderIfNeeded(header)

		require.Equal(t, 11, numCalls) // starting from next header + 10 given by constant
	})

	t.Run("header already in pool, should not request", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.RoundField = &testscommon.RoundHandlerMock{
			IndexCalled: func() int64 {
				return 20
			},
		}
		coreComponents.ProcessConfigsHandlerField = &testscommon.ProcessConfigsHandlerStub{
			GetMaxRoundsWithoutNewBlockReceivedByRoundCalled: func(round uint64) uint32 {
				return 5
			},
		}

		headersPool := &mock.HeadersCacherStub{
			GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) ([]data.HeaderHandler, [][]byte, error) {
				return make([]data.HeaderHandler, 0), [][]byte{}, nil
			},
		}
		dataPool := initDataPool()
		dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
			return headersPool
		}
		dataComponents.DataPool = dataPool

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		requestCalled := false
		arguments.RequestHandler = &testscommon.RequestHandlerStub{
			RequestShardHeaderByNonceCalled: func(shardID uint32, nonce uint64) {
				requestCalled = true
			},
		}

		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		header := &block.HeaderV3{
			Round:   10,
			ShardID: 2,
		}

		bp.RequestHeadersFromHeaderIfNeeded(header)

		require.False(t, requestCalled)
	})
}

func TestBaseProcessor_extractRootHashForCleanup(t *testing.T) {
	t.Parallel()

	t.Run("should return ErrNilLastExecutionResultHandler error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		rootHashHolder, err := bp.ExtractRootHashForCleanup(&block.HeaderV3{})
		require.Nil(t, rootHashHolder)
		require.Equal(t, process.ErrNilLastExecutionResultHandler, err)
	})

	t.Run("should work for HeaderV3", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		expectedRootHash := holders.NewDefaultRootHashesHolder([]byte("rootHash"))

		rootHashHolder, err := bp.ExtractRootHashForCleanup(&block.HeaderV3{
			LastExecutionResult: &block.ExecutionResultInfo{
				ExecutionResult: &block.BaseExecutionResult{
					RootHash: []byte("rootHash"),
				},
			},
		})
		require.Nil(t, err)
		require.Equal(t, expectedRootHash, rootHashHolder)
	})

	t.Run("should work for other HeaderV2", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		expectedRootHash := holders.NewDefaultRootHashesHolder([]byte("rootHash"))

		rootHashHolder, err := bp.ExtractRootHashForCleanup(&block.HeaderV2{
			ScheduledRootHash: []byte("rootHash"),
		})
		require.Nil(t, err)
		require.Equal(t, expectedRootHash, rootHashHolder)
	})

	t.Run("should work for other HeaderV1", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
		err := blkc.SetGenesisHeader(&block.Header{Nonce: 0})
		require.Nil(t, err)

		err = blkc.SetCurrentBlockHeaderAndRootHash(&block.Header{}, []byte("rootHash"))
		require.Nil(t, err)

		dataComponents.BlockChain = blkc
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		expectedRootHash := holders.NewDefaultRootHashesHolder([]byte("rootHash"))
		rootHashHolder, err := bp.ExtractRootHashForCleanup(&block.Header{})
		require.Nil(t, err)
		require.Equal(t, expectedRootHash, rootHashHolder)
	})
}

func TestBaseProcessor_saveProposedTxsToStorage(t *testing.T) {
	t.Parallel()

	t.Run("should call tx coordinator if not header/metaBlock v3", func(t *testing.T) {
		t.Parallel()

		saveTxsToStorageCalled := 0
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
			SaveTxsToStorageCalled: func(body *block.Body) {
				saveTxsToStorageCalled++
			},
		}
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		err = bp.SaveProposedTxsToStorage(&block.HeaderV2{}, &block.Body{})
		require.Nil(t, err)
		require.Equal(t, 1, saveTxsToStorageCalled)

		err = bp.SaveProposedTxsToStorage(&block.MetaBlock{}, &block.Body{})
		require.Nil(t, err)
		require.Equal(t, 2, saveTxsToStorageCalled)

		err = bp.SaveProposedTxsToStorage(&block.HeaderV3{}, &block.Body{})
		require.Nil(t, err)
		require.Equal(t, 2, saveTxsToStorageCalled)

		err = bp.SaveProposedTxsToStorage(&block.MetaBlockV3{}, &block.Body{})
		require.Nil(t, err)
		require.Equal(t, 2, saveTxsToStorageCalled)
	})

	t.Run("headerV3 should save txs from cache to storage", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		keys := [][]byte{[]byte("tx1"), []byte("tx2"), []byte("tx3"), []byte("tx4"), []byte("tx5")}
		txs := map[string]data.TransactionHandler{
			string(keys[0]): &transaction.Transaction{},
			string(keys[1]): &transaction.Transaction{},
			string(keys[2]): &transaction.Transaction{},
			string(keys[3]): &rewardTx.RewardTx{},
			string(keys[4]): &smartContractResult.SmartContractResult{},
		}
		marshalledTxs := make(map[string][]byte)
		for k, v := range txs {
			txsBytes, err := coreComponents.IntMarsh.Marshal(v)
			require.NoError(t, err)
			marshalledTxs[k] = txsBytes
		}

		dataPools := dataComponents.DataPool
		dataPools.Transactions().AddData(keys[0], txs[string(keys[0])], 100, "0")
		dataPools.Transactions().AddData(keys[1], txs[string(keys[1])], 100, "0")
		dataPools.Transactions().AddData(keys[2], txs[string(keys[2])], 100, "0")
		dataPools.RewardTransactions().AddData(keys[3], txs[string(keys[3])], 100, "0")
		dataPools.UnsignedTransactions().AddData(keys[4], txs[string(keys[4])], 100, "0")
		storer := dataComponents.Storage

		blockBody := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{keys[0], keys[1]},
					Type:     block.TxBlock,
				},
				{
					TxHashes: [][]byte{keys[2]},
					Type:     block.InvalidBlock,
				},
				{
					TxHashes: [][]byte{keys[3]},
					Type:     block.RewardsBlock,
				},
				{
					TxHashes: [][]byte{keys[4]},
					Type:     block.SmartContractResultBlock,
				},
			},
		}
		header := &block.HeaderV3{}

		err = bp.SaveProposedTxsToStorage(header, blockBody)
		require.Nil(t, err)

		val, err := storer.Get(dataRetriever.TransactionUnit, keys[0])
		require.NoError(t, err)
		require.Equal(t, marshalledTxs[string(keys[0])], val)

		val, err = storer.Get(dataRetriever.TransactionUnit, keys[1])
		require.NoError(t, err)
		require.Equal(t, marshalledTxs[string(keys[1])], val)

		val, err = storer.Get(dataRetriever.TransactionUnit, keys[2])
		require.NoError(t, err)
		require.Equal(t, marshalledTxs[string(keys[2])], val)

		val, err = storer.Get(dataRetriever.RewardTransactionUnit, keys[3])
		require.NoError(t, err)
		require.Equal(t, marshalledTxs[string(keys[3])], val)

		val, err = storer.Get(dataRetriever.UnsignedTransactionUnit, keys[4])
		require.NoError(t, err)
		require.Equal(t, marshalledTxs[string(keys[4])], val)
	})

	t.Run("headerV3 should save peer info from cache to storage", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		keys := [][]byte{[]byte("peer1"), []byte("peer2")}
		validatorInfos := map[string]*state.ShardValidatorInfo{
			string(keys[0]): {
				PublicKey: []byte("pubKey1"),
				ShardId:   0,
			},
			string(keys[1]): {
				PublicKey: []byte("pubKey2"),
				ShardId:   1,
			},
		}
		marshalledValidatorInfos := make(map[string][]byte)
		for k, v := range validatorInfos {
			validatorInfoBytes, err := coreComponents.IntMarsh.Marshal(v)
			require.NoError(t, err)
			marshalledValidatorInfos[k] = validatorInfoBytes
		}

		dataPools := dataComponents.DataPool
		dataPools.ValidatorsInfo().AddData(keys[0], validatorInfos[string(keys[0])], 100, "0")
		dataPools.ValidatorsInfo().AddData(keys[1], validatorInfos[string(keys[1])], 100, "0")
		storer := dataComponents.Storage

		blockBody := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{keys[0], keys[1]},
					Type:     block.PeerBlock,
				},
			},
		}
		header := &block.HeaderV3{}

		err = bp.SaveProposedTxsToStorage(header, blockBody)
		require.Nil(t, err)

		val, err := storer.Get(dataRetriever.UnsignedTransactionUnit, keys[0])
		require.NoError(t, err)
		require.Equal(t, marshalledValidatorInfos[string(keys[0])], val)

		val, err = storer.Get(dataRetriever.UnsignedTransactionUnit, keys[1])
		require.NoError(t, err)
		require.Equal(t, marshalledValidatorInfos[string(keys[1])], val)
	})

	t.Run("headerV3 should return error if peer info not found in cache", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		key := []byte("peer1")

		blockBody := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{key},
					Type:     block.PeerBlock,
				},
			},
		}
		header := &block.HeaderV3{}

		err = bp.SaveProposedTxsToStorage(header, blockBody)
		require.Equal(t, dataRetriever.ErrValidatorInfoNotFound, err)
	})

	t.Run("headerV3 should return error if invalid type in cache for peer block", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		key := []byte("peer1")
		// Add wrong type to the validator info pool
		dataPools := dataComponents.DataPool
		dataPools.ValidatorsInfo().AddData(key, &transaction.Transaction{}, 100, "0")

		blockBody := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{key},
					Type:     block.PeerBlock,
				},
			},
		}
		header := &block.HeaderV3{}

		err = bp.SaveProposedTxsToStorage(header, blockBody)
		require.ErrorIs(t, err, process.ErrInvalidTxInPool)
	})

	t.Run("headerV3 should return error if marshal fails for peer block", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		coreComponents.IntMarsh = &marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				// Fail marshal only for ShardValidatorInfo
				if _, ok := obj.(*state.ShardValidatorInfo); ok {
					return nil, expectedErr
				}
				return []byte("marshalled"), nil
			},
		}
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		key := []byte("peer1")
		validatorInfo := &state.ShardValidatorInfo{
			PublicKey: []byte("pubKey1"),
			ShardId:   0,
		}

		dataPools := dataComponents.DataPool
		dataPools.ValidatorsInfo().AddData(key, validatorInfo, 100, "0")

		blockBody := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{key},
					Type:     block.PeerBlock,
				},
			},
		}
		header := &block.HeaderV3{}

		err = bp.SaveProposedTxsToStorage(header, blockBody)
		require.Equal(t, expectedErr, err)
	})

	t.Run("headerV3 should return error if storer.Put fails for peer block", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		// Create a storer that fails on Put
		dataComponents.Storage = &storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return &storageStubs.StorerStub{
					PutCalled: func(key, data []byte) error {
						return expectedErr
					},
				}, nil
			},
		}

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		key := []byte("peer1")
		validatorInfo := &state.ShardValidatorInfo{
			PublicKey: []byte("pubKey1"),
			ShardId:   0,
		}

		dataPools := dataComponents.DataPool
		dataPools.ValidatorsInfo().AddData(key, validatorInfo, 100, "0")

		blockBody := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{key},
					Type:     block.PeerBlock,
				},
			},
		}
		header := &block.HeaderV3{}

		err = bp.SaveProposedTxsToStorage(header, blockBody)
		require.Equal(t, expectedErr, err)
	})
}

func TestBaseProcessor_checkContextBeforeExecution(t *testing.T) {
	t.Parallel()

	t.Run("should return error from getting the accountsDB root hash", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		dataComponents.BlockChain = &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() data.HeaderHandler {
				return &block.MetaBlockV3{
					Nonce: 1,
				}
			},
			GetLastExecutedBlockInfoCalled: func() (uint64, []byte, []byte) {
				return 1, []byte("hash1"), []byte("rootHash1")
			},
		}

		accounts := &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return nil, expectedErr
			},
		}

		arguments.AccountsDB[state.UserAccountsState] = accounts
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		err = bp.CheckContextBeforeExecution(&block.HeaderV3{
			Nonce:    2,
			PrevHash: []byte("hash1"),
		}, []byte("headerHash"))
		require.Equal(t, expectedErr, err)
	})

	t.Run("should return ErrBlockHashDoesNotMatch error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
		err := blkc.SetGenesisHeader(&block.Header{})
		require.NoError(t, err)

		blkc.SetLastExecutedBlockHeaderAndRootHash(
			&block.HeaderV3{},
			[]byte("hashX"),
			[]byte("rootHash"),
		)

		dataComponents.BlockChain = blkc
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		accounts := &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return []byte("rootHash"), nil
			},
		}
		arguments.AccountsDB[state.UserAccountsState] = accounts

		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		err = bp.CheckContextBeforeExecution(&block.HeaderV3{
			PrevHash: []byte("hash"),
		}, []byte("headerHash"))
		require.Equal(t, process.ErrBlockHashDoesNotMatch, err)
	})

	t.Run("should return ErrWrongNonceInBlock error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
		err := blkc.SetGenesisHeader(&block.Header{})
		require.NoError(t, err)

		blkc.SetLastExecutedBlockHeaderAndRootHash(
			&block.HeaderV3{
				Nonce: 0,
			},
			[]byte("hash"),
			[]byte("rootHash"),
		)

		dataComponents.BlockChain = blkc
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		accounts := &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return []byte("rootHash"), nil
			},
		}
		arguments.AccountsDB[state.UserAccountsState] = accounts

		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		err = bp.CheckContextBeforeExecution(&block.HeaderV3{
			PrevHash: []byte("hash"),
			Nonce:    2,
		}, []byte("headerHash"))
		require.Equal(t, process.ErrWrongNonceInBlock, err)
	})

	t.Run("should return ErrRootStateDoesNotMatch error", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
		err := blkc.SetGenesisHeader(&block.Header{})
		require.NoError(t, err)

		blkc.SetLastExecutedBlockHeaderAndRootHash(
			&block.HeaderV3{
				Nonce: 1,
			},
			[]byte("hash"),
			[]byte("rootHash"),
		)

		dataComponents.BlockChain = blkc
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		accounts := &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return []byte("rootHashX"), nil
			},
			RecreateTrieCalled: func(options common.RootHashHolder) error {
				return nil
			},
		}
		arguments.AccountsDB[state.UserAccountsState] = accounts

		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		err = bp.CheckContextBeforeExecution(&block.HeaderV3{
			PrevHash: []byte("hash"),
			Nonce:    2,
		}, []byte("headerHash"))
		require.Equal(t, process.ErrRootStateDoesNotMatch, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		blkc, _ := blockchain.NewBlockChain(&statusHandlerMock.AppStatusHandlerStub{})
		err := blkc.SetGenesisHeader(&block.Header{})
		require.NoError(t, err)

		blkc.SetLastExecutedBlockHeaderAndRootHash(
			&block.HeaderV3{
				Nonce: 1,
			},
			[]byte("hash"),
			[]byte("rootHash"),
		)

		dataComponents.BlockChain = blkc
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		accounts := &stateMock.AccountsStub{
			RootHashCalled: func() ([]byte, error) {
				return []byte("rootHash"), nil
			},
		}
		arguments.AccountsDB[state.UserAccountsState] = accounts

		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		err = bp.CheckContextBeforeExecution(&block.HeaderV3{
			PrevHash: []byte("hash"),
			Nonce:    2,
		}, []byte("headerHash"))
		require.Nil(t, err)
	})
}

func TestBaseProcess_collectMiniBlocks(t *testing.T) {
	t.Parallel()

	t.Run("if creating receipts hash fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		txCoordinatorMock := createTxCoordinatorMock()
		txCoordinatorMock.CreateReceiptsHashCalled = func() ([]byte, error) {
			return nil, expectedErr
		}

		arguments.TxCoordinator = &txCoordinatorMock
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		_, _, _, err = bp.CollectMiniBlocks([]byte("hash"), &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("should remove self receipt mini blocks", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		txCoordinatorMock := testscommon.TransactionCoordinatorMock{}
		txCoordinatorMock.CreatePostProcessMiniBlocksCalled = func() block.MiniBlockSlice {
			return block.MiniBlockSlice{
				{
					TxHashes: [][]byte{
						[]byte("txHash1"),
						[]byte("txHash2"),
					},
					Type:            block.ReceiptBlock, // this mini block should be removed
					ReceiverShardID: uint32(0),
					SenderShardID:   uint32(0),
				},
				{
					TxHashes: [][]byte{
						[]byte("txHash3"),
						[]byte("txHash4"),
					},
				},
			}
		}

		arguments.TxCoordinator = &txCoordinatorMock
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		miniBlockHeaderHandlers, totalTxCount, receiptHash, err := bp.CollectMiniBlocks([]byte("hash"), &block.Body{})
		require.NoError(t, err)
		require.Equal(t, 1, len(miniBlockHeaderHandlers))
		require.Equal(t, 2, totalTxCount)
		require.Equal(t, []byte("receiptHash"), receiptHash)
	})

	t.Run("if hashing fails, the error should be propagated", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		err := coreComponents.SetInternalMarshalizer(&marshallerMock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		})
		require.Nil(t, err)
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		txCoordinatorMock := testscommon.TransactionCoordinatorMock{}
		txCoordinatorMock.CreatePostProcessMiniBlocksCalled = func() block.MiniBlockSlice {
			return block.MiniBlockSlice{
				{
					TxHashes: [][]byte{
						[]byte("txHash1"),
						[]byte("txHash2"),
					},
					Type:            block.ReceiptBlock, // this mini block should be removed
					ReceiverShardID: uint32(0),
					SenderShardID:   uint32(0),
				},
				{
					TxHashes: [][]byte{
						[]byte("txHash3"),
						[]byte("txHash4"),
					},
				},
			}
		}

		arguments.TxCoordinator = &txCoordinatorMock
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		_, _, _, err = bp.CollectMiniBlocks([]byte("hash"), &block.Body{})
		require.Equal(t, expectedErr, err)
	})

	t.Run("sanitized mini blocks should be cached", func(t *testing.T) {
		t.Parallel()

		expectedValue := 0

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataPool := initDataPool()
		dataPool.ExecutedMiniBlocksCalled = func() storage.Cacher {
			return &cache.CacherStub{
				PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
					expectedValue++
					return false
				},
			}
		}
		dataComponents.DataPool = dataPool

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		txCoordinatorMock := testscommon.TransactionCoordinatorMock{}
		txCoordinatorMock.CreatePostProcessMiniBlocksCalled = func() block.MiniBlockSlice {
			return block.MiniBlockSlice{
				{
					TxHashes: [][]byte{
						[]byte("txHash1"),
						[]byte("txHash2"),
					},
					Type:            block.ReceiptBlock, // this mini block should be removed
					ReceiverShardID: uint32(0),
					SenderShardID:   uint32(0),
				},
				{
					TxHashes: [][]byte{
						[]byte("txHash3"),
						[]byte("txHash4"),
					},
				},
			}
		}

		arguments.TxCoordinator = &txCoordinatorMock
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		miniBlockHeaderHandlers, totalTxCount, receiptHash, err := bp.CollectMiniBlocks([]byte("hash"), &block.Body{})
		require.NoError(t, err)
		require.Equal(t, 1, len(miniBlockHeaderHandlers))
		require.Equal(t, 2, totalTxCount)
		require.Equal(t, []byte("receiptHash"), receiptHash)
		require.Equal(t, 2, expectedValue)
	})

	t.Run("existing intra shard mini blocks should be cached", func(t *testing.T) {
		t.Parallel()

		expectedValue := 0
		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		dataPool := initDataPool()
		dataPool.ExecutedMiniBlocksCalled = func() storage.Cacher {
			return &cache.CacherStub{
				PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
					expectedValue++
					return false
				},
			}
		}
		dataComponents.DataPool = dataPool

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		txCoordinatorMock := testscommon.TransactionCoordinatorMock{}
		txCoordinatorMock.GetCreatedInShardMiniBlocksCalled = func() []*block.MiniBlock {
			return []*block.MiniBlock{
				{
					TxHashes: [][]byte{
						[]byte("txHash1"),
					},
				},
			}
		}

		arguments.TxCoordinator = &txCoordinatorMock
		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		headerHash := []byte("headerHash")
		miniBlockHeaderHandlers, totalTxCount, receiptHash, err := bp.CollectMiniBlocks(headerHash, &block.Body{})
		require.NoError(t, err)
		require.Equal(t, 0, len(miniBlockHeaderHandlers))
		require.Equal(t, 0, totalTxCount)
		require.Equal(t, []byte("receiptHash"), receiptHash)

		require.Equal(t, 1, expectedValue)
	})
}

func TestBaseProcessor_CacheIntraShardMiniBlocks(t *testing.T) {
	t.Parallel()

	t.Run("should work with proto marshaller", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		_ = coreComponents.SetInternalMarshalizer(&marshal.GogoProtoMarshalizer{})

		executedMBs := cache.NewCacherStub()

		wasCalled := false
		executedMBs.PutCalled = func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
			wasCalled = true
			return
		}

		dataPool := initDataPool()
		dataPool.ExecutedMiniBlocksCalled = func() storage.Cacher {
			return executedMBs
		}
		dataComponents.DataPool = dataPool

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		bp, err := blproc.NewShardProcessor(arguments)
		require.NoError(t, err)

		headerHash := []byte("headerHash")
		miniBlocks := []*block.MiniBlock{}

		err = bp.CacheIntraShardMiniBlocks(headerHash, miniBlocks)
		require.Nil(t, err)

		require.True(t, wasCalled)
	})
}

func TestBaseProcessor_excludeRevertedExecutionResultsForHeader(t *testing.T) {
	t.Parallel()

	t.Run("should work in case of no pending execution results", func(t *testing.T) {
		t.Parallel()

		bp, err := blproc.ConstructPartialShardBlockProcessorForTest(map[string]interface{}{})
		require.NoError(t, err)

		header := &block.HeaderV3{}
		sanitizedPendingExecResults := bp.ExcludeRevertedExecutionResultsForHeader(
			header,
			[]data.BaseExecutionResultHandler{},
		)
		require.Equal(t, 0, len(sanitizedPendingExecResults))
	})

	t.Run("should remove last execution result if its header nonce is equal to the nonce of the created block", func(t *testing.T) {
		t.Parallel()

		bp, err := blproc.ConstructPartialShardBlockProcessorForTest(map[string]interface{}{})
		require.NoError(t, err)

		headerNonce := uint64(1)
		header := &block.HeaderV3{
			Nonce: headerNonce,
		}

		pendingExecutionResults := []data.BaseExecutionResultHandler{
			&block.BaseExecutionResult{
				HeaderNonce: headerNonce,
			},
		}

		sanitizedPendingExecResults := bp.ExcludeRevertedExecutionResultsForHeader(
			header,
			pendingExecutionResults,
		)
		require.Equal(t, 0, len(sanitizedPendingExecResults))
	})

	t.Run("should remove last execution result if if getting the execution result's header from storage fails", func(t *testing.T) {
		t.Parallel()

		bp, err := blproc.ConstructPartialShardBlockProcessorForTest(map[string]interface{}{
			"store": &storageStubs.ChainStorerStub{
				GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
					return nil, expectedErr
				},
			},
			"marshalizer": &testscommon.MarshallerStub{},
		})
		require.NoError(t, err)

		header := &block.HeaderV3{
			Nonce: 2,
		}

		pendingExecutionResults := []data.BaseExecutionResultHandler{
			&block.BaseExecutionResult{
				HeaderHash:  []byte("wrongHash"),
				HeaderNonce: 1,
			},
		}

		sanitizedPendingExecResults := bp.ExcludeRevertedExecutionResultsForHeader(
			header,
			pendingExecutionResults,
		)
		require.Equal(t, 0, len(sanitizedPendingExecResults))
	})

	t.Run("should not remove valid execution results", func(t *testing.T) {
		t.Parallel()

		bp, err := blproc.ConstructPartialShardBlockProcessorForTest(map[string]interface{}{
			"store": &storageStubs.ChainStorerStub{
				GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
					return &storageStubs.StorerStub{
						GetCalled: func(key []byte) ([]byte, error) {
							return nil, nil
						},
					}, nil
				},
			},
			"marshalizer": &testscommon.MarshallerStub{
				UnmarshalCalled: func(obj interface{}, buff []byte) error {
					return nil
				},
			},
		})
		require.NoError(t, err)

		header := &block.HeaderV3{
			Nonce: 2,
		}

		pendingExecutionResults := []data.BaseExecutionResultHandler{
			&block.BaseExecutionResult{
				HeaderHash:  []byte("hash"),
				HeaderNonce: 1,
			},
		}

		sanitizedPendingExecResults := bp.ExcludeRevertedExecutionResultsForHeader(
			header,
			pendingExecutionResults,
		)
		require.Equal(t, pendingExecutionResults, sanitizedPendingExecResults)
	})
}

func TestBaseProcessor_ProposedDirectSentTransactionsToBroadcast(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	wasProposedDirectSentTransactionsToBroadcastCalled := false
	arguments.TxCoordinator = &testscommon.TransactionCoordinatorMock{
		ProposedDirectSentTransactionsToBroadcastCalled: func(proposedBody data.BodyHandler) map[string][][]byte {
			wasProposedDirectSentTransactionsToBroadcastCalled = true
			return nil
		},
	}
	bp, _ := blproc.NewShardProcessor(arguments)

	_ = bp.ProposedDirectSentTransactionsToBroadcast(nil)
	require.True(t, wasProposedDirectSentTransactionsToBroadcastCalled)
}

func TestBaseProcessor_Close(t *testing.T) {
	t.Parallel()

	coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
	arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
	bp, _ := blproc.NewShardProcessor(arguments)

	require.NoError(t, bp.Close())
}

func TestBaseProcessor_WaitForExecutionResultsVerification(t *testing.T) {
	t.Parallel()

	t.Run("should return nil when verification succeeds on first call", func(t *testing.T) {
		t.Parallel()

		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.ExecutionResultsVerifier = &processMocks.ExecutionResultsVerifierMock{
			VerifyHeaderExecutionResultsCalled: func(header data.HeaderHandler) error {
				return nil
			},
		}
		bp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		header := &block.HeaderV3{Nonce: 1}
		err = bp.WaitForExecutionResultsVerification(header, func() time.Duration { return time.Second })
		require.Nil(t, err)
	})

	t.Run("should return non-retryable error immediately", func(t *testing.T) {
		t.Parallel()

		callCount := atomic.Int32{}
		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.ExecutionResultsVerifier = &processMocks.ExecutionResultsVerifierMock{
			VerifyHeaderExecutionResultsCalled: func(header data.HeaderHandler) error {
				callCount.Add(1)
				return process.ErrExecutionResultDoesNotMatch
			},
		}
		bp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		header := &block.HeaderV3{Nonce: 1}
		err = bp.WaitForExecutionResultsVerification(header, func() time.Duration { return time.Second })
		require.ErrorIs(t, err, process.ErrExecutionResultDoesNotMatch)
		require.Equal(t, int32(1), callCount.Load())
	})

	t.Run("should retry on mismatch then succeed", func(t *testing.T) {
		t.Parallel()

		callCount := atomic.Int32{}
		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.ExecutionResultsVerifier = &processMocks.ExecutionResultsVerifierMock{
			VerifyHeaderExecutionResultsCalled: func(header data.HeaderHandler) error {
				count := callCount.Add(1)
				if count < 3 {
					return process.ErrExecutionResultsNumberMismatch
				}
				return nil
			},
		}
		bp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		header := &block.HeaderV3{Nonce: 1}
		err = bp.WaitForExecutionResultsVerification(header, func() time.Duration { return time.Second })
		require.Nil(t, err)
		require.Equal(t, int32(3), callCount.Load())
	})

	t.Run("should timeout and return mismatch error when haveTime expires", func(t *testing.T) {
		t.Parallel()

		callCount := atomic.Int32{}
		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.ExecutionResultsVerifier = &processMocks.ExecutionResultsVerifierMock{
			VerifyHeaderExecutionResultsCalled: func(header data.HeaderHandler) error {
				callCount.Add(1)
				return process.ErrExecutionResultsNumberMismatch
			},
		}
		bp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		header := &block.HeaderV3{Nonce: 1}
		deadline := time.Now().Add(25 * time.Millisecond)
		err = bp.WaitForExecutionResultsVerification(header, func() time.Duration { return time.Until(deadline) })
		require.ErrorIs(t, err, process.ErrExecutionResultsNumberMismatch)
		require.Greater(t, callCount.Load(), int32(1))
	})

	t.Run("should return mismatch error immediately when haveTime returns zero", func(t *testing.T) {
		t.Parallel()

		callCount := atomic.Int32{}
		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.ExecutionResultsVerifier = &processMocks.ExecutionResultsVerifierMock{
			VerifyHeaderExecutionResultsCalled: func(header data.HeaderHandler) error {
				callCount.Add(1)
				return process.ErrExecutionResultsNumberMismatch
			},
		}
		bp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		header := &block.HeaderV3{Nonce: 1}
		err = bp.WaitForExecutionResultsVerification(header, func() time.Duration { return 0 })
		require.ErrorIs(t, err, process.ErrExecutionResultsNumberMismatch)
		require.Equal(t, int32(1), callCount.Load())
	})

	t.Run("should return mismatch error immediately when haveTime returns negative", func(t *testing.T) {
		t.Parallel()

		callCount := atomic.Int32{}
		arguments := CreateMockArguments(createComponentHolderMocks())
		arguments.ExecutionResultsVerifier = &processMocks.ExecutionResultsVerifierMock{
			VerifyHeaderExecutionResultsCalled: func(header data.HeaderHandler) error {
				callCount.Add(1)
				return process.ErrExecutionResultsNumberMismatch
			},
		}
		bp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		header := &block.HeaderV3{Nonce: 1}
		err = bp.WaitForExecutionResultsVerification(header, func() time.Duration { return -time.Second })
		require.ErrorIs(t, err, process.ErrExecutionResultsNumberMismatch)
		require.Equal(t, int32(1), callCount.Load())
	})
}

func TestBaseProcessor_PruneTrieAsyncHeader(t *testing.T) {
	t.Parallel()

	// header 1

	headerHash1 := []byte("headerHash1")
	rootHash10 := []byte("rootHash10")
	rootHash11 := []byte("rootHash11")

	baseExecRes10 := &block.BaseExecutionResult{RootHash: rootHash10}
	baseExecRes11 := &block.BaseExecutionResult{RootHash: rootHash11}
	executionResultsHandlers := []data.BaseExecutionResultHandler{
		&block.ExecutionResult{
			BaseExecutionResult: baseExecRes10,
		},
		&block.ExecutionResult{
			BaseExecutionResult: baseExecRes11,
		},
	}
	header1 := &block.HeaderV3{
		Nonce: 8,
		LastExecutionResult: &block.ExecutionResultInfo{
			ExecutionResult: baseExecRes11,
		},
	}
	_ = header1.SetExecutionResultsHandlers(executionResultsHandlers)

	// header 2

	headerHash2 := []byte("headerHash2")
	rootHash20 := []byte("rootHash20")
	rootHash21 := []byte("rootHash21")
	rootHash22 := []byte("rootHash22")
	rootHash23 := []byte("rootHash23")

	baseExecRes20 := &block.BaseExecutionResult{RootHash: rootHash20}
	baseExecRes21 := &block.BaseExecutionResult{RootHash: rootHash21}
	baseExecRes22 := &block.BaseExecutionResult{RootHash: rootHash22}
	baseExecRes23 := &block.BaseExecutionResult{RootHash: rootHash23}
	executionResultsHandlers2 := []data.BaseExecutionResultHandler{
		&block.ExecutionResult{
			BaseExecutionResult: baseExecRes20,
		},
		&block.ExecutionResult{
			BaseExecutionResult: baseExecRes21,
		},
		&block.ExecutionResult{
			BaseExecutionResult: baseExecRes22,
		},
		&block.ExecutionResult{
			BaseExecutionResult: baseExecRes23,
		},
	}

	header2 := &block.HeaderV3{
		Nonce:    9,
		PrevHash: headerHash1,
		LastExecutionResult: &block.ExecutionResultInfo{
			ExecutionResult: baseExecRes23,
		},
	}
	_ = header2.SetExecutionResultsHandlers(executionResultsHandlers2)

	// header 3

	headerHash3 := []byte("headerHash3")
	rootHash30 := []byte("rootHash30")
	rootHash31 := []byte("rootHash31")

	baseExecRes30 := &block.BaseExecutionResult{RootHash: rootHash30}
	baseExecRes31 := &block.BaseExecutionResult{RootHash: rootHash31}
	executionResultsHandlers3 := []data.BaseExecutionResultHandler{
		&block.ExecutionResult{
			BaseExecutionResult: baseExecRes30,
		},
		&block.ExecutionResult{
			BaseExecutionResult: baseExecRes31,
		},
	}

	header3 := &block.HeaderV3{
		Nonce:    10,
		PrevHash: headerHash2,
		LastExecutionResult: &block.ExecutionResultInfo{
			ExecutionResult: baseExecRes31,
		},
	}
	_ = header3.SetExecutionResultsHandlers(executionResultsHandlers3)

	// header 4

	headerHash4 := []byte("headerHash4")
	header4 := &block.HeaderV3{
		Nonce:    11,
		PrevHash: headerHash3,
		LastExecutionResult: &block.ExecutionResultInfo{
			ExecutionResult: baseExecRes31,
		},
	}

	// header 5

	headerHash5 := []byte("headerHash5")
	header5 := &block.HeaderV3{
		Nonce:    12,
		PrevHash: headerHash4,
		LastExecutionResult: &block.ExecutionResultInfo{
			ExecutionResult: baseExecRes31,
		},
	}

	// header 6

	headerHash6 := []byte("headerHash6")
	rootHash60 := []byte("rootHash60")
	rootHash61 := []byte("rootHash61")

	baseExecRes60 := &block.BaseExecutionResult{RootHash: rootHash60}
	baseExecRes61 := &block.BaseExecutionResult{RootHash: rootHash61}
	executionResultsHandlers6 := []data.BaseExecutionResultHandler{
		&block.ExecutionResult{
			BaseExecutionResult: baseExecRes60,
		},
		&block.ExecutionResult{
			BaseExecutionResult: baseExecRes61,
		},
	}

	header6 := &block.HeaderV3{
		Nonce:    13,
		PrevHash: headerHash5,
		LastExecutionResult: &block.ExecutionResultInfo{
			ExecutionResult: baseExecRes61,
		},
	}
	_ = header6.SetExecutionResultsHandlers(executionResultsHandlers6)

	// header 7

	headerHash7 := []byte("headerHash7")
	rootHash70 := []byte("rootHash70")

	baseExecRes70 := &block.BaseExecutionResult{RootHash: rootHash70}
	baseExecRes71 := &block.BaseExecutionResult{RootHash: rootHash70} // no roothash change
	executionResultsHandlers7 := []data.BaseExecutionResultHandler{
		&block.ExecutionResult{
			BaseExecutionResult: baseExecRes70,
		},
		&block.ExecutionResult{
			BaseExecutionResult: baseExecRes71,
		},
	}

	header7 := &block.HeaderV3{
		Nonce:    14,
		PrevHash: headerHash6,
		LastExecutionResult: &block.ExecutionResultInfo{
			ExecutionResult: baseExecRes71,
		},
	}
	_ = header7.SetExecutionResultsHandlers(executionResultsHandlers7)

	// header 8

	headerHash8 := []byte("headerHash8")

	baseExecRes80 := &block.BaseExecutionResult{RootHash: rootHash70} // not roothash change
	header8 := &block.HeaderV3{
		Nonce:    15,
		PrevHash: headerHash7,
		LastExecutionResult: &block.ExecutionResultInfo{
			ExecutionResult: baseExecRes80,
		},
	}

	// header 9

	headerHash9 := []byte("headerHash9")
	rootHash91 := []byte("rootHash91")

	baseExecRes90 := &block.BaseExecutionResult{RootHash: rootHash70} // no roothash change
	baseExecRes91 := &block.BaseExecutionResult{RootHash: rootHash91} // roothash changed
	executionResultsHandlers9 := []data.BaseExecutionResultHandler{
		&block.ExecutionResult{
			BaseExecutionResult: baseExecRes90,
		},
		&block.ExecutionResult{
			BaseExecutionResult: baseExecRes91,
		},
	}

	header9 := &block.HeaderV3{
		Nonce:    16,
		PrevHash: headerHash8,
		LastExecutionResult: &block.ExecutionResultInfo{
			ExecutionResult: baseExecRes91,
		},
	}
	_ = header9.SetExecutionResultsHandlers(executionResultsHandlers9)

	t.Run("last pruned header not set, should trigger provided header", func(t *testing.T) {
		t.Parallel()

		cancelPruneCalled := false
		pruneTrieCalled := false

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			IsPruningEnabledCalled: func() bool {
				return true
			},
			CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
				cancelPruneCalled = true
			},
			PruneTrieCalled: func(rootHash []byte, identifier state.TriePruningIdentifier, handler state.PruningHandler) {
				pruneTrieCalled = true
			},
		}

		headersPool := &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, headerHash1) {
					return header1, nil
				}

				return nil, errors.New("header not found")
			},
		}
		dataPool := initDataPool()
		dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
			return headersPool
		}
		dataComponents.DataPool = dataPool

		arguments.ForkDetector = &mock.ForkDetectorMock{
			GetHighestSettledBlockInfoCalled: func() (uint64, []byte) {
				return header1.GetNonce(), headerHash1
			},
		}

		bp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		require.Nil(t, bp.GetLastPrunedHash())

		_ = header1.SetExecutionResultsHandlers(executionResultsHandlers)
		bp.PruneTrieAsyncHeader()

		require.True(t, cancelPruneCalled)
		require.True(t, pruneTrieCalled)

		require.Equal(t, headerHash1, bp.GetLastPrunedHash())
	})

	t.Run("header nonce lower than last pruned header, should not trigger", func(t *testing.T) {
		t.Parallel()

		cancelPruneCalled := false
		pruneTrieCalled := false

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			IsPruningEnabledCalled: func() bool {
				return true
			},
			CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
				cancelPruneCalled = true
			},
			PruneTrieCalled: func(rootHash []byte, identifier state.TriePruningIdentifier, handler state.PruningHandler) {
				pruneTrieCalled = true
			},
		}

		arguments.ForkDetector = &mock.ForkDetectorMock{
			GetHighestSettledBlockInfoCalled: func() (uint64, []byte) {
				return header2.GetNonce(), headerHash2
			},
		}

		bp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		bp.SetLastPrunedNonce(10)
		bp.SetLastPrunedHash(headerHash3)

		bp.PruneTrieAsyncHeader()
		require.False(t, cancelPruneCalled)
		require.False(t, pruneTrieCalled)

		require.Equal(t, headerHash3, bp.GetLastPrunedHash())
	})

	t.Run("intermediate headers with included execution results", func(t *testing.T) {
		t.Parallel()

		cancelPruneRootHashes := make([][]byte, 0)
		pruneTrieRootHashes := make([][]byte, 0)

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			IsPruningEnabledCalled: func() bool {
				return true
			},
			CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
				cancelPruneRootHashes = append(cancelPruneRootHashes, rootHash)
			},
			PruneTrieCalled: func(rootHash []byte, identifier state.TriePruningIdentifier, handler state.PruningHandler) {
				pruneTrieRootHashes = append(pruneTrieRootHashes, rootHash)
			},
		}

		headersPool := &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, headerHash1) {
					return header1, nil
				}
				if bytes.Equal(hash, headerHash2) {
					return header2, nil
				}
				if bytes.Equal(hash, headerHash3) {
					return header3, nil
				}

				return nil, errors.New("header not found")
			},
		}
		dataPool := initDataPool()
		dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
			return headersPool
		}
		dataComponents.DataPool = dataPool

		arguments.ForkDetector = &mock.ForkDetectorMock{
			GetHighestSettledBlockInfoCalled: func() (uint64, []byte) {
				return header3.GetNonce(), headerHash3
			},
		}

		bp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		bp.SetLastPrunedHash(headerHash1)

		bp.PruneTrieAsyncHeader()

		require.Equal(t, 6, len(cancelPruneRootHashes))
		require.Equal(t, rootHash11, cancelPruneRootHashes[0])
		require.Equal(t, rootHash20, cancelPruneRootHashes[1])
		require.Equal(t, rootHash21, cancelPruneRootHashes[2])
		require.Equal(t, rootHash22, cancelPruneRootHashes[3])
		require.Equal(t, rootHash23, cancelPruneRootHashes[4])
		require.Equal(t, rootHash30, cancelPruneRootHashes[5])

		require.Equal(t, 6, len(pruneTrieRootHashes))
		require.Equal(t, rootHash11, pruneTrieRootHashes[0])
		require.Equal(t, rootHash20, pruneTrieRootHashes[1])
		require.Equal(t, rootHash21, pruneTrieRootHashes[2])
		require.Equal(t, rootHash22, pruneTrieRootHashes[3])
		require.Equal(t, rootHash23, pruneTrieRootHashes[4])
		require.Equal(t, rootHash30, pruneTrieRootHashes[5])

		require.Equal(t, headerHash3, bp.GetLastPrunedHash())

		// another call for the same current header should not trigger prune
		bp.PruneTrieAsyncHeader()

		require.Equal(t, 6, len(cancelPruneRootHashes))
		require.Equal(t, 6, len(pruneTrieRootHashes))
		require.Equal(t, headerHash3, bp.GetLastPrunedHash())
	})

	t.Run("intermediate headers without included execution results", func(t *testing.T) {
		t.Parallel()

		cancelPruneRootHashes := make([][]byte, 0)
		pruneTrieRootHashes := make([][]byte, 0)

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			IsPruningEnabledCalled: func() bool {
				return true
			},
			CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
				cancelPruneRootHashes = append(cancelPruneRootHashes, rootHash)
			},
			PruneTrieCalled: func(rootHash []byte, identifier state.TriePruningIdentifier, handler state.PruningHandler) {
				pruneTrieRootHashes = append(pruneTrieRootHashes, rootHash)
			},
		}

		headersPool := &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, headerHash3) {
					return header3, nil
				}
				if bytes.Equal(hash, headerHash4) {
					return header4, nil
				}
				if bytes.Equal(hash, headerHash5) {
					return header5, nil
				}
				if bytes.Equal(hash, headerHash6) {
					return header6, nil
				}

				return nil, errors.New("header not found")
			},
		}
		dataPool := initDataPool()
		dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
			return headersPool
		}
		dataComponents.DataPool = dataPool

		arguments.ForkDetector = &mock.ForkDetectorMock{
			GetHighestSettledBlockInfoCalled: func() (uint64, []byte) {
				return header6.GetNonce(), headerHash6
			},
		}

		bp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		bp.SetLastPrunedHash(headerHash3)

		bp.PruneTrieAsyncHeader()

		require.Equal(t, 2, len(cancelPruneRootHashes))
		require.Equal(t, rootHash31, cancelPruneRootHashes[0])
		require.Equal(t, rootHash60, cancelPruneRootHashes[1])

		require.Equal(t, 2, len(pruneTrieRootHashes))
		require.Equal(t, rootHash31, pruneTrieRootHashes[0])
		require.Equal(t, rootHash60, pruneTrieRootHashes[1])

		require.Equal(t, headerHash6, bp.GetLastPrunedHash())

		// another call for the same current header should not trigger prune
		bp.PruneTrieAsyncHeader()

		require.Equal(t, 2, len(cancelPruneRootHashes))
		require.Equal(t, 2, len(pruneTrieRootHashes))
		require.Equal(t, headerHash6, bp.GetLastPrunedHash())
	})

	t.Run("intermediate headers with included execution results, no roothash change", func(t *testing.T) {
		t.Parallel()

		cancelPruneRootHashes := make([][]byte, 0)
		pruneTrieRootHashes := make([][]byte, 0)

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)

		arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			IsPruningEnabledCalled: func() bool {
				return true
			},
			CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
				cancelPruneRootHashes = append(cancelPruneRootHashes, rootHash)
			},
			PruneTrieCalled: func(rootHash []byte, identifier state.TriePruningIdentifier, handler state.PruningHandler) {
				pruneTrieRootHashes = append(pruneTrieRootHashes, rootHash)
			},
		}

		headersPool := &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, headerHash6) {
					return header6, nil
				}
				if bytes.Equal(hash, headerHash7) {
					return header7, nil
				}
				if bytes.Equal(hash, headerHash8) {
					return header8, nil
				}
				if bytes.Equal(hash, headerHash9) {
					return header9, nil
				}

				return nil, errors.New("header not found")
			},
		}
		dataPool := initDataPool()
		dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
			return headersPool
		}
		dataComponents.DataPool = dataPool

		arguments.ForkDetector = &mock.ForkDetectorMock{
			GetHighestSettledBlockInfoCalled: func() (uint64, []byte) {
				return header9.GetNonce(), headerHash9
			},
		}

		bp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		bp.SetLastPrunedHash(headerHash5)

		bp.PruneTrieAsyncHeader()

		require.Equal(t, 3, len(cancelPruneRootHashes))
		require.Equal(t, rootHash60, cancelPruneRootHashes[0])
		require.Equal(t, rootHash61, cancelPruneRootHashes[1])
		require.Equal(t, rootHash70, cancelPruneRootHashes[2])

		require.Equal(t, 3, len(pruneTrieRootHashes))
		require.Equal(t, rootHash60, pruneTrieRootHashes[0])
		require.Equal(t, rootHash61, pruneTrieRootHashes[1])
		require.Equal(t, rootHash70, pruneTrieRootHashes[2])

		require.Equal(t, headerHash9, bp.GetLastPrunedHash())

		// another call for the same current header should not trigger prune
		bp.PruneTrieAsyncHeader()

		require.Equal(t, 3, len(cancelPruneRootHashes))
		require.Equal(t, 3, len(pruneTrieRootHashes))
		require.Equal(t, headerHash9, bp.GetLastPrunedHash())
	})

	t.Run("no settled block yet, should not prune", func(t *testing.T) {
		t.Parallel()

		pruneTrieCalled := false

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			IsPruningEnabledCalled: func() bool {
				return true
			},
			PruneTrieCalled: func(rootHash []byte, identifier state.TriePruningIdentifier, handler state.PruningHandler) {
				pruneTrieCalled = true
			},
		}
		arguments.ForkDetector = &mock.ForkDetectorMock{
			GetHighestSettledBlockInfoCalled: func() (uint64, []byte) {
				return 0, nil
			},
		}

		bp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		bp.PruneTrieAsyncHeader()

		require.False(t, pruneTrieCalled)
		require.Nil(t, bp.GetLastPrunedHash())
	})

	t.Run("settled header not resolvable, pruning postponed", func(t *testing.T) {
		t.Parallel()

		pruneTrieCalled := false

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			IsPruningEnabledCalled: func() bool {
				return true
			},
			PruneTrieCalled: func(rootHash []byte, identifier state.TriePruningIdentifier, handler state.PruningHandler) {
				pruneTrieCalled = true
			},
		}

		headersPool := &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				return nil, errors.New("header not found")
			},
		}
		dataPool := initDataPool()
		dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
			return headersPool
		}
		dataComponents.DataPool = dataPool

		arguments.ForkDetector = &mock.ForkDetectorMock{
			GetHighestSettledBlockInfoCalled: func() (uint64, []byte) {
				return header1.GetNonce(), headerHash1
			},
		}

		bp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		bp.PruneTrieAsyncHeader()

		require.False(t, pruneTrieCalled)
		require.Nil(t, bp.GetLastPrunedHash())
	})

	t.Run("settled behind committed tip retains its execution base root", func(t *testing.T) {
		t.Parallel()

		pruneTrieRootHashes := make([][]byte, 0)

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()

		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			IsPruningEnabledCalled: func() bool {
				return true
			},
			PruneTrieCalled: func(rootHash []byte, identifier state.TriePruningIdentifier, handler state.PruningHandler) {
				pruneTrieRootHashes = append(pruneTrieRootHashes, rootHash)
			},
		}

		headersPool := &mock.HeadersCacherStub{
			GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
				if bytes.Equal(hash, headerHash1) {
					return header1, nil
				}
				if bytes.Equal(hash, headerHash2) {
					return header2, nil
				}
				if bytes.Equal(hash, headerHash3) {
					return header3, nil
				}

				return nil, errors.New("header not found")
			},
		}
		dataPool := initDataPool()
		dataPool.HeadersCalled = func() dataRetriever.HeadersPool {
			return headersPool
		}
		dataComponents.DataPool = dataPool

		// the committed tip is header3, the settled checkpoint lags at header2
		settledNonce := header2.GetNonce()
		settledHash := headerHash2
		arguments.ForkDetector = &mock.ForkDetectorMock{
			GetHighestSettledBlockInfoCalled: func() (uint64, []byte) {
				return settledNonce, settledHash
			},
		}

		bp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		bp.SetLastPrunedHash(headerHash1)

		bp.PruneTrieAsyncHeader()

		// header2's last notarized root (rootHash23) and everything newer must survive: a legal
		// rollback to the settled block re-executes from rootHash23
		require.Equal(t, [][]byte{rootHash11, rootHash20, rootHash21, rootHash22}, pruneTrieRootHashes)
		require.Equal(t, headerHash2, bp.GetLastPrunedHash())

		// once settlement advances to the tip, the retained roots become prunable
		settledNonce = header3.GetNonce()
		settledHash = headerHash3

		bp.PruneTrieAsyncHeader()

		require.Equal(t, [][]byte{rootHash11, rootHash20, rootHash21, rootHash22, rootHash23, rootHash30}, pruneTrieRootHashes)
		require.Equal(t, headerHash3, bp.GetLastPrunedHash())
	})
}

func TestComputeEWLResetThreshold(t *testing.T) {
	t.Parallel()

	t.Run("gap 0 should return minimum baseline", func(t *testing.T) {
		t.Parallel()
		// gap=0 -> expected=0*2=0, 0*130/100=0, +10 = 10
		require.Equal(t, 10, blproc.ComputeEWLResetThreshold(0))
	})
	t.Run("gap 1", func(t *testing.T) {
		t.Parallel()
		// gap=1 -> expected=1*2=2, 2*130/100=2, +10 = 12
		require.Equal(t, 12, blproc.ComputeEWLResetThreshold(1))
	})
	t.Run("default gap 10", func(t *testing.T) {
		t.Parallel()
		// gap=10 -> expected=10*2=20, 20*130/100=26, +10 = 36
		require.Equal(t, 36, blproc.ComputeEWLResetThreshold(10))
	})
	t.Run("gap above cap should be clamped", func(t *testing.T) {
		t.Parallel()
		// gap=500 clamped to 250 -> expected=250*2=500, 500*130/100=650, +10 = 660
		require.Equal(t, 660, blproc.ComputeEWLResetThreshold(500))
		require.Equal(t, 660, blproc.ComputeEWLResetThreshold(1000))
	})
	t.Run("gap at cap boundary", func(t *testing.T) {
		t.Parallel()
		require.Equal(t, 660, blproc.ComputeEWLResetThreshold(250))
		require.Equal(t, blproc.ComputeEWLResetThreshold(250), blproc.ComputeEWLResetThreshold(251))
	})
}

func TestCancelPruneForRootHashTransition(t *testing.T) {
	t.Parallel()

	t.Run("different hashes should call CancelPrune for both", func(t *testing.T) {
		t.Parallel()
		cancelPruneCalls := make([]struct {
			rootHash   []byte
			identifier state.TriePruningIdentifier
		}, 0)
		accountsStub := &stateMock.AccountsStub{
			CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
				cancelPruneCalls = append(cancelPruneCalls, struct {
					rootHash   []byte
					identifier state.TriePruningIdentifier
				}{rootHash, identifier})
			},
		}

		blproc.CancelPruneForRootHashTransition(accountsStub, []byte("prev"), []byte("curr"))

		require.Len(t, cancelPruneCalls, 2)
		require.Equal(t, []byte("curr"), cancelPruneCalls[0].rootHash)
		require.Equal(t, state.NewRoot, cancelPruneCalls[0].identifier)
		require.Equal(t, []byte("prev"), cancelPruneCalls[1].rootHash)
		require.Equal(t, state.OldRoot, cancelPruneCalls[1].identifier)
	})
	t.Run("equal hashes should not call CancelPrune", func(t *testing.T) {
		t.Parallel()
		accountsStub := &stateMock.AccountsStub{
			CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
				require.Fail(t, "CancelPrune should not be called for equal hashes")
			},
		}
		blproc.CancelPruneForRootHashTransition(accountsStub, []byte("same"), []byte("same"))
	})
	t.Run("empty prev hash should not call CancelPrune", func(t *testing.T) {
		t.Parallel()
		accountsStub := &stateMock.AccountsStub{
			CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
				require.Fail(t, "CancelPrune should not be called when prev is empty")
			},
		}
		blproc.CancelPruneForRootHashTransition(accountsStub, nil, []byte("curr"))
		blproc.CancelPruneForRootHashTransition(accountsStub, []byte{}, []byte("curr"))
	})
	t.Run("empty current hash should not call CancelPrune", func(t *testing.T) {
		t.Parallel()
		accountsStub := &stateMock.AccountsStub{
			CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
				require.Fail(t, "CancelPrune should not be called when current is empty")
			},
		}
		blproc.CancelPruneForRootHashTransition(accountsStub, []byte("prev"), nil)
		blproc.CancelPruneForRootHashTransition(accountsStub, []byte("prev"), []byte{})
	})
}

func TestCleanupDismissedEWLEntries(t *testing.T) {
	t.Parallel()

	t.Run("empty dismissed queue should only run size check", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			IsPruningEnabledCalled:           func() bool { return true },
			GetEvictionWaitingListSizeCalled: func() int { return 0 },
		}
		arguments.ExecutionManager = &processMocks.ExecutionManagerMock{
			PopDismissedResultsCalled: func() []executionTrack.DismissedBatch { return nil },
		}

		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		// should not panic, should not call CancelPrune
		sp.CleanupDismissedEWLEntries()
	})
	t.Run("dismissed batches should trigger CancelPrune and reset last pruned header", func(t *testing.T) {
		t.Parallel()

		cancelPruneCalls := 0

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			IsPruningEnabledCalled: func() bool { return true },
			CancelPruneCalled: func(rootHash []byte, identifier state.TriePruningIdentifier) {
				cancelPruneCalls++
			},
			GetEvictionWaitingListSizeCalled: func() int { return 0 },
		}
		popCalled := false
		arguments.ExecutionManager = &processMocks.ExecutionManagerMock{
			PopDismissedResultsCalled: func() []executionTrack.DismissedBatch {
				if popCalled {
					return nil
				}
				popCalled = true
				return []executionTrack.DismissedBatch{
					{
						AnchorResult: &block.ExecutionResult{
							BaseExecutionResult: &block.BaseExecutionResult{RootHash: []byte("R0")},
						},
						Results: []data.BaseExecutionResultHandler{
							&block.ExecutionResult{
								BaseExecutionResult: &block.BaseExecutionResult{RootHash: []byte("R1")},
							},
							&block.ExecutionResult{
								BaseExecutionResult: &block.BaseExecutionResult{RootHash: []byte("R2")},
							},
						},
					},
				}
			},
		}

		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		sp.SetLastPrunedHash([]byte("someHash"))
		sp.SetLastPrunedNonce(100)

		sp.CleanupDismissedEWLEntries()

		// Two transitions: R0->R1 and R1->R2, each producing 2 CancelPrune calls = 4 total
		require.Equal(t, 4, cancelPruneCalls)
		// Last pruned header should be reset
		require.Nil(t, sp.GetLastPrunedHash())
	})
}

func TestCheckEWLSizeAndReset(t *testing.T) {
	t.Parallel()

	t.Run("ewl size below threshold should not trigger reset", func(t *testing.T) {
		t.Parallel()

		resetCalled := false

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			IsPruningEnabledCalled:           func() bool { return true },
			GetEvictionWaitingListSizeCalled: func() int { return 5 },
			ResetPruningCalled: func() {
				resetCalled = true
			},
		}
		arguments.ExecutionManager = &processMocks.ExecutionManagerMock{
			PopDismissedResultsCalled: func() []executionTrack.DismissedBatch { return nil },
		}

		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		// default gap=10 -> threshold=36, ewlSize=5 < 36
		sp.CheckEWLSizeAndReset()
		require.False(t, resetCalled)
	})
	t.Run("ewl size above threshold should trigger reset and clear last pruned header", func(t *testing.T) {
		t.Parallel()

		resetCalled := false

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			IsPruningEnabledCalled:           func() bool { return true },
			GetEvictionWaitingListSizeCalled: func() int { return 1000 },
			ResetPruningCalled: func() {
				resetCalled = true
			},
		}
		arguments.ExecutionManager = &processMocks.ExecutionManagerMock{
			PopDismissedResultsCalled: func() []executionTrack.DismissedBatch { return nil },
		}

		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		sp.SetLastPrunedHash([]byte("someHash"))
		sp.SetLastPrunedNonce(50)

		// default gap=10 -> threshold=36, ewlSize=1000 > 36
		sp.CheckEWLSizeAndReset()
		require.True(t, resetCalled)
		require.Nil(t, sp.GetLastPrunedHash())
	})
	t.Run("pruning disabled should skip reset even if size would exceed", func(t *testing.T) {
		t.Parallel()

		coreComponents, dataComponents, bootstrapComponents, statusComponents := createComponentHolderMocks()
		arguments := CreateMockArguments(coreComponents, dataComponents, bootstrapComponents, statusComponents)
		arguments.AccountsDB[state.UserAccountsState] = &stateMock.AccountsStub{
			IsPruningEnabledCalled: func() bool { return false },
			GetEvictionWaitingListSizeCalled: func() int {
				require.Fail(t, "should not check EWL size when pruning is disabled")
				return 0
			},
			ResetPruningCalled: func() {
				require.Fail(t, "should not reset when pruning is disabled")
			},
		}
		arguments.ExecutionManager = &processMocks.ExecutionManagerMock{
			PopDismissedResultsCalled: func() []executionTrack.DismissedBatch { return nil },
		}

		sp, err := blproc.NewShardProcessor(arguments)
		require.Nil(t, err)

		sp.CheckEWLSizeAndReset()
	})
}
