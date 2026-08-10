package shardedData

import (
	"github.com/multiversx/mx-chain-go/storage"
)

type immunityCache interface {
	storage.Cacher
	ImmunizeKeys(keys [][]byte, nonce uint64) (numNowTotal, numFutureTotal int)
	SetOldestImmuneNonce(nonce uint64)
	RemoveWithResult(key []byte) bool
	NumBytes() int
	Diagnose(deep bool)
}
