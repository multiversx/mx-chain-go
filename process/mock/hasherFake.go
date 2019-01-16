package mock

import "crypto/sha256"

// HasherFake that will be used for testing
type HasherFake struct {
}

// Compute will output the SHA's equivalent of the input string
func (sha HasherFake) Compute(s string) []byte {
	h := sha256.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}

// EmptyHash will return the equivalent of empty string SHA's
func (sha HasherFake) EmptyHash() []byte {
	if len(sha256EmptyHash) == 0 {
		sha256EmptyHash = sha.Compute("")
	}
	return sha256EmptyHash
}

// Size return the required size in bytes
func (HasherFake) Size() int {
	return sha256.Size
}
