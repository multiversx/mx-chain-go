package bn

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// factory

type Factory *factory

func (fct *factory) BlockChain() *blockchain.BlockChain {
	return fct.blockChain
}

func (fct *factory) SetBlockChain(blockChain *blockchain.BlockChain) {
	fct.blockChain = blockChain
}

func (fct *factory) BlockProcessor() process.BlockProcessor {
	return fct.blockProcessor
}

func (fct *factory) SetBlockProcessor(blockProcessor process.BlockProcessor) {
	fct.blockProcessor = blockProcessor
}

func (fct *factory) Bootstraper() process.Bootstraper {
	return fct.bootstraper
}

func (fct *factory) SetBootsraper(bootstraper process.Bootstraper) {
	fct.bootstraper = bootstraper
}

func (fct *factory) ChronologyHandler() consensus.ChronologyHandler {
	return fct.chronologyHandler
}

func (fct *factory) SetChronologyHandler(chronologyHandler consensus.ChronologyHandler) {
	fct.chronologyHandler = chronologyHandler
}

func (fct *factory) ConsensusState() *spos.ConsensusState {
	return fct.consensusState
}

func (fct *factory) SetConsensusState(consensusState *spos.ConsensusState) {
	fct.consensusState = consensusState
}

func (fct *factory) Hasher() hashing.Hasher {
	return fct.hasher
}

func (fct *factory) SetHasher(hasher hashing.Hasher) {
	fct.hasher = hasher
}

func (fct *factory) Marshalizer() marshal.Marshalizer {
	return fct.marshalizer
}

func (fct *factory) SetMarshalizer(marshalizer marshal.Marshalizer) {
	fct.marshalizer = marshalizer
}

func (fct *factory) MultiSigner() crypto.MultiSigner {
	return fct.multiSigner
}

func (fct *factory) SetMultiSigner(multiSigner crypto.MultiSigner) {
	fct.multiSigner = multiSigner
}

func (fct *factory) Rounder() consensus.Rounder {
	return fct.rounder
}

func (fct *factory) SetRounder(rounder consensus.Rounder) {
	fct.rounder = rounder
}

func (fct *factory) ShardCoordinator() sharding.ShardCoordinator {
	return fct.shardCoordinator
}

func (fct *factory) SetShardCoordinator(shardCoordinator sharding.ShardCoordinator) {
	fct.shardCoordinator = shardCoordinator
}

func (fct *factory) SyncTimer() ntp.SyncTimer {
	return fct.syncTimer
}

func (fct *factory) SetSyncTimer(syncTimer ntp.SyncTimer) {
	fct.syncTimer = syncTimer
}

func (fct *factory) ValidatorGroupSelector() consensus.ValidatorGroupSelector {
	return fct.validatorGroupSelector
}

func (fct *factory) SetValidatorGroupSelector(validatorGroupSelector consensus.ValidatorGroupSelector) {
	fct.validatorGroupSelector = validatorGroupSelector
}

func (fct *factory) Worker() *worker {
	return fct.worker
}

func (fct *factory) SetWorker(worker *worker) {
	fct.worker = worker
}

func (fct *factory) GenerateStartRoundSubround() error {
	return fct.generateStartRoundSubround()
}

func (fct *factory) GenerateBlockSubround() error {
	return fct.generateBlockSubround()
}

func (fct *factory) GenerateCommitmentHashSubround() error {
	return fct.generateCommitmentHashSubround()
}

func (fct *factory) GenerateBitmapSubround() error {
	return fct.generateBitmapSubround()
}

func (fct *factory) GenerateCommitmentSubround() error {
	return fct.generateCommitmentSubround()
}

func (fct *factory) GenerateSignatureSubround() error {
	return fct.generateSignatureSubround()
}

func (fct *factory) GenerateEndRoundSubround() error {
	return fct.generateEndRoundSubround()
}

// subround

