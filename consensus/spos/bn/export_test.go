package bn

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/validators/groupSelectors"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// Factory

func (fct *Factory) BlockChain() *blockchain.BlockChain {
	return fct.blockChain
}

func (fct *Factory) SetBlockChain(blockChain *blockchain.BlockChain) {
	fct.blockChain = blockChain
}

func (fct *Factory) BlockProcessor() process.BlockProcessor {
	return fct.blockProcessor
}

func (fct *Factory) SetBlockProcessor(blockProcessor process.BlockProcessor) {
	fct.blockProcessor = blockProcessor
}

func (fct *Factory) Bootstraper() process.Bootstraper {
	return fct.bootstraper
}

func (fct *Factory) SetBootsraper(bootstraper process.Bootstraper) {
	fct.bootstraper = bootstraper
}

func (fct *Factory) ChronologyHandler() chronology.ChronologyHandler {
	return fct.chronologyHandler
}

func (fct *Factory) SetChronologyHandler(chronologyHandler chronology.ChronologyHandler) {
	fct.chronologyHandler = chronologyHandler
}

func (fct *Factory) ConsensusState() *spos.ConsensusState {
	return fct.consensusState
}

func (fct *Factory) SetConsensusState(consensusState *spos.ConsensusState) {
	fct.consensusState = consensusState
}

func (fct *Factory) Hasher() hashing.Hasher {
	return fct.hasher
}

func (fct *Factory) SetHasher(hasher hashing.Hasher) {
	fct.hasher = hasher
}

func (fct *Factory) Marshalizer() marshal.Marshalizer {
	return fct.marshalizer
}

func (fct *Factory) SetMarshalizer(marshalizer marshal.Marshalizer) {
	fct.marshalizer = marshalizer
}

func (fct *Factory) MultiSigner() crypto.MultiSigner {
	return fct.multiSigner
}

func (fct *Factory) SetMultiSigner(multiSigner crypto.MultiSigner) {
	fct.multiSigner = multiSigner
}

func (fct *Factory) Rounder() round.Rounder {
	return fct.rounder
}

func (fct *Factory) SetRounder(rounder round.Rounder) {
	fct.rounder = rounder
}

func (fct *Factory) ShardCoordinator() sharding.ShardCoordinator {
	return fct.shardCoordinator
}

func (fct *Factory) SetShardCoordinator(shardCoordinator sharding.ShardCoordinator) {
	fct.shardCoordinator = shardCoordinator
}

func (fct *Factory) SyncTimer() ntp.SyncTimer {
	return fct.syncTimer
}

func (fct *Factory) SetSyncTimer(syncTimer ntp.SyncTimer) {
	fct.syncTimer = syncTimer
}

func (fct *Factory) ValidatorGroupSelector() groupSelectors.ValidatorGroupSelector {
	return fct.validatorGroupSelector
}

func (fct *Factory) SetValidatorGroupSelector(validatorGroupSelector groupSelectors.ValidatorGroupSelector) {
	fct.validatorGroupSelector = validatorGroupSelector
}

func (fct *Factory) Worker() *Worker {
	return fct.worker
}

func (fct *Factory) SetWorker(worker *Worker) {
	fct.worker = worker
}

func (fct *Factory) GenerateStartRoundSubround() bool {
	return fct.generateStartRoundSubround()
}

func (fct *Factory) GenerateBlockSubround() bool {
	return fct.generateBlockSubround()
}

func (fct *Factory) GenerateCommitmentHashSubround() bool {
	return fct.generateCommitmentHashSubround()
}

func (fct *Factory) GenerateBitmapSubround() bool {
	return fct.generateBitmapSubround()
}

func (fct *Factory) GenerateCommitmentSubround() bool {
	return fct.generateCommitmentSubround()
}

func (fct *Factory) GenerateSignatureSubround() bool {
	return fct.generateSignatureSubround()
}

func (fct *Factory) GenerateEndRoundSubround() bool {
	return fct.generateEndRoundSubround()
}

// subround

func (sr *subround) SetJobFunction(job func() bool) {
	sr.job = job
}

func (sr *subround) SetCheckFunction(check func() bool) {
	sr.check = check
}

// Worker

func (wrk *Worker) Bootstraper() process.Bootstraper {
	return wrk.bootstraper
}

func (wrk *Worker) SetBootstraper(bootstraper process.Bootstraper) {
	wrk.bootstraper = bootstraper
}

func (wrk *Worker) ConsensusState() *spos.ConsensusState {
	return wrk.consensusState
}

func (wrk *Worker) SetConsensusState(consensusState *spos.ConsensusState) {
	wrk.consensusState = consensusState
}

func (wrk *Worker) KeyGenerator() crypto.KeyGenerator {
	return wrk.keyGenerator
}

