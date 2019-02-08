package bn

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

func (wrk *Worker) CheckSignaturesValidity(bitmap []byte) error {
	return wrk.checkSignaturesValidity(bitmap)
}

func (wrk *Worker) GenCommitmentHash() ([]byte, error) {
	return wrk.genCommitmentHash()
}

func (wrk *Worker) CheckCommitmentsValidity(bitmap []byte) error {
	return wrk.checkCommitmentsValidity(bitmap)
}

func (wrk *Worker) ShouldDropConsensusMessage(cnsDta *spos.ConsensusData) bool {
	return wrk.shouldDropConsensusMessage(cnsDta)
}

func (wrk *Worker) CheckSignature(cnsData *spos.ConsensusData) error {
	return wrk.checkSignature(cnsData)
}

func (wrk *Worker) ProcessReceivedBlock(cnsDta *spos.ConsensusData) bool {
	return wrk.processReceivedBlock(cnsDta)
}

func (wrk *Worker) InitReceivedMessages() {
	wrk.initReceivedMessages()
}

func (wrk *Worker) InitMessageChannels() {
	wrk.initMessageChannels()
}

func (wrk *Worker) CleanReceivedMessages() {
	wrk.cleanReceivedMessages()
}

func (wrk *Worker) ExecuteMessage(cnsDtaList []*spos.ConsensusData) {
	wrk.executeMessage(cnsDtaList)
}

func GetMessageTypeName(messageType MessageType) string {
	return getMessageTypeName(messageType)
}

func GetSubroundName(subroundId chronology.SubroundId) string {
	return getSubroundName(subroundId)
}

func (wrk *Worker) SendConsensusMessage(cnsDta *spos.ConsensusData) bool {
	return wrk.sendConsensusMessage(cnsDta)
}

func (wrk *Worker) DoAdvanceJob() bool {
	return wrk.doAdvanceJob()
}

func (wrk *Worker) DoBitmapJob() bool {
	return wrk.doBitmapJob()
}

func (wrk *Worker) ReceivedBitmap(cnsDta *spos.ConsensusData) bool {
	return wrk.receivedBitmap(cnsDta)
}

func (wrk *Worker) IsValidatorInBitmap(validator string) bool {
	return wrk.isValidatorInBitmap(validator)
}

func (wrk *Worker) IsSelfInBitmap() bool {
	return wrk.isSelfInBitmap()
}

func (wrk *Worker) CheckBitmapConsensus() bool {
	return wrk.checkBitmapConsensus()
}

func (wrk *Worker) IsBitmapReceived(threshold int) bool {
	return wrk.isBitmapReceived(threshold)
}

func (wrk *Worker) ExtendBitmap() {
	wrk.extendBitmap()
}

func (wrk *Worker) DoBlockJob() bool {
	return wrk.doBlockJob()
}

func (wrk *Worker) SendBlockBody() bool {
	return wrk.sendBlockBody()
}

func (wrk *Worker) SendBlockHeader() bool {
	return wrk.sendBlockHeader()
}

func (wrk *Worker) ReceivedBlockBody(cnsDta *spos.ConsensusData) bool {
	return wrk.receivedBlockBody(cnsDta)
}

func (wrk *Worker) DecodeBlockBody(dta []byte) *block.TxBlockBody {
	return wrk.decodeBlockBody(dta)
}

func (wrk *Worker) ReceivedBlockHeader(cnsDta *spos.ConsensusData) bool {
	return wrk.receivedBlockHeader(cnsDta)
}

func (wrk *Worker) DecodeBlockHeader(dta []byte) *block.Header {
	return wrk.decodeBlockHeader(dta)
}

func (wrk *Worker) CheckBlockConsensus() bool {
	return wrk.checkBlockConsensus()
}

func (wrk *Worker) IsBlockReceived(threshold int) bool {
	return wrk.isBlockReceived(threshold)
}

func (wrk *Worker) ExtendBlock() {
	wrk.extendBlock()
}

func (wrk *Worker) CheckIfBlockIsValid(receivedHeader *block.Header) bool {
	return wrk.checkIfBlockIsValid(receivedHeader)
}

func (wrk *Worker) PrintBlockCM() {
	wrk.printBlockCM()
}

func (wrk *Worker) DoCommitmentJob() bool {
	return wrk.doCommitmentJob()
}

func (wrk *Worker) ReceivedCommitment(cnsDta *spos.ConsensusData) bool {
	return wrk.receivedCommitment(cnsDta)
}

func (wrk *Worker) CheckCommitmentConsensus() bool {
	return wrk.checkCommitmentConsensus()
}

func (wrk *Worker) CommitmentsCollected(threshold int) bool {
	return wrk.commitmentsCollected(threshold)
}

func (wrk *Worker) ExtendCommitment() {
	wrk.extendCommitment()
}

func (wrk *Worker) PrintCommitmentCM() {
	wrk.printCommitmentCM()
}

func (wrk *Worker) DoCommitmentHashJob() bool {
	return wrk.doCommitmentHashJob()
}

