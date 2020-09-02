package track

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

type shardBlockTrack struct {
	*baseBlockTrack
}

// NewShardBlockTrack creates an object for tracking the received shard blocks
func NewShardBlockTrack(arguments ArgShardTracker) (*shardBlockTrack, error) {
	err := checkTrackerNilParameters(arguments.ArgBaseTracker)
	if err != nil {
		return nil, err
	}

	bbt, err := createBaseBlockTrack(arguments.ArgBaseTracker)
	if err != nil {
		return nil, err
	}

	err = bbt.initNotarizedHeaders(arguments.StartHeaders)
	if err != nil {
		return nil, err
	}

	sbt := shardBlockTrack{
		baseBlockTrack: bbt,
	}

	argBlockProcessor := ArgBlockProcessor{
		HeaderValidator:               arguments.HeaderValidator,
		RequestHandler:                arguments.RequestHandler,
		ShardCoordinator:              arguments.ShardCoordinator,
		BlockTracker:                  &sbt,
		CrossNotarizer:                bbt.crossNotarizer,
		SelfNotarizer:                 bbt.selfNotarizer,
		CrossNotarizedHeadersNotifier: bbt.crossNotarizedHeadersNotifier,
		SelfNotarizedHeadersNotifier:  bbt.selfNotarizedHeadersNotifier,
		FinalMetachainHeadersNotifier: bbt.finalMetachainHeadersNotifier,
		Rounder:                       arguments.Rounder,
	}

	blockProcessorObject, err := NewBlockProcessor(argBlockProcessor)
	if err != nil {
		return nil, err
	}

	sbt.blockProcessor = blockProcessorObject
	sbt.headers = make(map[uint32]map[uint64][]*HeaderInfo)
	sbt.headersPool.RegisterHandler(sbt.receivedHeader)
	sbt.headersPool.Clear()

	return &sbt, nil
}

// GetSelfHeaders gets a slice of self headers from a given metablock
func (sbt *shardBlockTrack) GetSelfHeaders(headerHandler data.HeaderHandler) []*HeaderInfo {
	selfHeadersInfo := make([]*HeaderInfo, 0)

	metaBlock, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		log.Debug("GetSelfHeaders", "error", process.ErrWrongTypeAssertion)
		return selfHeadersInfo
	}

	for _, shardInfo := range metaBlock.ShardInfo {
		if shardInfo.ShardID != sbt.shardCoordinator.SelfId() {
			continue
		}

		header, err := process.GetShardHeader(shardInfo.HeaderHash, sbt.headersPool, sbt.marshalizer, sbt.store)
		if err != nil {
			log.Trace("GetSelfHeaders.GetShardHeader", "error", err.Error())

			header, err = sbt.getTrackedShardHeaderWithNonceAndHash(shardInfo.ShardID, shardInfo.Nonce, shardInfo.HeaderHash)
			if err != nil {
				log.Trace("GetSelfHeaders.getTrackedShardHeaderWithNonceAndHash", "error", err.Error())
				continue
			}
		}

		selfHeadersInfo = append(selfHeadersInfo, &HeaderInfo{Hash: shardInfo.HeaderHash, Header: header})
	}

	return selfHeadersInfo
}

func (sbt *shardBlockTrack) getTrackedShardHeaderWithNonceAndHash(
	shardID uint32,
	nonce uint64,
	hash []byte,
) (*block.Header, error) {

	headers, headersHashes := sbt.GetTrackedHeadersWithNonce(shardID, nonce)
	for i := 0; i < len(headers); i++ {
		if bytes.Compare(headersHashes[i], hash) != 0 {
			continue
		}

		header, ok := headers[i].(*block.Header)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		return header, nil
	}

	return nil, process.ErrMissingHeader
}

// CleanupInvalidCrossHeaders cleans headers added to the block tracker that have become invalid after processing
func (sbt *shardBlockTrack) CleanupInvalidCrossHeaders(_ uint32, _ uint64) {
	// no rule for shard
}

// ComputeLongestSelfChain computes the longest chain from self shard
func (sbt *shardBlockTrack) ComputeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, err := sbt.selfNotarizer.GetLastNotarizedHeader(core.MetachainShardId)
	if err != nil {
		log.Warn("ComputeLongestSelfChain.GetLastNotarizedHeader", "error", err.Error())
		return nil, nil, nil, nil
	}

	headers, hashes := sbt.ComputeLongestChain(sbt.shardCoordinator.SelfId(), lastSelfNotarizedHeader)
	return lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, headers, hashes
}

// ComputeCrossInfo computes the cross info from a given slice of metablocks
func (sbt *shardBlockTrack) ComputeCrossInfo(headers []data.HeaderHandler) {
	lenHeaders := len(headers)
	if lenHeaders == 0 {
		return
	}

	metaBlock, ok := headers[lenHeaders-1].(*block.MetaBlock)
	if !ok {
		log.Debug("ComputeCrossInfo", "error", process.ErrWrongTypeAssertion)
		return
	}

	for _, shardInfo := range metaBlock.ShardInfo {
		sbt.blockBalancer.SetNumPendingMiniBlocks(shardInfo.ShardID, shardInfo.NumPendingMiniBlocks)
		sbt.blockBalancer.SetLastShardProcessedMetaNonce(shardInfo.ShardID, shardInfo.LastIncludedMetaNonce)
	}

	log.Debug("compute cross info from meta block",
		"epoch", metaBlock.Epoch,
		"round", metaBlock.Round,
		"nonce", metaBlock.Nonce,
	)

	for shardID := uint32(0); shardID < sbt.shardCoordinator.NumberOfShards(); shardID++ {
		log.Debug("cross info",
			"shard", shardID,
			"pending miniblocks", sbt.blockBalancer.GetNumPendingMiniBlocks(shardID),
			"last meta nonce processed", sbt.blockBalancer.GetLastShardProcessedMetaNonce(shardID))
	}
}