func (sr *subround) SetJobFunction(job func() bool) {
	sr.job = job
}

func (sr *subround) SetCheckFunction(check func() bool) {
	sr.check = check
}

// worker

type Worker *worker

func (wrk *worker) Bootstraper() process.Bootstraper {
	return wrk.bootstraper
}

func (wrk *worker) SetBootstraper(bootstraper process.Bootstraper) {
	wrk.bootstraper = bootstraper
}

func (wrk *worker) ConsensusState() *spos.ConsensusState {
	return wrk.consensusState
}

func (wrk *worker) SetConsensusState(consensusState *spos.ConsensusState) {
	wrk.consensusState = consensusState
}

func (wrk *worker) KeyGenerator() crypto.KeyGenerator {
	return wrk.keyGenerator
}

func (wrk *worker) SetKeyGenerator(keyGenerator crypto.KeyGenerator) {
	wrk.keyGenerator = keyGenerator
}

func (wrk *worker) Marshalizer() marshal.Marshalizer {
	return wrk.marshalizer
}

func (wrk *worker) SetMarshalizer(marshalizer marshal.Marshalizer) {
	wrk.marshalizer = marshalizer
}

func (wrk *worker) Rounder() consensus.Rounder {
	return wrk.rounder
}

func (wrk *worker) SetRounder(rounder consensus.Rounder) {
	wrk.rounder = rounder
}

func (wrk *worker) CheckSignature(cnsData *spos.ConsensusMessage) error {
	return wrk.checkSignature(cnsData)
}

func (wrk *worker) ExecuteMessage(cnsDtaList []*spos.ConsensusMessage) {
	wrk.executeMessage(cnsDtaList)
}

func GetSubroundName(subroundId int) string {
	return getSubroundName(subroundId)
}

func (wrk *worker) InitReceivedMessages() {
	wrk.initReceivedMessages()
}

func (wrk *worker) SendConsensusMessage(cnsDta *spos.ConsensusMessage) bool {
	return wrk.sendConsensusMessage(cnsDta)
}

func (wrk *worker) BroadcastTxBlockBody2(blockBody *block.TxBlockBody) error {
	return wrk.broadcastTxBlockBody(blockBody)
}

func (wrk *worker) BroadcastHeader2(header *block.Header) error {
	return wrk.broadcastHeader(header)
}

func (wrk *worker) Extend(subroundId int) {
	wrk.extend(subroundId)
}

func (wrk *worker) ReceivedSyncState(isNodeSynchronized bool) {
	wrk.receivedSyncState(isNodeSynchronized)
}

func (wrk *worker) ReceivedMessages() map[MessageType][]*spos.ConsensusMessage {
	wrk.mutReceivedMessages.RLock()
	defer wrk.mutReceivedMessages.RUnlock()

	return wrk.receivedMessages
}

func (wrk *worker) SetReceivedMessages(messageType MessageType, cnsDta []*spos.ConsensusMessage) {
	wrk.mutReceivedMessages.Lock()
	wrk.receivedMessages[messageType] = cnsDta
	wrk.mutReceivedMessages.Unlock()
}

func (wrk *worker) NilReceivedMessages() {
	wrk.mutReceivedMessages.Lock()
	wrk.receivedMessages = nil
	wrk.mutReceivedMessages.Unlock()
}

func (wrk *worker) ReceivedMessagesCalls() map[MessageType]func(*spos.ConsensusMessage) bool {
	wrk.mutReceivedMessagesCalls.RLock()
	defer wrk.mutReceivedMessagesCalls.RUnlock()

	return wrk.receivedMessagesCalls
}

func (wrk *worker) SetReceivedMessagesCalls(messageType MessageType, f func(*spos.ConsensusMessage) bool) {
	wrk.mutReceivedMessagesCalls.Lock()
	wrk.receivedMessagesCalls[messageType] = f
	wrk.mutReceivedMessagesCalls.Unlock()
}

