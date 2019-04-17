package bn

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// factory

type Factory *factory

func (fct *factory) BlockChain() data.ChainHandler {
	return fct.consensusDataContainer.Blockchain()
}

func (fct *factory) BlockProcessor() process.BlockProcessor {
	return fct.consensusDataContainer.BlockProcessor()
}

func (fct *factory) Bootstraper() process.Bootstrapper {
	return fct.consensusDataContainer.BootStrapper()
}

func (fct *factory) ChronologyHandler() consensus.ChronologyHandler {
	return fct.consensusDataContainer.Chronology()
}

func (fct *factory) ConsensusState() *spos.ConsensusState {
	return fct.consensusState
}

func (fct *factory) Hasher() hashing.Hasher {
	return fct.consensusDataContainer.Hasher()
}

func (fct *factory) Marshalizer() marshal.Marshalizer {
	return fct.consensusDataContainer.Marshalizer()
}

func (fct *factory) MultiSigner() crypto.MultiSigner {
	return fct.consensusDataContainer.MultiSigner()
}

func (fct *factory) Rounder() consensus.Rounder {
	return fct.consensusDataContainer.Rounder()
}

func (fct *factory) ShardCoordinator() sharding.Coordinator {
	return fct.consensusDataContainer.ShardCoordinator()
}

func (fct *factory) SyncTimer() ntp.SyncTimer {
	return fct.consensusDataContainer.SyncTimer()
}

