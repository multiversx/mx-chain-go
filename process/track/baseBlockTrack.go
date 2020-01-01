package track

import (
	"bytes"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("process/track")

type headerInfo struct {
	hash   []byte
	header data.HeaderHandler
}

type baseBlockTrack struct {
	hasher            hashing.Hasher
	headerValidator   process.HeaderConstructionValidator
	marshalizer       marshal.Marshalizer
	rounder           consensus.Rounder
	shardCoordinator  sharding.Coordinator
	metaBlocksPool    storage.Cacher
	shardHeadersPool  storage.Cacher
	headersNoncesPool dataRetriever.Uint64SyncMapCacher
	store             dataRetriever.StorageService

	mutHeaders sync.RWMutex
	headers    map[uint32]map[uint64][]*headerInfo

	blockProcessor        blockProcessorHandler
	crossNotarizedHeaders blockNotarizerHandler
	selfNotarizedHeaders  blockNotarizerHandler

	crossNotarizedHeadersNotifier blockNotifierHandler
	selfNotarizedHeadersNotifier  blockNotifierHandler
}

func (bbt *baseBlockTrack) receivedShardHeader(shardHeaderHash []byte) {
	shardHeader, err := process.GetShardHeaderFromPool(shardHeaderHash, bbt.shardHeadersPool)
	if err != nil {
		log.Trace("GetShardHeaderFromPool", "error", err.Error())
		return
	}

	log.Debug("received shard header from network in block tracker",
		"shard", shardHeader.GetShardID(),
		"round", shardHeader.GetRound(),
		"nonce", shardHeader.GetNonce(),
		"hash", shardHeaderHash,
	)

	if bbt.isHeaderOutOfRange(shardHeader) {
		return
	}

	bbt.addHeader(shardHeader, shardHeaderHash)
	bbt.blockProcessor.processReceivedHeader(shardHeader)
}

func (bbt *baseBlockTrack) receivedMetaBlock(metaBlockHash []byte) {
	metaBlock, err := process.GetMetaHeaderFromPool(metaBlockHash, bbt.metaBlocksPool)
	if err != nil {
		log.Trace("GetMetaHeaderFromPool", "error", err.Error())
		return
	}

	log.Debug("received meta block from network in block tracker",
		"shard", metaBlock.GetShardID(),
		"round", metaBlock.GetRound(),
		"nonce", metaBlock.GetNonce(),
		"hash", metaBlockHash,
	)

	if bbt.isHeaderOutOfRange(metaBlock) {
		return
	}

	bbt.addHeader(metaBlock, metaBlockHash)
	bbt.blockProcessor.processReceivedHeader(metaBlock)
}

func (bbt *baseBlockTrack) isHeaderOutOfRange(header data.HeaderHandler) bool {
	var lastNotarizedHeaderNonce uint64

	isHeaderForSelfShard := header.GetShardID() == bbt.shardCoordinator.SelfId()
	if isHeaderForSelfShard {
		lastNotarizedHeaderNonce = bbt.selfNotarizedHeaders.getLastNotarizedHeaderNonce(header.GetShardID())
	} else {
		lastNotarizedHeaderNonce = bbt.crossNotarizedHeaders.getLastNotarizedHeaderNonce(header.GetShardID())
	}

	if header.GetNonce() > lastNotarizedHeaderNonce+process.MaxNonceDifferences {
		log.Debug("received header is out of range",
			"received nonce", header.GetNonce(),
			"last notarized nonce", lastNotarizedHeaderNonce,
		)
		return true
	}

	return false
}

func (bbt *baseBlockTrack) addHeader(header data.HeaderHandler, hash []byte) {
	if check.IfNil(header) {
		return
	}

	shardID := header.GetShardID()
	nonce := header.GetNonce()

	bbt.mutHeaders.Lock()
	defer bbt.mutHeaders.Unlock()

	headersForShard, ok := bbt.headers[shardID]
	if !ok {
		headersForShard = make(map[uint64][]*headerInfo)
		bbt.headers[shardID] = headersForShard
	}

	for _, headerInfo := range headersForShard[nonce] {
		if bytes.Equal(headerInfo.hash, hash) {
			return
		}
	}

	headersForShard[nonce] = append(headersForShard[nonce], &headerInfo{hash: hash, header: header})
}

// AddTrackedHeader adds header to the tracker lists
func (bbt *baseBlockTrack) AddTrackedHeader(header data.HeaderHandler, hash []byte) {
	bbt.addHeader(header, hash)
}

// AddCrossNotarizedHeader adds cross notarized header to the tracker lists
func (bbt *baseBlockTrack) AddCrossNotarizedHeader(
	shardID uint32,
	crossNotarizedHeader data.HeaderHandler,
	crossNotarizedHeaderHash []byte,
) {
	bbt.crossNotarizedHeaders.addNotarizedHeader(shardID, crossNotarizedHeader, crossNotarizedHeaderHash)
}

// AddSelfNotarizedHeader adds self notarized headers to the tracker lists
func (bbt *baseBlockTrack) AddSelfNotarizedHeader(
	shardID uint32,
	selfNotarizedHeader data.HeaderHandler,
	selfNotarizedHeaderHash []byte,
) {
	bbt.selfNotarizedHeaders.addNotarizedHeader(shardID, selfNotarizedHeader, selfNotarizedHeaderHash)
}

// CleanupHeadersForShardBehindNonce removes from local pools old headers for a given shard
func (bbt *baseBlockTrack) CleanupHeadersForShardBehindNonce(
	shardID uint32,
	selfNotarizedNonce uint64,
	crossNotarizedNonce uint64,
) {
	bbt.selfNotarizedHeaders.cleanupNotarizedHeadersBehindNonce(shardID, selfNotarizedNonce)
	nonce := selfNotarizedNonce

	if shardID != bbt.shardCoordinator.SelfId() {
		bbt.crossNotarizedHeaders.cleanupNotarizedHeadersBehindNonce(shardID, crossNotarizedNonce)
		nonce = crossNotarizedNonce
	}

	bbt.cleanupHeadersForShardBehindNonce(shardID, nonce)
}

func (bbt *baseBlockTrack) cleanupHeadersForShardBehindNonce(shardID uint32, nonce uint64) {
	bbt.mutHeaders.Lock()
	defer bbt.mutHeaders.Unlock()

	if nonce == 0 {
		return
	}

	headersForShard, ok := bbt.headers[shardID]
	if !ok {
		return
	}

	for headersNonce := range headersForShard {
		if headersNonce < nonce {
			delete(headersForShard, headersNonce)
		}
	}
}

// ComputeLongestChain returns the longest valid chain for a given shard from a given header
func (bbt *baseBlockTrack) ComputeLongestChain(
	shardID uint32,
	header data.HeaderHandler,
) ([]data.HeaderHandler, [][]byte) {

	return bbt.blockProcessor.computeLongestChain(shardID, header)
}

// DisplayTrackedHeaders displays tracked headers
func (bbt *baseBlockTrack) DisplayTrackedHeaders() {
	for shardID := uint32(0); shardID < bbt.shardCoordinator.NumberOfShards(); shardID++ {
		bbt.displayHeadersForShard(shardID)
	}

	bbt.displayHeadersForShard(sharding.MetachainShardId)
}

// GetCrossNotarizedHeader returns a cross notarized header for a given shard with a given offset, behind last cross notarized header
func (bbt *baseBlockTrack) GetCrossNotarizedHeader(
	shardID uint32,
	offset uint64,
) (data.HeaderHandler, []byte, error) {

	return bbt.crossNotarizedHeaders.getNotarizedHeader(shardID, offset)
}

// GetLastCrossNotarizedHeader returns last cross notarized header for a given shard
func (bbt *baseBlockTrack) GetLastCrossNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	return bbt.crossNotarizedHeaders.getLastNotarizedHeader(shardID)
}

