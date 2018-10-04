package mock

import "crypto/sha256"

var sha256EmptyHash []byte

type MockHasher struct {
}

func (sha MockHasher) Compute(s string) []byte {
	h := sha256.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}

func (sha MockHasher) EmptyHash() []byte {
	if len(sha256EmptyHash) == 0 {
		sha256EmptyHash = sha.Compute("")
	}
	return sha256EmptyHash
}

func (MockHasher) Size() int {
	return sha256.Size
}
