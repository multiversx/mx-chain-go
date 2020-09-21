package track

import (
	"bytes"
	"fmt"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ process.ValidityAttester = (*baseBlockTrack)(nil)

var log = logger.GetOrCreate("process/track")

// HeaderInfo holds the information about a header
type HeaderInfo struct {
	Hash   []byte
	Header data.HeaderHandler
}

type baseBlockTrack struct {
	hasher           hashing.Hasher
	headerValidator  process.HeaderConstructionValidator
	marshalizer      marshal.Marshalizer
	rounder          process.Rounder
	shardCoordinator sharding.Coordinator
	headersPool      dataRetriever.HeadersPool
	store            dataRetriever.StorageService

	blockProcessor                        blockProcessorHandler
	crossNotarizer                        blockNotarizerHandler
	selfNotarizer                         blockNotarizerHandler
	crossNotarizedHeadersNotifier         blockNotifierHandler
	selfNotarizedFromCrossHeadersNotifier blockNotifierHandler
	selfNotarizedHeadersNotifier          blockNotifierHandler
	finalMetachainHeadersNotifier         blockNotifierHandler
	blockBalancer                         blockBalancerHandler
	whitelistHandler                      process.WhiteListHandler

	mutHeaders                  sync.RWMutex
	headers                     map[uint32]map[uint64][]*HeaderInfo
	maxNumHeadersToKeepPerShard int
}

func createBaseBlockTrack(arguments ArgBaseTracker) (*baseBlockTrack, error) {
	err := checkTrackerNilParameters(arguments)
	if err != nil {
		return nil, err
	}

	maxNumHeadersToKeepPerShard := arguments.PoolsHolder.Headers().MaxSize()

	crossNotarizer, err := NewBlockNotarizer(arguments.Hasher, arguments.Marshalizer, arguments.ShardCoordinator)
	if err != nil {
		return nil, err
	}

	selfNotarizer, err := NewBlockNotarizer(arguments.Hasher, arguments.Marshalizer, arguments.ShardCoordinator)
	if err != nil {
		return nil, err
	}

	crossNotarizedHeadersNotifier, err := NewBlockNotifier()
	if err != nil {
		return nil, err
	}

	selfNotarizedFromCrossHeadersNotifier, err := NewBlockNotifier()
	if err != nil {
		return nil, err
	}

	selfNotarizedHeadersNotifier, err := NewBlockNotifier()
	if err != nil {
		return nil, err
	}

	finalMetachainHeadersNotifier, err := NewBlockNotifier()
	if err != nil {
		return nil, err
	}

	blockBalancerInstance, err := NewBlockBalancer()
	if err != nil {
		return nil, err
	}

	bbt := &baseBlockTrack{
		hasher:                                arguments.Hasher,
		headerValidator:                       arguments.HeaderValidator,
		marshalizer:                           arguments.Marshalizer,
		rounder:                               arguments.Rounder,
		shardCoordinator:                      arguments.ShardCoordinator,
		headersPool:                           arguments.PoolsHolder.Headers(),
		store:                                 arguments.Store,
		crossNotarizer:                        crossNotarizer,
		selfNotarizer:                         selfNotarizer,
		crossNotarizedHeadersNotifier:         crossNotarizedHeadersNotifier,
		selfNotarizedFromCrossHeadersNotifier: selfNotarizedFromCrossHeadersNotifier,
		selfNotarizedHeadersNotifier:          selfNotarizedHeadersNotifier,
		finalMetachainHeadersNotifier:         finalMetachainHeadersNotifier,
		blockBalancer:                         blockBalancerInstance,
		maxNumHeadersToKeepPerShard:           maxNumHeadersToKeepPerShard,
		whitelistHandler:                      arguments.WhitelistHandler,
	}

	return bbt, nil
}

func (bbt *baseBlockTrack) receivedHeader(headerHandler data.HeaderHandler, headerHash []byte) {
	if headerHandler.GetShardID() == core.MetachainShardId {
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
		"epoch", shardHeader.GetEpoch(),
		"round", shardHeader.GetRound(),
		"nonce", shardHeader.GetNonce(),
		"hash", shardHeaderHash,
	)

	if !bbt.ShouldAddHeader(headerHandler) {
		log.Trace("received shard header is out of range", "nonce", headerHandler.GetNonce())
		return
	}

	bbt.doWhitelistWithShardHeaderIfNeeded(shardHeader)

	bbt.addHeader(shardHeader, shardHeaderHash)
	bbt.blockProcessor.ProcessReceivedHeader(shardHeader)
}

func (bbt *baseBlockTrack) receivedMetaBlock(headerHandler data.HeaderHandler, metaBlockHash []byte) {
	metaBlock, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		log.Warn("cannot convert data.HeaderHandler in *block.Metablock")
		return
	}

	log.Debug("received meta block from network in block tracker",
		"shard", metaBlock.GetShardID(),
		"epoch", metaBlock.GetEpoch(),
		"round", metaBlock.GetRound(),
		"nonce", metaBlock.GetNonce(),
		"hash", metaBlockHash,
	)

	if !bbt.ShouldAddHeader(headerHandler) {
		log.Trace("received meta block is out of range", "nonce", headerHandler.GetNonce())
		return
	}

	bbt.doWhitelistWithMetaBlockIfNeeded(metaBlock)

	bbt.addHeader(metaBlock, metaBlockHash)
	bbt.blockProcessor.ProcessReceivedHeader(metaBlock)
}

