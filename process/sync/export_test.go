package sync

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

func (boot *ShardBootstrap) RequestHeaderWithNonce(nonce uint64) {
	boot.requestHeaderWithNonce(nonce)
}

func (boot *ShardBootstrap) GetMiniBlocks(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	return boot.miniBlocksResolver.GetMiniBlocks(hashes)
}

func (boot *MetaBootstrap) ReceivedHeaders(header data.HeaderHandler, key []byte) {
	boot.processReceivedHeader(header, key)
}

func (boot *ShardBootstrap) ReceivedHeaders(header data.HeaderHandler, key []byte) {
	boot.processReceivedHeader(header, key)
}

func (boot *ShardBootstrap) RollBack(revertUsingForkNonce bool) error {
	return boot.rollBack(revertUsingForkNonce)
}

func (boot *MetaBootstrap) RollBack(revertUsingForkNonce bool) error {
	return boot.rollBack(revertUsingForkNonce)
}

func (bfd *baseForkDetector) GetHeaders(nonce uint64) []*headerInfo {
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

func (bfd *baseForkDetector) LastCheckpointNonce() uint64 {
	return bfd.lastCheckpoint().nonce
}

func (bfd *baseForkDetector) LastCheckpointRound() uint64 {
	return bfd.lastCheckpoint().round
}

func (bfd *baseForkDetector) SetFinalCheckpoint(nonce uint64, round uint64, hash []byte) {
	bfd.setFinalCheckpoint(&checkpointInfo{nonce: nonce, round: round, hash: hash})
}

func (bfd *baseForkDetector) FinalCheckpointNonce() uint64 {
	return bfd.finalCheckpoint().nonce
}

func (bfd *baseForkDetector) FinalCheckpointRound() uint64 {
	return bfd.finalCheckpoint().round
}

func (bfd *baseForkDetector) CheckBlockValidity(header *block.Header, headerHash []byte, state process.BlockHeaderState) error {
	return bfd.checkBlockBasicValidity(header, headerHash, state)
}

func (bfd *baseForkDetector) RemovePastHeaders() {
	bfd.removePastHeaders()
}

func (bfd *baseForkDetector) RemoveInvalidReceivedHeaders() {
	bfd.removeInvalidReceivedHeaders()
}

func (bfd *baseForkDetector) ComputeProbableHighestNonce() uint64 {
	return bfd.computeProbableHighestNonce()
}

func (bfd *baseForkDetector) IsConsensusStuck() bool {
	return bfd.isConsensusStuck()
}

func (hi *headerInfo) Hash() []byte {
	return hi.hash
}

func (hi *headerInfo) GetBlockHeaderState() process.BlockHeaderState {
	return hi.state
}

func (boot *ShardBootstrap) NotifySyncStateListeners() {
	boot.notifySyncStateListeners(boot.isNodeSynchronized)
}

func (boot *MetaBootstrap) NotifySyncStateListeners() {
	boot.notifySyncStateListeners(boot.isNodeSynchronized)
}

func (boot *ShardBootstrap) SyncStateListeners() []func(bool) {
	return boot.syncStateListeners
}

func (boot *MetaBootstrap) SyncStateListeners() []func(bool) {
	return boot.syncStateListeners
}

func (boot *ShardBootstrap) SetForkNonce(nonce uint64) {
	boot.forkInfo.Nonce = nonce
}

func (boot *MetaBootstrap) SetForkNonce(nonce uint64) {
	boot.forkInfo.Nonce = nonce
}

func (boot *ShardBootstrap) IsForkDetected() bool {
	return boot.forkInfo.IsDetected
}

func (boot *MetaBootstrap) IsForkDetected() bool {
	return boot.forkInfo.IsDetected
}

func (boot *MetaBootstrap) GetNotarizedInfo(
	lastNotarized map[uint32]*HdrInfo,
	finalNotarized map[uint32]*HdrInfo,
	blockWithLastNotarized map[uint32]uint64,
	blockWithFinalNotarized map[uint32]uint64,
	startNonce uint64,
) *notarizedInfo {
	return &notarizedInfo{
		lastNotarized:           lastNotarized,
		finalNotarized:          finalNotarized,
		blockWithLastNotarized:  blockWithLastNotarized,
		blockWithFinalNotarized: blockWithFinalNotarized,
		startNonce:              startNonce,
	}
}

func (boot *baseBootstrap) ProcessReceivedHeader(headerHandler data.HeaderHandler, headerHash []byte) {
	boot.processReceivedHeader(headerHandler, headerHash)
}

func (boot *ShardBootstrap) RequestMiniBlocksFromHeaderWithNonceIfMissing(headerHandler data.HeaderHandler) {
	boot.requestMiniBlocksFromHeaderWithNonceIfMissing(headerHandler)
}

func (bfd *baseForkDetector) IsHeaderReceivedTooLate(header data.HeaderHandler, state process.BlockHeaderState, finality int64) bool {
	return bfd.isHeaderReceivedTooLate(header, state, finality)
}

func (bfd *baseForkDetector) SetProbableHighestNonce(nonce uint64) {
	bfd.setProbableHighestNonce(nonce)
}

func (sfd *shardForkDetector) ComputeFinalCheckpoint() {
	sfd.computeFinalCheckpoint()
}

func (bfd *baseForkDetector) AddCheckPoint(round uint64, nonce uint64, hash []byte) {
	bfd.addCheckpoint(&checkpointInfo{round: round, nonce: nonce, hash: hash})
}

func (bfd *baseForkDetector) ComputeGenesisTimeFromHeader(headerHandler data.HeaderHandler) int64 {
	return bfd.computeGenesisTimeFromHeader(headerHandler)
}

func (boot *baseBootstrap) InitNotarizedMap() map[uint32]*HdrInfo {
	return make(map[uint32]*HdrInfo, 0)
}

func (boot *baseBootstrap) SetNotarizedMap(notarizedMap map[uint32]*HdrInfo, shardId uint32, nonce uint64, hash []byte) {
	hdrInfo, ok := notarizedMap[shardId]
	if !ok {
		notarizedMap[shardId] = &HdrInfo{Nonce: nonce, Hash: hash}
		return
	}

	hdrInfo.Nonce = nonce
	hdrInfo.Hash = hash
}
