package blake2b

import (
	"golang.org/x/crypto/blake2b"
)

var b2bEmptyHash []byte

// Blake2b defines a Blake2b object
type Blake2b struct {
}

// Compute method returns a hash of the given string
func (b2b Blake2b) Compute(s string) []byte {
	h, _ := blake2b.New256(nil)
	h.Write([]byte(s))
	return h.Sum(nil)
}

// EmptyHash method returns a hash of the empty string
func (b2b Blake2b) EmptyHash() []byte {
	if len(b2bEmptyHash) == 0 {
		b2bEmptyHash = b2b.Compute("")
	}
	return b2bEmptyHash
}

// Size method returns the blake2b hash size
func (Blake2b) Size() int {
	return blake2b.Size256
}
