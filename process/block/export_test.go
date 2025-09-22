package block

import (
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/scheduled"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/executionTrack"
	"github.com/multiversx/mx-chain-go/process/estimator"
	"github.com/multiversx/mx-chain-go/process/missingData"

	"github.com/multiversx/mx-chain-go/common/graceperiod"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/process/coordinator"
	"github.com/multiversx/mx-chain-go/process/factory/containers"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dblookupext"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/outport"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
)

// ComputeHeaderHash -
func (bp *baseProcessor) ComputeHeaderHash(hdr data.HeaderHandler) ([]byte, error) {
	return core.CalculateHash(bp.marshalizer, bp.hasher, hdr)
}

// VerifyStateRoot -
func (bp *baseProcessor) VerifyStateRoot(rootHash []byte) bool {
	return bp.verifyStateRoot(rootHash)
}

// CheckBlockValidity -
func (bp *baseProcessor) CheckBlockValidity(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {
	return bp.checkBlockValidity(headerHandler, bodyHandler)
}

// RemoveHeadersBehindNonceFromPools -
func (bp *baseProcessor) RemoveHeadersBehindNonceFromPools(
	shouldRemoveBlockBody bool,
	shardId uint32,
	nonce uint64,
) {
	bp.removeHeadersBehindNonceFromPools(shouldRemoveBlockBody, shardId, nonce)
}

// GetPruningHandler -
func (bp *baseProcessor) GetPruningHandler(finalHeaderNonce uint64) state.PruningHandler {
	return bp.getPruningHandler(finalHeaderNonce)
}

// SetLastRestartNonce -
func (bp *baseProcessor) SetLastRestartNonce(lastRestartNonce uint64) {
	bp.lastRestartNonce = lastRestartNonce
}

// CommitTrieEpochRootHashIfNeeded -
func (bp *baseProcessor) CommitTrieEpochRootHashIfNeeded(metaBlock *block.MetaBlock, rootHash []byte) error {
	return bp.commitTrieEpochRootHashIfNeeded(metaBlock, rootHash)
}

// CreateMiniBlocks -
func (sp *shardProcessor) CreateMiniBlocks(haveTime func() bool) (*block.Body, map[string]*processedMb.ProcessedMiniBlockInfo, error) {
	return sp.createMiniBlocks(haveTime, []byte("random"))
}

// GetOrderedProcessedMetaBlocksFromHeader -
func (sp *shardProcessor) GetOrderedProcessedMetaBlocksFromHeader(header data.HeaderHandler) ([]data.HeaderHandler, error) {
	return sp.getOrderedProcessedMetaBlocksFromHeader(header)
}

// UpdateCrossShardInfo -
func (sp *shardProcessor) UpdateCrossShardInfo(processedMetaHdrs []data.HeaderHandler) error {
	return sp.updateCrossShardInfo(processedMetaHdrs)
}

// UpdateStateStorage -
func (sp *shardProcessor) UpdateStateStorage(finalHeaders []data.HeaderHandler, currentHeader data.HeaderHandler, currentHeaderHash []byte) {
	currShardHeader, ok := currentHeader.(data.ShardHeaderHandler)
	if !ok {
		return
	}
	sp.updateState(finalHeaders, currShardHeader, currentHeaderHash)
}

// NewShardProcessorEmptyWith3shards -
func NewShardProcessorEmptyWith3shards(
	tdp dataRetriever.PoolsHolder,
	genesisBlocks map[uint32]data.HeaderHandler,
	blockChain data.ChainHandler,
) (*shardProcessor, error) {
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
	nodesCoordinator := shardingMocks.NewNodesCoordinatorMock()

	argsHeaderValidator := ArgsHeaderValidator{
		Hasher:              &hashingMocks.HasherMock{},
		Marshalizer:         &mock.MarshalizerMock{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	hdrValidator, _ := NewHeaderValidator(argsHeaderValidator)

	accountsDb := make(map[state.AccountsDbIdentifier]state.AccountsAdapter)
	accountsDb[state.UserAccountsState] = &stateMock.AccountsStub{}
	gracePeriod, _ := graceperiod.NewEpochChangeGracePeriod([]config.EpochChangeGracePeriodByEpoch{{EnableEpoch: 0, GracePeriodInRounds: 1}})
	coreComponents := &mock.CoreComponentsMock{
		IntMarsh:                           &mock.MarshalizerMock{},
		Hash:                               &hashingMocks.HasherMock{},
		UInt64ByteSliceConv:                &mock.Uint64ByteSliceConverterMock{},
		StatusField:                        &statusHandlerMock.AppStatusHandlerStub{},
		RoundField:                         &mock.RoundHandlerMock{},
		ProcessStatusHandlerField:          &testscommon.ProcessStatusHandlerStub{},
		EpochNotifierField:                 &epochNotifier.EpochNotifierStub{},
		EnableEpochsHandlerField:           enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
		RoundNotifierField:                 &epochNotifier.RoundNotifierStub{},
		EnableRoundsHandlerField:           &testscommon.EnableRoundsHandlerStub{},
		EpochChangeGracePeriodHandlerField: gracePeriod,
	}
	dataComponents := &mock.DataComponentsMock{
		Storage:    &storageStubs.ChainStorerStub{},
		DataPool:   tdp,
		BlockChain: blockChain,
	}
	boostrapComponents := &mock.BootstrapComponentsMock{
		Coordinator:          shardCoordinator,
		HdrIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		VersionedHdrFactory:  &testscommon.VersionedHeaderFactoryStub{},
	}
	statusComponents := &mock.StatusComponentsMock{
		Outport: &outport.OutportStub{},
	}
	statusCoreComponents := &factory.StatusCoreComponentsStub{
		AppStatusHandlerField: &statusHandlerMock.AppStatusHandlerStub{},
	}
	preprocessors := containers.NewPreProcessorsContainer()
	blockDataRequesterArgs := coordinator.BlockDataRequestArgs{
		RequestHandler:      &testscommon.RequestHandlerStub{},
		MiniBlockPool:       dataComponents.Datapool().MiniBlocks(),
		PreProcessors:       preprocessors,
		ShardCoordinator:    boostrapComponents.ShardCoordinator(),
		EnableEpochsHandler: coreComponents.EnableEpochsHandler(),
	}
	// second instance for proposal missing data fetching to avoid interferences
	proposalBlockDataRequester, _ := coordinator.NewBlockDataRequester(blockDataRequesterArgs)

	mbSelectionSession, _ := NewMiniBlocksSelectionSession(
		boostrapComponents.ShardCoordinator().SelfId(),
		coreComponents.InternalMarshalizer(),
		coreComponents.Hasher(),
	)

	executionResultsTracker := executionTrack.NewExecutionResultsTracker()
	execResultsVerifier, _ := NewExecutionResultsVerifier(dataComponents.BlockChain, executionResultsTracker)
	inclusionEstimator := estimator.NewExecutionResultInclusionEstimator(
		config.ExecutionResultInclusionEstimatorConfig{
			SafetyMargin:       110,
			MaxResultsPerBlock: 20,
		},
		coreComponents.RoundHandler(),
	)

	missingDataArgs := missingData.ResolverArgs{
		HeadersPool:        dataComponents.DataPool.Headers(),
		ProofsPool:         dataComponents.DataPool.Proofs(),
		RequestHandler:     &testscommon.RequestHandlerStub{},
		BlockDataRequester: proposalBlockDataRequester,
	}
	missingDataResolver, _ := missingData.NewMissingDataResolver(missingDataArgs)

	arguments := ArgShardProcessor{
		ArgBaseProcessor: ArgBaseProcessor{
			CoreComponents:       coreComponents,
			DataComponents:       dataComponents,
			BootstrapComponents:  boostrapComponents,
			StatusComponents:     statusComponents,
			StatusCoreComponents: statusCoreComponents,
			AccountsDB:           accountsDb,
			ForkDetector:         &mock.ForkDetectorMock{},
			NodesCoordinator:     nodesCoordinator,
			FeeHandler:           &mock.FeeAccumulatorStub{},
			RequestHandler:       &testscommon.RequestHandlerStub{},
			BlockChainHook:       &testscommon.BlockChainHookStub{},
			TxCoordinator:        &testscommon.TransactionCoordinatorMock{},
			EpochStartTrigger:    &mock.EpochStartTriggerStub{},
			HeaderValidator:      hdrValidator,
			BootStorer: &mock.BoostrapStorerMock{
				PutCalled: func(round int64, bootData bootstrapStorage.BootstrapData) error {
					return nil
				},
			},
			BlockTracker:                       mock.NewBlockTrackerMock(shardCoordinator, genesisBlocks),
			BlockSizeThrottler:                 &mock.BlockSizeThrottlerStub{},
			Version:                            "softwareVersion",
			HistoryRepository:                  &dblookupext.HistoryRepositoryStub{},
			GasHandler:                         &mock.GasHandlerMock{},
			OutportDataProvider:                &outport.OutportDataProviderStub{},
			ScheduledTxsExecutionHandler:       &testscommon.ScheduledTxsExecutionStub{},
			ProcessedMiniBlocksTracker:         &testscommon.ProcessedMiniBlocksTrackerStub{},
			ReceiptsRepository:                 &testscommon.ReceiptsRepositoryStub{},
			BlockProcessingCutoffHandler:       &testscommon.BlockProcessingCutoffStub{},
			ManagedPeersHolder:                 &testscommon.ManagedPeersHolderStub{},
			SentSignaturesTracker:              &testscommon.SentSignatureTrackerStub{},
			HeadersForBlock:                    &testscommon.HeadersForBlockMock{},
			MiniBlocksSelectionSession:         mbSelectionSession,
			ExecutionResultsVerifier:           execResultsVerifier,
			MissingDataResolver:                missingDataResolver,
			ExecutionResultsInclusionEstimator: inclusionEstimator,
			ExecutionResultsTracker:            executionResultsTracker,
		},
	}
	shardProc, err := NewShardProcessor(arguments)
	return shardProc, err
}

// GetDataPool -
func (mp *metaProcessor) GetDataPool() dataRetriever.PoolsHolder {
	return mp.dataPool
}

// IsHdrMissing -
func (mp *metaProcessor) IsHdrMissing(hdrHash []byte) bool {
	hdrInfoValue, ok := mp.hdrsForCurrBlock.GetHeaderInfo(string(hdrHash))
	if !ok {
		return true
	}

	return check.IfNil(hdrInfoValue.GetHeader())
}

// CreateShardInfo -
func (mp *metaProcessor) CreateShardInfo() ([]data.ShardDataHandler, error) {
	return mp.createShardInfo()
}

// SaveMetricCrossCheckBlockHeight -
func (mp *metaProcessor) SaveMetricCrossCheckBlockHeight() {
	mp.saveMetricCrossCheckBlockHeight()
}

// NotarizedHdrs -
func (bp *baseProcessor) NotarizedHdrs() map[uint32][]data.HeaderHandler {
	lastCrossNotarizedHeaders := make(map[uint32][]data.HeaderHandler)
	for shardID := uint32(0); shardID < bp.shardCoordinator.NumberOfShards(); shardID++ {
		lastCrossNotarizedHeaderForShard := bp.LastNotarizedHdrForShard(shardID)
		if !check.IfNil(lastCrossNotarizedHeaderForShard) {
			lastCrossNotarizedHeaders[shardID] = append(lastCrossNotarizedHeaders[shardID], lastCrossNotarizedHeaderForShard)
		}
	}

	lastCrossNotarizedHeaderForShard := bp.LastNotarizedHdrForShard(core.MetachainShardId)
	if !check.IfNil(lastCrossNotarizedHeaderForShard) {
		lastCrossNotarizedHeaders[core.MetachainShardId] = append(lastCrossNotarizedHeaders[core.MetachainShardId], lastCrossNotarizedHeaderForShard)
	}

	return lastCrossNotarizedHeaders
}

// LastNotarizedHdrForShard -
func (bp *baseProcessor) LastNotarizedHdrForShard(shardID uint32) data.HeaderHandler {
	lastCrossNotarizedHeaderForShard, _, _ := bp.blockTracker.GetLastCrossNotarizedHeader(shardID)
	if check.IfNil(lastCrossNotarizedHeaderForShard) {
		return nil
	}

	return lastCrossNotarizedHeaderForShard
}

// SetMarshalizer -
func (bp *baseProcessor) SetMarshalizer(marshal marshal.Marshalizer) {
	bp.marshalizer = marshal
}

// SetHasher -
func (bp *baseProcessor) SetHasher(hasher hashing.Hasher) {
	bp.hasher = hasher
}

// SetHeaderValidator -
func (bp *baseProcessor) SetHeaderValidator(validator process.HeaderConstructionValidator) {
	bp.headerValidator = validator
}

// RequestHeadersIfMissing -
func (bp *baseProcessor) RequestHeadersIfMissing(sortedHdrs []data.HeaderHandler, shardId uint32) error {
	return bp.requestHeadersIfMissing(sortedHdrs, shardId)
}

// SetShardBlockFinality -
func (mp *metaProcessor) SetShardBlockFinality(val uint32) {
	mp.shardBlockFinality = val
}

// SaveLastNotarizedHeader -
func (mp *metaProcessor) SaveLastNotarizedHeader(header *block.MetaBlock) error {
	return mp.saveLastNotarizedHeader(header)
}

// CheckShardHeadersValidity -
func (mp *metaProcessor) CheckShardHeadersValidity(header *block.MetaBlock) (map[uint32]data.HeaderHandler, error) {
	return mp.checkShardHeadersValidity(header)
}

// CheckShardHeadersFinality -
func (mp *metaProcessor) CheckShardHeadersFinality(highestNonceHdrs map[uint32]data.HeaderHandler) error {
	return mp.checkShardHeadersFinality(highestNonceHdrs)
}

// CheckHeaderBodyCorrelation -
func (mp *metaProcessor) CheckHeaderBodyCorrelation(hdr data.HeaderHandler, body *block.Body) error {
	return mp.checkHeaderBodyCorrelation(hdr.GetMiniBlockHeaderHandlers(), body)
}

// IsHdrConstructionValid -
func (bp *baseProcessor) IsHdrConstructionValid(currHdr, prevHdr data.HeaderHandler) error {
	return bp.headerValidator.IsHeaderConstructionValid(currHdr, prevHdr)
}

// UpdateShardsHeadersNonce -
func (mp *metaProcessor) UpdateShardsHeadersNonce(key uint32, value uint64) {
	mp.updateShardHeadersNonce(key, value)
}

// GetShardsHeadersNonce -
func (mp *metaProcessor) GetShardsHeadersNonce() *sync.Map {
	return mp.shardsHeadersNonce
}

// SaveLastNotarizedHeader -
func (sp *shardProcessor) SaveLastNotarizedHeader(shardId uint32, processedHdrs []data.HeaderHandler) error {
	return sp.saveLastNotarizedHeader(shardId, processedHdrs)
}

// CheckHeaderBodyCorrelation -
func (sp *shardProcessor) CheckHeaderBodyCorrelation(hdr data.HeaderHandler, body *block.Body) error {
	return sp.checkHeaderBodyCorrelation(hdr.GetMiniBlockHeaderHandlers(), body)
}

// CheckAndRequestIfMetaHeadersMissing -
func (sp *shardProcessor) CheckAndRequestIfMetaHeadersMissing() {
	sp.checkAndRequestIfMetaHeadersMissing()
}

// GetHashAndHdrStruct -
func (sp *shardProcessor) GetHashAndHdrStruct(header data.HeaderHandler, hash []byte) *hashAndHdr {
	return &hashAndHdr{header, hash}
}

// CheckMetaHeadersValidityAndFinality -
func (sp *shardProcessor) CheckMetaHeadersValidityAndFinality() error {
	return sp.checkMetaHeadersValidityAndFinality()
}

// CreateAndProcessMiniBlocksDstMe -
func (sp *shardProcessor) CreateAndProcessMiniBlocksDstMe(
	haveTime func() bool,
) (block.MiniBlockSlice, uint32, uint32, error) {
	createAndProcessInfo, err := sp.createAndProcessMiniBlocksDstMe(haveTime)
	if err != nil {
		return nil, 0, 0, err
	}

	return createAndProcessInfo.miniBlocks, createAndProcessInfo.numHdrsAdded, createAndProcessInfo.numTxsAdded, err
}

// DisplayLogInfo -
func (sp *shardProcessor) DisplayLogInfo(
	header data.HeaderHandler,
	body *block.Body,
	headerHash []byte,
	numShards uint32,
	selfId uint32,
	dataPool dataRetriever.PoolsHolder,
	blockTracker process.BlockTracker,
) {
	sp.txCounter.displayLogInfo(header, body, headerHash, numShards, selfId, dataPool, blockTracker)
}

// GetHighestHdrForOwnShardFromMetachain -
func (sp *shardProcessor) GetHighestHdrForOwnShardFromMetachain(processedHdrs []data.HeaderHandler) ([]data.HeaderHandler, [][]byte, error) {
	return sp.getHighestHdrForOwnShardFromMetachain(processedHdrs)
}

// RestoreMetaBlockIntoPool -
func (sp *shardProcessor) RestoreMetaBlockIntoPool(
	miniBlockHashes map[string]uint32,
	metaBlockHashes [][]byte,
	headerHandler data.HeaderHandler,
) error {
	return sp.restoreMetaBlockIntoPool(headerHandler, miniBlockHashes, metaBlockHashes)
}

// GetAllMiniBlockDstMeFromMeta -
func (sp *shardProcessor) GetAllMiniBlockDstMeFromMeta(
	header data.ShardHeaderHandler,
) (map[string][]byte, error) {
	return sp.getAllMiniBlockDstMeFromMeta(header)
}

// CreateBlockStarted -
func (bp *baseProcessor) CreateBlockStarted() error {
	return bp.createBlockStarted()
}

// AddProcessedCrossMiniBlocksFromHeader -
func (sp *shardProcessor) AddProcessedCrossMiniBlocksFromHeader(header data.HeaderHandler) error {
	return sp.addProcessedCrossMiniBlocksFromHeader(header)
}

// VerifyCrossShardMiniBlockDstMe -
func (mp *metaProcessor) VerifyCrossShardMiniBlockDstMe(header *block.MetaBlock) error {
	return mp.verifyCrossShardMiniBlockDstMe(header)
}

// ApplyBodyToHeader -
func (mp *metaProcessor) ApplyBodyToHeader(metaHdr data.MetaHeaderHandler, body *block.Body) (data.BodyHandler, error) {
	return mp.applyBodyToHeader(metaHdr, body)
}

// ApplyBodyToHeader -
func (sp *shardProcessor) ApplyBodyToHeader(shardHdr data.ShardHeaderHandler, body *block.Body, processedMiniBlocksDestMeInfo map[string]*processedMb.ProcessedMiniBlockInfo) (*block.Body, error) {
	return sp.applyBodyToHeader(shardHdr, body, processedMiniBlocksDestMeInfo)
}

// CreateBlockBody -
func (mp *metaProcessor) CreateBlockBody(metaBlock data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error) {
	return mp.createBlockBody(metaBlock, haveTime)
}

// CreateBlockBody -
func (sp *shardProcessor) CreateBlockBody(shardHdr data.HeaderHandler, haveTime func() bool) (data.BodyHandler, map[string]*processedMb.ProcessedMiniBlockInfo, error) {
	return sp.createBlockBody(shardHdr, haveTime)
}

// CheckEpochCorrectnessCrossChain -
func (sp *shardProcessor) CheckEpochCorrectnessCrossChain() error {
	return sp.checkEpochCorrectnessCrossChain()
}

// CheckEpochCorrectness -
func (sp *shardProcessor) CheckEpochCorrectness(header *block.Header) error {
	return sp.checkEpochCorrectness(header)
}

// GetBootstrapHeadersInfo -
func (sp *shardProcessor) GetBootstrapHeadersInfo(
	selfNotarizedHeaders []data.HeaderHandler,
	selfNotarizedHeadersHashes [][]byte,
) []bootstrapStorage.BootstrapHeaderInfo {
	return sp.getBootstrapHeadersInfo(selfNotarizedHeaders, selfNotarizedHeadersHashes)
}

// RequestMetaHeadersIfNeeded -
func (sp *shardProcessor) RequestMetaHeadersIfNeeded(hdrsAdded uint32, lastMetaHdr data.HeaderHandler) {
	sp.requestMetaHeadersIfNeeded(hdrsAdded, lastMetaHdr)
}

// RequestShardHeadersIfNeeded -
func (mp *metaProcessor) RequestShardHeadersIfNeeded(hdrsAddedForShard map[uint32]uint32, lastShardHdr map[uint32]data.HeaderHandler) {
	mp.requestShardHeadersIfNeeded(hdrsAddedForShard, lastShardHdr)
}

// AddHeaderIntoTrackerPool -
func (bp *baseProcessor) AddHeaderIntoTrackerPool(nonce uint64, shardID uint32) {
	bp.addHeaderIntoTrackerPool(nonce, shardID)
}

// UpdateState -
func (bp *baseProcessor) UpdateState(
	finalHeader data.HeaderHandler,
	rootHash []byte,
	prevRootHash []byte,
	accounts state.AccountsAdapter,
) {
	bp.updateStateStorage(finalHeader, rootHash, prevRootHash, accounts)
}

// GasAndFeesDelta -
func GasAndFeesDelta(initialGasAndFees, finalGasAndFees scheduled.GasAndFees) scheduled.GasAndFees {
	return gasAndFeesDelta(initialGasAndFees, finalGasAndFees)
}

// RequestEpochStartInfo -
func (sp *shardProcessor) RequestEpochStartInfo(header data.ShardHeaderHandler, haveTime func() time.Duration) error {
	return sp.requestEpochStartInfo(header, haveTime)
}

// ProcessEpochStartMetaBlock -
func (mp *metaProcessor) ProcessEpochStartMetaBlock(
	header *block.MetaBlock,
	body *block.Body,
) error {
	return mp.processEpochStartMetaBlock(header, body)
}

// UpdateEpochStartHeader -
func (mp *metaProcessor) UpdateEpochStartHeader(metaHdr *block.MetaBlock) error {
	return mp.updateEpochStartHeader(metaHdr)
}

// CreateEpochStartBody -
func (mp *metaProcessor) CreateEpochStartBody(metaBlock *block.MetaBlock) (data.BodyHandler, error) {
	return mp.createEpochStartBody(metaBlock)
}

// GetIndexOfFirstMiniBlockToBeExecuted -
func (bp *baseProcessor) GetIndexOfFirstMiniBlockToBeExecuted(header data.HeaderHandler) int {
	return bp.getIndexOfFirstMiniBlockToBeExecuted(header)
}

// GetFinalMiniBlocks -
func (bp *baseProcessor) GetFinalMiniBlocks(header data.HeaderHandler, body *block.Body) (*block.Body, error) {
	return bp.getFinalMiniBlocks(header, body)
}

// GetScheduledMiniBlocksFromMe -
func GetScheduledMiniBlocksFromMe(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) (block.MiniBlockSlice, error) {
	return getScheduledMiniBlocksFromMe(headerHandler, bodyHandler)
}

// CheckScheduledMiniBlocksValidity -
func (bp *baseProcessor) CheckScheduledMiniBlocksValidity(headerHandler data.HeaderHandler) error {
	return bp.checkScheduledMiniBlocksValidity(headerHandler)
}

// SetMiniBlockHeaderReservedField -
func (bp *baseProcessor) SetMiniBlockHeaderReservedField(
	miniBlock *block.MiniBlock,
	miniBlockHeaderHandler data.MiniBlockHeaderHandler,
	processedMiniBlocksDestMeInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) error {
	return bp.setMiniBlockHeaderReservedField(miniBlock, miniBlockHeaderHandler, processedMiniBlocksDestMeInfo)
}

// GetFinalMiniBlockHeaders -
func (mp *metaProcessor) GetFinalMiniBlockHeaders(miniBlockHeaderHandlers []data.MiniBlockHeaderHandler) []data.MiniBlockHeaderHandler {
	return mp.getFinalMiniBlockHeaders(miniBlockHeaderHandlers)
}

// CheckProcessorNilParameters -
func CheckProcessorNilParameters(arguments ArgBaseProcessor) error {
	return checkProcessorParameters(arguments)
}

// SetIndexOfFirstTxProcessed -
func (bp *baseProcessor) SetIndexOfFirstTxProcessed(miniBlockHeaderHandler data.MiniBlockHeaderHandler) error {
	return bp.setIndexOfFirstTxProcessed(miniBlockHeaderHandler)
}

// SetIndexOfLastTxProcessed -
func (bp *baseProcessor) SetIndexOfLastTxProcessed(
	miniBlockHeaderHandler data.MiniBlockHeaderHandler,
	processedMiniBlocksDestMeInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) error {
	return bp.setIndexOfLastTxProcessed(miniBlockHeaderHandler, processedMiniBlocksDestMeInfo)
}

// GetProcessedMiniBlocksTracker -
func (bp *baseProcessor) GetProcessedMiniBlocksTracker() process.ProcessedMiniBlocksTracker {
	return bp.processedMiniBlocksTracker
}

// SetProcessingTypeAndConstructionStateForScheduledMb -
func (bp *baseProcessor) SetProcessingTypeAndConstructionStateForScheduledMb(
	miniBlockHeaderHandler data.MiniBlockHeaderHandler,
	processedMiniBlocksDestMeInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) error {
	return bp.setProcessingTypeAndConstructionStateForScheduledMb(miniBlockHeaderHandler, processedMiniBlocksDestMeInfo)
}

// SetProcessingTypeAndConstructionStateForNormalMb -
func (bp *baseProcessor) SetProcessingTypeAndConstructionStateForNormalMb(
	miniBlockHeaderHandler data.MiniBlockHeaderHandler,
	processedMiniBlocksDestMeInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) error {
	return bp.setProcessingTypeAndConstructionStateForNormalMb(miniBlockHeaderHandler, processedMiniBlocksDestMeInfo)
}

// RollBackProcessedMiniBlockInfo -
func (sp *shardProcessor) RollBackProcessedMiniBlockInfo(miniBlockHeader data.MiniBlockHeaderHandler, miniBlockHash []byte) {
	sp.rollBackProcessedMiniBlockInfo(miniBlockHeader, miniBlockHash)
}

// SetProcessedMiniBlocksInfo -
func (sp *shardProcessor) SetProcessedMiniBlocksInfo(miniBlockHashes [][]byte, metaBlockHash string, metaBlock *block.MetaBlock) {
	sp.setProcessedMiniBlocksInfo(miniBlockHashes, metaBlockHash, metaBlock)
}

// GetIndexOfLastTxProcessedInMiniBlock -
func (sp *shardProcessor) GetIndexOfLastTxProcessedInMiniBlock(miniBlockHash []byte, metaBlock *block.MetaBlock) int32 {
	return getIndexOfLastTxProcessedInMiniBlock(miniBlockHash, metaBlock)
}

// RollBackProcessedMiniBlocksInfo -
func (sp *shardProcessor) RollBackProcessedMiniBlocksInfo(headerHandler data.HeaderHandler, mapMiniBlockHashes map[string]uint32) {
	sp.rollBackProcessedMiniBlocksInfo(headerHandler, mapMiniBlockHashes)
}

// CheckConstructionStateAndIndexesCorrectness -
func (bp *baseProcessor) CheckConstructionStateAndIndexesCorrectness(mbh data.MiniBlockHeaderHandler) error {
	return checkConstructionStateAndIndexesCorrectness(mbh)
}

// GetAllMarshalledTxs -
func (mp *metaProcessor) GetAllMarshalledTxs(body *block.Body) map[string][][]byte {
	return mp.getAllMarshalledTxs(body)
}

// SetNonceOfFirstCommittedBlock -
func (bp *baseProcessor) SetNonceOfFirstCommittedBlock(nonce uint64) {
	bp.setNonceOfFirstCommittedBlock(nonce)
}

// CheckSentSignaturesAtCommitTime -
func (bp *baseProcessor) CheckSentSignaturesAtCommitTime(header data.HeaderHandler) error {
	return bp.checkSentSignaturesAtCommitTime(header)
}

// GetHdrForBlock -
func (mp *metaProcessor) GetHdrForBlock() HeadersForBlock {
	return mp.hdrsForCurrBlock
}

// GetHdrForBlock -
func (sp *shardProcessor) GetHdrForBlock() HeadersForBlock {
	return sp.hdrsForCurrBlock
}

// SelectIncomingMiniBlocks -
func (sp *shardProcessor) SelectIncomingMiniBlocks(
	lastCrossNotarizedMetaHdr data.HeaderHandler,
	orderedMetaBlocks []data.HeaderHandler,
	orderedMetaBlocksHashes [][]byte,
	haveTime func() bool,
) error {
	return sp.selectIncomingMiniBlocks(lastCrossNotarizedMetaHdr, orderedMetaBlocks, orderedMetaBlocksHashes, haveTime)
}

// DisplayHeader -
func DisplayHeader(
	headerHandler data.HeaderHandler,
	headerProof data.HeaderProofHandler,
) []*display.LineData {
	return displayHeader(headerHandler, headerProof)
}

// GetLastBaseExecutionResultHandler -
func GetLastBaseExecutionResultHandler(header data.HeaderHandler) (data.BaseExecutionResultHandler, error) {
	return getLastBaseExecutionResultHandler(header)
}

// CreateBaseProcessorWithMockedTracker -
func CreateBaseProcessorWithMockedTracker(tracker process.BlockTracker) *baseProcessor {
	return &baseProcessor{
		blockTracker: tracker,
	}
}

// ComputeOwnShardStuckIfNeeded -
func (bp *baseProcessor) ComputeOwnShardStuckIfNeeded(header data.HeaderHandler) error {
	return bp.computeOwnShardStuckIfNeeded(header)
}

// SetMiniBlockSelectionSession -
func (bp *baseProcessor) SetMiniBlockSelectionSession(session MiniBlocksSelectionSession) {
	bp.miniBlocksSelectionSession = session
}

// CheckHeaderBodyCorrelationProposal -
func (bp *baseProcessor) CheckHeaderBodyCorrelationProposal(miniBlockHeaders []data.MiniBlockHeaderHandler, body *block.Body) error {
	return bp.checkHeaderBodyCorrelationProposal(miniBlockHeaders, body)
}

// VerifyCrossShardMiniBlockDstMe -
func (sp *shardProcessor) VerifyCrossShardMiniBlockDstMe(header data.ShardHeaderHandler) error {
	return sp.verifyCrossShardMiniBlockDstMe(header)
}

// AddCrossShardMiniBlocksDstMeToMap -
func (sp *shardProcessor) AddCrossShardMiniBlocksDstMeToMap(
	header data.ShardHeaderHandler,
	referencedMetaBlockHash []byte,
	referencedMetaHeaderHandler data.HeaderHandler,
	lastCrossNotarizedHeader data.HeaderHandler,
	miniBlockMetaHashes map[string][]byte,
) error {
	return sp.addCrossShardMiniBlocksDstMeToMap(header, referencedMetaBlockHash, referencedMetaHeaderHandler, lastCrossNotarizedHeader, miniBlockMetaHashes)
}

// CheckInclusionEstimationForExecutionResults -
func (sp *shardProcessor) CheckInclusionEstimationForExecutionResults(header data.HeaderHandler) error {
	return sp.checkInclusionEstimationForExecutionResults(header)
}

// CheckMetaHeadersValidityAndFinalityProposal -
func (sp *shardProcessor) CheckMetaHeadersValidityAndFinalityProposal(header data.ShardHeaderHandler) error {
	return sp.checkMetaHeadersValidityAndFinalityProposal(header)
}
