package v2

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
func (fct *factory) ConsensusState() spos.ConsensusStateHandler {
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

// SubroundStartRound defines an alias for the subroundStartRound structure
type SubroundStartRound = *subroundStartRound

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

// GetSentSignatureTracker returns the subroundStartRound's SentSignaturesTracker instance
func (sr *subroundStartRound) GetSentSignatureTracker() spos.SentSignaturesTracker {
	return sr.sentSignatureTracker
}

// subroundBlock

// SubroundBlock defines an alias for the subroundBlock structure
type SubroundBlock = *subroundBlock

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
	return sr.processReceivedBlock(context.Background(), cnsDta.RoundIndex, cnsDta.PubKey)
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
func (sr *subroundBlock) ReceivedBlockHeader(header data.HeaderHandler) {
	sr.receivedBlockHeader(header)
}

// GetLeaderForHeader returns the leader based on header info
func (sr *subroundBlock) GetLeaderForHeader(headerHandler data.HeaderHandler) ([]byte, error) {
	return sr.getLeaderForHeader(headerHandler)
}

// subroundSignature

// SubroundSignature defines an alias to the subroundSignature structure
type SubroundSignature = *subroundSignature

// DoSignatureJob method does the job of the subround Signature
func (sr *subroundSignature) DoSignatureJob() bool {
	return sr.doSignatureJob(context.Background())
}

// DoSignatureConsensusCheck method checks if the consensus in the subround Signature is achieved
func (sr *subroundSignature) DoSignatureConsensusCheck() bool {
	return sr.doSignatureConsensusCheck()
}

// subroundEndRound

// SubroundEndRound defines a type for the subroundEndRound structure
type SubroundEndRound = *subroundEndRound

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

// DoEndRoundJobByLeader calls the unexported doEndRoundJobByNode function
func (sr *subroundEndRound) DoEndRoundJobByNode() bool {
	return sr.doEndRoundJobByNode()
}

// CreateAndBroadcastProof calls the unexported createAndBroadcastHeaderFinalInfo function
func (sr *subroundEndRound) CreateAndBroadcastProof(signature []byte, bitmap []byte) {
	_ = sr.createAndBroadcastProof(signature, bitmap, "sender")
}

// ReceivedProof calls the unexported receivedProof function
func (sr *subroundEndRound) ReceivedProof(proof consensus.ProofHandler) {
	sr.receivedProof(proof)
}

// IsOutOfTime calls the unexported isOutOfTime function
func (sr *subroundEndRound) IsOutOfTime() bool {
	return sr.isOutOfTime()
}

// VerifyNodesOnAggSigFail calls the unexported verifyNodesOnAggSigFail function
func (sr *subroundEndRound) VerifyNodesOnAggSigFail(ctx context.Context) ([]string, error) {
	return sr.verifyNodesOnAggSigFail(ctx)
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
func (sr *subroundEndRound) VerifyInvalidSigners(invalidSigners []byte) ([]string, error) {
	return sr.verifyInvalidSigners(invalidSigners)
}

// GetMinConsensusGroupIndexOfManagedKeys calls the unexported getMinConsensusGroupIndexOfManagedKeys function
func (sr *subroundEndRound) GetMinConsensusGroupIndexOfManagedKeys() int {
	return sr.getMinConsensusGroupIndexOfManagedKeys()
}

// CreateAndBroadcastInvalidSigners calls the unexported createAndBroadcastInvalidSigners function
func (sr *subroundEndRound) CreateAndBroadcastInvalidSigners(invalidSigners []byte) {
	sr.createAndBroadcastInvalidSigners(invalidSigners, nil, "sender")
}

// GetFullMessagesForInvalidSigners calls the unexported getFullMessagesForInvalidSigners function
func (sr *subroundEndRound) GetFullMessagesForInvalidSigners(invalidPubKeys []string) ([]byte, error) {
	return sr.getFullMessagesForInvalidSigners(invalidPubKeys)
}

// GetSentSignatureTracker returns the subroundEndRound's SentSignaturesTracker instance
func (sr *subroundEndRound) GetSentSignatureTracker() spos.SentSignaturesTracker {
	return sr.sentSignatureTracker
}

// ChangeEpoch calls the unexported changeEpoch function
func (sr *subroundStartRound) ChangeEpoch(epoch uint32) {
	sr.changeEpoch(epoch)
}

// IndexRoundIfNeeded calls the unexported indexRoundIfNeeded function
func (sr *subroundStartRound) IndexRoundIfNeeded(pubKeys []string) {
	sr.indexRoundIfNeeded(pubKeys)
}

// SendSignatureForManagedKey calls the unexported sendSignatureForManagedKey function
func (sr *subroundSignature) SendSignatureForManagedKey(idx int, pk string) bool {
	return sr.sendSignatureForManagedKey(idx, pk)
}

// DoSignatureJobForManagedKeys calls the unexported doSignatureJobForManagedKeys function
func (sr *subroundSignature) DoSignatureJobForManagedKeys(ctx context.Context) bool {
	return sr.doSignatureJobForManagedKeys(ctx)
}

// ReceivedSignature method is called when a signature is received through the signature channel
func (sr *subroundEndRound) ReceivedSignature(cnsDta *consensus.Message) bool {
	return sr.receivedSignature(context.Background(), cnsDta)
}

// WaitForProof -
func (sr *subroundEndRound) WaitForProof() bool {
	return sr.waitForProof()
}

// GetEquivalentProofSender -
func (sr *subroundEndRound) GetEquivalentProofSender() string {
	return sr.getEquivalentProofSender()
}

// SendProof -
func (sr *subroundEndRound) SendProof() (bool, error) {
	return sr.sendProof()
}
