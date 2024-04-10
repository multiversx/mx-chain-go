package keyBuilder

import (
	"github.com/multiversx/mx-chain-go/common"
)

type disabledKeyBuilder struct {
}

// NewDisabledKeyBuilder creates a new disabled key builder. This should be used when the key is not needed
func NewDisabledKeyBuilder() *disabledKeyBuilder {
	return &disabledKeyBuilder{}
}

// BuildKey does nothing for this implementation
func (dkb *disabledKeyBuilder) BuildKey(_ []byte) {

}

// GetKey returns an empty byte array for this implementation
func (dkb *disabledKeyBuilder) GetKey() ([]byte, error) {
	return []byte{}, nil
}

// ShallowClone returns a new disabled key builder
func (dkb *disabledKeyBuilder) ShallowClone() common.KeyBuilder {
	return &disabledKeyBuilder{}
}

// DeepClone returns a new disabled key builder
func (dkb *disabledKeyBuilder) DeepClone() common.KeyBuilder {
	return &disabledKeyBuilder{}
}

// Size returns 0 for this implementation
func (dkb *disabledKeyBuilder) Size() uint {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (dkb *disabledKeyBuilder) IsInterfaceNil() bool {
	return dkb == nil
}
