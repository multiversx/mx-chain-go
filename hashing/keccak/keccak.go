package keccak

import (
	"golang.org/x/crypto/sha3"
)

var keccakEmptyHash []byte

type Keccak struct {
}

func (Keccak) Compute(s string) []byte {
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte(s))
	return h.Sum(nil)
}

func (k Keccak) EmptyHash() []byte {
	if len(keccakEmptyHash) == 0 {
		keccakEmptyHash = k.Compute("")
	}
	return keccakEmptyHash
}

func (Keccak) Size() int {
	return sha3.NewLegacyKeccak256().Size()
}
