package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/state"
)

// AddressConverterStub -
type AddressConverterStub struct {
	CreateAddressFromPublicKeyBytesHandler func(pubKey []byte) (state.AddressContainer, error)
	ConvertToHexHandler                    func(addressContainer state.AddressContainer) (string, error)
	CreateAddressFromHexHandler            func(hexAddress string) (state.AddressContainer, error)
	PrepareAddressBytesHandler             func(addressBytes []byte) ([]byte, error)
	AddressLenHandler                      func() int
}

// CreateAddressFromPublicKeyBytes -
func (ac AddressConverterStub) CreateAddressFromPublicKeyBytes(pubKey []byte) (state.AddressContainer, error) {
	return ac.CreateAddressFromPublicKeyBytesHandler(pubKey)
}

// ConvertToHex -
func (ac AddressConverterStub) ConvertToHex(addressContainer state.AddressContainer) (string, error) {
	return ac.ConvertToHexHandler(addressContainer)
}

// CreateAddressFromHex -
func (ac AddressConverterStub) CreateAddressFromHex(hexAddress string) (state.AddressContainer, error) {
	return ac.CreateAddressFromHexHandler(hexAddress)
}

// PrepareAddressBytes -
func (ac AddressConverterStub) PrepareAddressBytes(addressBytes []byte) ([]byte, error) {
	return ac.PrepareAddressBytesHandler(addressBytes)
}

// AddressLen -
func (ac AddressConverterStub) AddressLen() int {
	return ac.AddressLenHandler()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ac *AddressConverterStub) IsInterfaceNil() bool {
	if ac == nil {
		return true
	}
	return false
}
