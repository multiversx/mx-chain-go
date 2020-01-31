package bls

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// factory

type Factory *factory

func (fct *factory) BlockChain() data.ChainHandler {
	return fct.consensusCore.Blockchain()
}

func (fct *factory) BlockProcessor() process.BlockProcessor {
	return fct.consensusCore.BlockProcessor()
}

func (fct *factory) Bootstrapper() process.Bootstrapper {
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

func (fct *factory) NodesCoordinator() sharding.NodesCoordinator {
	return fct.consensusCore.NodesCoordinator()
}

func (fct *factory) Worker() spos.WorkerHandler {
	return fct.worker
}

func (fct *factory) SetWorker(worker spos.WorkerHandler) {
	fct.worker = worker
}

func (fct *factory) GenerateStartRoundSubround() error {
	return fct.generateStartRoundSubround()
}

func (fct *factory) GenerateBlockSubround() error {
	return fct.generateBlockSubround()
}

func (fct *factory) GenerateSignatureSubround() error {
	return fct.generateSignatureSubround()
}

func (fct *factory) GenerateEndRoundSubround() error {
	return fct.generateEndRoundSubround()
}

func (fct *factory) AppStatusHandler() core.AppStatusHandler {
	return fct.appStatusHandler
}

func (fct *factory) Indexer() indexer.Indexer {
	return fct.indexer
}

// subroundStartRound

type SubroundStartRound *subroundStartRound

func (sr *subroundStartRound) DoStartRoundJob() bool {
	return sr.doStartRoundJob()
}

func (sr *subroundStartRound) DoStartRoundConsensusCheck() bool {
	return sr.doStartRoundConsensusCheck()
}

func (sr *subroundStartRound) GenerateNextConsensusGroup(roundIndex int64) error {
	return sr.generateNextConsensusGroup(roundIndex)
}

func (sr *subroundStartRound) InitCurrentRound() bool {
	return sr.initCurrentRound()
}

// subroundBlock

type SubroundBlock *subroundBlock

func (sr *subroundBlock) BlockChain() data.ChainHandler {
	return sr.Blockchain()
}

func (sr *subroundBlock) DoBlockJob() bool {
	return sr.doBlockJob()
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

func (sr *subroundBlock) CreateBody(hdr data.HeaderHandler) (data.BodyHandler, error) {
	return sr.createBody(hdr)
}

func (sr *subroundBlock) SendBlockBody(body data.BodyHandler) bool {
	return sr.sendBlockBody(body)
}

func (sr *subroundBlock) SendBlockHeader(header data.HeaderHandler) bool {
	return sr.sendBlockHeader(header)
}

func (sr *subroundBlock) ComputeSubroundProcessingMetric(startTime time.Time, metric string) {
	sr.computeSubroundProcessingMetric(startTime, metric)
}

func (sr *subroundBlock) ReceivedBlockBody(cnsDta *consensus.Message) bool {
	return sr.receivedBlockBody(cnsDta)
}

func (sr *subroundBlock) ReceivedBlockHeader(cnsDta *consensus.Message) bool {
	return sr.receivedBlockHeader(cnsDta)
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

func (sr *subroundSignature) SignaturesCollected(threshold int) (bool, int) {
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

func GetStringValue(messageType consensus.MessageType) string {
	return getStringValue(messageType)
}
