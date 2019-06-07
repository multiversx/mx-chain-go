package addressConverters

import (
	"encoding/hex"
	"strings"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

// HashAddressConverter is used to convert the address from/to different structures
type HashAddressConverter struct {
	hasher     hashing.Hasher
	addressLen int
	prefix     string
}

// NewHashAddressConverter creates a new instance of HashAddressConverter
func NewHashAddressConverter(hasher hashing.Hasher, addressLen int, prefix string) (*HashAddressConverter, error) {
	if hasher == nil {
		return nil, state.ErrNilHasher
	}

	if addressLen < 0 {
		return nil, state.ErrNegativeValue
	}

	return &HashAddressConverter{
		hasher:     hasher,
		addressLen: addressLen,
		prefix:     prefix,
	}, nil
}

// CreateAddressFromPublicKeyBytes hashes the bytes received as parameters, trimming if necessary
// and outputs a new AddressContainer obj
func (hac *HashAddressConverter) CreateAddressFromPublicKeyBytes(pubKey []byte) (state.AddressContainer, error) {
	if pubKey == nil {
		return nil, state.ErrNilPubKeysBytes
	}

	if len(pubKey) < hac.addressLen {
		return nil, state.NewErrorWrongSize(hac.addressLen, len(pubKey))
	}

	//compute hash
	hash := hac.hasher.Compute(string(pubKey))

	//check size, trimming as necessary
	if len(hash) > hac.addressLen {
		hash = hash[len(hash)-hac.addressLen:]
	}

	return state.NewAddress(hash), nil
}

// ConvertToHex returns the hex string representation of the address.
func (hac *HashAddressConverter) ConvertToHex(addressContainer state.AddressContainer) (string, error) {
	if addressContainer == nil {
		return "", state.ErrNilAddressContainer
	}

	return hac.prefix + hex.EncodeToString(addressContainer.Bytes()), nil
}

// CreateAddressFromHex creates the address from hex string
func (hac *HashAddressConverter) CreateAddressFromHex(hexAddress string) (state.AddressContainer, error) {
	if len(hexAddress) == 0 {
		return nil, state.ErrEmptyAddress
	}

	//to lower
	hexAddress = strings.ToLower(hexAddress)

	//check if it has prefix, trimming as necessary
	if strings.HasPrefix(hexAddress, strings.ToLower(hac.prefix)) {
		hexAddress = hexAddress[len(hac.prefix):]
	}

	//check lengths
	if len(hexAddress) != hac.addressLen*2 {
		return nil, state.NewErrorWrongSize(hac.addressLen*2, len(hexAddress))
	}

	//decode hex
	buff := make([]byte, hac.addressLen)
	_, err := hex.Decode(buff, []byte(hexAddress))

	if err != nil {
		return nil, err
	}

	return state.NewAddress(buff), nil
}

// PrepareAddressBytes checks and returns the slice compatible to the address format
func (hac *HashAddressConverter) PrepareAddressBytes(addressBytes []byte) ([]byte, error) {
	if addressBytes == nil {
		return nil, state.ErrNilAddressContainer
	}

	if len(addressBytes) == 0 {
		return nil, state.ErrEmptyAddress
	}

	if len(addressBytes) != hac.addressLen {
		return nil, state.NewErrorWrongSize(hac.addressLen, len(addressBytes))
	}

	return addressBytes, nil
}

// AddressLen returns the address length
func (hac *HashAddressConverter) AddressLen() int {
	return hac.addressLen
}
