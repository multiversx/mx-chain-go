package fnv

import (
	"hash/fnv"

	"github.com/ElrondNetwork/elrond-go/hashing"
)

var _ hashing.Hasher = (*Fnv)(nil)

var fnvEmptyHash []byte

// Fnv is a fnv128a implementation of the hasher interface.
type Fnv struct {
}

// Compute takes a string, and returns the fnv128a hash of that string
func (f Fnv) Compute(s string) []byte {
	if len(s) == 0 && len(fnvEmptyHash) != 0 {
		return f.EmptyHash()
	}
	h := fnv.New128a()
	_, _ = h.Write([]byte(s))
	return h.Sum(nil)
}

// EmptyHash returns the fnv128a hash of the empty string
func (f Fnv) EmptyHash() []byte {
	if len(fnvEmptyHash) == 0 {
		fnvEmptyHash = f.Compute("")
	}
	return fnvEmptyHash
}

// Size returns the size, in number of bytes, of a fnv128a hash
func (Fnv) Size() int {
	return fnv.New128a().Size()
}

// IsInterfaceNil returns true if there is no value under the interface
func (f Fnv) IsInterfaceNil() bool {
	return false
}
