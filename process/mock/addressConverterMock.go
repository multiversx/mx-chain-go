package mock

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
)

type AddressConverterMock struct {
	Fail                                          bool
	CreateAddressFromPublicKeyBytesRetErrForValue []byte
}

func (acm *AddressConverterMock) CreateAddressFromPublicKeyBytes(pubKey []byte) (state.AddressContainer, error) {
	if acm.Fail {
		return nil, errors.New("failure")
	}

	if acm.CreateAddressFromPublicKeyBytesRetErrForValue != nil {
		if bytes.Equal(acm.CreateAddressFromPublicKeyBytesRetErrForValue, pubKey) {
			return nil, errors.New("error required")
		}
	}

	return NewAddressMock(pubKey), nil
}

func (acm *AddressConverterMock) ConvertToHex(addressContainer state.AddressContainer) (string, error) {
	if acm.Fail {
		return "", errors.New("failure")
	}

	return hex.EncodeToString(addressContainer.Bytes()), nil
}

func (acm *AddressConverterMock) CreateAddressFromHex(hexAddress string) (state.AddressContainer, error) {
	if acm.Fail {
		return nil, errors.New("failure")
	}

	panic("implement me")
}

func (acm *AddressConverterMock) PrepareAddressBytes(addressBytes []byte) ([]byte, error) {
	if acm.Fail {
		return nil, errors.New("failure")
	}

	panic("implement me")
}
