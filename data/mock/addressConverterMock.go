package mock

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/ElrondNetwork/elrond-go/data/state"
)

var errFailure = errors.New("failure")

// AddressConverterMock -
type AddressConverterMock struct {
	Fail                                          bool
	CreateAddressFromPublicKeyBytesRetErrForValue []byte
}

// CreateAddressFromPublicKeyBytes -
func (acm *AddressConverterMock) CreateAddressFromPublicKeyBytes(pubKey []byte) (state.AddressContainer, error) {
	if acm.Fail {
		return nil, errFailure
	}

	if acm.CreateAddressFromPublicKeyBytesRetErrForValue != nil {
		if bytes.Equal(acm.CreateAddressFromPublicKeyBytesRetErrForValue, pubKey) {
			return nil, errors.New("error required")
		}
	}

	return NewAddressMockFromBytes(pubKey), nil
}

// ConvertToHex -
func (acm *AddressConverterMock) ConvertToHex(addressContainer state.AddressContainer) (string, error) {
	if acm.Fail {
		return "", errFailure
	}

	return hex.EncodeToString(addressContainer.Bytes()), nil
}

// CreateAddressFromHex -
func (acm *AddressConverterMock) CreateAddressFromHex(_ string) (state.AddressContainer, error) {
	if acm.Fail {
		return nil, errFailure
	}

	panic("implement me")
}

// PrepareAddressBytes -
func (acm *AddressConverterMock) PrepareAddressBytes(_ []byte) ([]byte, error) {
	if acm.Fail {
		return nil, errFailure
	}

	panic("implement me")
}

// AddressLen -
func (acm *AddressConverterMock) AddressLen() int {
	return 32
}

// IsInterfaceNil returns true if there is no value under the interface
func (acm *AddressConverterMock) IsInterfaceNil() bool {
	return acm == nil
}
