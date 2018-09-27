package hashing

import (
	"crypto/sha256"
)

var sha256EmptyHash []byte

type Sha256 struct {
}

func (sha Sha256) Compute(s string) []byte {
	h := sha256.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}

func (sha Sha256) EmptyHash() []byte {
	if len(sha256EmptyHash) == 0 {
		sha256EmptyHash = sha.Compute("")
	}
	return sha256EmptyHash
}

func (Sha256) Size() int {
	return sha256.Size
}
