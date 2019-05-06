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
	return fct.consensusCore.Blockchain()
}

func (fct *factory) BlockProcessor() process.BlockProcessor {
	return fct.consensusCore.BlockProcessor()
}

func (fct *factory) Bootstraper() process.Bootstrapper {
	return fct.consensusCore.BootStrapper()
}

func (fct *factory) ChronologyHandler() consensus.ChronologyHandler {
	return fct.consensusCore.Chronology()
}

func (fct *factory) ConsensusState() *spos.ConsensusState {
	return fct.consensusState
}

func (fct *factory) Hasher() hashing.Hasher {
	return fct.consensusCore.Hasher()
}

func (fct *factory) Marshalizer() marshal.Marshalizer {
	return fct.consensusCore.Marshalizer()
}

func (fct *factory) MultiSigner() crypto.MultiSigner {
	return fct.consensusCore.MultiSigner()
}

func (fct *factory) Rounder() consensus.Rounder {
	return fct.consensusCore.Rounder()
}

func (fct *factory) ShardCoordinator() sharding.Coordinator {
	return fct.consensusCore.ShardCoordinator()
}

func (fct *factory) SyncTimer() ntp.SyncTimer {
	return fct.consensusCore.SyncTimer()
}

func (fct *factory) ValidatorGroupSelector() consensus.ValidatorGroupSelector {
	return fct.consensusCore.ValidatorGroupSelector()
}

func (fct *factory) Worker() spos.WorkerHandler {
	return fct.worker
}

func (fct *factory) SetWorker(worker *spos.Worker) {
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

//worker

type BNConsensusService *worker

// subround

func (sr *subround) SetJobFunction(job func() bool) {
	sr.job = job
}

func (sr *subround) SetCheckFunction(check func() bool) {
	sr.check = check
}

// subroundStartRound

type SubroundStartRound *subroundStartRound

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

func (sr *subroundBlock) DoBlockJob() bool {
	return sr.doBlockJob()
}

func (sr *subroundBlock) ReceivedBlockBody(cnsDta *consensus.Message) bool {
	return sr.receivedBlockBody(cnsDta)
}

func (sr *subroundBlock) DecodeBlockBody(dta []byte) block.Body {
	return sr.decodeBlockBody(dta)
}

func (sr *subroundBlock) ReceivedBlockHeader(cnsDta *consensus.Message) bool {
	return sr.receivedBlockHeader(cnsDta)
}

func (sr *subroundBlock) DecodeBlockHeader(dta []byte) *block.Header {
	return sr.decodeBlockHeader(dta)
}

func (sr *subroundBlock) ProcessReceivedBlock(cnsDta *consensus.Message) bool {
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

func (sr *subroundCommitmentHash) DoCommitmentHashJob() bool {
	return sr.doCommitmentHashJob()
}

func (sr *subroundCommitmentHash) ReceivedCommitmentHash(cnsDta *consensus.Message) bool {
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

func (sr *subroundBitmap) DoBitmapJob() bool {
	return sr.doBitmapJob()
}

func (sr *subroundBitmap) ReceivedBitmap(cnsDta *consensus.Message) bool {
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

func (sr *subroundCommitment) DoCommitmentJob() bool {
	return sr.doCommitmentJob()
}

func (sr *subroundCommitment) ReceivedCommitment(cnsDta *consensus.Message) bool {
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

func (sr *subroundSignature) DoSignatureJob() bool {
	return sr.doSignatureJob()
}

func (sr *subroundSignature) ReceivedSignature(cnsDta *consensus.Message) bool {
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

func GetStringValue(messageType consensus.MessageType) string {
	return getStringValue(messageType)
}
