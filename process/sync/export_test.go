package sync

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
)

func (boot *ShardBootstrap) RequestHeaderWithNonce(nonce uint64) {
	boot.requestHeaderWithNonce(nonce)
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

func (bfd *basicForkDetector) LastCheckpointNonce() uint64 {
	return bfd.lastCheckpoint().nonce
}

func (bfd *basicForkDetector) LastCheckpointRound() uint64 {
	return bfd.lastCheckpoint().round
}

func (bfd *basicForkDetector) SetFinalCheckpoint(nonce uint64, round uint64) {
	bfd.setFinalCheckpoint(&checkpointInfo{nonce: nonce, round: round})
}

func (bfd *basicForkDetector) FinalCheckpointNonce() uint64 {
	return bfd.finalCheckpoint().nonce
}

func (bfd *basicForkDetector) FinalCheckpointRound() uint64 {
	return bfd.finalCheckpoint().round
}

func (bfd *basicForkDetector) CheckBlockValidity(header *block.Header, state process.BlockHeaderState) error {
	return bfd.checkBlockValidity(header, state)
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

func (bfd *basicForkDetector) GetProbableHighestNonce(headersInfo []*headerInfo) uint64 {
	return bfd.getProbableHighestNonce(headersInfo)
}

func (hi *headerInfo) Hash() []byte {
	return hi.hash
}

func (hi *headerInfo) GetBlockHeaderState() process.BlockHeaderState {
	return hi.state
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

func (boot *ShardBootstrap) SetForkNonce(nonce uint64) {
	boot.forkNonce = nonce
}

func (boot *MetaBootstrap) SetForkNonce(nonce uint64) {
	boot.forkNonce = nonce
}

func (boot *ShardBootstrap) IsForkDetected() bool {
	return boot.isForkDetected
}

func (boot *MetaBootstrap) IsForkDetected() bool {
	return boot.isForkDetected
}

func (boot *baseBootstrap) ProcessReceivedHeader(headerHandler data.HeaderHandler, headerHash []byte) {
	boot.processReceivedHeader(headerHandler, headerHash)
}

func (boot *baseBootstrap) LoadBlocks(
	blockFinality uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	return boot.loadBlocks(
		blockFinality,
		blockUnit,
		hdrNonceHashDataUnit)
}

func (boot *baseBootstrap) ApplyBlock(
	shardId uint32,
	nonce uint64,
) error {
	return boot.applyBlock(shardId, nonce)
}

func (boot *baseBootstrap) RemoveBlockHeader(
	nonce uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	return boot.removeBlockHeader(nonce, blockUnit, hdrNonceHashDataUnit)
}

func (boot *ShardBootstrap) RemoveBlockBody(
	nonce uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	return boot.removeBlockBody(nonce, blockUnit, hdrNonceHashDataUnit)
}

func (boot *MetaBootstrap) RemoveBlockBody(
	nonce uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	return boot.removeBlockBody(nonce, blockUnit, hdrNonceHashDataUnit)
}

func (boot *ShardBootstrap) ApplyNotarizedBlocks(
	finalNotarized map[uint32]uint64,
	lastNotarized map[uint32]uint64,
) error {
	return boot.applyNotarizedBlocks(finalNotarized, lastNotarized)
}

func (boot *MetaBootstrap) ApplyNotarizedBlocks(
	finalNotarized map[uint32]uint64,
	lastNotarized map[uint32]uint64,
) error {
	return boot.applyNotarizedBlocks(finalNotarized, lastNotarized)
}

func (boot *ShardBootstrap) SyncFromStorer(
	blockFinality uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
	notarizedBlockFinality uint64,
	notarizedHdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	return boot.syncFromStorer(blockFinality, blockUnit, hdrNonceHashDataUnit, notarizedBlockFinality)
}

func (boot *MetaBootstrap) SyncFromStorer(
	blockFinality uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
	notarizedBlockFinality uint64,
	notarizedHdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	return boot.syncFromStorer(blockFinality, blockUnit, hdrNonceHashDataUnit, notarizedBlockFinality)
}

func (boot *ShardBootstrap) SetStorageBootstrapper(sb storageBootstrapper) {
	boot.storageBootstrapper = sb
}

func (boot *MetaBootstrap) SetStorageBootstrapper(sb storageBootstrapper) {
	boot.storageBootstrapper = sb
}

type StorageBootstrapperMock struct {
	GetHeaderCalled                 func(shardId uint32, nonce uint64) (data.HeaderHandler, []byte, error)
	GetBlockBodyCalled              func(data.HeaderHandler) (data.BodyHandler, error)
	RemoveBlockBodyCalled           func(nonce uint64, blockUnit dataRetriever.UnitType, hdrNonceHashDataUnit dataRetriever.UnitType) error
	GetNonceWithLastNotarizedCalled func(currentNonce uint64) (startNonce uint64, finalNotarized map[uint32]uint64, lastNotarized map[uint32]uint64)
	ApplyNotarizedBlocksCalled      func(finalNotarized map[uint32]uint64, lastNotarized map[uint32]uint64) error
	CleanupNotarizedStorageCalled   func(lastNotarized map[uint32]uint64)
}

func (sbm *StorageBootstrapperMock) getHeader(shardId uint32, nonce uint64) (data.HeaderHandler, []byte, error) {
	return sbm.GetHeaderCalled(shardId, nonce)
}

func (sbm *StorageBootstrapperMock) getBlockBody(headerHandler data.HeaderHandler) (data.BodyHandler, error) {
	return sbm.GetBlockBodyCalled(headerHandler)
}

func (sbm *StorageBootstrapperMock) removeBlockBody(
	nonce uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
) error {

	return sbm.RemoveBlockBodyCalled(nonce, blockUnit, hdrNonceHashDataUnit)
}

func (sbm *StorageBootstrapperMock) getNonceWithLastNotarized(
	currentNonce uint64,
) (startNonce uint64, finalNotarized map[uint32]uint64, lastNotarized map[uint32]uint64) {

	return sbm.GetNonceWithLastNotarizedCalled(currentNonce)
}

func (sbm *StorageBootstrapperMock) applyNotarizedBlocks(
	finalNotarized map[uint32]uint64,
	lastNotarized map[uint32]uint64,
) error {

	return sbm.ApplyNotarizedBlocksCalled(finalNotarized, lastNotarized)
}

func (sbm *StorageBootstrapperMock) cleanupNotarizedStorage(lastNotarized map[uint32]uint64) {
	sbm.CleanupNotarizedStorageCalled(lastNotarized)
}
