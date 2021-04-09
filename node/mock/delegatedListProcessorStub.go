package mock

import "github.com/ElrondNetwork/elrond-go/data/api"

// DelegatedListProcessorStub -
type DelegatedListProcessorStub struct {
	GetDelegatorsListCalled func() ([]*api.Delegator, error)
}

// GetDelegatorsList -
func (dlps *DelegatedListProcessorStub) GetDelegatorsList() ([]*api.Delegator, error) {
	if dlps.GetDelegatorsListCalled != nil {
		return dlps.GetDelegatorsListCalled()
	}

	return nil, nil
}

// IsInterfaceNil -
func (dlps *DelegatedListProcessorStub) IsInterfaceNil() bool {
	return dlps == nil
}
