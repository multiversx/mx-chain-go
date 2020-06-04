package sharding

import (
	"encoding/hex"
	"fmt"
)

const intSize = 8

var _ Validator = (*validator)(nil)

type validator struct {
	pubKey  []byte
	chances uint32
	index   uint32
}

// NewValidator creates a new instance of a validator
func NewValidator(pubKey []byte, chances uint32, index uint32) (*validator, error) {
	if pubKey == nil {
		return nil, ErrNilPubKey
	}

	return &validator{
		pubKey:  pubKey,
		chances: chances,
		index:   index,
	}, nil
}

// PubKey returns the validator's public key
func (v *validator) PubKey() []byte {
	return v.pubKey
}

// Chances returns the validator's chances
func (v *validator) Chances() uint32 {
	return v.chances
}

// Index returns the validators index
func (v *validator) Index() uint32 {
	return v.index
}

// String returns the toString representation of the validator
func (v *validator) String() string {
	return fmt.Sprintf("%s %v %v", hex.EncodeToString(v.pubKey), v.index, v.chances)
}

// Size returns the size in bytes held by an instance of this struct
func (v *validator) Size() int {
	return len(v.pubKey) + intSize
}

// IsInterfaceNil returns true if there is no value under the interface
func (v *validator) IsInterfaceNil() bool {
	return v == nil
}
