package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

// RatingsProcessorStub -
type RatingsProcessorStub struct {
	IndexRatingsForEpochStartMetaBlockCalled func(metaBlock data.MetaHeaderHandler) error
}

// IndexRatingsForEpochStartMetaBlock -
func (r *RatingsProcessorStub) IndexRatingsForEpochStartMetaBlock(metaBlock data.MetaHeaderHandler) error {
	if r.IndexRatingsForEpochStartMetaBlockCalled != nil {
		return r.IndexRatingsForEpochStartMetaBlockCalled(metaBlock)
	}

	return nil
}

// IsInterfaceNil -
func (r *RatingsProcessorStub) IsInterfaceNil() bool {
	return r == nil
}
