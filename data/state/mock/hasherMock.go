package mock

import "crypto/sha256"

var sha256EmptyHash []byte

type HasherMock struct {
}

func (sha HasherMock) Compute(s string) []byte {
	h := sha256.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}

func (sha HasherMock) EmptyHash() []byte {
	if len(sha256EmptyHash) == 0 {
		sha256EmptyHash = sha.Compute("")
	}
	return sha256EmptyHash
}

func (HasherMock) Size() int {
	return sha256.Size
}
