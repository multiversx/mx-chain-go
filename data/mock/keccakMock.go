package mock

import (
	"golang.org/x/crypto/sha3"
)

var keccakEmptyHash []byte

// KeccakMock is a sha3-Keccak implementation of the hasher interface.
type KeccakMock struct {
}

// Compute takes a string, and returns the sha3-Keccak hash of that string
func (k KeccakMock) Compute(s string) []byte {
	if len(s) == 0 && len(keccakEmptyHash) != 0 {
		return k.EmptyHash()
	}
	h := sha3.NewLegacyKeccak256()
	_, _ = h.Write([]byte(s))
	return h.Sum(nil)
}

// EmptyHash returns the sha3-Keccak hash of the empty string
func (k KeccakMock) EmptyHash() []byte {
	if len(keccakEmptyHash) == 0 {
		keccakEmptyHash = k.Compute("")
	}
	return keccakEmptyHash
}

// Size returns the size, in number of bytes, of a sha3-Keccak hash
func (KeccakMock) Size() int {
	return sha3.NewLegacyKeccak256().Size()
}

// IsInterfaceNil returns true if there is no value under the interface
func (k *KeccakMock) IsInterfaceNil() bool {
	return k == nil
}
