package sync

import (
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

func (bfd *basicForkDetector) CheckpointNonce() uint64 {
	return bfd.checkpointNonce
}

func (bfd *basicForkDetector) SetLastCheckpointNonce(nonce uint64) {
	bfd.lastCheckpointNonce = nonce
}

func (bfd *basicForkDetector) LastCheckpointNonce() uint64 {
	return bfd.lastCheckpointNonce
}

func (bfd *basicForkDetector) SetCheckpointRound(round int32) {
	bfd.checkpointRound = round
}

func (bfd *basicForkDetector) CheckpointRound() int32 {
	return bfd.checkpointRound
}

func (bfd *basicForkDetector) SetLastCheckpointRound(round int32) {
	bfd.lastCheckpointRound = round
}

func (bfd *basicForkDetector) LastCheckpointRound() int32 {
	return bfd.lastCheckpointRound
}

func (bfd *basicForkDetector) Append(hdrInfo *headerInfo) {
	bfd.append(hdrInfo)
}

func (bfd *basicForkDetector) CheckBlockValidity(header *block.Header) error {
	return bfd.checkBlockValidity(header)
}

func (bfd *basicForkDetector) RemovePastHeaders() {
	bfd.removePastHeaders()
}

func (bfd *basicForkDetector) RemoveInvalidHeaders() {
	bfd.removeInvalidHeaders()
}

func (bfd *basicForkDetector) ComputeProbableHighestNonce() uint64 {
	return bfd.computeProbableHighestNonce()
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

func (boot *ShardBootstrap) SetIsForkDetected(isForkDetected bool) {
	boot.isForkDetected = isForkDetected
}

func (boot *MetaBootstrap) SetIsForkDetected(isForkDetected bool) {
	boot.isForkDetected = isForkDetected
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
