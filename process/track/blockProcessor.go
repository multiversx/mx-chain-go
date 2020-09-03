package track

import (
	"sort"

	"github.com/ElrondNetwork/elrond-go/core"

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
	selfNotarizer                 blockNotarizerHandler
	crossNotarizedHeadersNotifier blockNotifierHandler
	selfNotarizedHeadersNotifier  blockNotifierHandler
	finalMetachainHeadersNotifier blockNotifierHandler
	rounder                       process.Rounder

	blockFinality uint64
}

// NewBlockProcessor creates a block processor object which implements blockProcessorHandler interface
func NewBlockProcessor(arguments ArgBlockProcessor) (*blockProcessor, error) {
	err := checkBlockProcessorNilParameters(arguments)
	if err != nil {
		return nil, err
	}

	bp := blockProcessor{
		headerValidator:               arguments.HeaderValidator,
		requestHandler:                arguments.RequestHandler,
		shardCoordinator:              arguments.ShardCoordinator,
		blockTracker:                  arguments.BlockTracker,
		crossNotarizer:                arguments.CrossNotarizer,
		selfNotarizer:                 arguments.SelfNotarizer,
		crossNotarizedHeadersNotifier: arguments.CrossNotarizedHeadersNotifier,
		selfNotarizedHeadersNotifier:  arguments.SelfNotarizedHeadersNotifier,
		finalMetachainHeadersNotifier: arguments.FinalMetachainHeadersNotifier,
		rounder:                       arguments.Rounder,
	}

	bp.blockFinality = process.BlockFinality

	return &bp, nil
}

// ProcessReceivedHeader processes the header which has been received
func (bp *blockProcessor) ProcessReceivedHeader(header data.HeaderHandler) {
	if check.IfNil(header) {
		return
	}

	if header.GetShardID() == core.MetachainShardId {
		bp.doJobOnReceivedMetachainHeader()
	}

	isHeaderForSelfShard := header.GetShardID() == bp.shardCoordinator.SelfId()
	if isHeaderForSelfShard {
		bp.doJobOnReceivedHeader(header.GetShardID())
	} else {
		bp.doJobOnReceivedCrossNotarizedHeader(header.GetShardID())
	}
}

func (bp *blockProcessor) doJobOnReceivedHeader(shardID uint32) {
	_, _, selfNotarizedHeaders, selfNotarizedHeadersHashes := bp.blockTracker.ComputeLongestSelfChain()

	if len(selfNotarizedHeaders) > 0 {
		bp.selfNotarizedHeadersNotifier.CallHandlers(shardID, selfNotarizedHeaders, selfNotarizedHeadersHashes)
	}
}

func (bp *blockProcessor) doJobOnReceivedCrossNotarizedHeader(shardID uint32) {
	_, _, crossNotarizedHeaders, crossNotarizedHeadersHashes := bp.computeLongestChainFromLastCrossNotarized(shardID)
	selfNotarizedHeaders, selfNotarizedHeadersHashes := bp.computeSelfNotarizedHeaders(crossNotarizedHeaders)
	bp.blockTracker.ComputeCrossInfo(crossNotarizedHeaders)

	if len(crossNotarizedHeaders) > 0 {
		bp.crossNotarizedHeadersNotifier.CallHandlers(shardID, crossNotarizedHeaders, crossNotarizedHeadersHashes)
	}

	if len(selfNotarizedHeaders) > 0 {
		bp.selfNotarizedHeadersNotifier.CallHandlers(shardID, selfNotarizedHeaders, selfNotarizedHeadersHashes)
	}
}

func (bp *blockProcessor) doJobOnReceivedMetachainHeader() {
	var header data.HeaderHandler
	var headerHash []byte
	var err error

	if bp.shardCoordinator.SelfId() == core.MetachainShardId {
		header, headerHash, err = bp.selfNotarizer.GetLastNotarizedHeader(bp.shardCoordinator.SelfId())
		if err != nil {
			log.Warn("blockProcessor.doJobOnReceivedMetachainHeader", "error", err.Error())
			return
		}
	} else {
		header, headerHash, err = bp.crossNotarizer.GetLastNotarizedHeader(core.MetachainShardId)
		if err != nil {
			log.Warn("blockProcessor.doJobOnReceivedMetachainHeader", "error", err.Error())
			return
		}
	}

	sortedHeaders, _ := bp.blockTracker.SortHeadersFromNonce(core.MetachainShardId, header.GetNonce()+1)
	if len(sortedHeaders) == 0 {
		return
	}

	finalMetachainHeaders := make([]data.HeaderHandler, 0)
	finalMetachainHeadersHashes := make([][]byte, 0)

	err = bp.checkHeaderFinality(header, sortedHeaders, 0)
	if err == nil {
		finalMetachainHeaders = append(finalMetachainHeaders, header)
		finalMetachainHeadersHashes = append(finalMetachainHeadersHashes, headerHash)
	}

	headers, headerHashes := bp.ComputeLongestChain(core.MetachainShardId, header)

	finalMetachainHeaders = append(finalMetachainHeaders, headers...)
	finalMetachainHeadersHashes = append(finalMetachainHeadersHashes, headerHashes...)

	if len(finalMetachainHeaders) > 0 {
		bp.finalMetachainHeadersNotifier.CallHandlers(core.MetachainShardId, finalMetachainHeaders, finalMetachainHeadersHashes)
	}
}

