package track

import (
	"bytes"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
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

	blockProcessor                blockProcessorHandler
	crossNotarizer                blockNotarizerHandler
	selfNotarizer                 blockNotarizerHandler
	crossNotarizedHeadersNotifier blockNotifierHandler
	selfNotarizedHeadersNotifier  blockNotifierHandler

	mutHeaders sync.RWMutex
	headers    map[uint32]map[uint64][]*headerInfo
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
		lastNotarizedHeaderNonce = bbt.selfNotarizer.getLastNotarizedHeaderNonce(header.GetShardID())
	} else {
		lastNotarizedHeaderNonce = bbt.crossNotarizer.getLastNotarizedHeaderNonce(header.GetShardID())
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

	bbt.mutHeaders.Lock()
	defer bbt.mutHeaders.Unlock()

	shardID := header.GetShardID()
	nonce := header.GetNonce()

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

// AddCrossNotarizedHeader adds cross notarized header to the tracker lists
func (bbt *baseBlockTrack) AddCrossNotarizedHeader(
	shardID uint32,
	crossNotarizedHeader data.HeaderHandler,
	crossNotarizedHeaderHash []byte,
) {
	bbt.crossNotarizer.addNotarizedHeader(shardID, crossNotarizedHeader, crossNotarizedHeaderHash)
}

// AddSelfNotarizedHeader adds self notarized headers to the tracker lists
func (bbt *baseBlockTrack) AddSelfNotarizedHeader(
	shardID uint32,
	selfNotarizedHeader data.HeaderHandler,
	selfNotarizedHeaderHash []byte,
) {
	bbt.selfNotarizer.addNotarizedHeader(shardID, selfNotarizedHeader, selfNotarizedHeaderHash)
}

// AddTrackedHeader adds tracked headers to the tracker lists
func (bbt *baseBlockTrack) AddTrackedHeader(header data.HeaderHandler, hash []byte) {
	bbt.addHeader(header, hash)
}

// CleanupHeadersBehindNonce removes from local pools old headers for a given shard
func (bbt *baseBlockTrack) CleanupHeadersBehindNonce(
	shardID uint32,
	selfNotarizedNonce uint64,
	crossNotarizedNonce uint64,
) {
	bbt.selfNotarizer.cleanupNotarizedHeadersBehindNonce(shardID, selfNotarizedNonce)
	nonce := selfNotarizedNonce

	if shardID != bbt.shardCoordinator.SelfId() {
		bbt.crossNotarizer.cleanupNotarizedHeadersBehindNonce(shardID, crossNotarizedNonce)
		nonce = crossNotarizedNonce
	}

	bbt.cleanupTrackedHeadersBehindNonce(shardID, nonce)
}

func (bbt *baseBlockTrack) cleanupTrackedHeadersBehindNonce(shardID uint32, nonce uint64) {
	if nonce == 0 {
		return
	}

	bbt.mutHeaders.Lock()
	defer bbt.mutHeaders.Unlock()

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
func (bbt *baseBlockTrack) ComputeLongestChain(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte) {
	return bbt.blockProcessor.computeLongestChain(shardID, header)
}

// DisplayTrackedHeaders displays tracked headers
func (bbt *baseBlockTrack) DisplayTrackedHeaders() {
	for shardID := uint32(0); shardID < bbt.shardCoordinator.NumberOfShards(); shardID++ {
		bbt.displayHeadersForShard(shardID)
	}

	bbt.displayHeadersForShard(core.MetachainShardId)
}

func (bbt *baseBlockTrack) displayHeadersForShard(shardID uint32) {
	bbt.displayTrackedHeadersForShard(shardID, "tracked headers")
	bbt.crossNotarizer.displayNotarizedHeaders(shardID, "cross notarized headers")
	bbt.selfNotarizer.displayNotarizedHeaders(shardID, "self notarized headers")
}

func (bbt *baseBlockTrack) displayTrackedHeadersForShard(shardID uint32, message string) {
	headers, hashes := bbt.sortHeadersFromNonce(shardID, 0)
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

// GetCrossNotarizedHeader returns a cross notarized header for a given shard with a given offset, behind last cross notarized header
func (bbt *baseBlockTrack) GetCrossNotarizedHeader(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
	return bbt.crossNotarizer.getNotarizedHeader(shardID, offset)
}

// GetLastCrossNotarizedHeader returns last cross notarized header for a given shard
func (bbt *baseBlockTrack) GetLastCrossNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	return bbt.crossNotarizer.getLastNotarizedHeader(shardID)
}

// GetTrackedHeaders returns tracked headers for a given shard
func (bbt *baseBlockTrack) GetTrackedHeaders(shardID uint32) ([]data.HeaderHandler, [][]byte) {
	return bbt.sortHeadersFromNonce(shardID, 0)
}

func (bbt *baseBlockTrack) sortHeadersFromNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
	bbt.mutHeaders.RLock()
	defer bbt.mutHeaders.RUnlock()

	headersForShard, ok := bbt.headers[shardID]
	if !ok {
		return nil, nil
	}

	sortedHeadersInfo := make([]*headerInfo, 0)

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

// GetTrackedHeadersWithNonce returns tracked headers for a given shard and nonce
func (bbt *baseBlockTrack) GetTrackedHeadersWithNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
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
	header := bbt.getLastHeader(shardId)
	if check.IfNil(header) {
		return false
	}

	isShardStuck := bbt.rounder.Index()-int64(header.GetRound()) >= process.MaxRoundsWithoutCommittedBlock
	return isShardStuck
}

func (bbt *baseBlockTrack) getLastHeader(shardID uint32) data.HeaderHandler {
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

// RemoveLastNotarizedHeaders removes last notarized headers from tracker list
func (bbt *baseBlockTrack) RemoveLastNotarizedHeaders() {
	bbt.crossNotarizer.removeLastNotarizedHeader()
	bbt.selfNotarizer.removeLastNotarizedHeader()
}

// RestoreHeadersToGenesis restores notarized headers from tracker lists to genesis
func (bbt *baseBlockTrack) RestoreHeadersToGenesis() {
	bbt.crossNotarizer.restoreNotarizedHeadersToGenesis()
	bbt.selfNotarizer.restoreNotarizedHeadersToGenesis()
	bbt.restoreTrackedHeadersToGenesis()
}

func (bbt *baseBlockTrack) restoreTrackedHeadersToGenesis() {
	bbt.mutHeaders.Lock()
	bbt.headers = make(map[uint32]map[uint64][]*headerInfo)
	bbt.mutHeaders.Unlock()
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

func (bbt *baseBlockTrack) initNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	err := bbt.crossNotarizer.initNotarizedHeaders(startHeaders)
	if err != nil {
		return err
	}

	err = bbt.selfNotarizer.initNotarizedHeaders(startHeaders)
	if err != nil {
		return err
	}

	return nil
}
