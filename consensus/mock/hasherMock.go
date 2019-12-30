package mock

import "crypto/sha256"

var sha256EmptyHash []byte

// HasherMock that will be used for testing
type HasherMock struct {
}

// Compute will output the SHA's equivalent of the input string
func (sha HasherMock) Compute(s string) []byte {
	h := sha256.New()
	_, _ = h.Write([]byte(s))
	return h.Sum(nil)
}

// EmptyHash will return the equivalent of empty string SHA's
func (sha HasherMock) EmptyHash() []byte {
	if len(sha256EmptyHash) == 0 {
		sha256EmptyHash = sha.Compute("")
	}
	return sha256EmptyHash
}

// Size return the required size in bytes
func (HasherMock) Size() int {
	return sha256.Size
}

// IsInterfaceNil returns true if there is no value under the interface
func (sha HasherMock) IsInterfaceNil() bool {
	return false
}
