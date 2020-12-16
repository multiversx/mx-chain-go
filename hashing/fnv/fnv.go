package fnv

import (
	fnvLib "hash/fnv"

	"github.com/ElrondNetwork/elrond-go/hashing"
)

var _ hashing.Hasher = (*fnv)(nil)

// fnv is a fnv128a implementation of the hasher interface.
type fnv struct {
	emptyHash []byte
}

// NewFnv initializes the empty hash and returns a new instance of the fnv hasher
func NewFnv() *fnv {
	return &fnv{
		emptyHash: computeEmptyHash(),
	}
}

// Compute takes a string, and returns the fnv128a hash of that string
func (f *fnv) Compute(s string) []byte {
	if len(s) == 0 {
		return f.getEmptyHash()
	}
	h := fnvLib.New128a()
	_, _ = h.Write([]byte(s))
	return h.Sum(nil)
}

// Size returns the size, in number of bytes, of a fnv128a hash
func (f *fnv) Size() int {
	return fnvLib.New128a().Size()
}

func (f *fnv) getEmptyHash() []byte {
	hashCopy := make([]byte, len(f.emptyHash))
	copy(hashCopy, f.emptyHash)

	return hashCopy
}

func computeEmptyHash() []byte {
	h := fnvLib.New128a()
	_, _ = h.Write([]byte(""))
	return h.Sum(nil)
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *fnv) IsInterfaceNil() bool {
	return f == nil
}
