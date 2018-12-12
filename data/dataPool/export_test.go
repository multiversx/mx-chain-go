package dataPool

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

func (nthc *NonceToHashCacher) NonceToByteArray(nonce uint64) []byte {
	return nthc.nonceToByteArray(nonce)
}

func (nthc *NonceToHashCacher) ByteArrayToNonce(buff []byte) uint64 {
	return nthc.byteArrayToNonce(buff)
}

// NewTestNonceToHashCacher is the test constructor
func NewTestNonceToHashCacher(cacher storage.Cacher) *NonceToHashCacher {
	return &NonceToHashCacher{cacher: cacher}
}
