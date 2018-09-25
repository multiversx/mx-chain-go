package hashing

import (
	"encoding/hex"

	"golang.org/x/crypto/blake2b"
)

type Blake2b struct {
	Hash string
}

func (hasher *Blake2b) CalculateHash(obj interface{}) interface{} {
	h, err := blake2b.New256(nil)

	if err != nil {
		return err
	}

	h.Write([]byte(obj.(string)))
	hashed := h.Sum(nil)
	hasher.Hash = hex.EncodeToString(hashed)
	return hasher.Hash
}
