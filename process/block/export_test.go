package block

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
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

func DisplayHeader(headerHandler data.HeaderHandler) []*display.LineData {
	return displayHeader(headerHandler)
}

func (sp *shardProcessor) ReceivedMetaBlock(metaBlockHash []byte) {
	sp.receivedMetaBlock(metaBlockHash)
}

func (sp *shardProcessor) CreateMiniBlocks(noShards uint32, maxItemsInBlock uint32, round uint64, haveTime func() bool) (block.Body, error) {
	return sp.createMiniBlocks(noShards, maxItemsInBlock, round, haveTime)
}

func (sp *shardProcessor) GetProcessedMetaBlocksFromPool(body block.Body, header *block.Header) ([]data.HeaderHandler, error) {
	return sp.getProcessedMetaBlocksFromPool(body, header)
}

func (sp *shardProcessor) RemoveProcessedMetablocksFromPool(processedMetaHdrs []data.HeaderHandler) error {
	return sp.removeProcessedMetablocksFromPool(processedMetaHdrs)
}

func NewShardProcessorEmptyWith3shards(tdp dataRetriever.PoolsHolder, genesisBlocks map[uint32]data.HeaderHandler) (*shardProcessor, error) {
	shardProcessor, err := NewShardProcessor(
		&mock.ServiceContainerMock{},
		tdp,
		&mock.ChainStorerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		&mock.AccountsStub{},
		mock.NewMultiShardsCoordinatorMock(3),
		&mock.ForkDetectorMock{},
		&mock.BlocksTrackerMock{},
		&mock.PeerProcessorMock{},
		genesisBlocks,
		&mock.RequestHandlerMock{},
		&mock.TransactionCoordinatorMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	return shardProcessor, err
}

func NewMetaProcessorBasicSingleShard(mdp dataRetriever.MetaPoolsHolder, genesisBlocks map[uint32]data.HeaderHandler) (*metaProcessor, error) {
	mp, err := NewMetaProcessor(
		&mock.ServiceContainerMock{},
		&mock.AccountsStub{},
		mdp,
		&mock.ForkDetectorMock{},
		&mock.PeerProcessorMock{},
		mock.NewOneShardCoordinatorMock(),
		&mock.HasherStub{},
		&mock.MarshalizerMock{},
		&mock.ChainStorerMock{},
		genesisBlocks,
		&mock.RequestHandlerMock{},
		&mock.Uint64ByteSliceConverterMock{},
	)
	return mp, err
}

func (mp *metaProcessor) RequestBlockHeaders(header *block.MetaBlock) (uint32, uint32) {
	return mp.requestShardHeaders(header)
}

func (mp *metaProcessor) RemoveBlockInfoFromPool(header *block.MetaBlock) error {
	return mp.removeBlockInfoFromPool(header)
}

func (mp *metaProcessor) DisplayMetaBlock(header *block.MetaBlock) {
	mp.displayMetaBlock(header)
}

func (mp *metaProcessor) ReceivedHeader(hdrHash []byte) {
	mp.receivedHeader(hdrHash)
}

func (mp *metaProcessor) AddHdrHashToRequestedList(hdrHash []byte) {
	mp.mutRequestedShardHdrsHashes.Lock()
	defer mp.mutRequestedShardHdrsHashes.Unlock()

	if mp.requestedShardHdrsHashes == nil {
		mp.requestedShardHdrsHashes = make(map[string]bool)
		mp.allNeededShardHdrsFound = true
	}

	if mp.currHighestShardHdrsNonces == nil {
		mp.currHighestShardHdrsNonces = make(map[uint32]uint64, mp.shardCoordinator.NumberOfShards())
		for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
			mp.currHighestShardHdrsNonces[i] = uint64(0)
		}
	}

	mp.requestedShardHdrsHashes[string(hdrHash)] = true
	mp.allNeededShardHdrsFound = false
}

func (mp *metaProcessor) SetCurrHighestShardHdrsNonces(key uint32, value uint64) {
	mp.currHighestShardHdrsNonces[key] = value
}

func (mp *metaProcessor) IsHdrHashRequested(hdrHash []byte) bool {
	mp.mutRequestedShardHdrsHashes.Lock()
	defer mp.mutRequestedShardHdrsHashes.Unlock()

	_, found := mp.requestedShardHdrsHashes[string(hdrHash)]

	return found
}