func (bp *blockProcessor) computeLongestChainFromLastCrossNotarized(
	shardID uint32,
) (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {

	lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash, err := bp.crossNotarizer.GetLastNotarizedHeader(shardID)
	if err != nil {
		return nil, nil, nil, nil
	}

	headers, hashes := bp.ComputeLongestChain(shardID, lastCrossNotarizedHeader)
	return lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash, headers, hashes
}

func (bp *blockProcessor) computeSelfNotarizedHeaders(headers []data.HeaderHandler) ([]data.HeaderHandler, [][]byte) {
	selfNotarizedHeadersInfo := make([]*HeaderInfo, 0)

	for _, header := range headers {
		selfHeadersInfo := bp.blockTracker.GetSelfHeaders(header)
		if len(selfHeadersInfo) > 0 {
			selfNotarizedHeadersInfo = append(selfNotarizedHeadersInfo, selfHeadersInfo...)
		}
	}

	if len(selfNotarizedHeadersInfo) > 1 {
		sort.Slice(selfNotarizedHeadersInfo, func(i, j int) bool {
			return selfNotarizedHeadersInfo[i].Header.GetNonce() < selfNotarizedHeadersInfo[j].Header.GetNonce()
		})
	}

	selfNotarizedHeaders := make([]data.HeaderHandler, 0)
	selfNotarizedHeadersHashes := make([][]byte, 0)

	for _, selfNotarizedHeaderInfo := range selfNotarizedHeadersInfo {
		selfNotarizedHeaders = append(selfNotarizedHeaders, selfNotarizedHeaderInfo.Header)
		selfNotarizedHeadersHashes = append(selfNotarizedHeadersHashes, selfNotarizedHeaderInfo.Hash)
	}

	return selfNotarizedHeaders, selfNotarizedHeadersHashes
}

// ComputeLongestChain computes the longest chain for a given shard starting from a given header
func (bp *blockProcessor) ComputeLongestChain(shardID uint32, header data.HeaderHandler) ([]data.HeaderHandler, [][]byte) {
	headers := make([]data.HeaderHandler, 0)
	headersHashes := make([][]byte, 0)

	if check.IfNil(header) {
		return headers, headersHashes
	}

	var sortedHeaders []data.HeaderHandler
	var sortedHeadersHashes [][]byte

	defer func() {
		bp.requestHeadersIfNeeded(header, sortedHeaders, headers)
	}()

	sortedHeaders, sortedHeadersHashes = bp.blockTracker.SortHeadersFromNonce(shardID, header.GetNonce()+1)
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
	sortedReceivedHeaders []data.HeaderHandler,
	longestChainHeaders []data.HeaderHandler,
) {
	if check.IfNil(lastNotarizedHeader) {
		return
	}

	shouldRequestHeaders := false

	defer func() {
		if !shouldRequestHeaders {
			latestValidHeader := bp.getLatestValidHeader(lastNotarizedHeader, longestChainHeaders)
			highestRound := bp.getHighestRoundInReceivedHeaders(latestValidHeader, sortedReceivedHeaders)
			bp.requestHeadersIfNothingNewIsReceived(lastNotarizedHeader.GetNonce(), latestValidHeader, highestRound)
		}
	}()

	numSortedReceivedHeaders := len(sortedReceivedHeaders)
	if numSortedReceivedHeaders == 0 {
		return
	}

	highestNonceReceived := sortedReceivedHeaders[numSortedReceivedHeaders-1].GetNonce()
	highestNonceInLongestChain := lastNotarizedHeader.GetNonce()
	numLongestChainHeaders := len(longestChainHeaders)
	if numLongestChainHeaders > 0 {
		highestNonceInLongestChain = longestChainHeaders[numLongestChainHeaders-1].GetNonce()
	}

	shouldRequestHeaders = highestNonceReceived > highestNonceInLongestChain+bp.blockFinality && numLongestChainHeaders == 0
	if !shouldRequestHeaders {
		return
	}

	log.Debug("requestHeadersIfNeeded",
		"shard", lastNotarizedHeader.GetShardID(),
		"last notarized nonce", lastNotarizedHeader.GetNonce(),
		"numSortedReceivedHeaders", numSortedReceivedHeaders,
		"numLongestChainHeaders", numLongestChainHeaders,
		"highest nonce received", highestNonceReceived,
		"highest nonce in longest chain", highestNonceInLongestChain)

	bp.requestHeaders(lastNotarizedHeader.GetShardID(), highestNonceInLongestChain+1)
}

func (bp *blockProcessor) getLatestValidHeader(
	lastNotarizedHeader data.HeaderHandler,
	longestChainHeaders []data.HeaderHandler,
) data.HeaderHandler {

	latestValidHeader := lastNotarizedHeader
	numLongestChainHeaders := len(longestChainHeaders)
	if numLongestChainHeaders > 0 {
		latestValidHeader = longestChainHeaders[numLongestChainHeaders-1]
	}

	return latestValidHeader
}

func (bp *blockProcessor) getHighestRoundInReceivedHeaders(
	latestValidHeader data.HeaderHandler,
	sortedReceivedHeaders []data.HeaderHandler,
) uint64 {

	if check.IfNil(latestValidHeader) {
		return 0
	}

	highestRound := latestValidHeader.GetRound()
	numSortedReceivedHeaders := len(sortedReceivedHeaders)
	if numSortedReceivedHeaders > 0 {
		highestRound = core.MaxUint64(highestRound, sortedReceivedHeaders[numSortedReceivedHeaders-1].GetRound())
	}

	return highestRound
}

func (bp *blockProcessor) requestHeadersIfNothingNewIsReceived(
	lastNotarizedHeaderNonce uint64,
	latestValidHeader data.HeaderHandler,
	highestRoundInReceivedHeaders uint64,
) {
	if check.IfNil(latestValidHeader) {
		return
	}

	shouldRequestHeaders := bp.rounder.Index()-int64(highestRoundInReceivedHeaders) > process.MaxRoundsWithoutNewBlockReceived &&
		int64(latestValidHeader.GetNonce())-int64(lastNotarizedHeaderNonce) <= process.MaxHeadersToRequestInAdvance
	if !shouldRequestHeaders {
		return
	}

	log.Debug("requestHeadersIfNothingNewIsReceived",
		"shard", latestValidHeader.GetShardID(),
		"latest valid header nonce", latestValidHeader.GetNonce(),
		"chronology round", bp.rounder.Index(),
		"highest round in received headers", highestRoundInReceivedHeaders)

	bp.requestHeaders(latestValidHeader.GetShardID(), latestValidHeader.GetNonce()+1)
}

func (bp *blockProcessor) requestHeaders(shardID uint32, fromNonce uint64) {
	toNonce := fromNonce + bp.blockFinality
	for nonce := fromNonce; nonce <= toNonce; nonce++ {
		log.Trace("requestHeaders.RequestHeaderByNonce",
			"shard", shardID,
			"nonce", nonce)

		if shardID == core.MetachainShardId {
			go bp.requestHandler.RequestMetaHeaderByNonce(nonce)
		} else {
			go bp.requestHandler.RequestShardHeaderByNonce(shardID, nonce)
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (bp *blockProcessor) IsInterfaceNil() bool {
	return bp == nil
}

func checkBlockProcessorNilParameters(arguments ArgBlockProcessor) error {
	if check.IfNil(arguments.HeaderValidator) {
		return process.ErrNilHeaderValidator
	}
	if check.IfNil(arguments.RequestHandler) {
		return process.ErrNilRequestHandler
	}
	if check.IfNil(arguments.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(arguments.BlockTracker) {
		return ErrNilBlockTrackerHandler
	}
	if check.IfNil(arguments.CrossNotarizer) {
		return ErrNilCrossNotarizer
	}
	if check.IfNil(arguments.SelfNotarizer) {
		return ErrNilSelfNotarizer
	}
	if check.IfNil(arguments.CrossNotarizedHeadersNotifier) {
		return ErrNilCrossNotarizedHeadersNotifier
	}
	if check.IfNil(arguments.SelfNotarizedHeadersNotifier) {
		return ErrNilSelfNotarizedHeadersNotifier
	}
	if check.IfNil(arguments.FinalMetachainHeadersNotifier) {
		return ErrNilFinalMetachainHeadersNotifier
	}
	if check.IfNil(arguments.Rounder) {
		return ErrNilRounder
	}

	return nil
}
