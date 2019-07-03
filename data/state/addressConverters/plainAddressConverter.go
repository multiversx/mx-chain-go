package addressConverters

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/btcsuite/btcutil/bech32"
)

// PlainAddressConverter is used to convert the address from/to different structures
type PlainAddressConverter struct {
	addressLen int
	prefix     string
}

// NewPlainAddressConverter creates a new instance of HashAddressConverter
func NewPlainAddressConverter(addressLen int, prefix string) (*PlainAddressConverter, error) {
	if addressLen < 0 {
		return nil, state.ErrNegativeValue
	}

	return &PlainAddressConverter{
		addressLen: addressLen,
		prefix:     prefix,
	}, nil
}

// CreateAddressFromPublicKeyBytes returns the bytes received as parameters, trimming if necessary
// and outputs a new AddressContainer obj
func (pac *PlainAddressConverter) CreateAddressFromPublicKeyBytes(pubKey []byte) (state.AddressContainer, error) {
	if pubKey == nil {
		return nil, state.ErrNilPubKeysBytes
	}

	if len(pubKey) < pac.addressLen {
		return nil, state.NewErrorWrongSize(pac.addressLen, len(pubKey))
	}

	newPubKey := make([]byte, len(pubKey))
	copy(newPubKey, pubKey)

	//check size, trimming as necessary
	if len(newPubKey) > pac.addressLen {
		newPubKey = newPubKey[len(newPubKey)-pac.addressLen:]
	}

	return state.NewAddress(newPubKey), nil
}

// ConvertToHex returns the hex string representation of the address.
func (pac *PlainAddressConverter) ConvertToHex(addressContainer state.AddressContainer) (string, error) {
	if addressContainer == nil {
		return "", state.ErrNilAddressContainer
	}

	return pac.prefix + hex.EncodeToString(addressContainer.Bytes()), nil
}

// CreateAddressFromHex creates the address from hex string
func (pac *PlainAddressConverter) CreateAddressFromHex(hexAddress string) (state.AddressContainer, error) {
	if len(hexAddress) == 0 {
		return nil, state.ErrEmptyAddress
	}

	//to lower
	hexAddress = strings.ToLower(hexAddress)

	//check if it has prefix, trimming as necessary
	if strings.HasPrefix(hexAddress, strings.ToLower(pac.prefix)) {
		hexAddress = hexAddress[len(pac.prefix):]
	}

	//check lengths
	if len(hexAddress) != pac.addressLen*2 {
		return nil, state.NewErrorWrongSize(pac.addressLen*2, len(hexAddress))
	}

	//decode hex
	buff := make([]byte, pac.addressLen)
	_, err := hex.Decode(buff, []byte(hexAddress))
	if err != nil {
		return nil, err
	}

	return state.NewAddress(buff), nil
}

// PrepareAddressBytes checks and returns the slice compatible to the address format
func (pac *PlainAddressConverter) PrepareAddressBytes(addressBytes []byte) ([]byte, error) {
	if addressBytes == nil {
		return nil, state.ErrNilAddressContainer
	}

	if len(addressBytes) == 0 {
		return nil, state.ErrEmptyAddress
	}

	if len(addressBytes) != pac.addressLen {
		return nil, state.NewErrorWrongSize(pac.addressLen, len(addressBytes))
	}

	return addressBytes, nil
}

// AddressLen returns the address length
func (pac *PlainAddressConverter) AddressLen() int {
	return pac.addressLen
}

//the bech32 prefix will be erd, hrp separator is always 1
//addresses will look like this erd1ajkdna0mnj56qwl4avp0qpmgc664x4qrdjq3rs6ddpsyfj9gna3s4sjy2n

const erd = "erd"

// ConvertToBech32 returns the address in bech32 format
func (pac *PlainAddressConverter) ConvertToBech32(addressContainer state.AddressContainer) (string, error) {
	if addressContainer == nil {
		return "", state.ErrNilAddressContainer
	}

	conv, err := bech32.ConvertBits(addressContainer.Bytes(), 8, 5, true)
	if err == nil {
		enc, err := bech32.Encode(erd, conv)
		if err == nil {
			return enc, nil
		}
		return "", err
	}
	return "", err
}

// CreateAddressFromBech32 creates the address from bech32 string
func (pac *PlainAddressConverter) CreateAddressFromBech32(bech32Address string) (state.AddressContainer, error) {
	if len(bech32Address) == 0 {
		return nil, state.ErrEmptyAddress
	}

	_, dec, err := bech32.Decode(bech32Address)
	if err != nil {
		//TODO move error to state/errors.go
		return nil, errors.New("wrong bech32 string: " + bech32Address)
	}
	conv, err := bech32.ConvertBits(dec, 5, 8, false)
	if err != nil {
		//TODO move error to state/errors.go
		return nil, errors.New("can't convert bech32 string")
	}
	//return decoded
	return state.NewAddress(conv), nil
}
