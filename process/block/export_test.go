package block

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/display"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

func (bp *baseProcessor) ComputeHeaderHash(hdr data.HeaderHandler) ([]byte, error) {
	return bp.computeHeaderHash(hdr)
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

func (sp *shardProcessor) GetTransactionFromPool(senderShardID, destShardID uint32, txHash []byte) *transaction.Transaction {
	return sp.getTransactionFromPool(senderShardID, destShardID, txHash)
}

func (sp *shardProcessor) RequestBlockTransactions(body block.Body) int {
	return sp.requestBlockTransactions(body)
}

func (sp *shardProcessor) RequestBlockTransactionsForMiniBlock(mb *block.MiniBlock) int {
	return sp.requestBlockTransactionsForMiniBlock(mb)
}

func (sp *shardProcessor) WaitForTxHashes(waitTime time.Duration) {
	sp.waitForTxHashes(waitTime)
}

func (sp *shardProcessor) ReceivedTransaction(txHash []byte) {
	sp.receivedTransaction(txHash)
}

func (sp *shardProcessor) DisplayShardBlock(header *block.Header, txBlock block.Body) {
	sp.displayShardBlock(header, txBlock)
}

func SortTxByNonce(txShardStore storage.Cacher) ([]*transaction.Transaction, [][]byte, error) {
	return sortTxByNonce(txShardStore)
}

func (sp *shardProcessor) GetAllTxsFromMiniBlock(mb *block.MiniBlock, haveTime func() bool) ([]*transaction.Transaction, [][]byte, error) {
	return sp.getAllTxsFromMiniBlock(mb, haveTime)
}

func (sp *shardProcessor) ReceivedMiniBlock(miniBlockHash []byte) {
	sp.receivedMiniBlock(miniBlockHash)
}

func (sp *shardProcessor) ReceivedMetaBlock(metaBlockHash []byte) {
	sp.receivedMetaBlock(metaBlockHash)
}

func (sp *shardProcessor) AddTxHashToRequestedList(txHash []byte) {
	sp.mutTxsForBlock.Lock()
	defer sp.mutTxsForBlock.Unlock()

	if sp.txsForBlock == nil {
		sp.txsForBlock = make(map[string]*txInfo)
	}
	sp.txsForBlock[string(txHash)] = &txInfo{txShardInfo: &txShardInfo{}}
}

func (sp *shardProcessor) IsTxHashRequested(txHash []byte) bool {
	sp.mutTxsForBlock.Lock()
	defer sp.mutTxsForBlock.Unlock()

	return !sp.txsForBlock[string(txHash)].has
}

func (sp *shardProcessor) SetMissingTxs(missingTxs int) {
	sp.mutTxsForBlock.Lock()
	sp.missingTxs = missingTxs
	sp.mutTxsForBlock.Unlock()
}

func (sp *shardProcessor) ProcessMiniBlockComplete(miniBlock *block.MiniBlock, round uint32, haveTime func() bool) error {
	return sp.processMiniBlockComplete(miniBlock, round, haveTime)
}

func (sp *shardProcessor) CreateMiniBlocks(noShards uint32, maxTxInBlock int, round uint32, haveTime func() bool) (block.Body, error) {
	return sp.createMiniBlocks(noShards, maxTxInBlock, round, haveTime)
}

func (sp *shardProcessor) GetProcessedMetaBlocksFromPool(body block.Body) ([]data.HeaderHandler, error) {
	return sp.getProcessedMetaBlocksFromPool(body)
}

func (sp *shardProcessor) RemoveProcessedMetablocksFromPool(processedMetaHdrs []data.HeaderHandler) error {
	return sp.removeProcessedMetablocksFromPool(processedMetaHdrs)
}

func (sp *shardProcessor) RemoveTxBlockFromPools(blockBody block.Body) error {
	return sp.removeTxBlockFromPools(blockBody)
}

func (sp *shardProcessor) ChRcvAllTxs() chan bool {
	return sp.chRcvAllTxs
}

func (mp *metaProcessor) RequestBlockHeaders(header *block.MetaBlock) int {
	return mp.requestBlockHeaders(header)
}

func (mp *metaProcessor) WaitForBlockHeaders(waitTime time.Duration) {
	mp.waitForBlockHeaders(waitTime)
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
		mp.allNeededShardHdrsFound = false
	}

	if mp.currHighestShardHdrNonce == nil {
		mp.currHighestShardHdrNonce = make(map[uint32]uint64, mp.shardCoordinator.NumberOfShards())
		for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
			mp.currHighestShardHdrNonce[i] = uint64(0)
		}
	}

	mp.requestedShardHdrsHashes[string(hdrHash)] = true
}

func (mp *metaProcessor) IsHdrHashRequested(hdrHash []byte) bool {
	mp.mutRequestedShardHdrsHashes.Lock()
	defer mp.mutRequestedShardHdrsHashes.Unlock()

	_, found := mp.requestedShardHdrsHashes[string(hdrHash)]

	return found
}

func (mp *metaProcessor) CreateShardInfo(maxMiniBlocksInBlock uint32, round uint32, haveTime func() bool) ([]block.ShardData, error) {
	return mp.createShardInfo(maxMiniBlocksInBlock, round, haveTime)
}

func (bp *baseProcessor) LastNotarizedHdrs() map[uint32]data.HeaderHandler {
	return bp.lastNotarizedHdrs
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

func (mp *metaProcessor) CreateLastNotarizedHdrs(header *block.MetaBlock) error {
	return mp.createLastNotarizedHdrs(header)
}

func (mp *metaProcessor) CheckShardHeadersValidity(header *block.MetaBlock) (mapShardLastHeaders, error) {
	return mp.checkShardHeadersValidity(header)
}

func (mp *metaProcessor) CheckShardHeadersFinality(header *block.MetaBlock, highestNonceHdrs mapShardLastHeaders) error {
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

func NewBaseProcessor(shardCord sharding.Coordinator) *baseProcessor {
	return &baseProcessor{shardCoordinator: shardCord}
}

func (bp *baseProcessor) SaveLastNotarizedHeader(shardId uint32, processedHdrs []data.HeaderHandler) error {
	return bp.saveLastNotarizedHeader(shardId, processedHdrs)
}

func (sp *shardProcessor) CheckHeaderBodyCorrelation(hdr *block.Header, body block.Body) error {
	return sp.checkHeaderBodyCorrelation(hdr, body)
}

func (bp *baseProcessor) SetLastNotarizedHeadersSlice(startHeaders map[uint32]data.HeaderHandler, metaChainActive bool) error {
	return bp.setLastNotarizedHeadersSlice(startHeaders, metaChainActive)
}

func (sp *shardProcessor) CreateTxInfo(tx *transaction.Transaction, senderShardID uint32, receiverShardID uint32) *txInfo {
	txShardInfo := &txShardInfo{senderShardID: senderShardID, receiverShardID: receiverShardID}
	return &txInfo{tx: tx, txShardInfo: txShardInfo}
}

func (sp *shardProcessor) SetTxsForBlock(hash string, txInfo *txInfo) {
	sp.mutTxsForBlock.Lock()
	sp.txsForBlock[hash] = txInfo
	sp.mutTxsForBlock.Unlock()
}
