package hashing

import (
	"fmt"

	"golang.org/x/crypto/blake2b"
)

type Blake2bImpl struct {
	Nil []byte
}

func (Blake2bImpl) CalculateHash(obj interface{}) []byte {
	if obj == nil {
		fmt.Println("Nil object in blake2bimpl")
		return []byte{}
	}
	h, _ := blake2b.New256(nil)
	h.Write([]byte(obj.(string)))
	return h.Sum(nil)
}

func (Blake2bImpl) EmptyHash() []byte {
	return []byte{14, 87, 81, 192, 38, 229, 67, 178, 232, 171, 46, 176, 96, 153, 218, 161, 209, 229, 223, 71, 119, 143, 119, 135, 250, 171, 69, 205, 241, 47, 227, 168}
}

func (Blake2bImpl) HashSize() int {
	return 32
}
