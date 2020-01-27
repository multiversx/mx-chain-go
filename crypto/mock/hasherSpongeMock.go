package mock

import (
	"golang.org/x/crypto/blake2b"
)

var hasherSpongeEmptyHash []byte

const hashSize = 16

// HasherSpongeMock that will be used for testing
type HasherSpongeMock struct {
}

// Compute will output the SHA's equivalent of the input string
func (sha HasherSpongeMock) Compute(s string) []byte {
	h, _ := blake2b.New(hashSize, nil)
	_, _ = h.Write([]byte(s))
	return h.Sum(nil)
}

// EmptyHash will return the equivalent of empty string SHA's
func (sha HasherSpongeMock) EmptyHash() []byte {
	if len(hasherSpongeEmptyHash) == 0 {
		hasherSpongeEmptyHash = sha.Compute("")
	}
	return hasherSpongeEmptyHash
}

// Size returns the required size in bytes
func (HasherSpongeMock) Size() int {
	return hashSize
}

// IsInterfaceNil returns true if there is no value under the interface
func (sha HasherSpongeMock) IsInterfaceNil() bool {
	return false
}
