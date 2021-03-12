package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// RatingsProcessorStub -
type RatingsProcessorStub struct {
	IndexRatingsForEpochStartMetaBlockCalled func(metaBlock data.HeaderHandler) error
}

// IndexRatingsForEpochStartMetaBlock -
func (r *RatingsProcessorStub) IndexRatingsForEpochStartMetaBlock(metaBlock data.HeaderHandler) error {
	if r.IndexRatingsForEpochStartMetaBlockCalled != nil {
		return r.IndexRatingsForEpochStartMetaBlockCalled(metaBlock)
	}

	return nil
}

// IsInterfaceNil -
func (r *RatingsProcessorStub) IsInterfaceNil() bool {
	return r == nil
}
