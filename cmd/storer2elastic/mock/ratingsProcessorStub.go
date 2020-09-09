package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// RatingsProcessorStub -
type RatingsProcessorStub struct {
	IndexRatingsForEpochStartMetaBlockCalled func(metaBlock *block.MetaBlock) error
}

// IndexRatingsForEpochStartMetaBlock -
func (r *RatingsProcessorStub) IndexRatingsForEpochStartMetaBlock(metaBlock *block.MetaBlock) error {
	if r.IndexRatingsForEpochStartMetaBlockCalled != nil {
		return r.IndexRatingsForEpochStartMetaBlockCalled(metaBlock)
	}

	return nil
}

// IsInterfaceNil -
func (r *RatingsProcessorStub) IsInterfaceNil() bool {
	return r == nil
}
