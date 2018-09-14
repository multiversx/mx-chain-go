package hasher

import (
	"crypto/sha256"
	"encoding/hex"
)

type IHasher interface {
	CalculateHash(interface{}) interface{}
}

type HasherSha256 struct {
	Hash string
}

func (hasher *HasherSha256) CalculateHash(object interface{}) interface{} {
	h := sha256.New()
	h.Write([]byte(object.(string)))
	hashed := h.Sum(nil)
	hasher.Hash = hex.EncodeToString(hashed)
	return hasher.Hash
}