func (mp *metaProcessor) CreateShardInfo(maxItemsInBlock uint32, round uint64, haveTime func() bool) ([]block.ShardData, error) {
	return mp.createShardInfo(maxItemsInBlock, round, haveTime)
}

func (mp *metaProcessor) ProcessBlockHeaders(header *block.MetaBlock, round uint64, haveTime func() time.Duration) error {
	return mp.processBlockHeaders(header, round, haveTime)
}

func (mp *metaProcessor) RequestFinalMissingHeaders() uint32 {
	return mp.requestFinalMissingHeaders()
}

func (bp *baseProcessor) NotarizedHdrs() map[uint32][]data.HeaderHandler {
	return bp.notarizedHdrs
}

func (bp *baseProcessor) LastNotarizedHdrForShard(shardId uint32) data.HeaderHandler {
	return bp.lastNotarizedHdrForShard(shardId)
}

func (bp *baseProcessor) SetMarshalizer(marshal marshal.Marshalizer) {
	bp.marshalizer = marshal
}

func (bp *baseProcessor) SetHasher(hasher hashing.Hasher) {
	bp.hasher = hasher
}

func (mp *metaProcessor) SetNextKValidity(val uint32) {
	mp.mutRequestedShardHdrsHashes.Lock()
	mp.nextKValidity = val
	mp.mutRequestedShardHdrsHashes.Unlock()
}

func (mp *metaProcessor) SaveLastNotarizedHeader(header *block.MetaBlock) error {
	return mp.saveLastNotarizedHeader(header)
}

func (mp *metaProcessor) CheckShardHeadersValidity(header *block.MetaBlock) (map[uint32]data.HeaderHandler, error) {
	return mp.checkShardHeadersValidity(header)
}

func (mp *metaProcessor) CheckShardHeadersFinality(header *block.MetaBlock, highestNonceHdrs map[uint32]data.HeaderHandler) error {
	return mp.checkShardHeadersFinality(header, highestNonceHdrs)
}

func (bp *baseProcessor) IsHdrConstructionValid(currHdr, prevHdr data.HeaderHandler) error {
	return bp.isHdrConstructionValid(currHdr, prevHdr)
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
	return sp.checkHeaderBodyCorrelation(hdr, body)
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

func (sp *shardProcessor) RequestFinalMissingHeaders() uint32 {
	return sp.requestFinalMissingHeaders()
}

func (sp *shardProcessor) CheckMetaHeadersValidityAndFinality(hdr *block.Header) error {
	return sp.checkMetaHeadersValidityAndFinality(hdr)
}

func (sp *shardProcessor) GetOrderedMetaBlocks(round uint64) ([]*hashAndHdr, error) {
	return sp.getOrderedMetaBlocks(round)
}

func (sp *shardProcessor) CreateAndProcessCrossMiniBlocksDstMe(
	noShards uint32,
	maxItemsInBlock uint32,
	round uint64,
	haveTime func() bool,
) (block.MiniBlockSlice, [][]byte, uint32, error) {
	return sp.createAndProcessCrossMiniBlocksDstMe(noShards, maxItemsInBlock, round, haveTime)
}

func (bp *baseProcessor) SetBlockSizeThrottler(blockSizeThrottler process.BlockSizeThrottler) {
	bp.blockSizeThrottler = blockSizeThrottler
}

func (sp *shardProcessor) SetCurrHighestMetaHdrNonce(value uint64) {
	sp.currHighestMetaHdrNonce = value
}

func (sp *shardProcessor) DisplayLogInfo(
	header *block.Header,
	body block.Body,
	headerHash []byte,
	numShards uint32,
	selfId uint32,
	dataPool dataRetriever.PoolsHolder,
) {
	sp.txCounter.displayLogInfo(header, body, headerHash, numShards, selfId, dataPool)
}

func (sp *shardProcessor) GetHighestHdrForOwnShardFromMetachain(round uint64) (*block.Header, []byte, error) {
	return sp.getHighestHdrForOwnShardFromMetachain(round)
}

func (sp *shardProcessor) RestoreMetaBlockIntoPool(
	miniBlockHashes map[string]uint32,
	metaBlockHashes [][]byte,
) error {
	return sp.restoreMetaBlockIntoPool(miniBlockHashes, metaBlockHashes)
}

func (sp *shardProcessor) GetAllMiniBlockDstMeFromMeta(
	round uint64,
	metaHashes [][]byte,
) (map[string][]byte, error) {
	return sp.getAllMiniBlockDstMeFromMeta(round, metaHashes)
}
