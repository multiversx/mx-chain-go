package keyBuilder

import (
	"github.com/ElrondNetwork/elrond-go/common"
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

// Clone returns a new disabled key builder
func (dkb *disabledKeyBuilder) Clone() common.KeyBuilder {
	return &disabledKeyBuilder{}
}
