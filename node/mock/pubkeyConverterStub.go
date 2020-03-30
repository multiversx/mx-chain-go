package mock

import "github.com/ElrondNetwork/elrond-go/data/state"

// PubkeyConverterStub -
type PubkeyConverterStub struct {
	AddressLenCalled              func() int
	BytesCalled                   func(humanReadable string) ([]byte, error)
	StringCalled                  func(pkBytes []byte) (string, error)
	CreateAddressFromStringCalled func(humanReadable string) (state.AddressContainer, error)
	CreateAddressFromBytesCalled  func(pkBytes []byte) (state.AddressContainer, error)
}

// AddressLen -
func (pcs *PubkeyConverterStub) AddressLen() int {
	if pcs.AddressLenCalled != nil {
		return pcs.AddressLenCalled()
	}

	return 0
}

// Bytes -
func (pcs *PubkeyConverterStub) Bytes(humanReadable string) ([]byte, error) {
	if pcs.BytesCalled != nil {
		return pcs.BytesCalled(humanReadable)
	}

	return make([]byte, 0), nil
}

// String -
func (pcs *PubkeyConverterStub) String(pkBytes []byte) (string, error) {
	if pcs.StringCalled != nil {
		return pcs.StringCalled(pkBytes)
	}

	return "", nil
}

// CreateAddressFromString -
func (pcs *PubkeyConverterStub) CreateAddressFromString(humanReadable string) (state.AddressContainer, error) {
	if pcs.CreateAddressFromStringCalled != nil {
		return pcs.CreateAddressFromStringCalled(humanReadable)
	}

	return nil, nil
}

// CreateAddressFromBytes -
func (pcs *PubkeyConverterStub) CreateAddressFromBytes(pkBytes []byte) (state.AddressContainer, error) {
	if pcs.CreateAddressFromBytesCalled != nil {
		return pcs.CreateAddressFromBytesCalled(pkBytes)
	}

	return nil, nil
}

// IsInterfaceNil -
func (pcs *PubkeyConverterStub) IsInterfaceNil() bool {
	return pcs == nil
}
