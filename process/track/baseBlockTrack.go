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
	marshalizer       marshal.Marshalizer
	rounder           consensus.Rounder
	shardCoordinator  sharding.Coordinator
	metaBlocksPool    storage.Cacher
	shardHeadersPool  storage.Cacher
	headersNoncesPool dataRetriever.Uint64SyncMapCacher

	mutHeaders sync.RWMutex
	headers    map[uint32]map[uint64][]*headerInfo

	mutCrossNotarizedHeaders sync.RWMutex
	crossNotarizedHeaders    map[uint32][]*headerInfo

	mutSelfNotarizedHeaders sync.RWMutex
	selfNotarizedHeaders    map[uint32][]*headerInfo

	mutSelfNotarizedHeadersHandlers sync.RWMutex
	selfNotarizedHeadersHandlers    []func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)

	mutCrossNotarizedHeadersHandlers sync.RWMutex
	crossNotarizedHeadersHandlers    []func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)

	blockFinality uint64

	blockTracker blockTracker
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
	if check.IfNil(crossNotarizedHeader) {
		return
	}

	bbt.mutCrossNotarizedHeaders.Lock()
	bbt.crossNotarizedHeaders[shardID] = append(bbt.crossNotarizedHeaders[shardID], &headerInfo{header: crossNotarizedHeader, hash: crossNotarizedHeaderHash})
	if len(bbt.crossNotarizedHeaders[shardID]) > 1 {
		sort.Slice(bbt.crossNotarizedHeaders[shardID], func(i, j int) bool {
			return bbt.crossNotarizedHeaders[shardID][i].header.GetNonce() < bbt.crossNotarizedHeaders[shardID][j].header.GetNonce()
		})
	}
	bbt.mutCrossNotarizedHeaders.Unlock()
}

// AddSelfNotarizedHeader adds self notarized headers to the tracker lists
func (bbt *baseBlockTrack) AddSelfNotarizedHeader(
	shardID uint32,
	selfNotarizedHeader data.HeaderHandler,
	selfNotarizedHeaderHash []byte,
) {
	if check.IfNil(selfNotarizedHeader) {
		return
	}

	bbt.mutSelfNotarizedHeaders.Lock()
	bbt.selfNotarizedHeaders[shardID] = append(bbt.selfNotarizedHeaders[shardID], &headerInfo{header: selfNotarizedHeader, hash: selfNotarizedHeaderHash})
	if len(bbt.selfNotarizedHeaders[shardID]) > 1 {
		sort.Slice(bbt.selfNotarizedHeaders[shardID], func(i, j int) bool {
			return bbt.selfNotarizedHeaders[shardID][i].header.GetNonce() < bbt.selfNotarizedHeaders[shardID][j].header.GetNonce()
		})
	}
	bbt.mutSelfNotarizedHeaders.Unlock()
}

// CleanupHeadersForShardBehindNonce removes from local pools old headers for a given shard
func (bbt *baseBlockTrack) CleanupHeadersForShardBehindNonce(
	shardID uint32,
	selfNotarizedNonce uint64,
	crossNotarizedNonce uint64,
) {
	bbt.cleanupSelfNotarizedBlocksForShardBehindNonce(shardID, selfNotarizedNonce)
	nonce := selfNotarizedNonce

	if shardID != bbt.shardCoordinator.SelfId() {
		bbt.cleanupCrossNotarizedBlocksForShardBehindNonce(shardID, crossNotarizedNonce)
		nonce = crossNotarizedNonce
	}

	bbt.cleanupHeadersForShardBehindNonce(shardID, nonce)
}

func (bbt *baseBlockTrack) cleanupSelfNotarizedBlocksForShardBehindNonce(shardID uint32, nonce uint64) {
	bbt.mutSelfNotarizedHeaders.Lock()
	defer bbt.mutSelfNotarizedHeaders.Unlock()

	if nonce == 0 {
		return
	}

	selfNotarizedHeadersForShard, ok := bbt.selfNotarizedHeaders[shardID]
	if !ok {
		return
	}

	headersInfo := make([]*headerInfo, 0)
	for _, headerInfo := range selfNotarizedHeadersForShard {
		if headerInfo.header.GetNonce() < nonce {
			continue
		}

		headersInfo = append(headersInfo, headerInfo)
	}

	bbt.selfNotarizedHeaders[shardID] = headersInfo
}

