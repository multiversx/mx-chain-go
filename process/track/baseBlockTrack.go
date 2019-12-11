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
	headersNoncesPool dataRetriever.Uint64SyncMapCacher

	mutHeaders sync.RWMutex
	headers    map[uint32]map[uint64][]*headerInfo

	mutCrossNotarizedHeaders sync.RWMutex
	crossNotarizedHeaders    map[uint32][]data.HeaderHandler

	mutSelfNotarizedHeaders sync.RWMutex
	selfNotarizedHeaders    map[uint32][]data.HeaderHandler

	mutSelfNotarizedHandlers sync.RWMutex
	selfNotarizedHandlers    []func(headers []data.HeaderHandler)

	mutLongestChainHeadersIndexes sync.RWMutex
	longestChainHeadersIndexes    []int

	blockFinality uint32

	blockTracker blockTracker
}

// addHeader adds the given header to the received headers list
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

// RegisterSelfNotarizedHandler registers a new handler to be called when self notarized is changed
func (bbt *baseBlockTrack) RegisterSelfNotarizedHandler(handler func(headers []data.HeaderHandler)) {
	if handler == nil {
		log.Debug("attempt to register a nil handler to a tracker object")
		return
	}

	bbt.mutSelfNotarizedHandlers.Lock()
	bbt.selfNotarizedHandlers = append(bbt.selfNotarizedHandlers, handler)
	bbt.mutSelfNotarizedHandlers.Unlock()
}

func (bbt *baseBlockTrack) callSelfNotarizedHandlers(headers []data.HeaderHandler) {
	bbt.mutSelfNotarizedHandlers.RLock()
	for _, handler := range bbt.selfNotarizedHandlers {
		go handler(headers)
	}
	bbt.mutSelfNotarizedHandlers.RUnlock()
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

func (bbt *baseBlockTrack) setCrossNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	bbt.mutCrossNotarizedHeaders.Lock()
	defer bbt.mutCrossNotarizedHeaders.Unlock()

	if startHeaders == nil {
		return process.ErrCrossNotarizedHdrsSliceIsNil
	}

	bbt.crossNotarizedHeaders = make(map[uint32][]data.HeaderHandler)

	for _, startHeader := range startHeaders {
		shardID := startHeader.GetShardID()
		bbt.crossNotarizedHeaders[shardID] = append(bbt.crossNotarizedHeaders[shardID], startHeader)
	}

	//for i := uint32(0); i < bbt.shardCoordinator.NumberOfShards(); i++ {
	//	header, ok := startHeaders[i].(*block.Header)
	//	if !ok {
	//		return process.ErrWrongTypeAssertion
	//	}
	//	bbt.crossNotarizedHeaders[i] = append(bbt.crossNotarizedHeaders[i], header)
	//}
	//
	//metaBlock, ok := startHeaders[sharding.MetachainShardId].(*block.MetaBlock)
	//if !ok {
	//	return process.ErrWrongTypeAssertion
	//}
	//bbt.crossNotarizedHeaders[sharding.MetachainShardId] = append(bbt.crossNotarizedHeaders[sharding.MetachainShardId], metaBlock)

	return nil
}

func (bbt *baseBlockTrack) setSelfNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	bbt.mutSelfNotarizedHeaders.Lock()
	defer bbt.mutSelfNotarizedHeaders.Unlock()

	if startHeaders == nil {
		return process.ErrSelfNotarizedHdrsSliceIsNil
	}

	bbt.selfNotarizedHeaders = make(map[uint32][]data.HeaderHandler)

	for _, startHeader := range startHeaders {
		shardID := startHeader.GetShardID()
		bbt.selfNotarizedHeaders[shardID] = append(bbt.selfNotarizedHeaders[shardID], startHeader)
	}

	//for i := uint32(0); i < bbt.shardCoordinator.NumberOfShards(); i++ {
	//	header, ok := startHeaders[i].(*block.Header)
	//	if !ok {
	//		return process.ErrWrongTypeAssertion
	//	}
	//	bbt.selfNotarizedHeaders[i] = append(bbt.selfNotarizedHeaders[i], header)
	//}
	//
	//metaBlock, ok := startHeaders[sharding.MetachainShardId].(*block.MetaBlock)
	//if !ok {
	//	return process.ErrWrongTypeAssertion
	//}
	//bbt.selfNotarizedHeaders[sharding.MetachainShardId] = append(bbt.selfNotarizedHeaders[sharding.MetachainShardId], metaBlock)

	return nil
}

