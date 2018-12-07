package blake2b

import (
	"golang.org/x/crypto/blake2b"
)

var b2bEmptyHash []byte

// Blake2b is a blake2b implementation of the hasher interface.
type Blake2b struct {
}

// Compute takes a string, and returns the blake2b hash of that string
func (b2b Blake2b) Compute(s string) []byte {
	if len(s) == 0 && len(b2bEmptyHash) != 0 {
		return b2b.EmptyHash()
	}
	h, _ := blake2b.New256(nil)
	h.Write([]byte(s))
	return h.Sum(nil)
}

// EmptyHash returns the blake2b hash of the empty string
func (b2b Blake2b) EmptyHash() []byte {
	if len(b2bEmptyHash) == 0 {
		b2bEmptyHash = b2b.Compute("")
	}
	return b2bEmptyHash
}

// Size returns the size, in number of bytes, of a blake2b hash
func (Blake2b) Size() int {
	return blake2b.Size256
}
