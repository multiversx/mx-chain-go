package track

import (
	"github.com/multiversx/mx-chain-go/process"
)

type shardBlockTrackCreator struct{}

// NewShardBlockTrackCreator creates a shard block track creator
func NewShardBlockTrackCreator() *shardBlockTrackCreator {
	return &shardBlockTrackCreator{}
}

// CreateBlockTracker creates a shard block track
func (creator *shardBlockTrackCreator) CreateBlockTracker(args ArgBaseTracker) (process.BlockTracker, error) {
	return NewShardBlockTrack(ArgShardTracker{ArgBaseTracker: args})
}

// IsInterfaceNil checks if the underlying pointer is nil
func (creator *shardBlockTrackCreator) IsInterfaceNil() bool {
	return creator == nil
}
