package track

import (
	"sort"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type blockProcessor struct {
	headerValidator  process.HeaderConstructionValidator
	requestHandler   process.RequestHandler
	shardCoordinator sharding.Coordinator

	blockTracker                  blockTrackerHandler
	crossNotarizer                blockNotarizerHandler
	crossNotarizedHeadersNotifier blockNotifierHandler
	selfNotarizedHeadersNotifier  blockNotifierHandler

	blockFinality uint64
}

// NewBlockProcessor creates a block processor object which implements blockProcessorHandler interface
func NewBlockProcessor(
	headerValidator process.HeaderConstructionValidator,
	requestHandler process.RequestHandler,
	shardCoordinator sharding.Coordinator,
	blockTracker blockTrackerHandler,
	crossNotarizer blockNotarizerHandler,
	crossNotarizedHeadersNotifier blockNotifierHandler,
	selfNotarizedHeadersNotifier blockNotifierHandler,
) (*blockProcessor, error) {

	err := checkBlockProcessorNilParameters(
		headerValidator,
		requestHandler,
		shardCoordinator,
		blockTracker,
		crossNotarizer,
		crossNotarizedHeadersNotifier,
		selfNotarizedHeadersNotifier,
	)
	if err != nil {
		return nil, err
	}

	bp := blockProcessor{
		headerValidator:               headerValidator,
		requestHandler:                requestHandler,
		shardCoordinator:              shardCoordinator,
		blockTracker:                  blockTracker,
		crossNotarizer:                crossNotarizer,
		crossNotarizedHeadersNotifier: crossNotarizedHeadersNotifier,
		selfNotarizedHeadersNotifier:  selfNotarizedHeadersNotifier,
	}

	bp.blockFinality = process.BlockFinality

	return &bp, nil
}

func (bp *blockProcessor) processReceivedHeader(header data.HeaderHandler) {
	if check.IfNil(header) {
		return
	}

	isHeaderForSelfShard := header.GetShardID() == bp.shardCoordinator.SelfId()
	if isHeaderForSelfShard {
		bp.doJobOnReceivedHeader(header.GetShardID())
	} else {
		bp.doJobOnReceivedCrossNotarizedHeader(header.GetShardID())
	}
}

func (bp *blockProcessor) doJobOnReceivedHeader(shardID uint32) {
	_, _, selfNotarizedHeaders, selfNotarizedHeadersHashes := bp.blockTracker.computeLongestSelfChain()

	if len(selfNotarizedHeaders) > 0 {
		bp.selfNotarizedHeadersNotifier.callHandlers(shardID, selfNotarizedHeaders, selfNotarizedHeadersHashes)
	}
}

func (bp *blockProcessor) doJobOnReceivedCrossNotarizedHeader(shardID uint32) {
	_, _, crossNotarizedHeaders, crossNotarizedHeadersHashes := bp.computeLongestChainFromLastCrossNotarized(shardID)
	selfNotarizedHeaders, selfNotarizedHeadersHashes := bp.computeSelfNotarizedHeaders(crossNotarizedHeaders)

	if len(crossNotarizedHeaders) > 0 {
		bp.crossNotarizedHeadersNotifier.callHandlers(shardID, crossNotarizedHeaders, crossNotarizedHeadersHashes)
	}

	if len(selfNotarizedHeaders) > 0 {
		bp.selfNotarizedHeadersNotifier.callHandlers(shardID, selfNotarizedHeaders, selfNotarizedHeadersHashes)
	}
}

func (bp *blockProcessor) computeLongestChainFromLastCrossNotarized(
	shardID uint32,
) (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {

	lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash, err := bp.crossNotarizer.getLastNotarizedHeader(shardID)
	if err != nil {
		return nil, nil, nil, nil
	}

	headers, hashes := bp.computeLongestChain(shardID, lastCrossNotarizedHeader)
	return lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash, headers, hashes
}

func (bp *blockProcessor) computeSelfNotarizedHeaders(headers []data.HeaderHandler) ([]data.HeaderHandler, [][]byte) {
	selfNotarizedHeadersInfo := make([]*headerInfo, 0)

	for _, header := range headers {
		selfHeadersInfo := bp.blockTracker.getSelfHeaders(header)
		if len(selfHeadersInfo) > 0 {
			selfNotarizedHeadersInfo = append(selfNotarizedHeadersInfo, selfHeadersInfo...)
		}
	}

	if len(selfNotarizedHeadersInfo) > 1 {
		sort.Slice(selfNotarizedHeadersInfo, func(i, j int) bool {
			return selfNotarizedHeadersInfo[i].header.GetNonce() < selfNotarizedHeadersInfo[j].header.GetNonce()
		})
	}

	selfNotarizedHeaders := make([]data.HeaderHandler, 0)
	selfNotarizedHeadersHashes := make([][]byte, 0)

	for _, selfNotarizedHeaderInfo := range selfNotarizedHeadersInfo {
		selfNotarizedHeaders = append(selfNotarizedHeaders, selfNotarizedHeaderInfo.header)
		selfNotarizedHeadersHashes = append(selfNotarizedHeadersHashes, selfNotarizedHeaderInfo.hash)
	}

	return selfNotarizedHeaders, selfNotarizedHeadersHashes
}

func (bp *blockProcessor) computeLongestChain(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte) {
	headers := make([]data.HeaderHandler, 0)
	headersHashes := make([][]byte, 0)

	if check.IfNil(header) {
		return headers, headersHashes
	}

	sortedHeaders, sortedHeadersHashes := bp.blockTracker.sortHeadersFromNonce(shardID, header.GetNonce()+1)
	if len(sortedHeaders) == 0 {
		return headers, headersHashes
	}

	longestChainHeadersIndexes := make([]int, 0)
	headersIndexes := make([]int, 0)
	bp.getNextHeader(&longestChainHeadersIndexes, headersIndexes, header, sortedHeaders, 0)

	for _, index := range longestChainHeadersIndexes {
		headers = append(headers, sortedHeaders[index])
		headersHashes = append(headersHashes, sortedHeadersHashes[index])
	}

	bp.requestHeadersIfNeeded(header, sortedHeaders, headers)

	return headers, headersHashes
}

func (bp *blockProcessor) getNextHeader(
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

		err := bp.headerValidator.IsHeaderConstructionValid(currHeader, prevHeader)
		if err != nil {
			continue
		}

		err = bp.checkHeaderFinality(currHeader, sortedHeaders, i+1)
		if err != nil {
			continue
		}

		headersIndexes = append(headersIndexes, i)
		bp.getNextHeader(longestChainHeadersIndexes, headersIndexes, currHeader, sortedHeaders, i+1)
		headersIndexes = headersIndexes[:len(headersIndexes)-1]
	}
}

func (bp *blockProcessor) checkHeaderFinality(
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
		if numFinalityAttestingHeaders >= bp.blockFinality || currHeader.GetNonce() > prevHeader.GetNonce()+1 {
			break
		}

		err := bp.headerValidator.IsHeaderConstructionValid(currHeader, prevHeader)
		if err != nil {
			continue
		}

		prevHeader = currHeader
		numFinalityAttestingHeaders += 1
	}

	if numFinalityAttestingHeaders < bp.blockFinality {
		return process.ErrHeaderNotFinal
	}

	return nil
}