func (bbt *baseBlockTrack) addCrossNotarizedHeaders(shardID uint32, crossNotarizedHeaders []data.HeaderHandler) {
	bbt.mutCrossNotarizedHeaders.Lock()
	defer bbt.mutCrossNotarizedHeaders.Unlock()

	for _, crossNotarizedHeader := range crossNotarizedHeaders {
		bbt.crossNotarizedHeaders[shardID] = append(bbt.crossNotarizedHeaders[shardID], crossNotarizedHeader)
	}
}

func (bbt *baseBlockTrack) addSelfNotarizedHeaders(shardID uint32, selfNotarizedHeaders []data.HeaderHandler) {
	bbt.mutSelfNotarizedHeaders.Lock()
	defer bbt.mutSelfNotarizedHeaders.Unlock()

	for _, selfNotarizedHeader := range selfNotarizedHeaders {
		bbt.selfNotarizedHeaders[shardID] = append(bbt.selfNotarizedHeaders[shardID], selfNotarizedHeader)
	}
}

func (bbt *baseBlockTrack) getLastCrossNotarizedHeader(shardID uint32) (data.HeaderHandler, error) {
	bbt.mutCrossNotarizedHeaders.RLock()
	defer bbt.mutCrossNotarizedHeaders.RUnlock()

	if bbt.crossNotarizedHeaders == nil {
		return nil, process.ErrCrossNotarizedHdrsSliceIsNil
	}

	headerHandler := bbt.lastCrossNotarizedHdrForShard(shardID)
	if check.IfNil(headerHandler) {
		return nil, process.ErrCrossNotarizedHdrsSliceForShardIsNil
	}

	//err := bbt.checkHeaderTypeCorrect(shardID, headerHandler)
	//if err != nil {
	//	return nil, err
	//}

	return headerHandler, nil
}

func (bbt *baseBlockTrack) lastCrossNotarizedHdrForShard(shardID uint32) data.HeaderHandler {
	crossNotarizedHeadersCount := len(bbt.crossNotarizedHeaders[shardID])
	if crossNotarizedHeadersCount > 0 {
		return bbt.crossNotarizedHeaders[shardID][crossNotarizedHeadersCount-1]
	}

	return nil
}

func (bbt *baseBlockTrack) getLastSelfNotarizedHeader(shardID uint32) (data.HeaderHandler, error) {
	bbt.mutSelfNotarizedHeaders.RLock()
	defer bbt.mutSelfNotarizedHeaders.RUnlock()

	if bbt.selfNotarizedHeaders == nil {
		return nil, process.ErrSelfNotarizedHdrsSliceIsNil
	}

	headerHandler := bbt.lastSelfNotarizedHdrForShard(shardID)
	if check.IfNil(headerHandler) {
		return nil, process.ErrSelfNotarizedHdrsSliceForShardIsNil
	}

	//err := bbt.checkHeaderTypeCorrect(shardID, headerHandler)
	//if err != nil {
	//	return nil, err
	//}

	return headerHandler, nil
}

