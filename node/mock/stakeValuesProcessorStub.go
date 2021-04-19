package mock

import "github.com/ElrondNetwork/elrond-go/data/api"

// StakeValuesProcessorStub -
type StakeValuesProcessorStub struct {
	GetTotalStakedValueCalled func() (*api.StakeValues, error)
}

// GetTotalStakedValue -
func (svps *StakeValuesProcessorStub) GetTotalStakedValue() (*api.StakeValues, error) {
	if svps.GetTotalStakedValueCalled != nil {
		return svps.GetTotalStakedValueCalled()
	}

	return nil, nil
}

// IsInterfaceNil -
func (svps *StakeValuesProcessorStub) IsInterfaceNil() bool {
	return svps == nil
}
