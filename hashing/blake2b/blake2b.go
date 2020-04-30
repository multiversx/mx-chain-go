package blake2b

import (
	"hash"

	"github.com/ElrondNetwork/elrond-go/hashing"
	"golang.org/x/crypto/blake2b"
)

var _ hashing.Hasher = (*Blake2b)(nil)

// Blake2b is a blake2b implementation of the hasher interface.
type Blake2b struct {
	HashSize  int
	emptyHash []byte
}

func (b2b *Blake2b) getHasher() hash.Hash {
	if b2b.HashSize == 0 {
		h, _ := blake2b.New256(nil)
		return h
	}
	h, _ := blake2b.New(b2b.HashSize, nil)
	return h
}

func (b2b *Blake2b) setEmpty() {
	b2b.emptyHash = b2b.getHasher().Sum(nil)
}

// Compute takes a string, and returns the blake2b hash of that string
func (b2b *Blake2b) Compute(s string) []byte {
	if len(s) == 0 {
		if len(b2b.emptyHash) == 0 {
			b2b.setEmpty()
		}
		return b2b.emptyHash
	}

	h := b2b.getHasher()
	_, _ = h.Write([]byte(s))
	return h.Sum(nil)
}

// EmptyHash returns the blake2b hash of the empty string
func (b2b *Blake2b) EmptyHash() []byte {
	if len(b2b.emptyHash) == 0 {
		b2b.setEmpty()
	}
	return b2b.emptyHash
}

// Size returns the size, in number of bytes, of a blake2b hash
func (b2b *Blake2b) Size() int {
	if b2b.HashSize == 0 {
		return blake2b.Size256
	}

	return b2b.HashSize
}

// IsInterfaceNil returns true if there is no value under the interface
func (b2b *Blake2b) IsInterfaceNil() bool {
	return b2b == nil
}
