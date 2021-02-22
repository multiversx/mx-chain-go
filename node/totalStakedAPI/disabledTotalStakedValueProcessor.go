package totalStakedAPI

import "math/big"

type disabledTotalStakedValueProcessor struct{}

// NewDisabledTotalStakedValueProcessor -
func NewDisabledTotalStakedValueProcessor() (*disabledTotalStakedValueProcessor, error) {
	return new(disabledTotalStakedValueProcessor), nil
}

// GetTotalStakedValue -
func (d *disabledTotalStakedValueProcessor) GetTotalStakedValue() (*big.Int, error) {
	return nil, ErrCannotReturnTotalStakedFromShardNode
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledTotalStakedValueProcessor) IsInterfaceNil() bool {
	return d == nil
}
