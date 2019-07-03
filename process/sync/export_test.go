package sync

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
)

func (boot *ShardBootstrap) RequestHeader(nonce uint64) {
	boot.requestHeader(nonce)
}

func (boot *ShardBootstrap) GetHeaderFromPoolWithNonce(nonce uint64) (*block.Header, error) {
	return boot.getHeaderFromPoolWithNonce(nonce)
}

func (boot *MetaBootstrap) GetHeaderFromPoolWithNonce(nonce uint64) (*block.MetaBlock, error) {
	return boot.getHeaderFromPoolWithNonce(nonce)
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

func (bfd *basicForkDetector) CheckpointNonce() uint64 {
	return bfd.fork.checkpointNonce
}

func (bfd *basicForkDetector) SetLastCheckpointNonce(nonce uint64) {
	bfd.fork.lastCheckpointNonce = nonce
}

func (bfd *basicForkDetector) CheckpointRound() int32 {
	return bfd.fork.checkpointRound
}

func (bfd *basicForkDetector) SetLastCheckpointRound(round int32) {
	bfd.fork.lastCheckpointRound = round
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

func (boot *baseBootstrap) ProcessReceivedHeader(headerHandler data.HeaderHandler, headerHash []byte) {
	boot.processReceivedHeader(headerHandler, headerHash)
}

func (boot *baseBootstrap) LoadBlocks(
	blockFinality uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
	getHeader func(uint64) (data.HeaderHandler, []byte, error),
	getBlockBody func(data.HeaderHandler) (data.BodyHandler, error),
	removeBlockBody func(uint64, dataRetriever.UnitType, dataRetriever.UnitType) error,
) error {
	return boot.loadBlocks(blockFinality, blockUnit, hdrNonceHashDataUnit, getHeader, getBlockBody, removeBlockBody)
}

func (boot *baseBootstrap) ApplyBlock(
	nonce uint64,
	getHeader func(uint64) (data.HeaderHandler, []byte, error),
	getBlockBody func(data.HeaderHandler) (data.BodyHandler, error),
) error {
	return boot.applyBlock(nonce, getHeader, getBlockBody)
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

func (boot *baseBootstrap) LoadNotarizedBlocks(blockFinality uint64,
	hdrNonceHashDataUnit dataRetriever.UnitType,
	applyNotarisedBlock func(uint64, dataRetriever.UnitType) error,
) error {
	return boot.loadNotarizedBlocks(blockFinality, hdrNonceHashDataUnit, applyNotarisedBlock)
}

func (boot *ShardBootstrap) ApplyNotarizedBlock(nonce uint64, notarizedHdrNonceHashDataUnit dataRetriever.UnitType) error {
	return boot.applyNotarizedBlock(nonce, notarizedHdrNonceHashDataUnit)
}

func (boot *MetaBootstrap) ApplyNotarizedBlock(nonce uint64, notarizedHdrNonceHashDataUnit dataRetriever.UnitType) error {
	return boot.applyNotarizedBlock(nonce, notarizedHdrNonceHashDataUnit)
}

func (boot *baseBootstrap) RemoveNotarizedBlockHeader(
	nonce uint64,
	hdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	return boot.removeNotarizedBlockHeader(nonce, hdrNonceHashDataUnit)
}

func (boot *ShardBootstrap) SyncFromStorer(
	blockFinality uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
	notarizedBlockFinality uint64,
	notarizedHdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	return boot.syncFromStorer(blockFinality, blockUnit, hdrNonceHashDataUnit, notarizedBlockFinality, notarizedHdrNonceHashDataUnit)
}

func (boot *MetaBootstrap) SyncFromStorer(
	blockFinality uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
	notarizedBlockFinality uint64,
	notarizedHdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	return boot.syncFromStorer(blockFinality, blockUnit, hdrNonceHashDataUnit, notarizedBlockFinality, notarizedHdrNonceHashDataUnit)
}
