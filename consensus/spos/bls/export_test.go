package bls

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
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

func (fct *factory) ValidatorGroupSelector() sharding.NodesCoordinator {
	return fct.consensusCore.ValidatorGroupSelector()
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
