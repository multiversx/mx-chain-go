package bls

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
)

type subroundSignature struct {
	*spos.Subround

	appStatusHandler core.AppStatusHandler
}

// NewSubroundSignature creates a subroundSignature object
func NewSubroundSignature(
	baseSubround *spos.Subround,
	extend func(subroundId int),
) (*subroundSignature, error) {
	err := checkNewSubroundSignatureParams(
		baseSubround,
	)
	if err != nil {
		return nil, err
	}

	srSignature := subroundSignature{
		Subround:         baseSubround,
		appStatusHandler: statusHandler.NewNilStatusHandler(),
	}
	srSignature.Job = srSignature.doSignatureJob
	srSignature.Check = srSignature.doSignatureConsensusCheck
	srSignature.Extend = extend

	return &srSignature, nil
}

// SetAppStatusHandler method set appStatusHandler
func (sr *subroundSignature) SetAppStatusHandler(ash core.AppStatusHandler) error {
	if ash == nil || ash.IsInterfaceNil() {
		return spos.ErrNilAppStatusHandler
	}

	sr.appStatusHandler = ash
	return nil
}

func checkNewSubroundSignatureParams(
	baseSubround *spos.Subround,
) error {
	if baseSubround == nil {
		return spos.ErrNilSubround
	}
	if baseSubround.ConsensusState == nil {
		return spos.ErrNilConsensusState
	}

	err := spos.ValidateConsensusCore(baseSubround.ConsensusCoreHandler)

	return err
}

// doSignatureJob method does the job of the subround Signature
func (sr *subroundSignature) doSignatureJob() bool {
	if !sr.IsNodeInConsensusGroup(sr.SelfPubKey()) {
		return true
	}
	if !sr.CanDoSubroundJob(sr.Current()) {
		return false
	}

	sigPart, err := sr.MultiSigner().CreateSignatureShare(sr.GetData(), nil)
	if err != nil {
		debugError("CreateSignatureShare", err)
		return false
	}

	if !sr.IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		//TODO: Analyze it is possible to send message only to leader with O(1) instead of O(n)
		msg := consensus.NewConsensusMessage(
			sr.GetData(),
			sigPart,
			[]byte(sr.SelfPubKey()),
			nil,
			int(MtSignature),
			sr.Rounder().Index(),
			sr.ChainID(),
		)

		err = sr.BroadcastMessenger().BroadcastConsensusMessage(msg)
		if err != nil {
			debugError("BroadcastConsensusMessage", err)
			return false
		}

		log.Debug("step 2: signature has been sent",
			"type", "spos/bls",
			"time [s]", sr.SyncTimer().FormattedCurrentTime())
	}

	err = sr.SetSelfJobDone(sr.Current(), true)
	if err != nil {
		debugError("SetSelfJobDone", err)
		return false
	}

	return true
}

// receivedSignature method is called when a signature is received through the signature channel.
// If the signature is valid, than the jobDone map corresponding to the node which sent it,
// is set on true for the subround Signature
func (sr *subroundSignature) receivedSignature(cnsDta *consensus.Message) bool {
	if !sr.IsSelfLeaderInCurrentRound() {
		return false
	}

	node := string(cnsDta.PubKey)

	if !sr.IsConsensusDataSet() {
		return false
	}

	if !sr.IsConsensusDataEqual(cnsDta.BlockHeaderHash) {
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.Rounder().Index(), sr.Current()) {
		return false
	}

	// if this node is leader in this round and it already received 2/3 + 1 of signatures
	// it will ignore any others received later
	threshold := sr.Threshold(sr.Current())
	if ok, _ := sr.signaturesCollected(threshold); ok {
		return false
	}

	index, err := sr.ConsensusGroupIndex(node)
	if err != nil {
		debugError("ConsensusGroupIndex", err)
		return false
	}

	currentMultiSigner := sr.MultiSigner()
	err = currentMultiSigner.StoreSignatureShare(uint16(index), cnsDta.SubRoundData)
	if err != nil {
		debugError("StoreSignatureShare", err)
		return false
	}

	err = sr.SetJobDone(node, sr.Current(), true)
	if err != nil {
		debugError("SetJobDone", err)
		return false
	}

	threshold = sr.Threshold(sr.Current())
	if ok, n := sr.signaturesCollected(threshold); ok {
		log.Debug("step 2: signatures",
			"type", "spos/bls",
			"time [s]", sr.SyncTimer().FormattedCurrentTime(),
			"received", n,
			"total", len(sr.ConsensusGroup()))
	}

	sr.appStatusHandler.SetStringValue(core.MetricConsensusRoundState, "signed")
	return true
}

// doSignatureConsensusCheck method checks if the consensus in the subround Signature is achieved
func (sr *subroundSignature) doSignatureConsensusCheck() bool {
	if sr.RoundCanceled {
		return false
	}

	if sr.IsSubroundFinished(sr.Current()) {
		sr.appStatusHandler.SetStringValue(core.MetricConsensusRoundState, "signed")

		return true
	}

	isSelfLeader := sr.IsSelfLeaderInCurrentRound()
	isSelfInConsensusGroup := sr.IsNodeInConsensusGroup(sr.SelfPubKey())

	threshold := sr.Threshold(sr.Current())
	areSignaturesCollected, _ := sr.signaturesCollected(threshold)

	isJobDoneByLeader := isSelfLeader && areSignaturesCollected
	isJobDoneByConsensusNode := !isSelfLeader && isSelfInConsensusGroup && sr.IsSelfJobDone(sr.Current())

	isSubroundFinished := !isSelfInConsensusGroup || isJobDoneByConsensusNode || isJobDoneByLeader

	if isSubroundFinished {
		log.Debug("step 2: Subround has been finished",
			"type", "spos/bls",
			"subround", sr.Name(),
			"time [s]", sr.SyncTimer().FormattedCurrentTime())
		sr.SetStatus(sr.Current(), spos.SsFinished)

		sr.appStatusHandler.SetStringValue(core.MetricConsensusRoundState, "signed")

		return true
	}

	return false
}

// signaturesCollected method checks if the signatures received from the nodes, belonging to the current
// jobDone group, are more than the necessary given threshold
func (sr *subroundSignature) signaturesCollected(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		node := sr.ConsensusGroup()[i]

		isSignJobDone, err := sr.JobDone(node, sr.Current())
		if err != nil {
			debugError("JobDone", err)
			continue
		}

		if isSignJobDone {
			n++
		}
	}

	return n >= threshold, n
}
