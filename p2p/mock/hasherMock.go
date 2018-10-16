package mock

import (
	"crypto/sha256"
)

var hasherMockEmptyHash []byte

type HasherMock struct {
}

func (hm *HasherMock) Compute(s string) []byte {
	h := sha256.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}

func (hm *HasherMock) EmptyHash() []byte {
	if len(hasherMockEmptyHash) == 0 {
		hasherMockEmptyHash = hm.Compute("")
	}
	return hasherMockEmptyHash
}

func (*HasherMock) Size() int {
	return sha256.Size
}
