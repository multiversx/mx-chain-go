package sync

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
)

func (boot *ShardBootstrap) RequestHeaderWithNonce(nonce uint64) {
	boot.requestHeaderWithNonce(nonce)
}

func (boot *ShardBootstrap) GetMiniBlocks(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	return boot.miniBlocksResolver.GetMiniBlocks(hashes)
}

func (boot *MetaBootstrap) ReceivedHeaders(key []byte) {
	boot.receivedHeader(key)
}

func (boot *ShardBootstrap) ReceivedHeaders(key []byte) {
	boot.receivedHeaders(key)
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

func (bfd *baseForkDetector) SetFinalCheckpoint(nonce uint64, round uint64) {
	bfd.setFinalCheckpoint(&checkpointInfo{nonce: nonce, round: round})
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

func (bfd *baseForkDetector) ActivateForcedForkIfNeeded(
	header data.HeaderHandler,
	state process.BlockHeaderState,
) {
	bfd.activateForcedForkIfNeeded(header, state)
}

func (bfd *baseForkDetector) ShouldForceFork() bool {
	return bfd.shouldForceFork()
}

func (bfd *baseForkDetector) SetShouldForceFork(shouldForceFork bool) {
	bfd.setShouldForceFork(shouldForceFork)
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

func (boot *ShardBootstrap) RequestMiniBlocksFromHeaderWithNonceIfMissing(shardId uint32, nonce uint64) {
	boot.requestMiniBlocksFromHeaderWithNonceIfMissing(shardId, nonce)
}

type StorageBootstrapperMock struct {
	GetHeaderCalled               func(hash []byte) (data.HeaderHandler, error)
	GetBlockBodyCalled            func(header data.HeaderHandler) (data.BodyHandler, error)
	ApplyNotarizedBlocksCalled    func(lastNotarized map[uint32]*HdrInfo) error
	AddHeaderToForkDetectorCalled func(shardId uint32, nonce uint64, lastNotarizedMeta uint64)
}

func (sbm *StorageBootstrapperMock) getHeader(hash []byte) (data.HeaderHandler, error) {
	return sbm.GetHeaderCalled(hash)
}

func (sbm *StorageBootstrapperMock) getBlockBody(header data.HeaderHandler) (data.BodyHandler, error) {
	return sbm.GetBlockBodyCalled(header)
}

func (sbm *StorageBootstrapperMock) applyNotarizedBlocks(lastNotarized map[uint32]*HdrInfo) error {

	return sbm.ApplyNotarizedBlocksCalled(lastNotarized)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sbm *StorageBootstrapperMock) IsInterfaceNil() bool {
	if sbm == nil {
		return true
	}
	return false
}

func (bfd *baseForkDetector) IsHeaderReceivedTooLate(header data.HeaderHandler, state process.BlockHeaderState, finality int64) bool {
	return bfd.isHeaderReceivedTooLate(header, state, finality)
}

func (bfd *baseForkDetector) SetProbableHighestNonce(nonce uint64) {
	bfd.setProbableHighestNonce(nonce)
}

func (sfd *shardForkDetector) AddFinalHeaders(finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) {
	sfd.addFinalHeaders(finalHeaders, finalHeadersHashes)
}

func (bfd *baseForkDetector) AddCheckPoint(round uint64, nonce uint64) {
	bfd.addCheckpoint(&checkpointInfo{round: round, nonce: nonce})
}

func GetCacherWithHeaders(
	hdr1 data.HeaderHandler,
	hdr2 data.HeaderHandler,
	hash1 []byte,
	hash2 []byte,
) storage.Cacher {
	sds := &mock.CacherStub{
		RegisterHandlerCalled: func(func(key []byte)) {},
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(key, hash1) {
				return &hdr1, true
			}
			if bytes.Equal(key, hash2) {
				return &hdr2, true
			}

			return nil, false
		},
	}
	return sds
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
