package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type AddressConverterStub struct {
	CreateAddressFromPublicKeyBytesHandler func(pubKey []byte) (state.AddressContainer, error)
	ConvertToHexHandler                    func(addressContainer state.AddressContainer) (string, error)
	CreateAddressFromHexHandler            func(hexAddress string) (state.AddressContainer, error)
	PrepareAddressBytesHandler             func(addressBytes []byte) ([]byte, error)
	AddressLenHandler                      func() int
}

func (ac AddressConverterStub) CreateAddressFromPublicKeyBytes(pubKey []byte) (state.AddressContainer, error) {
	return ac.CreateAddressFromPublicKeyBytesHandler(pubKey)
}
func (ac AddressConverterStub) ConvertToHex(addressContainer state.AddressContainer) (string, error) {
	return ac.ConvertToHexHandler(addressContainer)
}
func (ac AddressConverterStub) CreateAddressFromHex(hexAddress string) (state.AddressContainer, error) {
	return ac.CreateAddressFromHexHandler(hexAddress)
}
func (ac AddressConverterStub) PrepareAddressBytes(addressBytes []byte) ([]byte, error) {
	return ac.PrepareAddressBytesHandler(addressBytes)
}

func (ac AddressConverterStub) AddressLen() int {
	return ac.AddressLenHandler()
}
