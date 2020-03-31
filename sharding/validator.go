package sharding

import "sync"

type validator struct {
	pubKey     []byte
	address    []byte
	mutChances sync.RWMutex
	chances    uint32
}

// NewValidator creates a new instance of a validator
func NewValidator(pubKey []byte, address []byte, chances uint32) (*validator, error) {
	if pubKey == nil {
		return nil, ErrNilPubKey
	}

	if address == nil {
		return nil, ErrNilAddress
	}

	return &validator{
		pubKey:  pubKey,
		address: address,
		chances: chances,
	}, nil
}

// PubKey returns the validator's public key
func (v *validator) PubKey() []byte {
	return v.pubKey
}

// Address returns the validator's address
func (v *validator) Address() []byte {
	return v.address
}

// Chances returns the validator's chances
func (v *validator) Chances() uint32 {
	v.mutChances.RLock()
	defer v.mutChances.RUnlock()
	return v.chances

}

// SetChances sets the validator's chances
func (v *validator) SetChances(chances uint32) {
	v.mutChances.Lock()
	v.chances = chances
	v.mutChances.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (v *validator) IsInterfaceNil() bool {
	return v == nil
}