// GetTrackedHeadersForShard returns tracked headers for a given shard
func (bbt *baseBlockTrack) GetTrackedHeadersForShard(shardID uint32) ([]data.HeaderHandler, [][]byte) {
	return bbt.sortHeadersForShardFromNonce(shardID, 0)
}

// GetTrackedHeadersForShardWithNonce returns tracked headers for a given shard and nonce
func (bbt *baseBlockTrack) GetTrackedHeadersForShardWithNonce(
	shardID uint32,
	nonce uint64,
) ([]data.HeaderHandler, [][]byte) {

	bbt.mutHeaders.RLock()
	defer bbt.mutHeaders.RUnlock()

	headersForShard, ok := bbt.headers[shardID]
	if !ok {
		return nil, nil
	}

	headersForShardWithNonce, ok := headersForShard[nonce]
	if !ok {
		return nil, nil
	}

	headers := make([]data.HeaderHandler, 0)
	headersHashes := make([][]byte, 0)

	for _, headerInfo := range headersForShardWithNonce {
		headers = append(headers, headerInfo.header)
		headersHashes = append(headersHashes, headerInfo.hash)
	}

	return headers, headersHashes
}

// IsShardStuck returns true if the given shard is stuck
func (bbt *baseBlockTrack) IsShardStuck(shardId uint32) bool {
	header := bbt.lastHeaderForShard(shardId)
	if check.IfNil(header) {
		return false
	}

	isShardStuck := bbt.rounder.Index()-int64(header.GetRound()) >= process.MaxRoundsWithoutCommittedBlock
	return isShardStuck
}

