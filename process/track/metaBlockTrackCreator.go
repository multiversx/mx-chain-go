package track

import (
	"github.com/multiversx/mx-chain-go/process"
)

type metaBlockTrackCreator struct {
	args ArgMetaTracker
}

// NewMetaBlockTrackCreator creates a shard block track creator
func NewMetaBlockTrackCreator(args ArgMetaTracker) *metaBlockTrackCreator {
	return &metaBlockTrackCreator{
		args: args,
	}
}

// CreateBlockTracker creates a shard block track
func (creator *metaBlockTrackCreator) CreateBlockTracker() (process.BlockTracker, error) {
	return NewMetaBlockTrack(creator.args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (creator *metaBlockTrackCreator) IsInterfaceNil() bool {
	return creator == nil
}
