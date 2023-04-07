package track

import (
	"github.com/multiversx/mx-chain-go/process"
)

type metaBlockTrackCreator struct{}

// NewMetaBlockTrackCreator creates a shard block track creator
func NewMetaBlockTrackCreator() *metaBlockTrackCreator {
	return &metaBlockTrackCreator{}
}

// CreateBlockTracker creates a shard block track
func (creator *metaBlockTrackCreator) CreateBlockTracker(args ArgBaseTracker) (process.BlockTracker, error) {
	return NewMetaBlockTrack(ArgMetaTracker{ArgBaseTracker: args})
}

// IsInterfaceNil checks if the underlying pointer is nil
func (creator *metaBlockTrackCreator) IsInterfaceNil() bool {
	return creator == nil
}
