package track

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type sideChainShardBlockTrack struct {
	*shardBlockTrack
}

// NewSideChainShardBlockTrack creates an object for tracking the received shard blocks
func NewSideChainShardBlockTrack(shardBlockTrack *shardBlockTrack) (*sideChainShardBlockTrack, error) {
	if shardBlockTrack == nil {
		return nil, process.ErrNilBlockTracker
	}

	sideChainBlockTrack := &sideChainShardBlockTrack{
		shardBlockTrack,
	}

	originalBlockProcessor, ok := sideChainBlockTrack.blockProcessor.(*blockProcessor)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	newBlockProcessor, err := NewSideChainBlockProcessor(originalBlockProcessor)
	if err != nil {
		return nil, err
	}

	sideChainBlockTrack.blockProcessor = newBlockProcessor
	return sideChainBlockTrack, nil
}

// ComputeLongestSelfChain computes the longest chain from self shard
func (scsbt *sideChainShardBlockTrack) ComputeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, err := scsbt.selfNotarizer.GetLastNotarizedHeader(scsbt.shardCoordinator.SelfId())
	if err != nil {
		log.Warn("ComputeLongestSelfChain.GetLastNotarizedHeader", "error", err.Error())
		return nil, nil, nil, nil
	}

	headers, hashes := scsbt.ComputeLongestChain(scsbt.shardCoordinator.SelfId(), lastSelfNotarizedHeader)
	return lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, headers, hashes
}

// GetSelfNotarizedHeader returns a self notarized header for self shard with a given offset, behind last self notarized header
func (scsbt *sideChainShardBlockTrack) GetSelfNotarizedHeader(_ uint32, offset uint64) (data.HeaderHandler, []byte, error) {
	return scsbt.selfNotarizer.GetNotarizedHeader(scsbt.shardCoordinator.SelfId(), offset)
}
