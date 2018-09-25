package sha256

import (
	"crypto/sha256"
	"encoding/hex"
)

type Sha256 struct {
}

func (hasher *Sha256) CalculateHash(obj interface{}) interface{} {
	h := sha256.New()
	h.Write([]byte(obj.(string)))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}
