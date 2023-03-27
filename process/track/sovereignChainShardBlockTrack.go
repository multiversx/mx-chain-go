package track

import (
	"errors"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignChainShardBlockTrack struct {
	*shardBlockTrack
}

// NewSovereignChainShardBlockTrack creates an object for tracking the received shard blocks
func NewSovereignChainShardBlockTrack(shardBlockTrack *shardBlockTrack) (*sovereignChainShardBlockTrack, error) {
	if shardBlockTrack == nil {
		return nil, process.ErrNilBlockTracker
	}

	scsbt := &sovereignChainShardBlockTrack{
		shardBlockTrack,
	}

	bp, ok := scsbt.blockProcessor.(*blockProcessor)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	scbp, err := NewSovereignChainBlockProcessor(bp)
	if err != nil {
		return nil, err
	}

	scsbt.blockProcessor = scbp
	scsbt.receivedHeaderFunc = scsbt.receivedHeader

	return scsbt, nil
}

// ComputeLongestSelfChain computes the longest chain from self shard
func (scsbt *sovereignChainShardBlockTrack) ComputeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, err := scsbt.selfNotarizer.GetLastNotarizedHeader(scsbt.shardCoordinator.SelfId())
	if err != nil {
		log.Warn("ComputeLongestSelfChain.GetLastNotarizedHeader", "error", err.Error())
		return nil, nil, nil, nil
	}

	headers, hashes := scsbt.ComputeLongestChain(scsbt.shardCoordinator.SelfId(), lastSelfNotarizedHeader)
	return lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, headers, hashes
}

// GetSelfNotarizedHeader returns a self notarized header for self shard with a given offset, behind last self notarized header
func (scsbt *sovereignChainShardBlockTrack) GetSelfNotarizedHeader(_ uint32, offset uint64) (data.HeaderHandler, []byte, error) {
	return scsbt.selfNotarizer.GetNotarizedHeader(scsbt.shardCoordinator.SelfId(), offset)
}

func (scsbt *sovereignChainShardBlockTrack) receivedHeader(headerHandler data.HeaderHandler, headerHash []byte) {
	extendedShardHeader, isExtendedShardHeaderReceived := headerHandler.(*block.ShardHeaderExtended)
	if isExtendedShardHeaderReceived {
		scsbt.receivedExtendedShardHeader(extendedShardHeader, headerHash)
		return
	}

	scsbt.receivedShardHeader(headerHandler, headerHash)
}

func (scsbt *sovereignChainShardBlockTrack) receivedExtendedShardHeader(
	extendedShardHeaderHandler data.ShardHeaderExtendedHandler,
	extendedShardHeaderHash []byte,
) {

	log.Debug("received extended shard header from network in block tracker",
		"shard", extendedShardHeaderHandler.GetShardID(),
		"epoch", extendedShardHeaderHandler.GetEpoch(),
		"round", extendedShardHeaderHandler.GetRound(),
		"nonce", extendedShardHeaderHandler.GetNonce(),
		"hash", extendedShardHeaderHash,
	)

	if !scsbt.shouldAddExtendedShardHeader(extendedShardHeaderHandler) {
		log.Trace("received extended shard header is out of range", "nonce", extendedShardHeaderHandler.GetNonce())
		return
	}

	if !scsbt.addHeader(extendedShardHeaderHandler, extendedShardHeaderHash, core.SovereignChainShardId) {
		log.Trace("received extended shard header was not added", "nonce", extendedShardHeaderHandler.GetNonce())
		return
	}

	scsbt.doWhitelistWithExtendedShardHeaderIfNeeded(extendedShardHeaderHandler)
	scsbt.blockProcessor.ProcessReceivedHeader(extendedShardHeaderHandler)
}

func (scsbt *sovereignChainShardBlockTrack) shouldAddExtendedShardHeader(extendedShardHeaderHandler data.ShardHeaderExtendedHandler) bool {
	lastNotarizedHeader, _, err := scsbt.crossNotarizer.GetLastNotarizedHeader(core.SovereignChainShardId)

	isFirstCrossNotarizedHeader := err != nil && errors.Is(err, process.ErrNotarizedHeadersSliceForShardIsNil)
	if isFirstCrossNotarizedHeader {
		return true
	}

	if err != nil {
		log.Debug("shouldAddExtendedShardHeader.GetLastNotarizedHeader",
			"shard", extendedShardHeaderHandler.GetShardID(),
			"error", err.Error())
		return false
	}

	lastNotarizedHeaderNonce := lastNotarizedHeader.GetNonce()

	isHeaderOutOfRange := extendedShardHeaderHandler.GetNonce() > lastNotarizedHeaderNonce+uint64(scsbt.maxNumHeadersToKeepPerShard)
	return !isHeaderOutOfRange
}

func (scsbt *sovereignChainShardBlockTrack) doWhitelistWithExtendedShardHeaderIfNeeded(extendedShardHeaderHandler data.ShardHeaderExtendedHandler) {
	if check.IfNil(extendedShardHeaderHandler) {
		return
	}
	if scsbt.isExtendedShardHeaderOutOfRange(extendedShardHeaderHandler) {
		return
	}

	miniBlockHandlers := extendedShardHeaderHandler.GetIncomingMiniBlockHandlers()

	keys := make([][]byte, 0)
	for _, miniBlockHandler := range miniBlockHandlers {
		miniBlock, ok := miniBlockHandler.(*block.MiniBlock)
		if !ok {
			log.Warn("doWhitelistWithExtendedShardHeaderIfNeeded", "error", process.ErrWrongTypeAssertion)
			continue
		}

		keys = append(keys, miniBlock.TxHashes...)
	}

	scsbt.whitelistHandler.Add(keys)
}

func (scsbt *sovereignChainShardBlockTrack) isExtendedShardHeaderOutOfRange(extendedShardHeaderHandler data.ShardHeaderExtendedHandler) bool {
	lastCrossNotarizedHeader, _, err := scsbt.GetLastCrossNotarizedHeader(core.SovereignChainShardId)

	isFirstCrossNotarizedHeader := err != nil && errors.Is(err, process.ErrNotarizedHeadersSliceForShardIsNil)
	if isFirstCrossNotarizedHeader {
		return false
	}

	if err != nil {
		log.Debug("isExtendedShardHeaderOutOfRange.GetLastCrossNotarizedHeader",
			"shard", extendedShardHeaderHandler.GetShardID(),
			"error", err.Error())
		return true
	}

	isExtendedShardHeaderOutOfRange := extendedShardHeaderHandler.GetNonce() > lastCrossNotarizedHeader.GetNonce()+process.MaxHeadersToWhitelistInAdvance
	return isExtendedShardHeaderOutOfRange
}

// ComputeLongestExtendedShardChainFromLastNotarized returns the longest valid chain for extended shard chain from its last cross notarized header
func (scsbt *sovereignChainShardBlockTrack) ComputeLongestExtendedShardChainFromLastNotarized() ([]data.HeaderHandler, [][]byte, error) {
	lastCrossNotarizedHeader, _, err := scsbt.GetLastCrossNotarizedHeader(core.SovereignChainShardId)
	if err != nil {
		return nil, nil, err
	}

	hdrsForShard, hdrsHashesForShard := scsbt.ComputeLongestChain(core.SovereignChainShardId, lastCrossNotarizedHeader)

	return hdrsForShard, hdrsHashesForShard, nil
}
