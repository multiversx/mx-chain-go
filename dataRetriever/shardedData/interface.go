package shardedData

import (
	"github.com/multiversx/mx-chain-go/storage"
)

type immunityCache interface {
	storage.Cacher
	ImmunizeKeys(keys [][]byte) (numNowTotal, numFutureTotal int)
	RemoveWithResult(key []byte) bool
	NumBytes() int
	Diagnose(deep bool)
}