func (bbt *baseBlockTrack) lastSelfNotarizedHdrForShard(shardID uint32) data.HeaderHandler {
	selfNotarizedHeadersCount := len(bbt.selfNotarizedHeaders[shardID])
	if selfNotarizedHeadersCount > 0 {
		return bbt.selfNotarizedHeaders[shardID][selfNotarizedHeadersCount-1]
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

func (bbt *baseBlockTrack) computeLongestChainFromLastCrossNotarized(shardID uint32) []data.HeaderHandler {
	lastCrossNotarizedHeader, err := bbt.getLastCrossNotarizedHeader(shardID)
	if err != nil {
		return nil
	}

	return bbt.ComputeLongestChain(shardID, lastCrossNotarizedHeader)
}

func (bbt *baseBlockTrack) computeLongestChainFromLastSelfNotarized(shardID uint32) []data.HeaderHandler {
	lastSelfNotarizedHeader, err := bbt.getLastSelfNotarizedHeader(shardID)
	if err != nil {
		return nil
	}

	return bbt.ComputeLongestChain(shardID, lastSelfNotarizedHeader)
}

func (bbt *baseBlockTrack) doJobOnReceivedCrossNotarizedBlock(shardID uint32) {
	crossNotarizedHeaders := bbt.computeLongestChainFromLastCrossNotarized(shardID)
	selfNotarizedHeaders := bbt.blockTracker.computeSelfNotarizedHeaders(crossNotarizedHeaders)

	bbt.addCrossNotarizedHeaders(shardID, crossNotarizedHeaders)
	bbt.addSelfNotarizedHeaders(shardID, selfNotarizedHeaders)

	//TODO: Should be analyzed if this nonce should be calculated differently
	nonce := bbt.getLastCrossNotarizedHeaderNonce(shardID)
	bbt.cleanupCrossNotarizedBlocksForShardBehindNonce(shardID, nonce)

	bbt.cleanupHeadersForShardBehindNonce(shardID, nonce)

	//TODO: Should be analyzed if this nonce should be calculated differently
	nonce = bbt.getLastSelfNotarizedHeaderNonce(shardID)
	bbt.cleanupSelfNotarizedBlocksForShardBehindNonce(shardID, nonce)

	log.Debug("############ display on received cross notarized block", "shard", shardID)
	bbt.displayHeadersForShard(shardID)
}

func (bbt *baseBlockTrack) doJobOnReceivedBlock(shardID uint32) {
	selfNotarizedHeaders := bbt.computeLongestChainFromLastSelfNotarized(shardID)
	bbt.addSelfNotarizedHeaders(shardID, selfNotarizedHeaders)

	//TODO: Should be analyzed if this nonce should be calculated differently
	nonce := bbt.getLastSelfNotarizedHeaderNonce(shardID)
	bbt.cleanupSelfNotarizedBlocksForShardBehindNonce(shardID, nonce)

	bbt.blockTracker.cleanupHeadersForSelfShard()

	log.Debug("############ display on received block", "shard", shardID)
	bbt.displayHeadersForShard(shardID)
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

func (bbt *baseBlockTrack) cleanupCrossNotarizedBlocksForShardBehindNonce(shardID uint32, nonce uint64) {
	bbt.mutCrossNotarizedHeaders.Lock()
	defer bbt.mutCrossNotarizedHeaders.Unlock()

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

func (bbt *baseBlockTrack) cleanupSelfNotarizedBlocksForShardBehindNonce(shardID uint32, nonce uint64) {
	bbt.mutSelfNotarizedHeaders.Lock()
	defer bbt.mutSelfNotarizedHeaders.Unlock()

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

func (bbt *baseBlockTrack) getLastCrossNotarizedHeaderNonce(shardID uint32) uint64 {
	lastCrossNotarizedHeader, err := bbt.getLastCrossNotarizedHeader(shardID)
	if err != nil {
		return 0
	}

	return lastCrossNotarizedHeader.GetNonce()
}

func (bbt *baseBlockTrack) getLastSelfNotarizedHeaderNonce(shardID uint32) uint64 {
	lastSelfNotarizedHeader, err := bbt.getLastSelfNotarizedHeader(shardID)
	if err != nil {
		return 0
	}

	return lastSelfNotarizedHeader.GetNonce()
}

func (bbt *baseBlockTrack) displayHeaders() {
	for shardID := range bbt.headers {
		bbt.displayHeadersForShard(shardID)
	}
}

func (bbt *baseBlockTrack) displayHeadersForShard(shardID uint32) {
	bbt.displayTrackedHeadersForShard(shardID)
	bbt.displayCrossNotarizedHeadersForShard(shardID)
	bbt.displaySelfNotarizedHeadersForShard(shardID)
}

func (bbt *baseBlockTrack) displayTrackedHeadersForShard(shardID uint32) {
	log.Debug(">>>>>>>>>> tracked headers", "shard", shardID)

	headers := bbt.sortHeadersForShardFromNonce(shardID, 0)
	for _, header := range headers {
		log.Debug("traked header info",
			"round", header.GetRound(),
			"nonce", header.GetNonce())
	}
}

func (bbt *baseBlockTrack) displayCrossNotarizedHeadersForShard(shardID uint32) {
	log.Debug(">>>>>>>>>> cross notarized headers", "shard", shardID)

	bbt.mutCrossNotarizedHeaders.RLock()

	crossNotarizedHeadersForShard, ok := bbt.crossNotarizedHeaders[shardID]
	if ok {
		for _, header := range crossNotarizedHeadersForShard {
			log.Debug("cross notarized header info",
				"round", header.GetRound(),
				"nonce", header.GetNonce())
		}
	}

	bbt.mutCrossNotarizedHeaders.RUnlock()
}

func (bbt *baseBlockTrack) displaySelfNotarizedHeadersForShard(shardID uint32) {
	log.Debug(">>>>>>>>>> self notarized headers", "shard", shardID)

	bbt.mutSelfNotarizedHeaders.RLock()

	selfNotarizedHeadersForShard, ok := bbt.selfNotarizedHeaders[shardID]
	if ok {
		for _, header := range selfNotarizedHeadersForShard {
			log.Debug("self notarized header info",
				"round", header.GetRound(),
				"nonce", header.GetNonce())
		}
	}

	bbt.mutSelfNotarizedHeaders.RUnlock()
}

func (bbt *baseBlockTrack) isHeaderOutOfRange(receivedNonce uint64, lastNotarizedNonce uint64) bool {
	isHeaderOutOfRange := receivedNonce > lastNotarizedNonce+process.MaxNonceDifferences
	return isHeaderOutOfRange
}

func (bbt *baseBlockTrack) addMissingShardHeaders(
	shardID uint32,
	fromNonce uint64,
	toNonce uint64,
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
) {

	for nonce := fromNonce; nonce <= toNonce; nonce++ {
		header, headerHash, err := process.GetShardHeaderFromPoolWithNonce(nonce, shardID, cacher, uint64SyncMapCacher)
		if err != nil {
			log.Trace("GetShardHeaderFromPoolWithNonce", "error", err.Error())
			return
		}

		log.Debug("get missing shard header",
			"shard", header.GetShardID(),
			"round", header.GetRound(),
			"nonce", header.GetNonce(),
			"hash", headerHash,
		)

		bbt.addHeader(header, headerHash)
	}
}

func (bbt *baseBlockTrack) addMissingMetaBlocks(
	fromNonce uint64,
	toNonce uint64,
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
) {

	for nonce := fromNonce; nonce <= toNonce; nonce++ {
		metaBlock, metaBlockHash, err := process.GetMetaHeaderFromPoolWithNonce(nonce, cacher, uint64SyncMapCacher)
		if err != nil {
			log.Trace("GetMetaHeaderFromPoolWithNonce", "error", err.Error())
			return
		}

		log.Debug("get missing meta block",
			"shard", metaBlock.GetShardID(),
			"round", metaBlock.GetRound(),
			"nonce", metaBlock.GetNonce(),
			"hash", metaBlockHash,
		)

		bbt.addHeader(metaBlock, metaBlockHash)
	}
}
