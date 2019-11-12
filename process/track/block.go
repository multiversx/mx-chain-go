package track

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type blockTrack struct {
	rounder consensus.Rounder

	mutLastHeaders sync.RWMutex
	lastHeaders    map[uint32]data.HeaderHandler
}

// NewBlockTrack creates an object for tracking the received blocks
func NewBlockTrack(rounder consensus.Rounder) (*blockTrack, error) {
	if check.IfNil(rounder) {
		return nil, process.ErrNilRounder
	}

	bt := blockTrack{rounder: rounder}
	bt.lastHeaders = make(map[uint32]data.HeaderHandler)

	return &bt, nil
}

// SetLastHeaderForShard sets the given header for the given shard as the last received if it has the highest round
func (bt *blockTrack) SetLastHeaderForShard(header data.HeaderHandler) {
	if check.IfNil(header) {
		return
	}

	bt.mutLastHeaders.Lock()
	defer bt.mutLastHeaders.Unlock()

	shardID := header.GetShardID()

	lastHeader, ok := bt.lastHeaders[shardID]
	if ok && lastHeader.GetRound() > header.GetRound() {
		return
	}

	bt.lastHeaders[shardID] = header
}

// LastHeaderForShard return the last header received (highest round) for the given shard
func (bt *blockTrack) LastHeaderForShard(shardId uint32) data.HeaderHandler {
	bt.mutLastHeaders.RLock()
	defer bt.mutLastHeaders.RUnlock()

	return bt.lastHeaders[shardId]
}

// IsShardStuck return true if the given shard is stuck
func (bt *blockTrack) IsShardStuck(shardId uint32) bool {
	header := bt.LastHeaderForShard(shardId)
	if check.IfNil(header) {
		return false
	}

	isShardStuck := bt.rounder.Index()-int64(header.GetRound()) > process.MaxRoundsWithoutCommittedBlock
	return isShardStuck
}

// IsInterfaceNil returns true if there is no value under the interface
func (bt *blockTrack) IsInterfaceNil() bool {
	return bt == nil
}