// ShouldAddHeader returns if the given header should be added or not in the tracker list (is out of the interest range)
func (bbt *baseBlockTrack) ShouldAddHeader(headerHandler data.HeaderHandler) bool {
	shardID := headerHandler.GetShardID()
	if shardID == bbt.shardCoordinator.SelfId() {
		return bbt.shouldAddHeaderForShard(headerHandler, bbt.selfNotarizer, core.MetachainShardId)
	}

	return bbt.shouldAddHeaderForShard(headerHandler, bbt.crossNotarizer, shardID)
}

func (bbt *baseBlockTrack) shouldAddHeaderForShard(
	headerHandler data.HeaderHandler,
	blockNotarizer blockNotarizerHandler,
	shardForFirstNotarizedHeader uint32,
) bool {
	lastNotarizedHeader, _, err := blockNotarizer.GetLastNotarizedHeader(headerHandler.GetShardID())
	if err != nil {
		log.Debug("shouldAddHeaderForShard.GetLastNotarizedHeader",
			"shard", headerHandler.GetShardID(),
			"error", err.Error())
		return false
	}

	lastNotarizedHeaderNonce := lastNotarizedHeader.GetNonce()

	isHeaderOutOfRange := headerHandler.GetNonce() > lastNotarizedHeaderNonce+uint64(bbt.maxNumHeadersToKeepPerShard)
	return !isHeaderOutOfRange
}

func (bbt *baseBlockTrack) addHeader(header data.HeaderHandler, hash []byte) {
	if check.IfNil(header) {
		return
	}

	shardID := header.GetShardID()
	nonce := header.GetNonce()

	bbt.mutHeaders.Lock()
	headersForShard, ok := bbt.headers[shardID]
	if !ok {
		headersForShard = make(map[uint64][]*HeaderInfo)
		bbt.headers[shardID] = headersForShard
	}

	for _, hdrInfo := range headersForShard[nonce] {
		if bytes.Equal(hdrInfo.Hash, hash) {
			bbt.mutHeaders.Unlock()
			return
		}
	}

	headersForShard[nonce] = append(headersForShard[nonce], &HeaderInfo{Hash: hash, Header: header})
	bbt.mutHeaders.Unlock()
}

// AddCrossNotarizedHeader adds cross notarized header to the tracker lists
func (bbt *baseBlockTrack) AddCrossNotarizedHeader(
	shardID uint32,
	crossNotarizedHeader data.HeaderHandler,
	crossNotarizedHeaderHash []byte,
) {
	bbt.crossNotarizer.AddNotarizedHeader(shardID, crossNotarizedHeader, crossNotarizedHeaderHash)
}

