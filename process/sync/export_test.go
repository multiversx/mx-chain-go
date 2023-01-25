package sync

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
)

// RequestHeaderWithNonce -
func (boot *ShardBootstrap) RequestHeaderWithNonce(nonce uint64) {
	boot.requestHeaderWithNonce(nonce)
}

// GetMiniBlocks -
func (boot *ShardBootstrap) GetMiniBlocks(hashes [][]byte) ([]*block.MiniblockAndHash, [][]byte) {
	return boot.miniBlocksProvider.GetMiniBlocks(hashes)
}

// ReceivedHeaders -
func (boot *MetaBootstrap) ReceivedHeaders(header data.HeaderHandler, key []byte) {
	boot.processReceivedHeader(header, key)
}

// ReceivedHeaders -
func (boot *ShardBootstrap) ReceivedHeaders(header data.HeaderHandler, key []byte) {
	boot.processReceivedHeader(header, key)
}

// RollBack -
func (boot *ShardBootstrap) RollBack(revertUsingForkNonce bool) error {
	return boot.rollBack(revertUsingForkNonce)
}

// RollBack -
func (boot *MetaBootstrap) RollBack(revertUsingForkNonce bool) error {
	return boot.rollBack(revertUsingForkNonce)
}

// GetHeaders -
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

// LastCheckpointNonce -
func (bfd *baseForkDetector) LastCheckpointNonce() uint64 {
	return bfd.lastCheckpoint().nonce
}

// LastCheckpointRound -
func (bfd *baseForkDetector) LastCheckpointRound() uint64 {
	return bfd.lastCheckpoint().round
}

// SetFinalCheckpoint -
func (bfd *baseForkDetector) SetFinalCheckpoint(nonce uint64, round uint64, hash []byte) {
	bfd.setFinalCheckpoint(&checkpointInfo{nonce: nonce, round: round, hash: hash})
}

// FinalCheckpointNonce -
func (bfd *baseForkDetector) FinalCheckpointNonce() uint64 {
	return bfd.finalCheckpoint().nonce
}

// FinalCheckpointRound -
func (bfd *baseForkDetector) FinalCheckpointRound() uint64 {
	return bfd.finalCheckpoint().round
}

// CheckBlockValidity -
func (bfd *baseForkDetector) CheckBlockValidity(header data.HeaderHandler, headerHash []byte) error {
	return bfd.checkBlockBasicValidity(header, headerHash)
}

// RemovePastHeaders -
func (bfd *baseForkDetector) RemovePastHeaders() {
	bfd.removePastHeaders()
}

// RemoveInvalidReceivedHeaders -
func (bfd *baseForkDetector) RemoveInvalidReceivedHeaders() {
	bfd.removeInvalidReceivedHeaders()
}

// ComputeProbableHighestNonce -
func (bfd *baseForkDetector) ComputeProbableHighestNonce() uint64 {
	return bfd.computeProbableHighestNonce()
}

// IsConsensusStuck -
func (bfd *baseForkDetector) IsConsensusStuck() bool {
	return bfd.isConsensusStuck()
}

// Hash -
func (hi *headerInfo) Hash() []byte {
	return hi.hash
}

// GetBlockHeaderState -
func (hi *headerInfo) GetBlockHeaderState() process.BlockHeaderState {
	return hi.state
}

// NotifySyncStateListeners -
func (boot *ShardBootstrap) NotifySyncStateListeners() {
	isNodeSynchronized := boot.GetNodeState() == common.NsSynchronized
	boot.notifySyncStateListeners(isNodeSynchronized)
}

// NotifySyncStateListeners -
func (boot *MetaBootstrap) NotifySyncStateListeners() {
	isNodeSynchronized := boot.GetNodeState() == common.NsSynchronized
	boot.notifySyncStateListeners(isNodeSynchronized)
}

// SyncStateListeners -
func (boot *ShardBootstrap) SyncStateListeners() []func(bool) {
	return boot.syncStateListeners
}

// SyncStateListeners -
func (boot *MetaBootstrap) SyncStateListeners() []func(bool) {
	return boot.syncStateListeners
}

// SetForkNonce -
func (boot *ShardBootstrap) SetForkNonce(nonce uint64) {
	boot.forkInfo.Nonce = nonce
}

// SetForkNonce -
func (boot *MetaBootstrap) SetForkNonce(nonce uint64) {
	boot.forkInfo.Nonce = nonce
}

// IsForkDetected -
func (boot *ShardBootstrap) IsForkDetected() bool {
	return boot.forkInfo.IsDetected
}

// IsForkDetected -
func (boot *MetaBootstrap) IsForkDetected() bool {
	return boot.forkInfo.IsDetected
}

