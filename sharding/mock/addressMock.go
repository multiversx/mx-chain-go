package mock

type AddressMock struct {
	Bts []byte
}

func (address *AddressMock) Bytes() []byte {
	return address.Bts
}
