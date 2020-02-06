package mock

// AddressMock -
type AddressMock struct {
	Bts []byte
}

// Bytes -
func (address *AddressMock) Bytes() []byte {
	return address.Bts
}

// IsInterfaceNil returns true if there is no value under the interface
func (address *AddressMock) IsInterfaceNil() bool {
	if address == nil {
		return true
	}
	return false
}