// AddSelfNotarizedHeader adds self notarized headers to the tracker lists
func (bbt *baseBlockTrack) AddSelfNotarizedHeader(
	shardID uint32,
	selfNotarizedHeader data.HeaderHandler,
	selfNotarizedHeaderHash []byte,
) {
	bbt.selfNotarizer.AddNotarizedHeader(shardID, selfNotarizedHeader, selfNotarizedHeaderHash)
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
	bbt.selfNotarizer.CleanupNotarizedHeadersBehindNonce(shardID, selfNotarizedNonce)
	nonce := selfNotarizedNonce

	if shardID != bbt.shardCoordinator.SelfId() {
		bbt.crossNotarizer.CleanupNotarizedHeadersBehindNonce(shardID, crossNotarizedNonce)
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
	return bbt.blockProcessor.ComputeLongestChain(shardID, header)
}

// ComputeLongestMetaChainFromLastNotarized returns the longest valid chain for metachain from its last cross notarized header
func (bbt *baseBlockTrack) ComputeLongestMetaChainFromLastNotarized() ([]data.HeaderHandler, [][]byte, error) {
	lastCrossNotarizedHeader, _, err := bbt.GetLastCrossNotarizedHeader(core.MetachainShardId)
	if err != nil {
		return nil, nil, err
	}

	hdrsForShard, hdrsHashesForShard := bbt.ComputeLongestChain(core.MetachainShardId, lastCrossNotarizedHeader)

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

	bbt.displayHeadersForShard(core.MetachainShardId)
}

func (bbt *baseBlockTrack) displayHeadersForShard(shardID uint32) {
	bbt.displayTrackedHeadersForShard(shardID, "tracked headers")
	bbt.crossNotarizer.DisplayNotarizedHeaders(shardID, "cross notarized headers")
	bbt.selfNotarizer.DisplayNotarizedHeaders(shardID, "self notarized headers")
}

func (bbt *baseBlockTrack) displayTrackedHeadersForShard(shardID uint32, message string) {
	headers, hashes := bbt.SortHeadersFromNonce(shardID, 0)
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
	return bbt.crossNotarizer.GetNotarizedHeader(shardID, offset)
}

// CheckBlockAgainstRounder verifies the provided header against the rounder's current round
func (bbt *baseBlockTrack) CheckBlockAgainstRounder(headerHandler data.HeaderHandler) error {
	if check.IfNil(headerHandler) {
		return process.ErrNilHeaderHandler
	}

	nextRound := bbt.rounder.Index() + 1
	if int64(headerHandler.GetRound()) > nextRound {
		return fmt.Errorf("%w header round: %d, next chronology round: %d",
			process.ErrHigherRoundInBlock,
			headerHandler.GetRound(),
			nextRound)
	}

	return nil
}

// CheckBlockAgainstFinal checks if the given header is valid related to the final header
func (bbt *baseBlockTrack) CheckBlockAgainstFinal(headerHandler data.HeaderHandler) error {
	if check.IfNil(headerHandler) {
		return process.ErrNilHeaderHandler
	}

	finalHeader, _, err := bbt.getFinalHeader(headerHandler.GetShardID())
	if err != nil {
		return fmt.Errorf("%w: header shard: %d, header round: %d, header nonce: %d",
			err,
			headerHandler.GetShardID(),
			headerHandler.GetRound(),
			headerHandler.GetNonce())
	}

	roundDif := int64(headerHandler.GetRound()) - int64(finalHeader.GetRound())
	nonceDif := int64(headerHandler.GetNonce()) - int64(finalHeader.GetNonce())

	if roundDif < 0 {
		return fmt.Errorf("%w for header round: %d, final header round: %d",
			process.ErrLowerRoundInBlock,
			headerHandler.GetRound(),
			finalHeader.GetRound())
	}
	if nonceDif < 0 {
		return fmt.Errorf("%w for header nonce: %d, final header nonce: %d",
			process.ErrLowerNonceInBlock,
			headerHandler.GetNonce(),
			finalHeader.GetNonce())
	}
	if roundDif < nonceDif {
		return fmt.Errorf("%w for "+
			"header round: %d, final header round: %d, round dif: %d"+
			"header nonce: %d, final header nonce: %d, nonce dif: %d",
			process.ErrHigherNonceInBlock,
			headerHandler.GetRound(),
			finalHeader.GetRound(),
			roundDif,
			headerHandler.GetNonce(),
			finalHeader.GetNonce(),
			nonceDif)
	}

	return nil
}

func (bbt *baseBlockTrack) getFinalHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	if shardID != bbt.shardCoordinator.SelfId() {
		return bbt.crossNotarizer.GetFirstNotarizedHeader(shardID)
	}

	return bbt.selfNotarizer.GetFirstNotarizedHeader(shardID)
}

