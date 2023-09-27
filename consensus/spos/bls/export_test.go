package bls

import (
	"context"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	cryptoCommon "github.com/multiversx/mx-chain-go/common/crypto"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/ntp"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

const ProcessingThresholdPercent = processingThresholdPercent
const DefaultMaxNumOfMessageTypeAccepted = defaultMaxNumOfMessageTypeAccepted
const MaxNumOfMessageTypeSignatureAccepted = maxNumOfMessageTypeSignatureAccepted

// factory

// Factory defines a type for the factory structure
type Factory *factory

// BlockChain gets the chain handler object
func (fct *factory) BlockChain() data.ChainHandler {
	return fct.consensusCore.Blockchain()
}

// BlockProcessor gets the block processor object
func (fct *factory) BlockProcessor() process.BlockProcessor {
	return fct.consensusCore.BlockProcessor()
}

// Bootstrapper gets the bootstrapper object
func (fct *factory) Bootstrapper() process.Bootstrapper {
	return fct.consensusCore.BootStrapper()
}

// ChronologyHandler gets the chronology handler object
func (fct *factory) ChronologyHandler() consensus.ChronologyHandler {
	return fct.consensusCore.Chronology()
}

// ConsensusState gets the consensus state struct pointer
func (fct *factory) ConsensusState() *spos.ConsensusState {
	return fct.consensusState
}

// Hasher gets the hasher object
func (fct *factory) Hasher() hashing.Hasher {
	return fct.consensusCore.Hasher()
}

// Marshalizer gets the marshalizer object
func (fct *factory) Marshalizer() marshal.Marshalizer {
	return fct.consensusCore.Marshalizer()
}

// MultiSigner gets the multi signer object
func (fct *factory) MultiSignerContainer() cryptoCommon.MultiSignerContainer {
	return fct.consensusCore.MultiSignerContainer()
}

// RoundHandler gets the roundHandler object
func (fct *factory) RoundHandler() consensus.RoundHandler {
	return fct.consensusCore.RoundHandler()
}

// ShardCoordinator gets the shard coordinator object
func (fct *factory) ShardCoordinator() sharding.Coordinator {
	return fct.consensusCore.ShardCoordinator()
}

// SyncTimer gets the sync timer object
func (fct *factory) SyncTimer() ntp.SyncTimer {
	return fct.consensusCore.SyncTimer()
}

// NodesCoordinator gets the nodes coordinator object
func (fct *factory) NodesCoordinator() nodesCoordinator.NodesCoordinator {
	return fct.consensusCore.NodesCoordinator()
}

// Worker gets the worker object
func (fct *factory) Worker() spos.WorkerHandler {
	return fct.worker
}

// SetWorker sets the worker object
func (fct *factory) SetWorker(worker spos.WorkerHandler) {
	fct.worker = worker
}

// GenerateStartRoundSubround generates the instance of subround StartRound and added it to the chronology subrounds list
func (fct *factory) GenerateStartRoundSubround() error {
	return fct.generateStartRoundSubround()
}

// GenerateBlockSubround generates the instance of subround Block and added it to the chronology subrounds list
func (fct *factory) GenerateBlockSubround() error {
	return fct.generateBlockSubround()
}

// GenerateSignatureSubround generates the instance of subround Signature and added it to the chronology subrounds list
func (fct *factory) GenerateSignatureSubround() error {
	return fct.generateSignatureSubround()
}

// GenerateEndRoundSubround generates the instance of subround EndRound and added it to the chronology subrounds list
func (fct *factory) GenerateEndRoundSubround() error {
	return fct.generateEndRoundSubround()
}

// AppStatusHandler gets the app status handler object
func (fct *factory) AppStatusHandler() core.AppStatusHandler {
	return fct.appStatusHandler
}

// Outport gets the outport object
func (fct *factory) Outport() outport.OutportHandler {
	return fct.outportHandler
}

// subroundStartRound

// SubroundStartRound defines a type for the subroundStartRound structure
type SubroundStartRound *subroundStartRound

// DoStartRoundJob method does the job of the subround StartRound
func (sr *subroundStartRound) DoStartRoundJob() bool {
	return sr.doStartRoundJob(context.Background())
}

// DoStartRoundConsensusCheck method checks if the consensus is achieved in the subround StartRound
func (sr *subroundStartRound) DoStartRoundConsensusCheck() bool {
	return sr.doStartRoundConsensusCheck()
}

// GenerateNextConsensusGroup generates the next consensu group based on current (random seed, shard id and round)
func (sr *subroundStartRound) GenerateNextConsensusGroup(roundIndex int64) error {
	return sr.generateNextConsensusGroup(roundIndex)
}

// InitCurrentRound inits all the stuff needed in the current round
func (sr *subroundStartRound) InitCurrentRound() bool {
	return sr.initCurrentRound()
}

// subroundBlock

// SubroundBlock defines a type for the subroundBlock structure
type SubroundBlock *subroundBlock

// Blockchain gets the ChainHandler stored in the ConsensusCore
func (sr *subroundBlock) BlockChain() data.ChainHandler {
	return sr.Blockchain()
}

// DoBlockJob method does the job of the subround Block
func (sr *subroundBlock) DoBlockJob() bool {
	return sr.doBlockJob(context.Background())
}

// ProcessReceivedBlock method processes the received proposed block in the subround Block
func (sr *subroundBlock) ProcessReceivedBlock(cnsDta *consensus.Message) bool {
	return sr.processReceivedBlock(context.Background(), cnsDta)
}

// DoBlockConsensusCheck method checks if the consensus in the subround Block is achieved
func (sr *subroundBlock) DoBlockConsensusCheck() bool {
	return sr.doBlockConsensusCheck()
}

// IsBlockReceived method checks if the block was received from the leader in the current round
func (sr *subroundBlock) IsBlockReceived(threshold int) bool {
	return sr.isBlockReceived(threshold)
}

// CreateHeader method creates the proposed block header in the subround Block
func (sr *subroundBlock) CreateHeader() (data.HeaderHandler, error) {
	return sr.createHeader()
}

// CreateBody method creates the proposed block body in the subround Block
func (sr *subroundBlock) CreateBlock(hdr data.HeaderHandler) (data.HeaderHandler, data.BodyHandler, error) {
	return sr.createBlock(hdr)
}

// SendBlockBody method sends the proposed block body in the subround Block
func (sr *subroundBlock) SendBlockBody(body data.BodyHandler, marshalizedBody []byte) bool {
	return sr.sendBlockBody(body, marshalizedBody)
}

// SendBlockHeader method sends the proposed block header in the subround Block
func (sr *subroundBlock) SendBlockHeader(header data.HeaderHandler, marshalizedHeader []byte) bool {
	return sr.sendBlockHeader(header, marshalizedHeader)
}

// ComputeSubroundProcessingMetric computes processing metric related to the subround Block
func (sr *subroundBlock) ComputeSubroundProcessingMetric(startTime time.Time, metric string) {
	sr.computeSubroundProcessingMetric(startTime, metric)
}

// ReceivedBlockBody method is called when a block body is received through the block body channel
func (sr *subroundBlock) ReceivedBlockBody(cnsDta *consensus.Message) bool {
	return sr.receivedBlockBody(context.Background(), cnsDta)
}

// ReceivedBlockHeader method is called when a block header is received through the block header channel
func (sr *subroundBlock) ReceivedBlockHeader(cnsDta *consensus.Message) bool {
	return sr.receivedBlockHeader(context.Background(), cnsDta)
}

// ReceivedBlockBodyAndHeader is called when both a header and block body have been received
func (sr *subroundBlock) ReceivedBlockBodyAndHeader(cnsDta *consensus.Message) bool {
	return sr.receivedBlockBodyAndHeader(context.Background(), cnsDta)
}

// subroundSignature

// SubroundSignature defines a type for the subroundSignature structure
type SubroundSignature *subroundSignature

// DoSignatureJob method does the job of the subround Signature
func (sr *subroundSignature) DoSignatureJob() bool {
	return sr.doSignatureJob(context.Background())
}

// ReceivedSignature method is called when a signature is received through the signature channel
func (sr *subroundSignature) ReceivedSignature(cnsDta *consensus.Message) bool {
	return sr.receivedSignature(context.Background(), cnsDta)
}

// DoSignatureConsensusCheck method checks if the consensus in the subround Signature is achieved
func (sr *subroundSignature) DoSignatureConsensusCheck() bool {
	return sr.doSignatureConsensusCheck()
}

// AreSignaturesCollected method checks if the number of signatures received from the nodes are more than the given threshold
func (sr *subroundSignature) AreSignaturesCollected(threshold int) (bool, int) {
	return sr.areSignaturesCollected(threshold)
}

// subroundEndRound

// SubroundEndRound defines a type for the subroundEndRound structure
type SubroundEndRound *subroundEndRound

// DoEndRoundJob method does the job of the subround EndRound
func (sr *subroundEndRound) DoEndRoundJob() bool {
	return sr.doEndRoundJob(context.Background())
}

// DoEndRoundConsensusCheck method checks if the consensus is achieved
func (sr *subroundEndRound) DoEndRoundConsensusCheck() bool {
	return sr.doEndRoundConsensusCheck()
}

// CheckSignaturesValidity method checks the signature validity for the nodes included in bitmap
func (sr *subroundEndRound) CheckSignaturesValidity(bitmap []byte) error {
	return sr.checkSignaturesValidity(bitmap)
}

// DoEndRoundJobByParticipant calls the unexported doEndRoundJobByParticipant function
func (sr *subroundEndRound) DoEndRoundJobByParticipant(cnsDta *consensus.Message) bool {
	return sr.doEndRoundJobByParticipant(cnsDta)
}

// DoEndRoundJobByLeader calls the unexported doEndRoundJobByLeader function
func (sr *subroundEndRound) DoEndRoundJobByLeader() bool {
	return sr.doEndRoundJobByLeader()
}

// HaveConsensusHeaderWithFullInfo calls the unexported haveConsensusHeaderWithFullInfo function
func (sr *subroundEndRound) HaveConsensusHeaderWithFullInfo(cnsDta *consensus.Message) (bool, data.HeaderHandler) {
	return sr.haveConsensusHeaderWithFullInfo(cnsDta)
}

// CreateAndBroadcastHeaderFinalInfo calls the unexported createAndBroadcastHeaderFinalInfo function
func (sr *subroundEndRound) CreateAndBroadcastHeaderFinalInfo() {
	sr.createAndBroadcastHeaderFinalInfo()
}

// ReceivedBlockHeaderFinalInfo calls the unexported receivedBlockHeaderFinalInfo function
func (sr *subroundEndRound) ReceivedBlockHeaderFinalInfo(cnsDta *consensus.Message) bool {
	return sr.receivedBlockHeaderFinalInfo(context.Background(), cnsDta)
}

// IsBlockHeaderFinalInfoValid calls the unexported isBlockHeaderFinalInfoValid function
func (sr *subroundEndRound) IsBlockHeaderFinalInfoValid(cnsDta *consensus.Message) bool {
	return sr.isBlockHeaderFinalInfoValid(cnsDta)
}

// IsConsensusHeaderReceived calls the unexported isConsensusHeaderReceived function
func (sr *subroundEndRound) IsConsensusHeaderReceived() (bool, data.HeaderHandler) {
	return sr.isConsensusHeaderReceived()
}

// IsOutOfTime calls the unexported isOutOfTime function
func (sr *subroundEndRound) IsOutOfTime() bool {
	return sr.isOutOfTime()
}

// VerifyNodesOnAggSigFail calls the unexported verifyNodesOnAggSigFail function
func (sr *subroundEndRound) VerifyNodesOnAggSigFail() ([]string, error) {
	return sr.verifyNodesOnAggSigFail()
}

// ComputeAggSigOnValidNodes calls the unexported computeAggSigOnValidNodes function
func (sr *subroundEndRound) ComputeAggSigOnValidNodes() ([]byte, []byte, error) {
	return sr.computeAggSigOnValidNodes()
}

// ReceivedInvalidSignersInfo calls the unexported receivedInvalidSignersInfo function
func (sr *subroundEndRound) ReceivedInvalidSignersInfo(cnsDta *consensus.Message) bool {
	return sr.receivedInvalidSignersInfo(context.Background(), cnsDta)
}

// VerifyInvalidSigners calls the unexported verifyInvalidSigners function
func (sr *subroundEndRound) VerifyInvalidSigners(invalidSigners []byte) error {
	return sr.verifyInvalidSigners(invalidSigners)
}

// GetMinConsensusGroupIndexOfManagedKeys calls the unexported getMinConsensusGroupIndexOfManagedKeys function
func (sr *subroundEndRound) GetMinConsensusGroupIndexOfManagedKeys() int {
	return sr.getMinConsensusGroupIndexOfManagedKeys()
}

// GetStringValue calls the unexported getStringValue function
func GetStringValue(messageType consensus.MessageType) string {
	return getStringValue(messageType)
}

// CreateAndBroadcastInvalidSigners calls the unexported createAndBroadcastInvalidSigners function
func (sr *subroundEndRound) CreateAndBroadcastInvalidSigners(invalidSigners []byte) {
	sr.createAndBroadcastInvalidSigners(invalidSigners)
}

// GetFullMessagesForInvalidSigners calls the unexported getFullMessagesForInvalidSigners function
func (sr *subroundEndRound) GetFullMessagesForInvalidSigners(invalidPubKeys []string) ([]byte, error) {
	return sr.getFullMessagesForInvalidSigners(invalidPubKeys)
}

// GetSentSignatureTracker returns the subroundEndRound's SentSignaturesTracker instance
func (sr *subroundEndRound) GetSentSignatureTracker() spos.SentSignaturesTracker {
	return sr.sentSignatureTracker
}

// GetSentSignatureTracker returns the subroundStartRound's SentSignaturesTracker instance
func (sr *subroundStartRound) GetSentSignatureTracker() spos.SentSignaturesTracker {
	return sr.sentSignatureTracker
}