func (fct *factory) ValidatorGroupSelector() consensus.ValidatorGroupSelector {
	return fct.consensusDataContainer.ValidatorGroupSelector()
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

func (sr *subround) SetConsensusState(state *spos.ConsensusState) {
	sr.consensusState = state
}

// worker

type Worker *worker

func (wrk *worker) Bootstraper() process.Bootstrapper {
	return wrk.bootstraper
}

func (wrk *worker) SetBootstraper(bootstraper process.Bootstrapper) {
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

func (wrk *worker) CheckSignature(cnsData *consensus.ConsensusMessage) error {
	return wrk.checkSignature(cnsData)
}

func (wrk *worker) ExecuteMessage(cnsDtaList []*consensus.ConsensusMessage) {
	wrk.executeMessage(cnsDtaList)
}

func GetSubroundName(subroundId int) string {
	return getSubroundName(subroundId)
}

func (wrk *worker) InitReceivedMessages() {
	wrk.initReceivedMessages()
}

func (wrk *worker) SendConsensusMessage(cnsDta *consensus.ConsensusMessage) bool {
	return wrk.sendConsensusMessage(cnsDta)
}

func (wrk *worker) Extend(subroundId int) {
	wrk.extend(subroundId)
}

func (wrk *worker) ReceivedSyncState(isNodeSynchronized bool) {
	wrk.receivedSyncState(isNodeSynchronized)
}

func (wrk *worker) ReceivedMessages() map[spos.MessageType][]*consensus.ConsensusMessage {
	wrk.mutReceivedMessages.RLock()
	defer wrk.mutReceivedMessages.RUnlock()

	return wrk.receivedMessages
}

func (wrk *worker) SetReceivedMessages(messageType spos.MessageType, cnsDta []*consensus.ConsensusMessage) {
	wrk.mutReceivedMessages.Lock()
	wrk.receivedMessages[messageType] = cnsDta
	wrk.mutReceivedMessages.Unlock()
}

func (wrk *worker) NilReceivedMessages() {
	wrk.mutReceivedMessages.Lock()
	wrk.receivedMessages = nil
	wrk.mutReceivedMessages.Unlock()
}

func (wrk *worker) ReceivedMessagesCalls() map[spos.MessageType]func(*consensus.ConsensusMessage) bool {
	wrk.mutReceivedMessagesCalls.RLock()
	defer wrk.mutReceivedMessagesCalls.RUnlock()

	return wrk.receivedMessagesCalls
}

func (wrk *worker) SetReceivedMessagesCalls(messageType spos.MessageType, f func(*consensus.ConsensusMessage) bool) {
	wrk.mutReceivedMessagesCalls.Lock()
	wrk.receivedMessagesCalls[messageType] = f
	wrk.mutReceivedMessagesCalls.Unlock()
}

func (wrk *worker) ExecuteMessageChannel() chan *consensus.ConsensusMessage {
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

func (sr *subroundStartRound) Bootstraper() process.Bootstrapper {
	return sr.consensusDataContainer.BootStrapper()
}

func (sr *subroundStartRound) ConsensusState() *spos.ConsensusState {
	return sr.consensusState
}

func (sr *subroundStartRound) Rounder() consensus.Rounder {
	return sr.consensusDataContainer.Rounder()
}

func (sr *subroundStartRound) SyncTimer() ntp.SyncTimer {
	return sr.consensusDataContainer.SyncTimer()
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

func (sr *subroundBlock) BlockChain() data.ChainHandler {
	return sr.Blockchain()
}

func (sr *subroundBlock) BlockProcessor() process.BlockProcessor {
	return sr.consensusDataContainer.BlockProcessor()
}

func (sr *subroundBlock) ConsensusState() *spos.ConsensusState {
	return sr.consensusState
}

func (sr *subroundBlock) Rounder() consensus.Rounder {
	return sr.consensusDataContainer.Rounder()
}

func (sr *subroundBlock) SyncTimer() ntp.SyncTimer {
	return sr.consensusDataContainer.SyncTimer()
}

func (sr *subroundBlock) DoBlockJob() bool {
	return sr.doBlockJob()
}

func (sr *subroundBlock) ReceivedBlockBody(cnsDta *consensus.ConsensusMessage) bool {
	return sr.receivedBlockBody(cnsDta)
}

func (sr *subroundBlock) DecodeBlockBody(dta []byte) block.Body {
	return sr.decodeBlockBody(dta)
}

func (sr *subroundBlock) ReceivedBlockHeader(cnsDta *consensus.ConsensusMessage) bool {
	return sr.receivedBlockHeader(cnsDta)
}

func (sr *subroundBlock) DecodeBlockHeader(dta []byte) *block.Header {
	return sr.decodeBlockHeader(dta)
}

func (sr *subroundBlock) ProcessReceivedBlock(cnsDta *consensus.ConsensusMessage) bool {
	return sr.processReceivedBlock(cnsDta)
}

func (sr *subroundBlock) DoBlockConsensusCheck() bool {
	return sr.doBlockConsensusCheck()
}

func (sr *subroundBlock) IsBlockReceived(threshold int) bool {
	return sr.isBlockReceived(threshold)
}

func (sr *subroundBlock) CreateHeader() (data.HeaderHandler, error) {
	return sr.createHeader()
}

// subroundCommitmentHash

type SubroundCommitmentHash *subroundCommitmentHash

func (sr *subroundCommitmentHash) ConsensusState() *spos.ConsensusState {
	return sr.consensusState
}

func (sr *subroundCommitmentHash) MultiSigner() crypto.MultiSigner {
	return sr.consensusDataContainer.MultiSigner()
}

func (sr *subroundCommitmentHash) Rounder() consensus.Rounder {
	return sr.consensusDataContainer.Rounder()
}

func (sr *subroundCommitmentHash) DoCommitmentHashJob() bool {
	return sr.doCommitmentHashJob()
}

func (sr *subroundCommitmentHash) ReceivedCommitmentHash(cnsDta *consensus.ConsensusMessage) bool {
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

func (sr *subroundBitmap) Rounder() consensus.Rounder {
	return sr.consensusDataContainer.Rounder()
}

func (sr *subroundBitmap) DoBitmapJob() bool {
	return sr.doBitmapJob()
}

func (sr *subroundBitmap) ReceivedBitmap(cnsDta *consensus.ConsensusMessage) bool {
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

func (sr *subroundCommitment) Rounder() consensus.Rounder {
	return sr.consensusDataContainer.Rounder()
}

func (sr *subroundCommitment) DoCommitmentJob() bool {
	return sr.doCommitmentJob()
}

func (sr *subroundCommitment) ReceivedCommitment(cnsDta *consensus.ConsensusMessage) bool {
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

func (sr *subroundSignature) MultiSigner() crypto.MultiSigner {
	return sr.consensusDataContainer.MultiSigner()
}

func (sr *subroundSignature) Rounder() consensus.Rounder {
	return sr.consensusDataContainer.Rounder()
}

func (sr *subroundSignature) DoSignatureJob() bool {
	return sr.doSignatureJob()
}

func (sr *subroundSignature) ReceivedSignature(cnsDta *consensus.ConsensusMessage) bool {
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
	return sr.consensusDataContainer.BlockProcessor()
}

func (sr *subroundEndRound) ConsensusState() *spos.ConsensusState {
	return sr.consensusState
}

func (sr *subroundEndRound) MultiSigner() crypto.MultiSigner {
	return sr.consensusDataContainer.MultiSigner()
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

func (sr *subroundEndRound) BroadcastBlock() func(data.BodyHandler, data.HeaderHandler) error {
	return sr.broadcastBlock
}

func (sr *subroundEndRound) SetBroadcastBlock(broadcastBlock func(data.BodyHandler, data.HeaderHandler) error) {
	sr.broadcastBlock = broadcastBlock
}

func (sr *subroundStartRound) InitCurrentRound() bool {
	return sr.initCurrentRound()
}

func GetStringValue(messageType spos.MessageType) string {
	return getStringValue(messageType)
}
