package mock

type AddressContainerStub struct {
	BytesHandler func() []byte
}

func (ac AddressContainerStub) Bytes() []byte {
	return ac.BytesHandler()
}
