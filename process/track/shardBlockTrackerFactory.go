package track

import "github.com/multiversx/mx-chain-go/process"

type shardBlockTrackerFactory struct {
}

// NewShardBlockTrackerFactory creates a new shard block tracker factory
func NewShardBlockTrackerFactory() (*shardBlockTrackerFactory, error) {
	return &shardBlockTrackerFactory{}, nil
}

// CreateBlockTracker creates a new block tracker for shard chain
func (sbtcf *shardBlockTrackerFactory) CreateBlockTracker(argBaseTracker ArgShardTracker) (process.BlockTracker, error) {
	return NewShardBlockTrack(argBaseTracker)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sbtcf *shardBlockTrackerFactory) IsInterfaceNil() bool {
	return sbtcf == nil
}
