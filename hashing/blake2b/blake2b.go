package blake2b

import (
	"hash"

	"github.com/ElrondNetwork/elrond-go/hashing"
	blake2bLib "golang.org/x/crypto/blake2b"
)

var _ hashing.Hasher = (*blake2b)(nil)

// blake2b is a blake2b implementation of the hasher interface.
type blake2b struct {
	// customHashSize holds the custom value for the hash size. if not set, it will be 0 and the
	// default hasher will be used
	customHashSize int
	emptyHash      []byte
}

// NewBlake2bWithSize returns a new instance of the blake2b with a custom size
func NewBlake2bWithSize(hashSize int) (*blake2b, error) {
	if hashSize < 0 {
		return nil, ErrInvalidHashSize
	}

	h := &blake2b{
		customHashSize: hashSize,
	}
	h.emptyHash = h.computeEmptyHash()

	return h, nil
}

// NewBlake2b returns a new instance of the blake2b hasher
func NewBlake2b() *blake2b {
	h := &blake2b{
		customHashSize: 0,
	}

	h.emptyHash = h.computeEmptyHash()

	return h
}

func (b2b *blake2b) getHasher() hash.Hash {
	if b2b.customHashSize == 0 {
		h, _ := blake2bLib.New256(nil)
		return h
	}
	h, _ := blake2bLib.New(b2b.customHashSize, nil)
	return h
}

// Compute takes a string, and returns the blake2b hash of that string
func (b2b *blake2b) Compute(s string) []byte {
	if len(s) == 0 {
		return b2b.getEmptyHash()
	}

	h := b2b.getHasher()
	_, _ = h.Write([]byte(s))
	return h.Sum(nil)
}

// Size returns the size, in number of bytes, of a blake2b hash
func (b2b *blake2b) Size() int {
	if b2b.customHashSize == 0 {
		return blake2bLib.Size256
	}

	return b2b.customHashSize
}

func (b2b *blake2b) getEmptyHash() []byte {
	hashCopy := make([]byte, len(b2b.emptyHash))
	copy(hashCopy, b2b.emptyHash)

	return hashCopy
}

func (b2b *blake2b) computeEmptyHash() []byte {
	h := b2b.getHasher()
	_, _ = h.Write([]byte(""))
	return h.Sum(nil)
}

// IsInterfaceNil returns true if there is no value under the interface
func (b2b *blake2b) IsInterfaceNil() bool {
	return b2b == nil
}
