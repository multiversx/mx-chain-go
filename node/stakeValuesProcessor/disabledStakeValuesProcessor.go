package stakeValuesProcessor

import "github.com/ElrondNetwork/elrond-go/data/api"

type disabledStaleValuesProcessor struct{}

// NewDisabledStakeValuesProcessor -
func NewDisabledStakeValuesProcessor() (*disabledStaleValuesProcessor, error) {
	return new(disabledStaleValuesProcessor), nil
}

// GetTotalStakedValue -
func (d *disabledStaleValuesProcessor) GetTotalStakedValue() (*api.StakeValues, error) {
	return nil, ErrCannotReturnTotalStakedFromShardNode
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledStaleValuesProcessor) IsInterfaceNil() bool {
	return d == nil
}
