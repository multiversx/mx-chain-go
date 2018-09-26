package hashing

import (
	"crypto/sha256"
	"fmt"
)

type Sha256Impl struct {
	Nil []byte
}

func (Sha256Impl) CalculateHash(obj interface{}) []byte {
	if obj == nil {
		fmt.Println("Nil object in sha256impl")
		return []byte{}
	}
	h := sha256.New()
	h.Write([]byte(obj.(string)))
	return h.Sum(nil)
}

func (Sha256Impl) EmptyHash() []byte {
	return []byte{227, 176, 196, 66, 152, 252, 28, 20, 154, 251, 244, 200, 153, 111, 185, 36, 39, 174, 65, 228, 100, 155, 147, 76, 164, 149, 153, 27, 120, 82, 184, 85}
}

func (Sha256Impl) HashSize() int {
	return 32
}
