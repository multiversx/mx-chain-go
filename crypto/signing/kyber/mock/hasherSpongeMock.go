package mock

import (
	"golang.org/x/crypto/blake2b"
)

var hasherSpongeEmptyHash []byte

// HasherMock that will be used for testing
type HasherSpongeMock struct {
}

// Compute will output the SHA's equivalent of the input string
func (sha HasherSpongeMock) Compute(s string) []byte {
	h, _ := blake2b.New(16, nil)
	h.Write([]byte(s))
	return h.Sum(nil)
}

// EmptyHash will return the equivalent of empty string SHA's
func (sha HasherSpongeMock) EmptyHash() []byte {
	if len(sha256EmptyHash) == 0 {
		hasherSpongeEmptyHash = sha.Compute("")
	}
	return hasherSpongeEmptyHash
}

// Size returns the required size in bytes
func (HasherSpongeMock) Size() int {
	return 16
}
