package track

import (
	"github.com/multiversx/mx-chain-go/process"
)

type shardBlockTrackCreator struct {
	args ArgShardTracker
}

// NewShardBlockTrackCreator creates a shard block track creator
func NewShardBlockTrackCreator(args ArgShardTracker) *shardBlockTrackCreator {
	return &shardBlockTrackCreator{
		args: args,
	}
}

// CreateBlockTracker creates a shard block track
func (creator *shardBlockTrackCreator) CreateBlockTracker() (process.BlockTracker, error) {
	return NewShardBlockTrack(creator.args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (creator *shardBlockTrackCreator) IsInterfaceNil() bool {
	return creator == nil
}
