package mock

import (
	"crypto/sha256"
)

var hasherMockEmptyHash []byte

// HasherMock is used in testing
type HasherMock struct {
}

// Computes the SHA256 of the provided string
func (hm *HasherMock) Compute(s string) []byte {
	h := sha256.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}

// EmptyHash returns the hash of the empty string
func (hm *HasherMock) EmptyHash() []byte {
	if len(hasherMockEmptyHash) == 0 {
		hasherMockEmptyHash = hm.Compute("")
	}
	return hasherMockEmptyHash
}

// Size return SHA256 output size: 32
func (*HasherMock) Size() int {
	return sha256.Size
}
