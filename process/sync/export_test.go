package sync

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

func (boot *Bootstrap) RequestHeader(nonce uint64) {
	boot.requestHeader(nonce)
}

func (boot *Bootstrap) GetHeaderFromPool(nonce uint64) *block.Header {
	return boot.getHeaderFromPoolHavingNonce(nonce)
}

func (boot *Bootstrap) GetTxBody(hash []byte) interface{} {
	return boot.getTxBody(hash)
}

func (boot *Bootstrap) ReceivedHeaders(key []byte) {
	boot.receivedHeaders(key)
}

func (boot *Bootstrap) ForkChoice(hdr *block.Header) error {
	return boot.forkChoice(hdr)
}

func (bfd *basicForkDetector) GetHeaders(nonce uint64) []*headerInfo {
	bfd.mutHeaders.Lock()
	defer bfd.mutHeaders.Unlock()

	headers := bfd.headers[nonce]

	if headers == nil {
		return nil
	}

	newHeaders := make([]*headerInfo, len(headers))
	copy(newHeaders, headers)

	return newHeaders
}

func (hi *headerInfo) Header() *block.Header {
	return hi.header
}

func (hi *headerInfo) Hash() []byte {
	return hi.hash
}

func (hi *headerInfo) IsReceived() bool {
	return hi.isReceived
}

func (boot *Bootstrap) NotifySyncStateListeners() {
	boot.notifySyncStateListeners()
}

func (boot *Bootstrap) SyncStateListeners() []func(bool) {
	return boot.syncStateListeners
}
