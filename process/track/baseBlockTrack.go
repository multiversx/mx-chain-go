package track

import (
	"bytes"
	"sync"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
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
	marshalizer      marshal.Marshalizer
	rounder          consensus.Rounder
	shardCoordinator sharding.Coordinator

	mutHeaders sync.RWMutex
	headers    map[uint32]map[uint64][]*headerInfo

	mutNotarizedHeaders sync.RWMutex
	notarizedHeaders    map[uint32][]data.HeaderHandler

	mutFinalizedHeaders sync.RWMutex
	finalizedHeaders    map[uint32][]data.HeaderHandler

	mutLongestChainHeadersIndexes sync.RWMutex
	longestChainHeadersIndexes    []int

	blockFinality uint32

	blockTracker blockTracker
}

// AddHeader adds the given header to the received headers list
func (bbt *baseBlockTrack) AddHeader(header data.HeaderHandler, hash []byte) {
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

// IsShardStuck returns true if the given shard is stuck
func (bbt *baseBlockTrack) IsShardStuck(shardId uint32) bool {
	header := bbt.LastHeaderForShard(shardId)
	if check.IfNil(header) {
		return false
	}

	isShardStuck := bbt.rounder.Index()-int64(header.GetRound()) >= process.MaxRoundsWithoutCommittedBlock
	return isShardStuck
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

// IsInterfaceNil returns true if there is no value under the interface
func (bbt *baseBlockTrack) IsInterfaceNil() bool {
	return bbt == nil
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

func (bbt *baseBlockTrack) setNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	bbt.mutNotarizedHeaders.Lock()
	defer bbt.mutNotarizedHeaders.Unlock()

	if startHeaders == nil {
		return process.ErrNotarizedHdrsSliceIsNil
	}

	bbt.notarizedHeaders = make(map[uint32][]data.HeaderHandler)

	for _, startHeader := range startHeaders {
		bbt.notarizedHeaders[startHeader.GetShardID()] = append(bbt.notarizedHeaders[startHeader.GetShardID()], startHeader)
	}

	//for i := uint32(0); i < bbt.shardCoordinator.NumberOfShards(); i++ {
	//	header, ok := startHeaders[i].(*block.Header)
	//	if !ok {
	//		return process.ErrWrongTypeAssertion
	//	}
	//	bbt.notarizedHeaders[i] = append(bbt.notarizedHeaders[i], header)
	//}
	//
	//metaBlock, ok := startHeaders[sharding.MetachainShardId].(*block.MetaBlock)
	//if !ok {
	//	return process.ErrWrongTypeAssertion
	//}
	//bbt.notarizedHeaders[sharding.MetachainShardId] = append(bbt.notarizedHeaders[sharding.MetachainShardId], metaBlock)

	return nil
}

func (bbt *baseBlockTrack) setFinalizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	bbt.mutFinalizedHeaders.Lock()
	defer bbt.mutFinalizedHeaders.Unlock()

	if startHeaders == nil {
		return process.ErrFinalizedHdrsSliceIsNil
	}

	bbt.finalizedHeaders = make(map[uint32][]data.HeaderHandler)

	for _, startHeader := range startHeaders {
		bbt.finalizedHeaders[startHeader.GetShardID()] = append(bbt.finalizedHeaders[startHeader.GetShardID()], startHeader)
	}

	//for i := uint32(0); i < bbt.shardCoordinator.NumberOfShards(); i++ {
	//	header, ok := startHeaders[i].(*block.Header)
	//	if !ok {
	//		return process.ErrWrongTypeAssertion
	//	}
	//	bbt.finalizedHeaders[i] = append(bbt.finalizedHeaders[i], header)
	//}
	//
	//metaBlock, ok := startHeaders[sharding.MetachainShardId].(*block.MetaBlock)
	//if !ok {
	//	return process.ErrWrongTypeAssertion
	//}
	//bbt.finalizedHeaders[sharding.MetachainShardId] = append(bbt.finalizedHeaders[sharding.MetachainShardId], metaBlock)

	return nil
}

func (bbt *baseBlockTrack) addNotarizedHeaders(notarizedHeaders []data.HeaderHandler) {
	bbt.mutNotarizedHeaders.Lock()
	defer bbt.mutNotarizedHeaders.Unlock()

	for _, notarizedHeader := range notarizedHeaders {
		bbt.notarizedHeaders[notarizedHeader.GetShardID()] = append(bbt.notarizedHeaders[notarizedHeader.GetShardID()], notarizedHeader)
	}
}

func (bbt *baseBlockTrack) addFinalizedHeaders(finalizedHeaders []data.HeaderHandler) {
	bbt.mutFinalizedHeaders.Lock()
	defer bbt.mutFinalizedHeaders.Unlock()

	for _, finalizedHeader := range finalizedHeaders {
		bbt.finalizedHeaders[finalizedHeader.GetShardID()] = append(bbt.finalizedHeaders[finalizedHeader.GetShardID()], finalizedHeader)
	}
}

func (bbt *baseBlockTrack) getLastNotarizedHeader(shardID uint32) (data.HeaderHandler, error) {
	bbt.mutNotarizedHeaders.RLock()
	defer bbt.mutNotarizedHeaders.RUnlock()

	if bbt.notarizedHeaders == nil {
		return nil, process.ErrNotarizedHdrsSliceIsNil
	}

	headerHandler := bbt.lastNotarizedHdrForShard(shardID)
	if check.IfNil(headerHandler) {
		return nil, process.ErrNotarizedHdrsSliceForShardIsNil
	}

	//err := bbt.checkHeaderTypeCorrect(shardID, headerHandler)
	//if err != nil {
	//	return nil, err
	//}

	return headerHandler, nil
}

func (bbt *baseBlockTrack) lastNotarizedHdrForShard(shardID uint32) data.HeaderHandler {
	notarizedHeadersCount := len(bbt.notarizedHeaders[shardID])
	if notarizedHeadersCount > 0 {
		return bbt.notarizedHeaders[shardID][notarizedHeadersCount-1]
	}

	return nil
}

func (bbt *baseBlockTrack) getLastFinalizedHeader(shardID uint32) (data.HeaderHandler, error) {
	bbt.mutFinalizedHeaders.RLock()
	defer bbt.mutFinalizedHeaders.RUnlock()

	if bbt.finalizedHeaders == nil {
		return nil, process.ErrFinalizedHdrsSliceIsNil
	}

	headerHandler := bbt.lastFinalizedHdrForShard(shardID)
	if check.IfNil(headerHandler) {
		return nil, process.ErrFinalizedHdrsSliceForShardIsNil
	}

	//err := bbt.checkHeaderTypeCorrect(shardID, headerHandler)
	//if err != nil {
	//	return nil, err
	//}

	return headerHandler, nil
}

func (bbt *baseBlockTrack) lastFinalizedHdrForShard(shardID uint32) data.HeaderHandler {
	finalizedHeadersCount := len(bbt.finalizedHeaders[shardID])
	if finalizedHeadersCount > 0 {
		return bbt.finalizedHeaders[shardID][finalizedHeadersCount-1]
	}

	return nil
}

//func (bbt *baseBlockTrack) checkHeaderTypeCorrect(shardID uint32, headerHandler data.HeaderHandler) error {
//	if shardID >= bbt.shardCoordinator.NumberOfShards() && shardID != sharding.MetachainShardId {
//		return process.ErrShardIdMissmatch
//	}
//
//	if shardID < bbt.shardCoordinator.NumberOfShards() {
//		_, ok := headerHandler.(*block.Header)
//		if !ok {
//			return process.ErrWrongTypeAssertion
//		}
//	}
//
//	if shardID == sharding.MetachainShardId {
//		_, ok := headerHandler.(*block.MetaBlock)
//		if !ok {
//			return process.ErrWrongTypeAssertion
//		}
//	}
//
//	return nil
//}

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

func (bbt *baseBlockTrack) isHeaderConstructionValid(currHeader, prevHeader data.HeaderHandler) error {
	if check.IfNil(prevHeader) {
		return process.ErrNilBlockHeader
	}
	if check.IfNil(currHeader) {
		return process.ErrNilBlockHeader
	}

	//// special case with genesis nonce - 0
	//if currHeader.GetNonce() == 0 {
	//	if prevHeader.GetNonce() != 0 {
	//		return process.ErrWrongNonceInBlock
	//	}
	//	if prevHeader.GetRootHash() != nil {
	//		return process.ErrRootStateDoesNotMatch
	//	}
	//
	//	return nil
	//}

	if prevHeader.GetRound() >= currHeader.GetRound() {
		log.Trace("round does not match",
			"shard", currHeader.GetShardID(),
			"local block round", prevHeader.GetRound(),
			"received round", currHeader.GetRound())
		return process.ErrLowerRoundInBlock
	}

	if currHeader.GetNonce() != prevHeader.GetNonce()+1 {
		log.Trace("nonce does not match",
			"shard", currHeader.GetShardID(),
			"local block nonce", prevHeader.GetNonce(),
			"received nonce", currHeader.GetNonce())
		return process.ErrWrongNonceInBlock
	}

	prevHash, err := core.CalculateHash(bbt.marshalizer, bbt.hasher, prevHeader)
	if err != nil {
		return err
	}

	if !bytes.Equal(currHeader.GetPrevHash(), prevHash) {
		log.Trace("block hash does not match",
			"shard", currHeader.GetShardID(),
			"local prev hash", prevHash,
			"received block with prev hash", currHeader.GetPrevHash(),
		)
		return process.ErrBlockHashDoesNotMatch
	}

	if !bytes.Equal(currHeader.GetPrevRandSeed(), prevHeader.GetRandSeed()) {
		log.Trace("random seed does not match",
			"shard", currHeader.GetShardID(),
			"local rand seed", prevHeader.GetRandSeed(),
			"received block with rand seed", currHeader.GetPrevRandSeed(),
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
	numFinalityAttestingHeaders := uint32(0)

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

func checkTrackerNilParameters(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
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

	return nil
}

func (bbt *baseBlockTrack) computeLongestChainFromLastNotarized(shardID uint32) []data.HeaderHandler {
	lastNotarizedHeader, err := bbt.getLastNotarizedHeader(shardID)
	if err != nil {
		return nil
	}

	return bbt.ComputeLongestChain(shardID, lastNotarizedHeader)
}

func (bbt *baseBlockTrack) doJobOnReceivedCrossNotarizedBlock(shardID uint32) {
	notarizedHeaders := bbt.computeLongestChainFromLastNotarized(shardID)
	finalizedHeaders := bbt.blockTracker.computeFinalizedHeaders(notarizedHeaders)
	bbt.addNotarizedHeaders(notarizedHeaders)
	bbt.addFinalizedHeaders(finalizedHeaders)

	nonce := bbt.getFinalNonce(shardID)
	bbt.cleanupHeadersForShardBehindNonce(shardID, nonce)
	bbt.cleanupCrossNotarizedBlocksForShardBehindNonce(shardID, nonce)
	bbt.displayHeadersForShard(shardID)
}

func (bbt *baseBlockTrack) doJobOnReceivedBlock(shardID uint32) {

}

func (bbt *baseBlockTrack) cleanupCrossNotarizedBlocksForShardBehindNonce(shardID uint32, nonce uint64) {
	bbt.mutNotarizedHeaders.Lock()
	defer bbt.mutNotarizedHeaders.Unlock()

	notarizedHeadersForShard, ok := bbt.notarizedHeaders[shardID]
	if !ok {
		return
	}

	headers := make([]data.HeaderHandler, 0)
	for _, header := range notarizedHeadersForShard {
		if header.GetNonce() < nonce {
			continue
		}

		headers = append(headers, header)
	}

	bbt.notarizedHeaders[shardID] = headers
}

func (bbt *baseBlockTrack) cleanupHeadersForShardBehindNonce(shardID uint32, nonce uint64) {
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

//TODO: final nonce should be calculated differently
func (bbt *baseBlockTrack) getFinalNonce(shardID uint32) uint64 {
	lastNotarizedHeader, err := bbt.getLastNotarizedHeader(shardID)
	if err != nil {
		return 0
	}

	return lastNotarizedHeader.GetNonce()
}

func (bbt *baseBlockTrack) displayHeaders() {
	for shardID := range bbt.headers {
		bbt.displayHeadersForShard(shardID)
	}
}

func (bbt *baseBlockTrack) displayHeadersForShard(shardID uint32) {
	//lastHeader, err := bbt.getLastNotarizedHeader(shardID)
	//if err != nil {
	//	log.Debug("get last notarized header", "error", err.Error())
	//	return
	//}

	log.Debug("############ tracked headers ############", "shard", shardID)

	//headers := bbt.sortHeadersForShardFromNonce(shardID, lastHeader.GetNonce())
	headers := bbt.sortHeadersForShardFromNonce(shardID, 0)
	for _, header := range headers {
		log.Debug("traked header info",
			"round", header.GetRound(),
			"nonce", header.GetNonce())
	}

	log.Debug("############ notarized headers ############", "shard", shardID)

	bbt.mutNotarizedHeaders.RLock()

	notarizedHeadersForShard, ok := bbt.notarizedHeaders[shardID]
	if ok {
		for _, header := range notarizedHeadersForShard {
			log.Debug("notarized header info",
				"round", header.GetRound(),
				"nonce", header.GetNonce())
		}
	}

	bbt.mutNotarizedHeaders.RUnlock()

	log.Debug("############ finalized headers ############", "shard", shardID)

	bbt.mutFinalizedHeaders.RLock()

	finalizedHeadersForShard, ok := bbt.finalizedHeaders[shardID]
	if ok {
		for _, header := range finalizedHeadersForShard {
			log.Debug("finalized header info",
				"round", header.GetRound(),
				"nonce", header.GetNonce())
		}
	}

	bbt.mutFinalizedHeaders.RUnlock()

	log.Debug("#########################################")
}
