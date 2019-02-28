package sync

import (
	"time"

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

func (boot *Bootstrap) SetHighestNonceReceived(highestNonceReceived uint64) {
	boot.highestNonceReceived = highestNonceReceived
}

func (boot *Bootstrap) SetIsForkDetected(isForkDetected bool) {
	boot.isForkDetected = isForkDetected
}

func (boot *Bootstrap) GetTimeStampForRound(roundIndex uint32) time.Time {
	return boot.getTimeStampForRound(roundIndex)
}

func (boot *Bootstrap) ShouldCreateEmptyBlock(nonce uint64) bool {
	return boot.shouldCreateEmptyBlock(nonce)
}

func (boot *Bootstrap) CreateAndBroadcastEmptyBlock() error {
	return boot.createAndBroadcastEmptyBlock()
}

func (boot *Bootstrap) BroadcastEmptyBlock(txBlockBody *block.TxBlockBody, header *block.Header) error {
	return boot.broadcastEmptyBlock(txBlockBody, header)
}