// CheckBlockAgainstWhiteList returns if the provided intercepted data (blocks) is whitelisted or not
func (bbt *baseBlockTrack) CheckBlockAgainstWhitelist(interceptedData process.InterceptedData) bool {
	return bbt.whitelistHandler.IsWhiteListed(interceptedData)
}

// GetLastCrossNotarizedHeader returns last cross notarized header for a given shard
func (bbt *baseBlockTrack) GetLastCrossNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	return bbt.crossNotarizer.GetLastNotarizedHeader(shardID)
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

// GetLastSelfNotarizedHeader returns last self notarized header for a given shard
func (bbt *baseBlockTrack) GetLastSelfNotarizedHeader(shardID uint32) (data.HeaderHandler, []byte, error) {
	return bbt.selfNotarizer.GetLastNotarizedHeader(shardID)
}

// GetSelfNotarizedHeader returns a self notarized header for a given shard with a given offset, behind last self notarized header
func (bbt *baseBlockTrack) GetSelfNotarizedHeader(shardID uint32, offset uint64) (data.HeaderHandler, []byte, error) {
	return bbt.selfNotarizer.GetNotarizedHeader(shardID, offset)
}

// GetTrackedHeaders returns tracked headers for a given shard
func (bbt *baseBlockTrack) GetTrackedHeaders(shardID uint32) ([]data.HeaderHandler, [][]byte) {
	return bbt.SortHeadersFromNonce(shardID, 0)
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

// SortHeadersFromNonce gets sorted tracked headers for a given shard from a given nonce
func (bbt *baseBlockTrack) SortHeadersFromNonce(shardID uint32, nonce uint64) ([]data.HeaderHandler, [][]byte) {
	bbt.mutHeaders.RLock()
	defer bbt.mutHeaders.RUnlock()

	headersForShard, ok := bbt.headers[shardID]
	if !ok {
		return nil, nil
	}

	sortedHeadersInfo := make([]*HeaderInfo, 0)
	for headersNonce, headersInfo := range headersForShard {
		if headersNonce < nonce {
			continue
		}

		sortedHeadersInfo = append(sortedHeadersInfo, headersInfo...)
	}

	if len(sortedHeadersInfo) > 1 {
		sort.Slice(sortedHeadersInfo, func(i, j int) bool {
			return sortedHeadersInfo[i].Header.GetNonce() < sortedHeadersInfo[j].Header.GetNonce()
		})
	}

	headers := make([]data.HeaderHandler, 0)
	headersHashes := make([][]byte, 0)

	for _, hdrInfo := range sortedHeadersInfo {
		headers = append(headers, hdrInfo.Header)
		headersHashes = append(headersHashes, hdrInfo.Hash)
	}

	return headers, headersHashes
}

// RemoveHeaderFromPool removes the header with the given shard and nonce from pool
func (bbt *baseBlockTrack) RemoveHeaderFromPool(shardID uint32, nonce uint64) {
	bbt.headersPool.RemoveHeaderByNonceAndShardId(nonce, shardID)
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

	for _, hdrInfo := range headersForShardWithNonce {
		headers = append(headers, hdrInfo.Header)
		headersHashes = append(headersHashes, hdrInfo.Hash)
	}

	return headers, headersHashes
}

// IsShardStuck returns true if the given shard is stuck
func (bbt *baseBlockTrack) IsShardStuck(shardID uint32) bool {
	if bbt.shardCoordinator.SelfId() == core.MetachainShardId {
		return false
	}

	if shardID == core.MetachainShardId {
		return bbt.isMetaStuck()
	}

	numPendingMiniBlocks := bbt.blockBalancer.GetNumPendingMiniBlocks(shardID)
	lastShardProcessedMetaNonce := bbt.blockBalancer.GetLastShardProcessedMetaNonce(shardID)

	isMetaDifferenceTooLarge := false
	shouldCheckLastMetaNonceProcessed := shardID != core.MetachainShardId && lastShardProcessedMetaNonce > 0
	if shouldCheckLastMetaNonceProcessed {
		metaHeaders, _ := bbt.GetTrackedHeaders(core.MetachainShardId)
		numMetaHeaders := len(metaHeaders)
		if numMetaHeaders > 0 {
			lastMetaHeader := metaHeaders[numMetaHeaders-1]
			metaDiff := lastMetaHeader.GetNonce() - lastShardProcessedMetaNonce
			isMetaDifferenceTooLarge = metaDiff > process.MaxMetaNoncesBehind
		}
	}

	maxNumPendingMiniBlocks := process.MaxNumPendingMiniBlocksPerShard * bbt.shardCoordinator.NumberOfShards()
	isShardStuck := numPendingMiniBlocks >= maxNumPendingMiniBlocks || isMetaDifferenceTooLarge
	return isShardStuck
}

func (bbt *baseBlockTrack) isMetaStuck() bool {
	selfHdrNotarizedByItself, _, err := bbt.GetLastSelfNotarizedHeader(bbt.shardCoordinator.SelfId())
	if err != nil {
		log.Debug("isMetaStuck.GetLastSelfNotarizedHeader",
			"shard", bbt.shardCoordinator.SelfId(),
			"error", err.Error())
		return false
	}

	selfHdrNotarizedByMeta, _, err := bbt.GetLastSelfNotarizedHeader(core.MetachainShardId)
	if err != nil {
		log.Debug("isMetaStuck.GetLastSelfNotarizedHeader",
			"shard", core.MetachainShardId,
			"error", err.Error())
		return false
	}

	isMetaStuck := selfHdrNotarizedByItself.GetNonce() > selfHdrNotarizedByMeta.GetNonce()+process.MaxShardNoncesBehind
	return isMetaStuck
}

// RegisterCrossNotarizedHeadersHandler registers a new handler to be called when cross notarized header is changed
func (bbt *baseBlockTrack) RegisterCrossNotarizedHeadersHandler(
	handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte),
) {
	bbt.crossNotarizedHeadersNotifier.RegisterHandler(handler)
}

