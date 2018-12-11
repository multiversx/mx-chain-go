package syncBlock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

func (boot *bootstrap) ReceivedHeader(nonce uint64) {
	boot.receivedHeader(nonce)
}

func (boot *bootstrap) ReceivedBody(nonce uint64) {
	boot.receivedBody(nonce)
}

func (boot *bootstrap) ShouldSync() bool {
	return boot.shouldSync()
}

func (boot *bootstrap) GetHeaderFromPool(nonce uint64) *block.Header {
	return boot.getHeaderFromPool(nonce)
}

func (boot *bootstrap) GetBodyFromPool(nonce uint64) *block.Block {
	return boot.getBodyFromPool(nonce)
}

func (boot *bootstrap) RequestHeader(nonce uint64) {
	boot.requestHeader(nonce)
}

func (boot *bootstrap) RequestBody(nonce uint64) {
	boot.requestBody(nonce)
}

func (boot *bootstrap) RequestedHeaderNonce() int64 {
	return boot.requestedHeaderNonce()
}

func (boot *bootstrap) RequestedBodyNonce() int64 {
	return boot.requestedBodyNonce()
}
