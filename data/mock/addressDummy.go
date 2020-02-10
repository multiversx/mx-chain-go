package mock

// AddressDummy -
type AddressDummy struct {
	bytes []byte
	hash  []byte
}

// NewAddressDummy -
func NewAddressDummy(bytes, hash []byte) *AddressDummy {
	return &AddressDummy{
		bytes: bytes,
		hash:  hash,
	}
}

// Bytes -
func (ad *AddressDummy) Bytes() []byte {
	return ad.bytes
}

// Hash -
func (ad *AddressDummy) Hash() []byte {
	return ad.hash
}

// IsInterfaceNil returns true if there is no value under the interface
func (ad *AddressDummy) IsInterfaceNil() bool {
	if ad == nil {
		return true
	}
	return false
}
