package synchBlock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

func (bs *bootstrap) ReceivedHeader(nounce uint64) {
	bs.receivedHeader(nounce)
}

func (bs *bootstrap) ReceivedBody(nounce uint64) {
	bs.receivedBody(nounce)
}

func (bs *bootstrap) ShouldSynch() bool {
	return bs.shouldSynch()
}

func (bs *bootstrap) GetHeaderFromPool(nounce uint64) *block.Header {
	return bs.getHeaderFromPool(nounce)
}

func (bs *bootstrap) GetBodyFromPool(nounce uint64) *block.Block {
	return bs.getBodyFromPool(nounce)
}

func (bs *bootstrap) RequestHeader(nounce uint64) {
	bs.requestHeader(nounce)
}

func (bs *bootstrap) RequestBody(nounce uint64) {
	bs.requestBody(nounce)
}
