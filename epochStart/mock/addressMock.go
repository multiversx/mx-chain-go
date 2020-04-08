package mock

// AddressMock is the struct holding a mock address
type AddressMock struct {
	bytes []byte
}

// NewAddressMock creates a new Address with the same byte slice as the parameter received
func NewAddressMock(adr []byte) *AddressMock {
	return &AddressMock{bytes: adr}
}

// Bytes returns the data corresponding to this address
func (adr *AddressMock) Bytes() []byte {
	return adr.bytes
}

// IsInterfaceNil returns true if there is no value under the interface
func (adr *AddressMock) IsInterfaceNil() bool {
	return adr == nil
}
