package track

import (
	"bytes"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.GetOrCreate("process/track")

type headerInfo struct {
	hash   []byte
	header data.HeaderHandler
}

type baseBlockTrack struct {
	hasher           hashing.Hasher
	headerValidator  process.HeaderConstructionValidator
	marshalizer      marshal.Marshalizer
	rounder          consensus.Rounder
	shardCoordinator sharding.Coordinator
	headersPool      dataRetriever.HeadersPool
	store            dataRetriever.StorageService

	blockProcessor                blockProcessorHandler
	crossNotarizer                blockNotarizerHandler
	selfNotarizer                 blockNotarizerHandler
	crossNotarizedHeadersNotifier blockNotifierHandler
	selfNotarizedHeadersNotifier  blockNotifierHandler

	mutHeaders sync.RWMutex
	headers    map[uint32]map[uint64][]*headerInfo
}

func (bbt *baseBlockTrack) receivedHeader(headerHandler data.HeaderHandler, headerHash []byte) {
	if headerHandler.GetShardID() == sharding.MetachainShardId {
		bbt.receivedMetaBlock(headerHandler, headerHash)
		return
	}

	bbt.receivedShardHeader(headerHandler, headerHash)
}

func (bbt *baseBlockTrack) receivedShardHeader(headerHandler data.HeaderHandler, shardHeaderHash []byte) {
	shardHeader, ok := headerHandler.(*block.Header)
	if !ok {
		log.Warn("cannot convert data.HeaderHandler in *block.Header")
		return
	}

	log.Debug("received shard header from network in block tracker",
		"shard", shardHeader.GetShardID(),
		"round", shardHeader.GetRound(),
		"nonce", shardHeader.GetNonce(),
		"hash", shardHeaderHash,
	)

	bbt.addHeader(shardHeader, shardHeaderHash)
	bbt.blockProcessor.processReceivedHeader(shardHeader)
}

func (bbt *baseBlockTrack) receivedMetaBlock(headerHandler data.HeaderHandler, metaBlockHash []byte) {
	metaBlock, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		log.Warn("cannot convert data.HeaderHandler in *block.Metablock")
		return
	}

	log.Debug("received meta block from network in block tracker",
		"shard", metaBlock.GetShardID(),
		"round", metaBlock.GetRound(),
		"nonce", metaBlock.GetNonce(),
		"hash", metaBlockHash,
	)

	bbt.addHeader(metaBlock, metaBlockHash)
	bbt.blockProcessor.processReceivedHeader(metaBlock)
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

// ComputeLongestMetaChainFromLastNotarized returns the longest valid chain for metachain from its last cross notarized header
func (bbt *baseBlockTrack) ComputeLongestMetaChainFromLastNotarized() ([]data.HeaderHandler, [][]byte, error) {
	lastCrossNotarizedHeader, _, err := bbt.GetLastCrossNotarizedHeader(sharding.MetachainShardId)
	if err != nil {
		return nil, nil, err
	}

	hdrsForShard, hdrsHashesForShard := bbt.ComputeLongestChain(sharding.MetachainShardId, lastCrossNotarizedHeader)

	return hdrsForShard, hdrsHashesForShard, nil
}

// ComputeLongestShardsChainsFromLastNotarized returns the longest valid chains for all shards from theirs last cross notarized headers
func (bbt *baseBlockTrack) ComputeLongestShardsChainsFromLastNotarized() ([]data.HeaderHandler, [][]byte, map[uint32][]data.HeaderHandler, error) {
	hdrsMap := make(map[uint32][]data.HeaderHandler)
	hdrsHashesMap := make(map[uint32][][]byte)

	lastCrossNotarizedHeaders, err := bbt.GetLastCrossNotarizedHeadersForAllShards()
	if err != nil {
		return nil, nil, nil, err
	}

	maxHdrLen := 0
	for shardID := uint32(0); shardID < bbt.shardCoordinator.NumberOfShards(); shardID++ {
		hdrsForShard, hdrsHashesForShard := bbt.ComputeLongestChain(shardID, lastCrossNotarizedHeaders[shardID])

		hdrsMap[shardID] = append(hdrsMap[shardID], hdrsForShard...)
		hdrsHashesMap[shardID] = append(hdrsHashesMap[shardID], hdrsHashesForShard...)

		tmpHdrLen := len(hdrsForShard)
		if maxHdrLen < tmpHdrLen {
			maxHdrLen = tmpHdrLen
		}
	}

	orderedHeaders := make([]data.HeaderHandler, 0)
	orderedHeadersHashes := make([][]byte, 0)

	// copy from map to lists - equality between number of headers per shard
	for i := 0; i < maxHdrLen; i++ {
		for shardID := uint32(0); shardID < bbt.shardCoordinator.NumberOfShards(); shardID++ {
			hdrsForShard := hdrsMap[shardID]
			hdrsHashesForShard := hdrsHashesMap[shardID]
			if i >= len(hdrsForShard) {
				continue
			}

			orderedHeaders = append(orderedHeaders, hdrsForShard[i])
			orderedHeadersHashes = append(orderedHeadersHashes, hdrsHashesForShard[i])
		}
	}

	return orderedHeaders, orderedHeadersHashes, hdrsMap, nil
}

// DisplayTrackedHeaders displays tracked headers
func (bbt *baseBlockTrack) DisplayTrackedHeaders() {
	for shardID := uint32(0); shardID < bbt.shardCoordinator.NumberOfShards(); shardID++ {
		bbt.displayHeadersForShard(shardID)
	}

	bbt.displayHeadersForShard(sharding.MetachainShardId)
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

// GetLastCrossNotarizedHeadersForAllShards returns last cross notarized headers for all shards
func (bbt *baseBlockTrack) GetLastCrossNotarizedHeadersForAllShards() (map[uint32]data.HeaderHandler, error) {
	lastCrossNotarizedHeaders := make(map[uint32]data.HeaderHandler, bbt.shardCoordinator.NumberOfShards())

	// save last committed header for verification
	for shardID := uint32(0); shardID < bbt.shardCoordinator.NumberOfShards(); shardID++ {
		lastCrossNotarizedHeader, _, err := bbt.GetLastCrossNotarizedHeader(shardID)
		if err != nil {
			return nil, err
		}

		lastCrossNotarizedHeaders[shardID] = lastCrossNotarizedHeader
	}

	return lastCrossNotarizedHeaders, nil
}

// GetTrackedHeaders returns tracked headers for a given shard
func (bbt *baseBlockTrack) GetTrackedHeaders(shardID uint32) ([]data.HeaderHandler, [][]byte) {
	return bbt.sortHeadersFromNonce(shardID, 0)
}

// GetTrackedHeadersForAllShards returns tracked headers for all shards
func (bbt *baseBlockTrack) GetTrackedHeadersForAllShards() map[uint32][]data.HeaderHandler {
	trackedHeaders := make(map[uint32][]data.HeaderHandler)

	for shardID := uint32(0); shardID < bbt.shardCoordinator.NumberOfShards(); shardID++ {
		trackedHeadersForShard, _ := bbt.GetTrackedHeaders(shardID)
		trackedHeaders[shardID] = append(trackedHeaders[shardID], trackedHeadersForShard...)
	}

	return trackedHeaders
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

// RestoreToGenesis sets class variables to theirs initial values
func (bbt *baseBlockTrack) RestoreToGenesis() {
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
	if check.IfNil(arguments.RequestHandler) {
		return process.ErrNilRequestHandler
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