func (wrk *Worker) SetKeyGenerator(keyGenerator crypto.KeyGenerator) {
	wrk.keyGenerator = keyGenerator
}

func (wrk *Worker) Marshalizer() marshal.Marshalizer {
	return wrk.marshalizer
}

func (wrk *Worker) SetMarshalizer(marshalizer marshal.Marshalizer) {
	wrk.marshalizer = marshalizer
}

func (wrk *Worker) Rounder() round.Rounder {
	return wrk.rounder
}

func (wrk *Worker) SetRounder(rounder round.Rounder) {
	wrk.rounder = rounder
}

func (wrk *Worker) CheckSignature(cnsData *spos.ConsensusData) error {
	return wrk.checkSignature(cnsData)
}

func (wrk *Worker) ExecuteMessage(cnsDtaList []*spos.ConsensusData) {
	wrk.executeMessage(cnsDtaList)
}

func GetMessageTypeName(messageType MessageType) string {
	return getMessageTypeName(messageType)
}

func GetSubroundName(subroundId int) string {
	return getSubroundName(subroundId)
}

func (wrk *Worker) InitReceivedMessages() {
	wrk.initReceivedMessages()
}

func (wrk *Worker) SendConsensusMessage(cnsDta *spos.ConsensusData) bool {
	return wrk.sendConsensusMessage(cnsDta)
}

func (wrk *Worker) BroadcastTxBlockBody2(blockBody *block.TxBlockBody) error {
	return wrk.broadcastTxBlockBody(blockBody)
}

func (wrk *Worker) BroadcastHeader2(header *block.Header) error {
	return wrk.broadcastHeader(header)
}

func (wrk *Worker) Extend(subroundId int) {
	wrk.extend(subroundId)
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

func (sr *subroundBlock) Rounder() round.Rounder {
	return sr.rounder
}

func (sr *subroundBlock) SetRounder(rounder round.Rounder) {
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

func (sr *subroundBlock) ReceivedBlockBody(cnsDta *spos.ConsensusData) bool {
	return sr.receivedBlockBody(cnsDta)
}

func (sr *subroundBlock) DecodeBlockBody(dta []byte) *block.TxBlockBody {
	return sr.decodeBlockBody(dta)
}

func (sr *subroundBlock) ReceivedBlockHeader(cnsDta *spos.ConsensusData) bool {
	return sr.receivedBlockHeader(cnsDta)
}

func (sr *subroundBlock) DecodeBlockHeader(dta []byte) *block.Header {
	return sr.decodeBlockHeader(dta)
}

func (sr *subroundBlock) ProcessReceivedBlock(cnsDta *spos.ConsensusData) bool {
	return sr.processReceivedBlock(cnsDta)
}

func (sr *subroundBlock) DoBlockConsensusCheck() bool {
	return sr.doBlockConsensusCheck()
}

func (sr *subroundBlock) IsBlockReceived(threshold int) bool {
	return sr.isBlockReceived(threshold)
}

func (sr *subroundBlock) CheckIfBlockIsValid(receivedHeader *block.Header) bool {
	return sr.checkIfBlockIsValid(receivedHeader)
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

func (sr *subroundCommitmentHash) Rounder() round.Rounder {
	return sr.rounder
}

func (sr *subroundCommitmentHash) SetRounder(rounder round.Rounder) {
	sr.rounder = rounder
}

func (sr *subroundCommitmentHash) DoCommitmentHashJob() bool {
	return sr.doCommitmentHashJob()
}

func (sr *subroundCommitmentHash) ReceivedCommitmentHash(cnsDta *spos.ConsensusData) bool {
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

func (sr *subroundBitmap) Rounder() round.Rounder {
	return sr.rounder
}

func (sr *subroundBitmap) SetRounder(rounder round.Rounder) {
	sr.rounder = rounder
}

func (sr *subroundBitmap) DoBitmapJob() bool {
	return sr.doBitmapJob()
}

func (sr *subroundBitmap) ReceivedBitmap(cnsDta *spos.ConsensusData) bool {
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

func (sr *subroundCommitment) Rounder() round.Rounder {
	return sr.rounder
}

func (sr *subroundCommitment) SetRounder(rounder round.Rounder) {
	sr.rounder = rounder
}

func (sr *subroundCommitment) DoCommitmentJob() bool {
	return sr.doCommitmentJob()
}

func (sr *subroundCommitment) ReceivedCommitment(cnsDta *spos.ConsensusData) bool {
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

func (sr *subroundSignature) Rounder() round.Rounder {
	return sr.rounder
}

func (sr *subroundSignature) SetRounder(rounder round.Rounder) {
	sr.rounder = rounder
}

func (sr *subroundSignature) DoSignatureJob() bool {
	return sr.doSignatureJob()
}

func (sr *subroundSignature) ReceivedSignature(cnsDta *spos.ConsensusData) bool {
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