func (bbt *baseBlockTrack) cleanupCrossNotarizedBlocksForShardBehindNonce(shardID uint32, nonce uint64) {
	bbt.mutCrossNotarizedHeaders.Lock()
	defer bbt.mutCrossNotarizedHeaders.Unlock()

	if nonce == 0 {
		return
	}

	crossNotarizedHeadersForShard, ok := bbt.crossNotarizedHeaders[shardID]
	if !ok {
		return
	}

	headersInfo := make([]*headerInfo, 0)
	for _, headerInfo := range crossNotarizedHeadersForShard {
		if headerInfo.header.GetNonce() < nonce {
			continue
		}

		headersInfo = append(headersInfo, headerInfo)
	}

	bbt.crossNotarizedHeaders[shardID] = headersInfo
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
func (bbt *baseBlockTrack) ComputeLongestChain(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte) {
	headers := make([]data.HeaderHandler, 0)
	headersHashes := make([][]byte, 0)

	if check.IfNil(header) {
		return headers, headersHashes
	}

	sortedHeaders, sortedHeadersHashes := bbt.sortHeadersForShardFromNonce(shardID, header.GetNonce()+1)
	if len(sortedHeaders) == 0 {
		return headers, headersHashes
	}

	longestChainHeadersIndexes := make([]int, 0)
	headersIndexes := make([]int, 0)
	bbt.getNextHeader(&longestChainHeadersIndexes, headersIndexes, header, sortedHeaders, 0)

	for _, index := range longestChainHeadersIndexes {
		headers = append(headers, sortedHeaders[index])
		headersHashes = append(headersHashes, sortedHeadersHashes[index])
	}

	return headers, headersHashes
}

func (bbt *baseBlockTrack) sortHeadersForShardFromNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
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

func (bbt *baseBlockTrack) getNextHeader(
	longestChainHeadersIndexes *[]int,
	headersIndexes []int,
	prevHeader data.HeaderHandler,
	sortedHeaders []data.HeaderHandler,
	index int,
) {
	defer func() {
		if len(headersIndexes) > len(*longestChainHeadersIndexes) {
			*longestChainHeadersIndexes = headersIndexes
		}
	}()

	if check.IfNil(prevHeader) {
		return
	}

	for i := index; i < len(sortedHeaders); i++ {
		currHeader := sortedHeaders[i]
		if currHeader.GetNonce() > prevHeader.GetNonce()+1 {
			break
		}

		if currHeader.GetNonce() == prevHeader.GetNonce()+1 {
			err := bbt.isHeaderConstructionValid(currHeader, prevHeader)
			if err != nil {
				continue
			}

			err = bbt.checkHeaderFinality(currHeader, sortedHeaders, i+1)
			if err != nil {
				continue
			}

			headersIndexes = append(headersIndexes, i)
			bbt.getNextHeader(longestChainHeadersIndexes, headersIndexes, currHeader, sortedHeaders, i+1)
			headersIndexes = headersIndexes[:len(headersIndexes)-1]
		}
	}
}

func (bbt *baseBlockTrack) isHeaderConstructionValid(currHeader, prevHeader data.HeaderHandler) error {
	if check.IfNil(prevHeader) {
		return process.ErrNilBlockHeader
	}
	if check.IfNil(currHeader) {
		return process.ErrNilBlockHeader
	}

	if prevHeader.GetRound() >= currHeader.GetRound() {
		log.Trace("round does not match",
			"shard", currHeader.GetShardID(),
			"local header round", prevHeader.GetRound(),
			"received round", currHeader.GetRound())
		return process.ErrLowerRoundInBlock
	}

	if currHeader.GetNonce() != prevHeader.GetNonce()+1 {
		log.Trace("nonce does not match",
			"shard", currHeader.GetShardID(),
			"local header nonce", prevHeader.GetNonce(),
			"received nonce", currHeader.GetNonce())
		return process.ErrWrongNonceInBlock
	}

	prevHash, err := core.CalculateHash(bbt.marshalizer, bbt.hasher, prevHeader)
	if err != nil {
		return err
	}

	if !bytes.Equal(currHeader.GetPrevHash(), prevHash) {
		log.Trace("header hash does not match",
			"shard", currHeader.GetShardID(),
			"local header hash", prevHash,
			"received header with prev hash", currHeader.GetPrevHash(),
		)
		return process.ErrBlockHashDoesNotMatch
	}

	if !bytes.Equal(currHeader.GetPrevRandSeed(), prevHeader.GetRandSeed()) {
		log.Trace("header random seed does not match",
			"shard", currHeader.GetShardID(),
			"local header random seed", prevHeader.GetRandSeed(),
			"received header with prev random seed", currHeader.GetPrevRandSeed(),
		)
		return process.ErrRandSeedDoesNotMatch
	}

	return nil
}

func (bbt *baseBlockTrack) checkHeaderFinality(
	header data.HeaderHandler,
	sortedHeaders []data.HeaderHandler,
	index int,
) error {

	if check.IfNil(header) {
		return process.ErrNilBlockHeader
	}

	prevHeader := header
	numFinalityAttestingHeaders := uint64(0)

	for i := index; i < len(sortedHeaders); i++ {
		currHeader := sortedHeaders[i]
		if numFinalityAttestingHeaders >= bbt.blockFinality || currHeader.GetNonce() > prevHeader.GetNonce()+1 {
			break
		}

		if currHeader.GetNonce() == prevHeader.GetNonce()+1 {
			err := bbt.isHeaderConstructionValid(currHeader, prevHeader)
			if err != nil {
				continue
			}

			prevHeader = currHeader
			numFinalityAttestingHeaders += 1
		}
	}

	if numFinalityAttestingHeaders < bbt.blockFinality {
		return process.ErrHeaderNotFinal
	}

	return nil
}

// DisplayTrackedHeaders displays tracked headers
func (bbt *baseBlockTrack) DisplayTrackedHeaders() {
	for shardID := uint32(0); shardID < bbt.shardCoordinator.NumberOfShards(); shardID++ {
		bbt.displayHeadersForShard(shardID)
	}

	bbt.displayHeadersForShard(sharding.MetachainShardId)
}

// GetLastCrossNotarizedHeader returns last cross notarized header for a given shard
func (bbt *baseBlockTrack) GetLastCrossNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	bbt.mutCrossNotarizedHeaders.RLock()
	defer bbt.mutCrossNotarizedHeaders.RUnlock()

	if bbt.crossNotarizedHeaders == nil {
		return nil, nil, process.ErrCrossNotarizedHdrsSliceIsNil
	}

	headerInfo := bbt.lastCrossNotarizedHdrForShard(shardID)
	if headerInfo == nil {
		return nil, nil, process.ErrCrossNotarizedHdrsSliceForShardIsNil
	}

	return headerInfo.header, headerInfo.hash, nil
}

