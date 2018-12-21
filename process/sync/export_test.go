package sync

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

func (boot *bootstrap) RequestHeader(nonce uint64) {
	boot.requestHeader(nonce)
}

func (boot *bootstrap) ShouldSync() bool {
	return boot.shouldSync()
}

func (boot *bootstrap) GetHeaderFromPool(nonce uint64) *block.Header {
	return boot.getHeaderFromPool(nonce)
}

func (boot *bootstrap) GetTxBodyFromPool(hash []byte) interface{} {
	return boot.getTxBodyFromPool(hash)
}