// RegisterSelfNotarizedFromCrossHeadersHandler registers a new handler to be called when self notarized header,
// extracted from cross headers, is changed
func (bbt *baseBlockTrack) RegisterSelfNotarizedFromCrossHeadersHandler(
	handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte),
) {
	bbt.selfNotarizedFromCrossHeadersNotifier.RegisterHandler(handler)
}

// RegisterSelfNotarizedHeadersHandler registers a new handler to be called when self notarized header is changed
func (bbt *baseBlockTrack) RegisterSelfNotarizedHeadersHandler(
	handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte),
) {
	bbt.selfNotarizedHeadersNotifier.RegisterHandler(handler)
}

// RegisterFinalMetachainHeadersHandler registers a new handler to be called when a metachain header is final
func (bbt *baseBlockTrack) RegisterFinalMetachainHeadersHandler(
	handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte),
) {
	bbt.finalMetachainHeadersNotifier.RegisterHandler(handler)
}

// RemoveLastNotarizedHeaders removes last notarized headers from tracker list
func (bbt *baseBlockTrack) RemoveLastNotarizedHeaders() {
	bbt.crossNotarizer.RemoveLastNotarizedHeader()
	bbt.selfNotarizer.RemoveLastNotarizedHeader()
}

// RestoreToGenesis sets class variables to theirs initial values
func (bbt *baseBlockTrack) RestoreToGenesis() {
	bbt.crossNotarizer.RestoreNotarizedHeadersToGenesis()
	bbt.selfNotarizer.RestoreNotarizedHeadersToGenesis()
	bbt.restoreTrackedHeadersToGenesis()
}

func (bbt *baseBlockTrack) restoreTrackedHeadersToGenesis() {
	bbt.mutHeaders.Lock()
	bbt.headers = make(map[uint32]map[uint64][]*HeaderInfo)
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
	if check.IfNil(arguments.PoolsHolder) {
		return process.ErrNilPoolsHolder
	}
	if check.IfNil(arguments.PoolsHolder.Headers()) {
		return process.ErrNilHeadersDataPool
	}

	return nil
}

