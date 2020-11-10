package sha256

import (
	sha256Lib "crypto/sha256"

	"github.com/ElrondNetwork/elrond-go/hashing"
)

var _ hashing.Hasher = (*sha256)(nil)

// sha256 is a sha256 implementation of the hasher interface.
type sha256 struct {
	emptyHash []byte
}

// NewSha256 initializes the empty hash and returns a new instance of the sha256 hasher
func NewSha256() *sha256 {
	return &sha256{
		emptyHash: computeEmptyHash(),
	}
}

// Compute takes a string, and returns the sha256 hash of that string
func (s *sha256) Compute(str string) []byte {
	if len(str) == 0 {
		return s.emptyHash
	}
	h := sha256Lib.New()
	_, _ = h.Write([]byte(str))
	return h.Sum(nil)
}

// Size returns the size, in number of bytes, of a sha256 hash
func (s *sha256) Size() int {
	return sha256Lib.Size
}

func computeEmptyHash() []byte {
	h := sha256Lib.New()
	_, _ = h.Write([]byte(""))
	return h.Sum(nil)
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *sha256) IsInterfaceNil() bool {
	return s == nil
}
