package mock

// AddressMock is the struct holding a mock address
type AddressMock struct {
	bytes []byte
	hash  []byte
}

// NewAddressMock creates a new Address with the same byte slice as the parameter received
func NewAddressMock(adr []byte, hash []byte) *AddressMock {
	return &AddressMock{bytes: adr, hash: hash}
}

// Bytes returns the data corresponding to this address
func (adr *AddressMock) Bytes() []byte {
	return adr.bytes
}

// Hash return the address corresponding hash
func (adr *AddressMock) Hash() []byte {
	return adr.hash
}