func (bbt *baseBlockTrack) initNotarizedHeaders(startHeaders map[uint32]data.HeaderHandler) error {
	err := bbt.crossNotarizer.InitNotarizedHeaders(startHeaders)
	if err != nil {
		return err
	}

	selfStartHeader := startHeaders[bbt.shardCoordinator.SelfId()]
	selfStartHeaders := make(map[uint32]data.HeaderHandler)
	for shardID := range startHeaders {
		selfStartHeaders[shardID] = selfStartHeader
	}

	err = bbt.selfNotarizer.InitNotarizedHeaders(selfStartHeaders)
	if err != nil {
		return err
	}

	return nil
}

func (bbt *baseBlockTrack) doWhitelistWithMetaBlockIfNeeded(metablock *block.MetaBlock) {
	selfShardID := bbt.shardCoordinator.SelfId()
	if selfShardID == core.MetachainShardId {
		return
	}
	if metablock == nil {
		return
	}
	if bbt.isHeaderOutOfRange(metablock) {
		return
	}

	miniBlockHdrs := metablock.GetMiniBlockHeaders()
	keys := make([][]byte, 0)

	crossMbKeysMeta := getCrossShardMiniblockKeys(miniBlockHdrs, selfShardID, core.MetachainShardId)
	if len(crossMbKeysMeta) > 0 {
		keys = append(keys, crossMbKeysMeta...)
	}

	for _, shardData := range metablock.ShardInfo {
		if shardData.ShardID == selfShardID {
			continue
		}

		crossMbKeysShard := getCrossShardMiniblockKeys(shardData.ShardMiniBlockHeaders, selfShardID, shardData.ShardID)
		if len(crossMbKeysShard) > 0 {
			keys = append(keys, crossMbKeysShard...)
		}
	}

	bbt.whitelistHandler.Add(keys)
}

func (bbt *baseBlockTrack) doWhitelistWithShardHeaderIfNeeded(shardHeader *block.Header) {
	selfShardID := bbt.shardCoordinator.SelfId()
	if selfShardID != core.MetachainShardId {
		return
	}
	if shardHeader == nil {
		return
	}
	if bbt.isHeaderOutOfRange(shardHeader) {
		return
	}

	miniBlockHdrs := shardHeader.GetMiniBlockHeaders()
	keys := make([][]byte, 0)

	crossMbKeysShard := getCrossShardMiniblockKeys(miniBlockHdrs, selfShardID, shardHeader.ShardID)
	if len(crossMbKeysShard) > 0 {
		keys = append(keys, crossMbKeysShard...)
	}

	bbt.whitelistHandler.Add(keys)
}

func getCrossShardMiniblockKeys(miniBlockHdrs []block.MiniBlockHeader, selfShardID uint32, processingShard uint32) [][]byte {
	keys := make([][]byte, 0)
	for _, miniBlockHdr := range miniBlockHdrs {
		receiverShard := miniBlockHdr.GetReceiverShardID()
		receiverIsSelfShard := receiverShard == selfShardID || receiverShard == core.AllShardId && processingShard == core.MetachainShardId
		senderIsCrossShard := miniBlockHdr.GetSenderShardID() != selfShardID
		if receiverIsSelfShard && senderIsCrossShard {
			keys = append(keys, miniBlockHdr.Hash)
			log.Debug(
				"getCrossShardMiniblockKeys",
				"type", miniBlockHdr.GetType(),
				"sender", miniBlockHdr.GetSenderShardID(),
				"receiver", miniBlockHdr.GetReceiverShardID(),
				"hash", miniBlockHdr.GetHash(),
			)
		}
	}
	return keys
}

func (bbt *baseBlockTrack) isHeaderOutOfRange(headerHandler data.HeaderHandler) bool {
	lastCrossNotarizedHeader, _, err := bbt.GetLastCrossNotarizedHeader(headerHandler.GetShardID())
	if err != nil {
		log.Debug("isHeaderOutOfRange.GetLastCrossNotarizedHeader",
			"shard", headerHandler.GetShardID(),
			"error", err.Error())
		return true
	}

	isHeaderOutOfRange := headerHandler.GetNonce() > lastCrossNotarizedHeader.GetNonce()+process.MaxHeadersToWhitelistInAdvance
	return isHeaderOutOfRange
}
