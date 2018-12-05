package sha256

import (
	"crypto/sha256"
)

var sha256EmptyHash []byte

// Sha256 defines a Sha256 object
type Sha256 struct {
}

// Compute method returns a hash of the given string
func (sha Sha256) Compute(s string) []byte {
	h := sha256.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}

// EmptyHash method returns a hash of the empty string
func (sha Sha256) EmptyHash() []byte {
	if len(sha256EmptyHash) == 0 {
		sha256EmptyHash = sha.Compute("")
	}
	return sha256EmptyHash
}

// Size method returns the sha256 hash size
func (Sha256) Size() int {
	return sha256.Size
}
