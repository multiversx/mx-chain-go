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

	return &sideChainShardBlockTrack{
		shardBlockTrack,
	}, nil
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
