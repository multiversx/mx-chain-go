package track

import (
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignBlockTrackCreator struct {
	args ArgShardTracker
}

// NewSovereignBlockTrackCreator creates a sovereign block track creator
func NewSovereignBlockTrackCreator(args ArgShardTracker) *sovereignBlockTrackCreator {
	return &sovereignBlockTrackCreator{
		args: args,
	}
}

// CreateBlockTracker creates a new sovereign block track
func (creator *sovereignBlockTrackCreator) CreateBlockTracker() (process.BlockTracker, error) {
	blockTracker, err := NewShardBlockTrack(creator.args)
	if err != nil {
		return nil, err
	}

	return NewSovereignChainShardBlockTrack(blockTracker)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (creator *sovereignBlockTrackCreator) IsInterfaceNil() bool {
	return creator == nil
}
