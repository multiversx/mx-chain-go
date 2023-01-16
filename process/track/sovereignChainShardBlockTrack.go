package track

import (
	"github.com/multiversx/mx-chain-core-go/data"
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