// GetTrackedHeadersForShard returns tracked headers for a given shard
func (bbt *baseBlockTrack) GetTrackedHeadersForShard(shardID uint32) ([]data.HeaderHandler, [][]byte) {
	return bbt.sortHeadersForShardFromNonce(shardID, 0)
}

// GetTrackedHeadersForShardWithNonce returns tracked headers for a given shard and nonce
func (bbt *baseBlockTrack) GetTrackedHeadersForShardWithNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
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

// RemoveLastCrossNotarizedHeader removes last cross notarized header from tracker list
func (bbt *baseBlockTrack) RemoveLastCrossNotarizedHeader() {
	bbt.mutCrossNotarizedHeaders.Lock()
	for shardID := range bbt.crossNotarizedHeaders {
		notarizedHdrsCount := len(bbt.crossNotarizedHeaders[shardID])
		if notarizedHdrsCount > 1 {
			bbt.crossNotarizedHeaders[shardID] = bbt.crossNotarizedHeaders[shardID][:notarizedHdrsCount-1]
		}
	}
	bbt.mutCrossNotarizedHeaders.Unlock()
}

// RestoreHeadersToGenesis restores notarized tracker lists to genesis
func (bbt *baseBlockTrack) RestoreHeadersToGenesis() {
	bbt.restoreCrossNotarizedHeadersToGenesis()
	bbt.restoreSelfNotarizedHeadersToGenesis()
	bbt.restoreTrackedHeadersToGenesis()
}

// IsInterfaceNil returns true if there is no value under the interface
func (bbt *baseBlockTrack) IsInterfaceNil() bool {
	return bbt == nil
}

func checkTrackerNilParameters(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
) error {

	if check.IfNil(hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(rounder) {
		return process.ErrNilRounder
	}
	if check.IfNil(shardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(store) {
		return process.ErrNilStorage
	}

	return nil
}
