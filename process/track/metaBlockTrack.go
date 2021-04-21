package track

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

type metaBlockTrack struct {
	*baseBlockTrack
}

// NewMetaBlockTrack creates an object for tracking the received meta blocks
func NewMetaBlockTrack(arguments ArgMetaTracker) (*metaBlockTrack, error) {
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

	mbt := metaBlockTrack{
		baseBlockTrack: bbt,
	}

	argBlockProcessor := ArgBlockProcessor{
		HeaderValidator:                       arguments.HeaderValidator,
		RequestHandler:                        arguments.RequestHandler,
		ShardCoordinator:                      arguments.ShardCoordinator,
		BlockTracker:                          &mbt,
		CrossNotarizer:                        bbt.crossNotarizer,
		SelfNotarizer:                         bbt.selfNotarizer,
		CrossNotarizedHeadersNotifier:         bbt.crossNotarizedHeadersNotifier,
		SelfNotarizedFromCrossHeadersNotifier: bbt.selfNotarizedFromCrossHeadersNotifier,
		SelfNotarizedHeadersNotifier:          bbt.selfNotarizedHeadersNotifier,
		FinalMetachainHeadersNotifier:         bbt.finalMetachainHeadersNotifier,
		RoundHandler:                          arguments.RoundHandler,
	}

	blockProcessorObject, err := NewBlockProcessor(argBlockProcessor)
	if err != nil {
		return nil, err
	}

	mbt.blockProcessor = blockProcessorObject
	mbt.headers = make(map[uint32]map[uint64][]*HeaderInfo)
	mbt.headersPool.RegisterHandler(mbt.receivedHeader)
	mbt.headersPool.Clear()

	return &mbt, nil
}

// GetSelfHeaders gets a slice of self headers from a given header
func (mbt *metaBlockTrack) GetSelfHeaders(headerHandler data.HeaderHandler) []*HeaderInfo {
	selfMetaBlocksInfo := make([]*HeaderInfo, 0)

	header, ok := headerHandler.(*block.Header)
	if !ok {
		log.Debug("GetSelfHeaders", "error", process.ErrWrongTypeAssertion)
		return selfMetaBlocksInfo
	}

	for _, metaBlockHash := range header.MetaBlockHashes {
		metaBlock, err := process.GetMetaHeader(metaBlockHash, mbt.headersPool, mbt.marshalizer, mbt.store)
		if err != nil {
			log.Trace("GetSelfHeaders.GetMetaHeader", "error", err.Error())

			metaBlock, err = mbt.getTrackedMetaBlockWithHash(metaBlockHash)
			if err != nil {
				log.Trace("GetSelfHeaders.getTrackedMetaBlockWithHash", "error", err.Error())
				continue
			}
		}

		selfMetaBlocksInfo = append(selfMetaBlocksInfo, &HeaderInfo{Hash: metaBlockHash, Header: metaBlock})
	}

	return selfMetaBlocksInfo
}

func (mbt *metaBlockTrack) getTrackedMetaBlockWithHash(hash []byte) (*block.MetaBlock, error) {
	metaBlocks, metaBlocksHashes := mbt.GetTrackedHeaders(core.MetachainShardId)
	for i := 0; i < len(metaBlocks); i++ {
		if !bytes.Equal(metaBlocksHashes[i], hash) {
			continue
		}

		metaBlock, ok := metaBlocks[i].(*block.MetaBlock)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		return metaBlock, nil
	}

	return nil, process.ErrMissingHeader
}

// CleanupInvalidCrossHeaders cleans headers added to the block tracker that have become invalid after processing
// For metachain some shard headers may become invalid if the attesting header for a start of epoch block changes
// due to a rollback
func (mbt *metaBlockTrack) CleanupInvalidCrossHeaders(metaNewEpoch uint32, metaRoundAttestingEpoch uint64) {
	for shardID := uint32(0); shardID < mbt.shardCoordinator.NumberOfShards(); shardID++ {
		shardHeader, _, err := mbt.crossNotarizer.GetLastNotarizedHeader(shardID)
		if err != nil {
			log.Warn("get last notarized headers", "error", err.Error())
			continue
		}

		mbt.removeInvalidShardHeadersDueToEpochChange(shardID, shardHeader.GetNonce(), metaRoundAttestingEpoch, metaNewEpoch)
	}
}

func (mbt *metaBlockTrack) removeInvalidShardHeadersDueToEpochChange(
	shardID uint32,
	shardNotarizedNonce uint64,
	metaRoundAttestingEpoch uint64,
	metaNewEpoch uint32,
) {
	mbt.mutHeaders.Lock()
	defer mbt.mutHeaders.Unlock()

	for nonce, headersInfo := range mbt.headers[shardID] {
		if nonce <= shardNotarizedNonce {
			continue
		}

		newHeadersInfo := make([]*HeaderInfo, 0, len(headersInfo))

		for _, headerInfo := range headersInfo {
			round := headerInfo.Header.GetRound()
			epoch := headerInfo.Header.GetEpoch()
			isInvalidHeader := round > metaRoundAttestingEpoch+process.EpochChangeGracePeriod && epoch < metaNewEpoch
			if !isInvalidHeader {
				newHeadersInfo = append(newHeadersInfo, headerInfo)
			}
		}

		mbt.headers[shardID][nonce] = newHeadersInfo
	}
}

// ComputeLongestSelfChain computes the longest chain from self shard
func (mbt *metaBlockTrack) ComputeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, err := mbt.selfNotarizer.GetLastNotarizedHeader(mbt.shardCoordinator.SelfId())
	if err != nil {
		log.Warn("ComputeLongestSelfChain.GetLastNotarizedHeader", "error", err.Error())
		return nil, nil, nil, nil
	}

	headers, hashes := mbt.ComputeLongestChain(mbt.shardCoordinator.SelfId(), lastSelfNotarizedHeader)
	return lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, headers, hashes
}

// ComputeCrossInfo computes cross info from a given slice of headers
func (mbt *metaBlockTrack) ComputeCrossInfo(_ []data.HeaderHandler) {
}