func (wrk *worker) ExecuteMessageChannel() chan *spos.ConsensusMessage {
	return wrk.executeMessageChannel
}

func (wrk *worker) ConsensusStateChangedChannels() chan bool {
	return wrk.consensusStateChangedChannels
}

func (wrk *worker) SetConsensusStateChangedChannels(consensusStateChangedChannels chan bool) {
	wrk.consensusStateChangedChannels = consensusStateChangedChannels
}

// subroundStartRound

type SubroundStartRound *subroundStartRound

func (sr *subroundStartRound) Bootstraper() process.Bootstraper {
	return sr.bootstraper
}

func (sr *subroundStartRound) SetBootsraper(bootstraper process.Bootstraper) {
	sr.bootstraper = bootstraper
}

func (sr *subroundStartRound) ConsensusState() *spos.ConsensusState {
	return sr.consensusState
}

func (sr *subroundStartRound) SetConsensusState(consensusState *spos.ConsensusState) {
	sr.consensusState = consensusState
}

func (sr *subroundStartRound) Rounder() consensus.Rounder {
	return sr.rounder
}

func (sr *subroundStartRound) SetRounder(rounder consensus.Rounder) {
	sr.rounder = rounder
}

func (sr *subroundStartRound) SyncTimer() ntp.SyncTimer {
	return sr.syncTimer
}

func (sr *subroundStartRound) SetSyncTimer(syncTimer ntp.SyncTimer) {
	sr.syncTimer = syncTimer
}

func (sr *subroundStartRound) DoStartRoundJob() bool {
	return sr.doStartRoundJob()
}

func (sr *subroundStartRound) DoStartRoundConsensusCheck() bool {
	return sr.doStartRoundConsensusCheck()
}

func (sr *subroundStartRound) GenerateNextConsensusGroup(roundIndex int32) error {
	return sr.generateNextConsensusGroup(roundIndex)
}

// subroundBlock

type SubroundBlock *subroundBlock

func (sr *subroundBlock) BlockChain() *blockchain.BlockChain {
	return sr.blockChain
}

func (sr *subroundBlock) SetBlockChain(blockChain *blockchain.BlockChain) {
	sr.blockChain = blockChain
}

func (sr *subroundBlock) BlockProcessor() process.BlockProcessor {
	return sr.blockProcessor
}

func (sr *subroundBlock) SetBlockProcessor(blockProcessor process.BlockProcessor) {
	sr.blockProcessor = blockProcessor
}

func (sr *subroundBlock) ConsensusState() *spos.ConsensusState {
	return sr.consensusState
}

func (sr *subroundBlock) SetConsensusState(consensusState *spos.ConsensusState) {
	sr.consensusState = consensusState
}

func (sr *subroundBlock) Rounder() consensus.Rounder {
	return sr.rounder
}

func (sr *subroundBlock) SetRounder(rounder consensus.Rounder) {
	sr.rounder = rounder
}

func (sr *subroundBlock) SyncTimer() ntp.SyncTimer {
	return sr.syncTimer
}

func (sr *subroundBlock) SetSyncTimer(syncTimer ntp.SyncTimer) {
	sr.syncTimer = syncTimer
}

func (sr *subroundBlock) DoBlockJob() bool {
	return sr.doBlockJob()
}

func (sr *subroundBlock) ReceivedBlockBody(cnsDta *spos.ConsensusMessage) bool {
	return sr.receivedBlockBody(cnsDta)
}

func (sr *subroundBlock) DecodeBlockBody(dta []byte) *block.TxBlockBody {
	return sr.decodeBlockBody(dta)
}

func (sr *subroundBlock) ReceivedBlockHeader(cnsDta *spos.ConsensusMessage) bool {
	return sr.receivedBlockHeader(cnsDta)
}

func (sr *subroundBlock) DecodeBlockHeader(dta []byte) *block.Header {
	return sr.decodeBlockHeader(dta)
}

