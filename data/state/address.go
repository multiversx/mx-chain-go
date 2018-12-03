package state

import (
	"encoding/hex"
	"strings"

	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

// AdrLen will hold the address length in bytes
const (
	AdrLen    = 32
	HexPrefix = "0x"
)

// Address is the struct holding an Elrond address identifier
type Address struct {
	bytes []byte
	hash  []byte
}

// NewAddress creates a new Address with the same byte slice as the parameter received
// Before creating the new pointer it does some checks against the slice received
func NewAddress(adr []byte) (*Address, error) {
	address := Address{}

	cleanedBytes, err := checkAdrBytes(adr)

	if err != nil {
		return nil, err
	}

	address.bytes = cleanedBytes
	return &address, nil
}

// Bytes returns the data corresponding to this address
func (adr *Address) Bytes() []byte {
	return adr.bytes
}

// Hash return the address corresponding hash
// Value is instantiated once (at the first Hash() call
func (adr *Address) Hash(hasher hashing.Hasher) []byte {
	if adr.hash == nil {
		adr.hash = hasher.Compute(string(adr.bytes))
	}

	return adr.hash
}

// Hex returns the hex string representation of the address.
func (adr *Address) Hex() string {
	return HexPrefix + hex.EncodeToString(adr.Bytes())
}

// FromPubKeyBytes hashes the bytes received as parameters, trimming if necessary
// and outputs a new Address obj
func FromPubKeyBytes(pubKey []byte, hasher hashing.Hasher) (*Address, error) {
	if pubKey == nil {
		return nil, ErrNilPubKeysBytes
	}

	if hasher == nil {
		return nil, ErrNilHasher
	}

	if len(pubKey) < AdrLen {
		return nil, NewErrorWrongSize(AdrLen, len(pubKey))
	}

	//compute hash
	hash := hasher.Compute(string(pubKey))

	//check size, trimming as necessary
	if len(hash) > AdrLen {
		hash = hash[len(hash)-AdrLen:]
	}

	return NewAddress(hash)
}

func checkAdrBytes(adr []byte) ([]byte, error) {
	if adr == nil {
		return nil, ErrNilAddress
	}

	if len(adr) == 0 {
		return nil, ErrEmptyAddress
	}

	if len(adr) != AdrLen {
		return nil, NewErrorWrongSize(AdrLen, len(adr))
	}

	return adr, nil
}

// FromHexAddress generates a new Address starting from its string representation
// The prefix is optional
func FromHexAddress(adr string) (*Address, error) {
	if len(adr) == 0 {
		return nil, ErrEmptyAddress
	}

	//to lower
	adr = strings.ToLower(adr)

	//check if it has prefix, trimming as necessary
	if strings.HasPrefix(adr, strings.ToLower(HexPrefix)) {
		adr = adr[len(HexPrefix):]
	}

	//check lengths
	if len(adr) != AdrLen*2 {
		return nil, NewErrorWrongSize(AdrLen*2, len(adr))
	}

	//decode hex
	buff := make([]byte, AdrLen)
	_, err := hex.Decode(buff, []byte(adr))

	if err != nil {
		return nil, err
	}

	return NewAddress(buff)
}
