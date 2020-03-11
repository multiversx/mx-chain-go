package track

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
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

	if check.IfNil(arguments.PoolsHolder) {
		return nil, process.ErrNilPoolsHolder
	}
	if check.IfNil(arguments.PoolsHolder.Headers()) {
		return nil, process.ErrNilHeadersDataPool
	}

	crossNotarizer, err := NewBlockNotarizer(arguments.Hasher, arguments.Marshalizer)
	if err != nil {
		return nil, err
	}

	selfNotarizer, err := NewBlockNotarizer(arguments.Hasher, arguments.Marshalizer)
	if err != nil {
		return nil, err
	}

	crossNotarizedHeadersNotifier, err := NewBlockNotifier()
	if err != nil {
		return nil, err
	}

	selfNotarizedHeadersNotifier, err := NewBlockNotifier()
	if err != nil {
		return nil, err
	}

	blockBalancerInstance, err := NewBlockBalancer()
	if err != nil {
		return nil, err
	}

	bbt := &baseBlockTrack{
		hasher:                        arguments.Hasher,
		headerValidator:               arguments.HeaderValidator,
		marshalizer:                   arguments.Marshalizer,
		rounder:                       arguments.Rounder,
		shardCoordinator:              arguments.ShardCoordinator,
		headersPool:                   arguments.PoolsHolder.Headers(),
		store:                         arguments.Store,
		crossNotarizer:                crossNotarizer,
		selfNotarizer:                 selfNotarizer,
		crossNotarizedHeadersNotifier: crossNotarizedHeadersNotifier,
		selfNotarizedHeadersNotifier:  selfNotarizedHeadersNotifier,
		blockBalancer:                 blockBalancerInstance,
	}

	err = bbt.initNotarizedHeaders(arguments.StartHeaders)
	if err != nil {
		return nil, err
	}

	mbt := metaBlockTrack{
		baseBlockTrack: bbt,
	}

	argBlockProcessor := ArgBlockProcessor{
		HeaderValidator:               arguments.HeaderValidator,
		RequestHandler:                arguments.RequestHandler,
		ShardCoordinator:              arguments.ShardCoordinator,
		BlockTracker:                  &mbt,
		CrossNotarizer:                crossNotarizer,
		CrossNotarizedHeadersNotifier: crossNotarizedHeadersNotifier,
		SelfNotarizedHeadersNotifier:  selfNotarizedHeadersNotifier,
		Rounder:                       arguments.Rounder,
	}

	blockProcessorObject, err := NewBlockProcessor(argBlockProcessor)
	if err != nil {
		return nil, err
	}

	mbt.blockProcessor = blockProcessorObject

	mbt.headers = make(map[uint32]map[uint64][]*HeaderInfo)
	mbt.headersPool.RegisterHandler(mbt.receivedHeader)

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
			continue
		}

		selfMetaBlocksInfo = append(selfMetaBlocksInfo, &HeaderInfo{Hash: metaBlockHash, Header: metaBlock})
	}

	return selfMetaBlocksInfo
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

// ComputeNumPendingMiniBlocks computes the number of pending miniblocks from a given slice of headers
func (mbt *metaBlockTrack) ComputeNumPendingMiniBlocks(_ []data.HeaderHandler) {
}
