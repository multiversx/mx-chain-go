package mock

import "github.com/ElrondNetwork/elrond-go/data/state"

// PubkeyConverterStub -
type PubkeyConverterStub struct {
	LenCalled                     func() int
	DecodeCalled                  func(humanReadable string) ([]byte, error)
	EncodeCalled                  func(pkBytes []byte) string
	CreateAddressFromStringCalled func(humanReadable string) (state.AddressContainer, error)
	CreateAddressFromBytesCalled  func(pkBytes []byte) (state.AddressContainer, error)
}

// Len -
func (pcs *PubkeyConverterStub) Len() int {
	if pcs.LenCalled != nil {
		return pcs.LenCalled()
	}

	return 0
}

// Decode -
func (pcs *PubkeyConverterStub) Decode(humanReadable string) ([]byte, error) {
	if pcs.DecodeCalled != nil {
		return pcs.DecodeCalled(humanReadable)
	}

	return make([]byte, 0), nil
}

// Encode -
func (pcs *PubkeyConverterStub) Encode(pkBytes []byte) string {
	if pcs.EncodeCalled != nil {
		return pcs.EncodeCalled(pkBytes)
	}

	return ""
}

// CreateAddressFromString -
func (pcs *PubkeyConverterStub) CreateAddressFromString(humanReadable string) (state.AddressContainer, error) {
	if pcs.CreateAddressFromBytesCalled != nil {
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
