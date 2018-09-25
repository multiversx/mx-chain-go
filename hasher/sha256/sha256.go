package hasher

import (
	"crypto/sha256"
	"encoding/hex"
)

type Sha256Impl struct {
}

func (si Sha256Impl) CalculateHash(obj interface{}) interface{} {
	h := sha256.New()
	h.Write([]byte(obj.(string)))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}
