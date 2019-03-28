package mock

import (
	"encoding/hex"
	"strings"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/pkg/errors"
)

type AddressConverterFake struct {
	addressLen int
	prefix     string
}

func NewAddressConverterFake(addressLen int, prefix string) *AddressConverterFake {
	return &AddressConverterFake{
		addressLen: addressLen,
		prefix:     prefix,
	}
}

func (acf *AddressConverterFake) CreateAddressFromPublicKeyBytes(pubKey []byte) (state.AddressContainer, error) {
	newPubKey := make([]byte, len(pubKey))
	copy(newPubKey, pubKey)

	//check size, trimming as necessary
	if len(newPubKey) > acf.addressLen {
		newPubKey = newPubKey[len(newPubKey)-acf.addressLen:]
	}

	return state.NewAddress(newPubKey), nil
}

func (acf *AddressConverterFake) ConvertToHex(addressContainer state.AddressContainer) (string, error) {
	return acf.prefix + hex.EncodeToString(addressContainer.Bytes()), nil
}

func (acf *AddressConverterFake) CreateAddressFromHex(hexAddress string) (state.AddressContainer, error) {

	//to lower
	hexAddress = strings.ToLower(hexAddress)

	//check if it has prefix, trimming as necessary
	if strings.HasPrefix(hexAddress, strings.ToLower(acf.prefix)) {
		hexAddress = hexAddress[len(acf.prefix):]
	}

	//check lengths
	if len(hexAddress) != acf.addressLen*2 {
		return nil, errors.New("wrong size")
	}

	//decode hex
	buff := make([]byte, acf.addressLen)
	_, err := hex.Decode(buff, []byte(hexAddress))

	if err != nil {
		return nil, err
	}

	return state.NewAddress(buff), nil
}

func (acf *AddressConverterFake) PrepareAddressBytes(addressBytes []byte) ([]byte, error) {
	return addressBytes, nil
}