func (wrk *Worker) ReceivedCommitmentHash(cnsDta *spos.ConsensusData) bool {
	return wrk.receivedCommitmentHash(cnsDta)
}

func (wrk *Worker) CheckCommitmentHashConsensus() bool {
	return wrk.checkCommitmentHashConsensus()
}

func (wrk *Worker) IsCommitmentHashReceived(threshold int) bool {
	return wrk.isCommitmentHashReceived(threshold)
}

func (wrk *Worker) CommitmentHashesCollected(threshold int) bool {
	return wrk.commitmentHashesCollected(threshold)
}

func (wrk *Worker) ExtendCommitmentHash() {
	wrk.extendCommitmentHash()
}

func (wrk *Worker) PrintCommitmentHashCM() {
	wrk.printCommitmentHashCM()
}

func (wrk *Worker) DoEndRoundJob() bool {
	return wrk.doEndRoundJob()
}

func (wrk *Worker) CheckEndRoundConsensus() bool {
	return wrk.checkEndRoundConsensus()
}

func (wrk *Worker) ExtendEndRound() {
	wrk.extendEndRound()
}

func (wrk *Worker) DoSignatureJob() bool {
	return wrk.doSignatureJob()
}

func (wrk *Worker) ReceivedSignature(cnsDta *spos.ConsensusData) bool {
	return wrk.receivedSignature(cnsDta)
}

func (wrk *Worker) CheckSignatureConsensus() bool {
	return wrk.checkSignatureConsensus()
}

func (wrk *Worker) SignaturesCollected(threshold int) bool {
	return wrk.signaturesCollected(threshold)
}

func (wrk *Worker) ExtendSignature() {
	wrk.extendSignature()
}

func (wrk *Worker) PrintSignatureCM() {
	wrk.printSignatureCM()
}

func (wrk *Worker) DoStartRoundJob() bool {
	return wrk.doStartRoundJob()
}

func (wrk *Worker) CheckStartRoundConsensus() bool {
	return wrk.checkStartRoundConsensus()
}

func (wrk *Worker) ExtendStartRound() {
	wrk.extendStartRound()
}

func (wrk *Worker) SetKeyGen(keyGen crypto.KeyGenerator) {
	wrk.keyGen = keyGen
}

func (wrk *Worker) SetBlockProcessor(blockProcessor process.BlockProcessor) {
	wrk.blockProcessor = blockProcessor
}

func (wrk *Worker) SetBootstraper(boot process.Bootstraper) {
	wrk.boot = boot
}

func (wrk *Worker) MultiSigner() crypto.MultiSigner {
	return wrk.multiSigner
}

func (wrk *Worker) SetMultiSigner(multiSigner crypto.MultiSigner) {
	wrk.multiSigner = multiSigner
}

func (wrk *Worker) IsConsensusDataNotSet() bool {
	return wrk.isConsensusDataNotSet()
}

func (wrk *Worker) IsConsensusDataAlreadySet() bool {
	return wrk.isConsensusDataAlreadySet()
}

func (wrk *Worker) IsSelfJobDone(currentRound chronology.SubroundId) bool {
	return wrk.isSelfJobDone(currentRound)
}

func (wrk *Worker) IsJobDone(node string, currentRound chronology.SubroundId) bool {
	return wrk.isJobDone(node, currentRound)
}

func (wrk *Worker) IsCurrentRoundFinished(currentRound chronology.SubroundId) bool {
	return wrk.isCurrentRoundFinished(currentRound)
}

func (wrk *Worker) IsMessageReceivedFromItself(node string) bool {
	return wrk.isMessageReceivedFromItself(node)
}

func (wrk *Worker) IsMessageReceivedTooLate() bool {
	return wrk.isMessageReceivedTooLate()
}

func (wrk *Worker) IsMessageReceivedForOtherRound(roundIndex int32) bool {
	return wrk.isMessageReceivedForOtherRound(roundIndex)
}

func (wrk *Worker) IsBlockBodyAlreadyReceived() bool {
	return wrk.isBlockBodyAlreadyReceived()
}

func (wrk *Worker) IsHeaderAlreadyReceived() bool {
	return wrk.isHeaderAlreadyReceived()
}

func (wrk *Worker) CanDoSubroundJob(currentRound chronology.SubroundId) bool {
	return wrk.canDoSubroundJob(currentRound)
}

func (wrk *Worker) CanReceiveMessage(node string, roundIndex int32, currentSubround chronology.SubroundId) bool {
	return wrk.canReceiveMessage(node, roundIndex, currentSubround)
}

func (wrk *Worker) IsBlockSubroundUnfinished() bool {
	return wrk.isBlockSubroundUnfinished()
}

func (wrk *Worker) IsCommitmentHashSubroundUnfinished() bool {
	return wrk.isCommitmentHashSubroundUnfinished()
}

func (wrk *Worker) IsBitmapSubroundUnfinished() bool {
	return wrk.isBitmapSubroundUnfinished()
}

func (wrk *Worker) IsCommitmentSubroundUnfinished() bool {
	return wrk.isCommitmentSubroundUnfinished()
}
