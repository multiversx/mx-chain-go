package mock

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/data/api"
)

// DelegatedListProcessorStub -
type DelegatedListProcessorStub struct {
	GetDelegatorsListCalled func(ctx context.Context) ([]*api.Delegator, error)
}

// GetDelegatorsList -
func (dlps *DelegatedListProcessorStub) GetDelegatorsList(ctx context.Context) ([]*api.Delegator, error) {
	if dlps.GetDelegatorsListCalled != nil {
		return dlps.GetDelegatorsListCalled(ctx)
	}

	return nil, nil
}

// IsInterfaceNil -
func (dlps *DelegatedListProcessorStub) IsInterfaceNil() bool {
	return dlps == nil
}
