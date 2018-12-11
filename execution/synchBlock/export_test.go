package synchBlock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

func (bs *bootstrap) ReceivedHeader(nonce uint64) {
	bs.receivedHeader(nonce)
}

func (bs *bootstrap) ReceivedBody(nonce uint64) {
	bs.receivedBody(nonce)
}

func (bs *bootstrap) ShouldSynch() bool {
	return bs.shouldSynch()
}

func (bs *bootstrap) GetHeaderFromPool(nonce uint64) *block.Header {
	return bs.getHeaderFromPool(nonce)
}

func (bs *bootstrap) GetBodyFromPool(nonce uint64) *block.Block {
	return bs.getBodyFromPool(nonce)
}

func (bs *bootstrap) RequestHeader(nonce uint64) {
	bs.requestHeader(nonce)
}

func (bs *bootstrap) RequestBody(nonce uint64) {
	bs.requestBody(nonce)
}
