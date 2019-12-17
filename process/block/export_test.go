package block

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

func (bp *baseProcessor) ComputeHeaderHash(hdr data.HeaderHandler) ([]byte, error) {
	return core.CalculateHash(bp.marshalizer, bp.hasher, hdr)
}

func (bp *baseProcessor) VerifyStateRoot(rootHash []byte) bool {
	return bp.verifyStateRoot(rootHash)
}

func (bp *baseProcessor) CheckBlockValidity(
	chainHandler data.ChainHandler,
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {
	return bp.checkBlockValidity(chainHandler, headerHandler, bodyHandler)
}

func (sp *shardProcessor) ReceivedMetaBlock(metaBlockHash []byte) {
	sp.receivedMetaBlock(metaBlockHash)
}

func (sp *shardProcessor) CreateMiniBlocks(maxItemsInBlock uint32, round uint64, haveTime func() bool) (block.Body, error) {
	return sp.createMiniBlocks(maxItemsInBlock, round, haveTime)
}

func (sp *shardProcessor) GetOrderedProcessedMetaBlocksFromHeader(header *block.Header) ([]data.HeaderHandler, error) {
	return sp.getOrderedProcessedMetaBlocksFromHeader(header)
}

func (sp *shardProcessor) RemoveProcessedMetaBlocksFromPool(processedMetaHdrs []data.HeaderHandler) error {
	return sp.removeProcessedMetaBlocksFromPool(processedMetaHdrs)
}

func NewShardProcessorEmptyWith3shards(tdp dataRetriever.PoolsHolder, genesisBlocks map[uint32]data.HeaderHandler) (*shardProcessor, error) {
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(3)
	nodesCoordinator := mock.NewNodesCoordinatorMock()
	specialAddressHandler := mock.NewSpecialAddressHandlerMock(
		&mock.AddressConverterMock{},
		shardCoordinator,
		nodesCoordinator,
	)

	argsHeaderValidator := ArgsHeaderValidator{
		Hasher:      &mock.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
	}
	headerValidator, _ := NewHeaderValidator(argsHeaderValidator)

	arguments := ArgShardProcessor{
		ArgBaseProcessor: ArgBaseProcessor{
			Accounts:                     &mock.AccountsStub{},
			ForkDetector:                 &mock.ForkDetectorMock{},
			Hasher:                       &mock.HasherMock{},
			Marshalizer:                  &mock.MarshalizerMock{},
			Store:                        &mock.ChainStorerMock{},
			ShardCoordinator:             shardCoordinator,
			NodesCoordinator:             nodesCoordinator,
			SpecialAddressHandler:        specialAddressHandler,
			Uint64Converter:              &mock.Uint64ByteSliceConverterMock{},
			StartHeaders:                 genesisBlocks,
			RequestHandler:               &mock.RequestHandlerMock{},
			Core:                         &mock.ServiceContainerMock{},
			BlockChainHook:               &mock.BlockChainHookHandlerMock{},
			TxCoordinator:                &mock.TransactionCoordinatorMock{},
			ValidatorStatisticsProcessor: &mock.ValidatorStatisticsProcessorMock{},
			EpochStartTrigger:            &mock.EpochStartTriggerStub{},
			HeaderValidator:              headerValidator,
			Rounder:                      &mock.RounderMock{},
			BootStorer: &mock.BoostrapStorerMock{
				PutCalled: func(round int64, bootData bootstrapStorage.BootstrapData) error {
					return nil
				},
			},
		},
		DataPool:        tdp,
		TxsPoolsCleaner: &mock.TxPoolsCleanerMock{},
	}
	shardProcessor, err := NewShardProcessor(arguments)
	return shardProcessor, err
}

func (mp *metaProcessor) RequestBlockHeaders(header *block.MetaBlock) (uint32, uint32) {
	return mp.requestShardHeaders(header)
}

func (mp *metaProcessor) RemoveBlockInfoFromPool(header *block.MetaBlock) error {
	return mp.removeBlockInfoFromPool(header)
}

func (mp *metaProcessor) ReceivedShardHeader(shardHeaderHash []byte) {
	mp.receivedShardHeader(shardHeaderHash)
}

func (mp *metaProcessor) AddHdrHashToRequestedList(hdr *block.Header, hdrHash []byte) {
	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	defer mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	if mp.hdrsForCurrBlock.hdrHashAndInfo == nil {
		mp.hdrsForCurrBlock.hdrHashAndInfo = make(map[string]*hdrInfo)
	}

	if mp.hdrsForCurrBlock.highestHdrNonce == nil {
		mp.hdrsForCurrBlock.highestHdrNonce = make(map[uint32]uint64, mp.shardCoordinator.NumberOfShards())
	}

	mp.hdrsForCurrBlock.hdrHashAndInfo[string(hdrHash)] = &hdrInfo{hdr: hdr, usedInBlock: true}
	mp.hdrsForCurrBlock.missingHdrs++
}

func (mp *metaProcessor) IsHdrMissing(hdrHash []byte) bool {
	mp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	defer mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	hdrInfo, ok := mp.hdrsForCurrBlock.hdrHashAndInfo[string(hdrHash)]
	if !ok {
		return true
	}

	return hdrInfo.hdr == nil || hdrInfo.hdr.IsInterfaceNil()
}

func (mp *metaProcessor) CreateShardInfo(round uint64) ([]block.ShardData, error) {
	return mp.createShardInfo(round)
}

func (mp *metaProcessor) ProcessBlockHeaders(header *block.MetaBlock, round uint64, haveTime func() time.Duration) error {
	return mp.processBlockHeaders(header, round, haveTime)
}

func (mp *metaProcessor) CreateEpochStartForMetablock() (*block.EpochStart, error) {
	return mp.createEpochStartForMetablock()
}

func (mp *metaProcessor) GetLastFinalizedMetaHashForShard(shardHdr *block.Header) ([]byte, []byte, error) {
	return mp.getLastFinalizedMetaHashForShard(shardHdr)
}

func (mp *metaProcessor) RequestMissingFinalityAttestingShardHeaders() uint32 {
	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	defer mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	return mp.requestMissingFinalityAttestingShardHeaders()
}

func (bp *baseProcessor) NotarizedHdrs() map[uint32][]data.HeaderHandler {
	return bp.notarizedHdrs
}

func (bp *baseProcessor) LastNotarizedHdrForShard(shardId uint32) data.HeaderHandler {
	return bp.lastNotarizedHdrForShard(shardId)
}

func (bp *baseProcessor) RemoveLastNotarized() {
	bp.removeLastNotarized()
}

func (bp *baseProcessor) SetMarshalizer(marshal marshal.Marshalizer) {
	bp.marshalizer = marshal
}

func (bp *baseProcessor) SetHasher(hasher hashing.Hasher) {
	bp.hasher = hasher
}

func (bp *baseProcessor) SetHeaderValidator(validator process.HeaderConstructionValidator) {
	bp.headerValidator = validator
}

func (mp *metaProcessor) SetShardBlockFinality(val uint32) {
	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	mp.shardBlockFinality = val
	mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()
}

func (mp *metaProcessor) SaveLastNotarizedHeader(header *block.MetaBlock) error {
	return mp.saveLastNotarizedHeader(header)
}

func (mp *metaProcessor) CheckShardHeadersValidity(header *block.MetaBlock) (map[uint32]data.HeaderHandler, error) {
	return mp.checkShardHeadersValidity(header)
}

func (mp *metaProcessor) CheckShardHeadersFinality(highestNonceHdrs map[uint32]data.HeaderHandler) error {
	return mp.checkShardHeadersFinality(highestNonceHdrs)
}

func (bp *baseProcessor) IsHdrConstructionValid(currHdr, prevHdr data.HeaderHandler) error {
	return bp.headerValidator.IsHeaderConstructionValid(currHdr, prevHdr)
}

func (mp *metaProcessor) IsShardHeaderValidFinal(currHdr *block.Header, lastHdr *block.Header, sortedShardHdrs []*block.Header) (bool, []uint32) {
	return mp.isShardHeaderValidFinal(currHdr, lastHdr, sortedShardHdrs)
}

func (mp *metaProcessor) ChRcvAllHdrs() chan bool {
	return mp.chRcvAllHdrs
}

func (mp *metaProcessor) UpdateShardsHeadersNonce(key uint32, value uint64) {
	mp.updateShardHeadersNonce(key, value)
}

func (mp *metaProcessor) GetShardsHeadersNonce() *sync.Map {
	return mp.shardsHeadersNonce
}

func NewBaseProcessor(shardCord sharding.Coordinator) *baseProcessor {
	return &baseProcessor{shardCoordinator: shardCord}
}

func (bp *baseProcessor) SaveLastNotarizedHeader(shardId uint32, processedHdrs []data.HeaderHandler) error {
	return bp.saveLastNotarizedHeader(shardId, processedHdrs)
}

func (sp *shardProcessor) CheckHeaderBodyCorrelation(hdr *block.Header, body block.Body) error {
	return sp.checkHeaderBodyCorrelation(hdr.MiniBlockHeaders, body)
}

func (bp *baseProcessor) SetLastNotarizedHeadersSlice(startHeaders map[uint32]data.HeaderHandler) error {
	return bp.setLastNotarizedHeadersSlice(startHeaders)
}

func (sp *shardProcessor) CheckAndRequestIfMetaHeadersMissing(round uint64) {
	sp.checkAndRequestIfMetaHeadersMissing(round)
}

func (sp *shardProcessor) IsMetaHeaderFinal(currHdr data.HeaderHandler, sortedHdrs []*hashAndHdr, startPos int) bool {
	return sp.isMetaHeaderFinal(currHdr, sortedHdrs, startPos)
}

func (sp *shardProcessor) GetHashAndHdrStruct(header data.HeaderHandler, hash []byte) *hashAndHdr {
	return &hashAndHdr{header, hash}
}

func (sp *shardProcessor) RequestMissingFinalityAttestingHeaders() uint32 {
	sp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	defer sp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	return sp.requestMissingFinalityAttestingHeaders(
		sharding.MetachainShardId,
		sp.metaBlockFinality,
		sp.getMetaHeaderFromPoolWithNonce,
		sp.dataPool.MetaBlocks())
}

func (sp *shardProcessor) CheckMetaHeadersValidityAndFinality() error {
	return sp.checkMetaHeadersValidityAndFinality()
}

func (sp *shardProcessor) GetOrderedMetaBlocks(round uint64) ([]*hashAndHdr, error) {
	return sp.getOrderedMetaBlocks(round)
}

func (sp *shardProcessor) CreateAndProcessCrossMiniBlocksDstMe(
	maxItemsInBlock uint32,
	round uint64,
	haveTime func() bool,
) (block.MiniBlockSlice, uint32, uint32, error) {
	return sp.createAndProcessCrossMiniBlocksDstMe(maxItemsInBlock, round, haveTime)
}

func (bp *baseProcessor) SetBlockSizeThrottler(blockSizeThrottler process.BlockSizeThrottler) {
	bp.blockSizeThrottler = blockSizeThrottler
}

func (sp *shardProcessor) DisplayLogInfo(
	header *block.Header,
	body block.Body,
	headerHash []byte,
	numShards uint32,
	selfId uint32,
	dataPool dataRetriever.PoolsHolder,
	statusHandler core.AppStatusHandler,
) {
	sp.txCounter.displayLogInfo(header, body, headerHash, numShards, selfId, dataPool, statusHandler)
}

func (sp *shardProcessor) GetHighestHdrForOwnShardFromMetachain(processedHdrs []data.HeaderHandler) ([]data.HeaderHandler, [][]byte, error) {
	return sp.getHighestHdrForOwnShardFromMetachain(processedHdrs)
}

func (sp *shardProcessor) RestoreMetaBlockIntoPool(
	miniBlockHashes map[string]uint32,
	metaBlockHashes [][]byte,
) error {
	return sp.restoreMetaBlockIntoPool(miniBlockHashes, metaBlockHashes)
}

func (sp *shardProcessor) GetAllMiniBlockDstMeFromMeta(
	header *block.Header,
) (map[string][]byte, error) {
	return sp.getAllMiniBlockDstMeFromMeta(header)
}

func (sp *shardProcessor) IsMiniBlockProcessed(metaBlockHash []byte, miniBlockHash []byte) bool {
	return sp.isMiniBlockProcessed(metaBlockHash, miniBlockHash)
}

func (sp *shardProcessor) AddProcessedMiniBlock(metaBlockHash []byte, miniBlockHash []byte) {
	sp.addProcessedMiniBlock(metaBlockHash, miniBlockHash)
}

func (bp *baseProcessor) SetHdrForCurrentBlock(headerHash []byte, headerHandler data.HeaderHandler, usedInBlock bool) {
	bp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	bp.hdrsForCurrBlock.hdrHashAndInfo[string(headerHash)] = &hdrInfo{hdr: headerHandler, usedInBlock: usedInBlock}
	bp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()
}

func (bp *baseProcessor) SetHighestHdrNonceForCurrentBlock(shardId uint32, value uint64) {
	bp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	bp.hdrsForCurrBlock.highestHdrNonce[shardId] = value
	bp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()
}

func (bp *baseProcessor) CreateBlockStarted() {
	bp.createBlockStarted()
}

func (sp *shardProcessor) CreateBlockStarted() {
	sp.createBlockStarted()
}

func (sp *shardProcessor) AddProcessedCrossMiniBlocksFromHeader(header *block.Header) error {
	return sp.addProcessedCrossMiniBlocksFromHeader(header)
}

func (mp *metaProcessor) VerifyCrossShardMiniBlockDstMe(header *block.MetaBlock) error {
	return mp.verifyCrossShardMiniBlockDstMe(header)
}