func (sr *subroundBlock) ProcessReceivedBlock(cnsDta *spos.ConsensusMessage) bool {
	return sr.processReceivedBlock(cnsDta)
}

func (sr *subroundBlock) DoBlockConsensusCheck() bool {
	return sr.doBlockConsensusCheck()
}

func (sr *subroundBlock) IsBlockReceived(threshold int) bool {
	return sr.isBlockReceived(threshold)
}

// subroundCommitmentHash

type SubroundCommitmentHash *subroundCommitmentHash

func (sr *subroundCommitmentHash) ConsensusState() *spos.ConsensusState {
	return sr.consensusState
}

func (sr *subroundCommitmentHash) SetConsensusState(consensusState *spos.ConsensusState) {
	sr.consensusState = consensusState
}

func (sr *subroundCommitmentHash) MultiSigner() crypto.MultiSigner {
	return sr.multiSigner
}

func (sr *subroundCommitmentHash) SetMultiSigner(multiSigner crypto.MultiSigner) {
	sr.multiSigner = multiSigner
}

func (sr *subroundCommitmentHash) Rounder() consensus.Rounder {
	return sr.rounder
}

func (sr *subroundCommitmentHash) SetRounder(rounder consensus.Rounder) {
	sr.rounder = rounder
}

func (sr *subroundCommitmentHash) DoCommitmentHashJob() bool {
	return sr.doCommitmentHashJob()
}

func (sr *subroundCommitmentHash) ReceivedCommitmentHash(cnsDta *spos.ConsensusMessage) bool {
	return sr.receivedCommitmentHash(cnsDta)
}

func (sr *subroundCommitmentHash) DoCommitmentHashConsensusCheck() bool {
	return sr.doCommitmentHashConsensusCheck()
}

func (sr *subroundCommitmentHash) IsCommitmentHashReceived(threshold int) bool {
	return sr.isCommitmentHashReceived(threshold)
}

func (sr *subroundCommitmentHash) CommitmentHashesCollected(threshold int) bool {
	return sr.commitmentHashesCollected(threshold)
}

func (sr *subroundCommitmentHash) GenCommitmentHash() ([]byte, error) {
	return sr.genCommitmentHash()
}

// subroundBitmap

type SubroundBitmap *subroundBitmap

func (sr *subroundBitmap) ConsensusState() *spos.ConsensusState {
	return sr.consensusState
}

func (sr *subroundBitmap) SetConsensusState(consensusState *spos.ConsensusState) {
	sr.consensusState = consensusState
}

func (sr *subroundBitmap) Rounder() consensus.Rounder {
	return sr.rounder
}

func (sr *subroundBitmap) SetRounder(rounder consensus.Rounder) {
	sr.rounder = rounder
}

func (sr *subroundBitmap) DoBitmapJob() bool {
	return sr.doBitmapJob()
}

func (sr *subroundBitmap) ReceivedBitmap(cnsDta *spos.ConsensusMessage) bool {
	return sr.receivedBitmap(cnsDta)
}

func (sr *subroundBitmap) DoBitmapConsensusCheck() bool {
	return sr.doBitmapConsensusCheck()
}

func (sr *subroundBitmap) IsBitmapReceived(threshold int) bool {
	return sr.isBitmapReceived(threshold)
}

// subroundCommitment

type SubroundCommitment *subroundCommitment

func (sr *subroundCommitment) ConsensusState() *spos.ConsensusState {
	return sr.consensusState
}

func (sr *subroundCommitment) SetConsensusState(consensusState *spos.ConsensusState) {
	sr.consensusState = consensusState
}

func (sr *subroundCommitment) Rounder() consensus.Rounder {
	return sr.rounder
}

func (sr *subroundCommitment) SetRounder(rounder consensus.Rounder) {
	sr.rounder = rounder
}

func (sr *subroundCommitment) DoCommitmentJob() bool {
	return sr.doCommitmentJob()
}

