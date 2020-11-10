package keccak

import (
	"github.com/ElrondNetwork/elrond-go/hashing"
	"golang.org/x/crypto/sha3"
)

var _ hashing.Hasher = (*keccak)(nil)

// keccak is a sha3-keccak implementation of the hasher interface.
type keccak struct {
	emptyHash []byte
}

// NewKeccak initializes the empty hash and returns a new instance of the keccak hasher
func NewKeccak() *keccak {
	return &keccak{
		emptyHash: computeEmptyHash(),
	}
}

// Compute takes a string, and returns the sha3-keccak hash of that string
func (k *keccak) Compute(s string) []byte {
	if len(s) == 0 {
		return k.emptyHash
	}
	h := sha3.NewLegacyKeccak256()
	_, _ = h.Write([]byte(s))
	return h.Sum(nil)
}

// Size returns the size, in number of bytes, of a sha3-keccak hash
func (k *keccak) Size() int {
	return sha3.NewLegacyKeccak256().Size()
}

func computeEmptyHash() []byte {
	h := sha3.NewLegacyKeccak256()
	_, _ = h.Write([]byte(""))
	return h.Sum(nil)
}

// IsInterfaceNil returns true if there is no value under the interface
func (k *keccak) IsInterfaceNil() bool {
	return k == nil
}