func (bp *blockProcessor) requestHeadersIfNeeded(
	lastNotarizedHeader data.HeaderHandler,
	sortedHeaders []data.HeaderHandler,
	longestChainHeaders []data.HeaderHandler,
) {
	if check.IfNil(lastNotarizedHeader) {
		return
	}

	nbSortedHeaders := len(sortedHeaders)
	if nbSortedHeaders == 0 {
		return
	}

	highestNonceReceived := sortedHeaders[nbSortedHeaders-1].GetNonce()
	highestNonceInLongestChain := lastNotarizedHeader.GetNonce()
	nbLongestChainHeaders := len(longestChainHeaders)
	if nbLongestChainHeaders > 0 {
		highestNonceInLongestChain = longestChainHeaders[nbLongestChainHeaders-1].GetNonce()
	}

	log.Debug("requestHeadersIfNeeded",
		"shard", lastNotarizedHeader.GetShardID(),
		"last notarized nonce", lastNotarizedHeader.GetNonce(),
		"highest nonce received", highestNonceReceived,
		"highest nonce in longest chain", highestNonceInLongestChain)

	if highestNonceReceived <= highestNonceInLongestChain+bp.blockFinality {
		return
	}

	shardID := lastNotarizedHeader.GetShardID()
	fromNonce := highestNonceInLongestChain + 1
	toNonce := fromNonce + uint64(bp.blockFinality)
	for nonce := fromNonce; nonce <= toNonce; nonce++ {
		log.Debug("request header",
			"shard", shardID,
			"nonce", nonce)

		if shardID == sharding.MetachainShardId {
			go bp.requestHandler.RequestMetaHeaderByNonce(nonce)
		} else {
			go bp.requestHandler.RequestShardHeaderByNonce(shardID, nonce)
		}
	}
}

func checkBlockProcessorNilParameters(
	headerValidator process.HeaderConstructionValidator,
	requestHandler process.RequestHandler,
	shardCoordinator sharding.Coordinator,
	blockTracker blockTrackerHandler,
	crossNotarizer blockNotarizerHandler,
	crossNotarizedHeadersNotifier blockNotifierHandler,
	selfNotarizedHeadersNotifier blockNotifierHandler,
) error {
	if check.IfNil(headerValidator) {
		return process.ErrNilHeaderValidator
	}
	if check.IfNil(requestHandler) {
		return process.ErrNilRequestHandler
	}
	if check.IfNil(shardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if blockTracker == nil {
		return ErrNilBlockTrackerHandler
	}
	if crossNotarizer == nil {
		return ErrNilCrossNotarizer
	}
	if crossNotarizedHeadersNotifier == nil {
		return ErrCrossNotarizedHeadersNotifier
	}
	if selfNotarizedHeadersNotifier == nil {
		return ErrSelfNotarizedHeadersNotifier
	}

	return nil
}
