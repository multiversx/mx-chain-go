package sync

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

func (boot *ShardBootstrap) RequestHeader(nonce uint64) {
	boot.requestHeader(nonce)
}

func (boot *MetaBootstrap) RequestHeader(nonce uint64) {
	boot.requestHeader(nonce)
}

func (boot *ShardBootstrap) GetHeaderFromPool(nonce uint64) *block.Header {
	return boot.getHeaderFromPoolHavingNonce(nonce)
}

func (boot *MetaBootstrap) GetHeaderFromPool(nonce uint64) *block.MetaBlock {
	return boot.getHeaderFromPoolHavingNonce(nonce)
}

func (boot *ShardBootstrap) GetMiniBlocks(hashes [][]byte) interface{} {
	return boot.miniBlockResolver.GetMiniBlocks(hashes)
}

func (boot *MetaBootstrap) ReceivedHeaders(key []byte) {
	boot.receivedHeader(key)
}

func (boot *ShardBootstrap) ReceivedHeaders(key []byte) {
	boot.receivedHeaders(key)
}

func (boot *ShardBootstrap) ForkChoice() error {
	return boot.forkChoice()
}

func (boot *MetaBootstrap) ForkChoice() error {
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

func (boot *ShardBootstrap) NotifySyncStateListeners() {
	boot.notifySyncStateListeners()
}

func (boot *MetaBootstrap) NotifySyncStateListeners() {
	boot.notifySyncStateListeners()
}

func (boot *ShardBootstrap) SyncStateListeners() []func(bool) {
	return boot.syncStateListeners
}

func (boot *MetaBootstrap) SyncStateListeners() []func(bool) {
	return boot.syncStateListeners
}

func (boot *MetaBootstrap) HighestNonceReceived() uint64 {
	return boot.rcvHdrInfo.highestNonce
}

func (boot *ShardBootstrap) HighestNonceReceived() uint64 {
	return boot.rcvHdrInfo.highestNonce
}

func (boot *ShardBootstrap) SetHighestNonceReceived(highestNonceReceived uint64) {
	boot.rcvHdrInfo.highestNonce = highestNonceReceived
}

func (boot *MetaBootstrap) SetHighestNonceReceived(highestNonceReceived uint64) {
	boot.rcvHdrInfo.highestNonce = highestNonceReceived
}

func (boot *ShardBootstrap) SetIsForkDetected(isForkDetected bool) {
	boot.isForkDetected = isForkDetected
}

func (boot *MetaBootstrap) SetIsForkDetected(isForkDetected bool) {
	boot.isForkDetected = isForkDetected
}

func (boot *ShardBootstrap) GetTimeStampForRound(roundIndex uint32) time.Time {
	return boot.getTimeStampForRound(roundIndex)
}

func (boot *MetaBootstrap) GetTimeStampForRound(roundIndex uint32) time.Time {
	return boot.getTimeStampForRound(roundIndex)
}

func (boot *ShardBootstrap) ShouldCreateEmptyBlock(nonce uint64) bool {
	return boot.shouldCreateEmptyBlock(nonce)
}

func (boot *MetaBootstrap) ShouldCreateEmptyBlock(nonce uint64) bool {
	return boot.shouldCreateEmptyBlock(nonce)
}

func (boot *ShardBootstrap) CreateAndBroadcastEmptyBlock() error {
	return boot.createAndBroadcastEmptyBlock()
}

func (boot *MetaBootstrap) CreateAndBroadcastEmptyBlock() error {
	return boot.createAndBroadcastEmptyBlock()
}

func (boot *ShardBootstrap) BroadcastEmptyBlock(txBlockBody block.Body, header *block.Header) error {
	return boot.broadcastEmptyBlock(txBlockBody, header)
}

func (boot *MetaBootstrap) BroadcastEmptyBlock(txBlockBody block.Body, header *block.Header) error {
	return boot.broadcastEmptyBlock(txBlockBody, header)
}

func (boot *ShardBootstrap) SetIsNodeSynchronized(isNodeSyncronized bool) {
	boot.isNodeSynchronized = isNodeSyncronized
}

func (boot *MetaBootstrap) SetIsNodeSynchronized(isNodeSyncronized bool) {
	boot.isNodeSynchronized = isNodeSyncronized
}

func (boot *ShardBootstrap) SetRoundIndex(roundIndex int32) {
	boot.roundIndex = roundIndex
}

func (boot *MetaBootstrap) SetRoundIndex(roundIndex int32) {
	boot.roundIndex = roundIndex
}

func (boot *ShardBootstrap) SetForkNonce(nonce uint64) {
	boot.forkNonce = nonce
}

func (boot *MetaBootstrap) SetForkNonce(nonce uint64) {
	boot.forkNonce = nonce
}
