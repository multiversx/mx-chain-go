package sha256

import (
	"crypto/sha256"

	"github.com/ElrondNetwork/elrond-go/hashing"
)

var _ hashing.Hasher = (*Sha256)(nil)

var sha256EmptyHash []byte

// Sha256 is a sha256 implementation of the hasher interface.
type Sha256 struct {
}

// Compute takes a string, and returns the sha256 hash of that string
func (sha Sha256) Compute(s string) []byte {
	if len(s) == 0 && len(sha256EmptyHash) != 0 {
		return sha.EmptyHash()
	}
	h := sha256.New()
	_, _ = h.Write([]byte(s))
	return h.Sum(nil)
}

// EmptyHash returns the sha256 hash of the empty string
func (sha Sha256) EmptyHash() []byte {
	if len(sha256EmptyHash) == 0 {
		sha256EmptyHash = sha.Compute("")
	}
	return sha256EmptyHash
}

// Size returns the size, in number of bytes, of a sha256 hash
func (Sha256) Size() int {
	return sha256.Size
}

// IsInterfaceNil returns true if there is no value under the interface
func (sha Sha256) IsInterfaceNil() bool {
	return false
}