// GetNotarizedInfo -
func (boot *MetaBootstrap) GetNotarizedInfo(
	lastNotarized map[uint32]*hdrInfo,
	finalNotarized map[uint32]*hdrInfo,
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

// ProcessReceivedHeader -
func (boot *baseBootstrap) ProcessReceivedHeader(headerHandler data.HeaderHandler, headerHash []byte) {
	boot.processReceivedHeader(headerHandler, headerHash)
}

// RequestMiniBlocksFromHeaderWithNonceIfMissing -
func (boot *ShardBootstrap) RequestMiniBlocksFromHeaderWithNonceIfMissing(headerHandler data.HeaderHandler) {
	boot.requestMiniBlocksFromHeaderWithNonceIfMissing(headerHandler)
}

// IsHeaderReceivedTooLate -
func (bfd *baseForkDetector) IsHeaderReceivedTooLate(header data.HeaderHandler, state process.BlockHeaderState, finality int64) bool {
	return bfd.isHeaderReceivedTooLate(header, state, finality)
}

// SetProbableHighestNonce -
func (bfd *baseForkDetector) SetProbableHighestNonce(nonce uint64) {
	bfd.setProbableHighestNonce(nonce)
}

// ComputeFinalCheckpoint -
func (sfd *shardForkDetector) ComputeFinalCheckpoint() {
	sfd.computeFinalCheckpoint()
}

// AddCheckPoint -
func (bfd *baseForkDetector) AddCheckPoint(round uint64, nonce uint64, hash []byte) {
	bfd.addCheckpoint(&checkpointInfo{round: round, nonce: nonce, hash: hash})
}

// ComputeGenesisTimeFromHeader -
func (bfd *baseForkDetector) ComputeGenesisTimeFromHeader(headerHandler data.HeaderHandler) int64 {
	return bfd.computeGenesisTimeFromHeader(headerHandler)
}

// InitNotarizedMap -
func (boot *baseBootstrap) InitNotarizedMap() map[uint32]*hdrInfo {
	return make(map[uint32]*hdrInfo)
}

// SetNotarizedMap -
func (boot *baseBootstrap) SetNotarizedMap(notarizedMap map[uint32]*hdrInfo, shardId uint32, nonce uint64, hash []byte) {
	hdrInfoInstance, ok := notarizedMap[shardId]
	if !ok {
		notarizedMap[shardId] = &hdrInfo{Nonce: nonce, Hash: hash}
		return
	}

	hdrInfoInstance.Nonce = nonce
	hdrInfoInstance.Hash = hash
}

// SetNodeStateCalculated -
func (boot *baseBootstrap) SetNodeStateCalculated(state bool) {
	boot.mutNodeState.Lock()
	boot.isNodeStateCalculated = state
	boot.mutNodeState.Unlock()
}

// ComputeNodeState -
func (boot *baseBootstrap) ComputeNodeState() {
	boot.computeNodeState()
}

// DoJobOnSyncBlockFail -
func (boot *baseBootstrap) DoJobOnSyncBlockFail(bodyHandler data.BodyHandler, headerHandler data.HeaderHandler, err error) {
	boot.doJobOnSyncBlockFail(bodyHandler, headerHandler, err)
}

// SetNumSyncedWithErrorsForNonce -
func (boot *baseBootstrap) SetNumSyncedWithErrorsForNonce(nonce uint64, numSyncedWithErrors uint32) {
	boot.mutNonceSyncedWithErrors.Lock()
	boot.mapNonceSyncedWithErrors[nonce] = numSyncedWithErrors
	boot.mutNonceSyncedWithErrors.Unlock()
}

// GetNumSyncedWithErrorsForNonce -
func (boot *baseBootstrap) GetNumSyncedWithErrorsForNonce(nonce uint64) uint32 {
	boot.mutNonceSyncedWithErrors.RLock()
	numSyncedWithErrors := boot.mapNonceSyncedWithErrors[nonce]
	boot.mutNonceSyncedWithErrors.RUnlock()

	return numSyncedWithErrors
}

// GetMapNonceSyncedWithErrorsLen -
func (boot *baseBootstrap) GetMapNonceSyncedWithErrorsLen() int {
	boot.mutNonceSyncedWithErrors.RLock()
	mapNonceSyncedWithErrorsLen := len(boot.mapNonceSyncedWithErrors)
	boot.mutNonceSyncedWithErrors.RUnlock()

	return mapNonceSyncedWithErrorsLen
}

// CleanNoncesSyncedWithErrorsBehindFinal -
func (boot *baseBootstrap) CleanNoncesSyncedWithErrorsBehindFinal() {
	boot.cleanNoncesSyncedWithErrorsBehindFinal()
}

// IsInImportMode -
func (boot *baseBootstrap) IsInImportMode() bool {
	return boot.isInImportMode
}

// ProcessWaitTime -
func (boot *baseBootstrap) ProcessWaitTime() time.Duration {
	return boot.processWaitTime
}