func (sr *subroundCommitment) ReceivedCommitment(cnsDta *spos.ConsensusMessage) bool {
	return sr.receivedCommitment(cnsDta)
}

func (sr *subroundCommitment) DoCommitmentConsensusCheck() bool {
	return sr.doCommitmentConsensusCheck()
}

func (sr *subroundCommitment) CommitmentsCollected(threshold int) bool {
	return sr.commitmentsCollected(threshold)
}

// subroundSignature

type SubroundSignature *subroundSignature

func (sr *subroundSignature) ConsensusState() *spos.ConsensusState {
	return sr.consensusState
}

func (sr *subroundSignature) SetConsensusState(consensusState *spos.ConsensusState) {
	sr.consensusState = consensusState
}

func (sr *subroundSignature) MultiSigner() crypto.MultiSigner {
	return sr.multiSigner
}

func (sr *subroundSignature) SetMultiSigner(multiSigner crypto.MultiSigner) {
	sr.multiSigner = multiSigner
}

func (sr *subroundSignature) Rounder() consensus.Rounder {
	return sr.rounder
}

func (sr *subroundSignature) SetRounder(rounder consensus.Rounder) {
	sr.rounder = rounder
}

func (sr *subroundSignature) DoSignatureJob() bool {
	return sr.doSignatureJob()
}

func (sr *subroundSignature) ReceivedSignature(cnsDta *spos.ConsensusMessage) bool {
	return sr.receivedSignature(cnsDta)
}

func (sr *subroundSignature) DoSignatureConsensusCheck() bool {
	return sr.doSignatureConsensusCheck()
}

func (sr *subroundSignature) CheckCommitmentsValidity(bitmap []byte) error {
	return sr.checkCommitmentsValidity(bitmap)
}

func (sr *subroundSignature) SignaturesCollected(threshold int) bool {
	return sr.signaturesCollected(threshold)
}

// subroundEndRound

type SubroundEndRound *subroundEndRound

func (sr *subroundEndRound) BlockProcessor() process.BlockProcessor {
	return sr.blockProcessor
}

func (sr *subroundEndRound) SetBlockProcessor(blockProcessor process.BlockProcessor) {
	sr.blockProcessor = blockProcessor
}

func (sr *subroundEndRound) ConsensusState() *spos.ConsensusState {
	return sr.consensusState
}

func (sr *subroundEndRound) SetConsensusState(consensusState *spos.ConsensusState) {
	sr.consensusState = consensusState
}

func (sr *subroundEndRound) MultiSigner() crypto.MultiSigner {
	return sr.multiSigner
}

func (sr *subroundEndRound) SetMultiSigner(multiSigner crypto.MultiSigner) {
	sr.multiSigner = multiSigner
}

func (sr *subroundEndRound) DoEndRoundJob() bool {
	return sr.doEndRoundJob()
}

func (sr *subroundEndRound) DoEndRoundConsensusCheck() bool {
	return sr.doEndRoundConsensusCheck()
}

func (sr *subroundEndRound) CheckSignaturesValidity(bitmap []byte) error {
	return sr.checkSignaturesValidity(bitmap)
}

func (sr *subroundEndRound) BroadcastTxBlockBody() func(*block.TxBlockBody) error {
	return sr.broadcastTxBlockBody
}

func (sr *subroundEndRound) SetBroadcastTxBlockBody(broadcastTxBlockBody func(*block.TxBlockBody) error) {
	sr.broadcastTxBlockBody = broadcastTxBlockBody
}

func (sr *subroundEndRound) BroadcastHeader() func(*block.Header) error {
	return sr.broadcastHeader
}

func (sr *subroundEndRound) SetBroadcastHeader(broadcastHeader func(*block.Header) error) {
	sr.broadcastHeader = broadcastHeader
}

func (sr *subroundStartRound) InitCurrentRound() bool {
	return sr.initCurrentRound()
}
