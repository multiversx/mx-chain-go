package mock

import "github.com/ElrondNetwork/elrond-go/data/api"

// DirectStakedListProcessorStub -
type DirectStakedListProcessorStub struct {
	GetDirectStakedListCalled func() ([]*api.DirectStakedValue, error)
}

// GetDirectStakedList -
func (dslps *DirectStakedListProcessorStub) GetDirectStakedList() ([]*api.DirectStakedValue, error) {
	if dslps.GetDirectStakedListCalled != nil {
		return dslps.GetDirectStakedListCalled()
	}

	return nil, nil
}

// IsInterfaceNil -
func (dslps *DirectStakedListProcessorStub) IsInterfaceNil() bool {
	return dslps == nil
}
