package stakeValuesProc

import "github.com/ElrondNetwork/elrond-go/data/api"

type disabledTotalStakedValueProcessor struct{}

// NewDisabledTotalStakedValueProcessor -
func NewDisabledTotalStakedValueProcessor() (*disabledTotalStakedValueProcessor, error) {
	return new(disabledTotalStakedValueProcessor), nil
}

// GetTotalStakedValue -
func (d *disabledTotalStakedValueProcessor) GetTotalStakedValue() (*api.StakeValues, error) {
	return nil, ErrCannotReturnTotalStakedFromShardNode
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledTotalStakedValueProcessor) IsInterfaceNil() bool {
	return d == nil
}
