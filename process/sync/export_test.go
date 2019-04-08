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

func (boot *Bootstrap) GetMiniBlocks(hashes [][]byte) interface{} {
	return boot.miniBlockResolver.GetMiniBlocks(hashes)
}

func (boot *Bootstrap) ReceivedHeaders(key []byte) {
	boot.receivedHeaders(key)
}

func (boot *Bootstrap) ForkChoice() error {
	return boot.forkChoice()
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

func (bfd *basicForkDetector) SetCheckpointNonce(nonce uint64) {
	bfd.checkpointNonce = nonce
}

func (bfd *basicForkDetector) SetLastCheckpointNonce(nonce uint64) {
	bfd.lastCheckpointNonce = nonce
}

func (bfd *basicForkDetector) CheckpointNonce() uint64 {
	return bfd.checkpointNonce
}

func (bfd *basicForkDetector) Append(hdrInfo *headerInfo) {
	bfd.append(hdrInfo)
}

func (bfd *basicForkDetector) RemovePastHeaders(nonce uint64) {
	bfd.removePastHeaders(nonce)
}

func (hi *headerInfo) Nonce() uint64 {
	return hi.nonce
}

func (hi *headerInfo) Hash() []byte {
	return hi.hash
}

func (hi *headerInfo) IsProcessed() bool {
	return hi.isProcessed
}

func (hi *headerInfo) IsSigned() bool {
	return hi.isSigned
}

func (boot *Bootstrap) NotifySyncStateListeners() {
	boot.notifySyncStateListeners()
}

func (boot *Bootstrap) SyncStateListeners() []func(bool) {
	return boot.syncStateListeners
}

func (boot *Bootstrap) HighestNonceReceived() uint64 {
	return boot.rcvHdrInfo.highestNonce
}

func (boot *Bootstrap) SetHighestNonceReceived(highestNonceReceived uint64) {
	boot.rcvHdrInfo.highestNonce = highestNonceReceived
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

func (boot *Bootstrap) BroadcastEmptyBlock(txBlockBody block.Body, header *block.Header) error {
	return boot.broadcastEmptyBlock(txBlockBody, header)
}

func (boot *Bootstrap) SetIsNodeSynchronized(isNodeSyncronized bool) {
	boot.isNodeSynchronized = isNodeSyncronized
}

func (boot *Bootstrap) SetRoundIndex(roundIndex int32) {
	boot.roundIndex = roundIndex
}

func (boot *Bootstrap) SetForkNonce(nonce uint64) {
	boot.forkNonce = nonce
}
