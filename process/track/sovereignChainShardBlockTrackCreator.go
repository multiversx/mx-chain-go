package track

import (
	"github.com/multiversx/mx-chain-go/process"
)

type sovereignBlockTrackCreator struct{}

// NewSovereignBlockTrackCreator creates a sovereign block track creator
func NewSovereignBlockTrackCreator() *sovereignBlockTrackCreator {
	return &sovereignBlockTrackCreator{}
}

// CreateBlockTracker creates a new sovereign block track
func (creator *sovereignBlockTrackCreator) CreateBlockTracker(args ArgBaseTracker) (process.BlockTracker, error) {
	blockTracker, err := NewShardBlockTrack(ArgShardTracker{ArgBaseTracker: args})
	if err != nil {
		return nil, err
	}

	return NewSovereignChainShardBlockTrack(blockTracker)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (creator *sovereignBlockTrackCreator) IsInterfaceNil() bool {
	return creator == nil
}