// LastHeaderForShard returns the last header received (highest round) for the given shard
func (bbt *baseBlockTrack) lastHeaderForShard(shardID uint32) data.HeaderHandler {
	bbt.mutHeaders.RLock()
	defer bbt.mutHeaders.RUnlock()

	var lastHeaderForShard data.HeaderHandler

	headersForShard, ok := bbt.headers[shardID]
	if !ok {
		return lastHeaderForShard
	}

	maxRound := uint64(0)
	for _, headersInfo := range headersForShard {
		for _, headerInfo := range headersInfo {
			if headerInfo.header.GetRound() > maxRound {
				maxRound = headerInfo.header.GetRound()
				lastHeaderForShard = headerInfo.header
			}
		}
	}

	return lastHeaderForShard
}

// RegisterCrossNotarizedHeadersHandler registers a new handler to be called when cross notarized header is changed
func (bbt *baseBlockTrack) RegisterCrossNotarizedHeadersHandler(
	handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte),
) {
	bbt.crossNotarizedHeadersNotifier.registerHandler(handler)
}

// RegisterSelfNotarizedHeadersHandler registers a new handler to be called when self notarized header is changed
func (bbt *baseBlockTrack) RegisterSelfNotarizedHeadersHandler(
	handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte),
) {
	bbt.selfNotarizedHeadersNotifier.registerHandler(handler)
}

// RemoveLastCrossNotarizedHeader removes last cross notarized header from tracker list
func (bbt *baseBlockTrack) RemoveLastCrossNotarizedHeader() {
	bbt.crossNotarizedHeaders.removeLastNotarizedHeader()
}

// RemoveLastSelfNotarizedHeader removes last self notarized header from tracker list
func (bbt *baseBlockTrack) RemoveLastSelfNotarizedHeader() {
	bbt.selfNotarizedHeaders.removeLastNotarizedHeader()
}

// RestoreHeadersToGenesis restores notarized tracker lists to genesis
func (bbt *baseBlockTrack) RestoreHeadersToGenesis() {
	bbt.crossNotarizedHeaders.restoreNotarizedHeadersToGenesis()
	bbt.selfNotarizedHeaders.restoreNotarizedHeadersToGenesis()
	bbt.restoreTrackedHeadersToGenesis()
}

// IsInterfaceNil returns true if there is no value under the interface
func (bbt *baseBlockTrack) IsInterfaceNil() bool {
	return bbt == nil
}

func checkTrackerNilParameters(arguments ArgBaseTracker) error {

	if check.IfNil(arguments.Hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(arguments.HeaderValidator) {
		return process.ErrNilHeaderValidator
	}
	if check.IfNil(arguments.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(arguments.Rounder) {
		return process.ErrNilRounder
	}
	if check.IfNil(arguments.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(arguments.Store) {
		return process.ErrNilStorage
	}

	return nil
}

func (bbt *baseBlockTrack) restoreTrackedHeadersToGenesis() {
	bbt.mutHeaders.Lock()
	bbt.headers = make(map[uint32]map[uint64][]*headerInfo)
	bbt.mutHeaders.Unlock()
}

func (bbt *baseBlockTrack) displayHeadersForShard(shardID uint32) {
	bbt.displayTrackedHeadersForShard(shardID, "tracked headers")
	bbt.crossNotarizedHeaders.displayNotarizedHeaders(shardID, "cross notarized headers")
	bbt.selfNotarizedHeaders.displayNotarizedHeaders(shardID, "self notarized headers")
}

func (bbt *baseBlockTrack) displayTrackedHeadersForShard(shardID uint32, message string) {
	headers, hashes := bbt.sortHeadersForShardFromNonce(shardID, 0)
	shouldNotDisplay := len(headers) == 0 ||
		len(headers) == 1 && headers[0].GetNonce() == 0
	if shouldNotDisplay {
		return
	}

	log.Debug(message,
		"shard", shardID,
		"nb", len(headers))

	for index, header := range headers {
		log.Trace("tracked header info",
			"round", header.GetRound(),
			"nonce", header.GetNonce(),
			"hash", hashes[index])
	}
}

func (bbt *baseBlockTrack) sortHeadersForShardFromNonce(
	shardID uint32,
	nonce uint64,
) ([]data.HeaderHandler, [][]byte) {

	bbt.mutHeaders.RLock()
	defer bbt.mutHeaders.RUnlock()

	sortedHeadersInfo := make([]*headerInfo, 0)

	headersForShard, ok := bbt.headers[shardID]
	if !ok {
		return nil, nil
	}

	for headersNonce, headersInfo := range headersForShard {
		if headersNonce < nonce {
			continue
		}

		sortedHeadersInfo = append(sortedHeadersInfo, headersInfo...)
	}

	if len(sortedHeadersInfo) > 1 {
		sort.Slice(sortedHeadersInfo, func(i, j int) bool {
			return sortedHeadersInfo[i].header.GetNonce() < sortedHeadersInfo[j].header.GetNonce()
		})
	}

	headers := make([]data.HeaderHandler, 0)
	headersHashes := make([][]byte, 0)

	for _, headerInfo := range sortedHeadersInfo {
		headers = append(headers, headerInfo.header)
		headersHashes = append(headersHashes, headerInfo.hash)
	}

	return headers, headersHashes
}
