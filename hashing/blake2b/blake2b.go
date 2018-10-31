package blake2b

import (
	"golang.org/x/crypto/blake2b"
)

var b2bEmptyHash []byte

type Blake2b struct {
}

func (b2b Blake2b) Compute(s string) []byte {
	h, _ := blake2b.New256(nil)
	h.Write([]byte(s))
	return h.Sum(nil)
}

func (b2b Blake2b) EmptyHash() []byte {
	if len(b2bEmptyHash) == 0 {
		b2bEmptyHash = b2b.Compute("")
	}
	return b2bEmptyHash
}

func (Blake2b) Size() int {
	return blake2b.Size256
}
