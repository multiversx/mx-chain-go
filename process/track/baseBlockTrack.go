package track

import (
	"bytes"
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
	crossNotarizedHeaders    map[uint32][]data.HeaderHandler

	mutSelfNotarizedHeaders sync.RWMutex
	selfNotarizedHeaders    map[uint32][]data.HeaderHandler

	mutLongestChainHeadersIndexes sync.RWMutex
	longestChainHeadersIndexes    []int

	mutSelfNotarizedHeadersHandlers sync.RWMutex
	selfNotarizedHeadersHandlers    []func(headers []data.HeaderHandler, headersHashes [][]byte)

	blockFinality uint64

	blockTracker blockTracker
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

	headers := make([]data.HeaderHandler, 0)
	for _, header := range selfNotarizedHeadersForShard {
		if header.GetNonce() < nonce {
			continue
		}

		headers = append(headers, header)
	}

	bbt.selfNotarizedHeaders[shardID] = headers
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

	headers := make([]data.HeaderHandler, 0)
	for _, header := range crossNotarizedHeadersForShard {
		if header.GetNonce() < nonce {
			continue
		}

		headers = append(headers, header)
	}

	bbt.crossNotarizedHeaders[shardID] = headers
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
func (bbt *baseBlockTrack) ComputeLongestChain(shardID uint32, header data.HeaderHandler) []data.HeaderHandler {
	headers := make([]data.HeaderHandler, 0)

	sortedHeaders := bbt.sortHeadersForShardFromNonce(shardID, header.GetNonce()+1)
	if len(sortedHeaders) == 0 {
		return headers
	}

	bbt.initLongestChainHeadersIndexes()
	headersIndexes := make([]int, 0)
	bbt.getNextHeader(headersIndexes, header, sortedHeaders, 0)

	longestChainHeadersIndexes := bbt.getLongestChainHeadersIndexes()
	for _, index := range longestChainHeadersIndexes {
		headers = append(headers, sortedHeaders[index])
	}

	return headers
}

func (bbt *baseBlockTrack) initLongestChainHeadersIndexes() {
	bbt.mutLongestChainHeadersIndexes.Lock()
	bbt.longestChainHeadersIndexes = make([]int, 0)
	bbt.mutLongestChainHeadersIndexes.Unlock()
}

func (bbt *baseBlockTrack) getLongestChainHeadersIndexes() []int {
	bbt.mutLongestChainHeadersIndexes.RLock()
	longestChainHeadersIndexes := bbt.longestChainHeadersIndexes
	bbt.mutLongestChainHeadersIndexes.RUnlock()

	return longestChainHeadersIndexes
}

func (bbt *baseBlockTrack) setLongestChainHeadersIndexes(longestChainHeadersIndexes []int) {
	bbt.mutLongestChainHeadersIndexes.Lock()
	bbt.longestChainHeadersIndexes = longestChainHeadersIndexes
	bbt.mutLongestChainHeadersIndexes.Unlock()
}

func (bbt *baseBlockTrack) sortHeadersForShardFromNonce(shardID uint32, nonce uint64) []data.HeaderHandler {
	bbt.mutHeaders.RLock()
	defer bbt.mutHeaders.RUnlock()

	headersForShard, ok := bbt.headers[shardID]
	if !ok {
		return nil
	}

	headers := make([]data.HeaderHandler, 0)
	for headersNonce, headersInfo := range headersForShard {
		if headersNonce < nonce {
			continue
		}

		for _, headerInfo := range headersInfo {
			headers = append(headers, headerInfo.header)
		}
	}

	process.SortHeadersByNonce(headers)

	return headers
}

func (bbt *baseBlockTrack) getNextHeader(
	headersIndexes []int,
	prevHeader data.HeaderHandler,
	sortedHeaders []data.HeaderHandler,
	index int,
) {
	defer func() {
		if len(headersIndexes) > len(bbt.getLongestChainHeadersIndexes()) {
			bbt.setLongestChainHeadersIndexes(headersIndexes)
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
			bbt.getNextHeader(headersIndexes, currHeader, sortedHeaders, i+1)
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

// IsShardStuck returns true if the given shard is stuck
func (bbt *baseBlockTrack) IsShardStuck(shardId uint32) bool {
	header := bbt.LastHeaderForShard(shardId)
	if check.IfNil(header) {
		return false
	}

	isShardStuck := bbt.rounder.Index()-int64(header.GetRound()) >= process.MaxRoundsWithoutCommittedBlock
	return isShardStuck
}

// LastHeaderForShard returns the last header received (highest round) for the given shard
func (bbt *baseBlockTrack) LastHeaderForShard(shardID uint32) data.HeaderHandler {
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
