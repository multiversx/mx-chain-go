package mock

import (
	"context"

	"github.com/ElrondNetwork/elrond-go-core/data/api"
)

// DirectStakedListProcessorStub -
type DirectStakedListProcessorStub struct {
	GetDirectStakedListCalled func(ctx context.Context) ([]*api.DirectStakedValue, error)
}

// GetDirectStakedList -
func (dslps *DirectStakedListProcessorStub) GetDirectStakedList(ctx context.Context) ([]*api.DirectStakedValue, error) {
	if dslps.GetDirectStakedListCalled != nil {
		return dslps.GetDirectStakedListCalled(ctx)
	}

	return nil, nil
}

// IsInterfaceNil -
func (dslps *DirectStakedListProcessorStub) IsInterfaceNil() bool {
	return dslps == nil
}
