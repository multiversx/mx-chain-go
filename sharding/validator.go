package sharding

type validator struct {
	pubKey  []byte
	address []byte
}

// NewValidator creates a new instance of a validator
func NewValidator(pubKey []byte, address []byte) (*validator, error) {
	if pubKey == nil {
		return nil, ErrNilPubKey
	}

	if address == nil {
		return nil, ErrNilAddress
	}

	return &validator{
		pubKey:  pubKey,
		address: address,
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

// IsInterfaceNil returns true if there is no value under the interface
func (v *validator) IsInterfaceNil() bool {
	return v == nil
}
