package state

import (
	"encoding/hex"
	"strings"
)

// PlainAddressConverter is used to convert the address from/to different structures
type PlainAddressConverter struct {
	addressLen int
	prefix     string
}

// NewPlainAddressConverter creates a new instance of HashAddressConverter
func NewPlainAddressConverter(addressLen int, prefix string) (*PlainAddressConverter, error) {
	if addressLen < 0 {
		return nil, ErrNegativeValue
	}

	return &PlainAddressConverter{
		addressLen: addressLen,
		prefix:     prefix,
	}, nil
}

// CreateAddressFromPublicKeyBytes returns the bytes received as parameters, trimming if necessary
// and outputs a new AddressContainer obj
func (pac *PlainAddressConverter) CreateAddressFromPublicKeyBytes(pubKey []byte) (AddressContainer, error) {
	if pubKey == nil {
		return nil, ErrNilPubKeysBytes
	}

	if len(pubKey) < pac.addressLen {
		return nil, NewErrorWrongSize(pac.addressLen, len(pubKey))
	}

	newPubKey := make([]byte, len(pubKey))
	copy(newPubKey, pubKey)

	//check size, trimming as necessary
	if len(newPubKey) > pac.addressLen {
		newPubKey = newPubKey[len(newPubKey)-pac.addressLen:]
	}

	return NewAddress(newPubKey), nil
}

// ConvertToHex returns the hex string representation of the address.
func (pac *PlainAddressConverter) ConvertToHex(addressContainer AddressContainer) (string, error) {
	if addressContainer == nil {
		return "", ErrNilAddressContainer
	}

	return pac.prefix + hex.EncodeToString(addressContainer.Bytes()), nil
}

// CreateAddressFromHex creates the address from hex string
func (pac *PlainAddressConverter) CreateAddressFromHex(hexAddress string) (AddressContainer, error) {
	if len(hexAddress) == 0 {
		return nil, ErrEmptyAddress
	}

	//to lower
	hexAddress = strings.ToLower(hexAddress)

	//check if it has prefix, trimming as necessary
	if strings.HasPrefix(hexAddress, strings.ToLower(pac.prefix)) {
		hexAddress = hexAddress[len(pac.prefix):]
	}

	//check lengths
	if len(hexAddress) != pac.addressLen*2 {
		return nil, NewErrorWrongSize(pac.addressLen*2, len(hexAddress))
	}

	//decode hex
	buff := make([]byte, pac.addressLen)
	_, err := hex.Decode(buff, []byte(hexAddress))

	if err != nil {
		return nil, err
	}

	return NewAddress(buff), nil
}

// PrepareAddressBytes checks and returns the slice compatible to the address format
func (pac *PlainAddressConverter) PrepareAddressBytes(addressBytes []byte) ([]byte, error) {
	if addressBytes == nil {
		return nil, ErrNilAddressContainer
	}

	if len(addressBytes) == 0 {
		return nil, ErrEmptyAddress
	}

	if len(addressBytes) != pac.addressLen {
		return nil, NewErrorWrongSize(pac.addressLen, len(addressBytes))
	}

	return addressBytes, nil
}
