package mock

type AddressMock struct {
	Bts []byte
}

func (address *AddressMock) Bytes() []byte {
	return address.Bts
}

// IsInterfaceNil returns if there is no value under the interface
func (address *AddressMock) IsInterfaceNil() bool {
	if address == nil {
		return true
	}

	return false
}
