package bls

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

type subroundSignature struct {
	*spos.Subround

	sendConsensusMessage func(*consensus.Message) bool
}

// NewSubroundSignature creates a subroundSignature object
func NewSubroundSignature(
	subround *spos.Subround,
	sendConsensusMessage func(*consensus.Message) bool,
	extend func(subroundId int),
) (*subroundSignature, error) {
	err := checkNewSubroundSignatureParams(
		subround,
		sendConsensusMessage,
	)
	if err != nil {
		return nil, err
	}

	srSignature := subroundSignature{
		subround,
		sendConsensusMessage,
	}
	srSignature.Job = srSignature.doSignatureJob
	srSignature.Check = srSignature.doSignatureConsensusCheck
	srSignature.Extend = extend

	return &srSignature, nil
}

func checkNewSubroundSignatureParams(
	subround *spos.Subround,
	sendConsensusMessage func(*consensus.Message) bool,
) error {
	if subround == nil {
		return spos.ErrNilSubround
	}

	if subround.ConsensusState == nil {
		return spos.ErrNilConsensusState
	}

	if sendConsensusMessage == nil {
		return spos.ErrNilSendConsensusMessageFunction
	}

	err := spos.ValidateConsensusCore(subround.ConsensusCoreHandler)

	return err
}

// doSignatureJob method does the Job of the signatuure Subround
func (sr *subroundSignature) doSignatureJob() bool {
	//if !sr.IsSelfJobDone(SrBitmap) { // is NOT self in the leader's bitmap?
	//	return false
	//}

	if !sr.CanDoSubroundJob(SrSignature) {
		return false
	}

	// Start (new add)
	sigPart, err := sr.MultiSigner().CreateSignatureShare(sr.GetData(), []byte("bls"))
	if err != nil {
		log.Error(err.Error())
		return false
	}
	// End (new add)

	msg := consensus.NewConsensusMessage(
		sr.GetData(),
		sigPart,
		[]byte(sr.SelfPubKey()),
		nil,
		int(MtSignature),
		uint64(sr.Rounder().TimeStamp().Unix()),
		sr.Rounder().Index())

	// Start (new add)
	if !sr.IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		if !sr.sendConsensusMessage(msg) {
			return false
		}

		log.Info(fmt.Sprintf("%sStep 2: signature has been sent\n", sr.SyncTimer().FormattedCurrentTime()))

		sr.RoundCanceled = true
	}
	// End (new add)

	err = sr.SetSelfJobDone(SrSignature, true)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

// receivedSignature method is called when a Signature is received through the Signature channel.
// If the Signature is valid, than the jobDone map corresponding to the node which sent it,
// is set on true for the Subround Signature
func (sr *subroundSignature) receivedSignature(cnsDta *consensus.Message) bool {
	node := string(cnsDta.PubKey)

	if !sr.IsConsensusDataSet() {
		return false
	}

	if !sr.IsConsensusDataEqual(cnsDta.BlockHeaderHash) {
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.Rounder().Index(), SrSignature) {
		return false
	}

	// Start (new add)
	// if this node is leader in this round and it already received 2/3 + 1 of signatures
	// it will ignore any others received later
	if sr.IsSelfLeaderInCurrentRound() {
		threshold := sr.Threshold(SrSignature)
		if sr.signaturesCollected(threshold) {
			return false
		}
	}
	// End (new add)

	index, err := sr.ConsensusGroupIndex(node)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	currentMultiSigner := sr.MultiSigner()
	err = currentMultiSigner.StoreSignatureShare(uint16(index), cnsDta.SubRoundData)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sr.SetJobDone(node, SrSignature, true)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	if sr.IsSelfLeaderInCurrentRound() { // (new add)
		threshold := sr.Threshold(SrSignature)
		if sr.signaturesCollected(threshold) {
			n := sr.ComputeSize(SrSignature)
			log.Info(fmt.Sprintf("%sStep 2: received %d from %d signatures\n",
				sr.SyncTimer().FormattedCurrentTime(), n, len(sr.ConsensusGroup())))
		}
	}

	return true
}

// doSignatureConsensusCheck method checks if the consensus in the <SIGNATURE> Subround is achieved
func (sr *subroundSignature) doSignatureConsensusCheck() bool {
	if sr.RoundCanceled {
		return false
	}

	if sr.Status(SrSignature) == spos.SsFinished {
		return true
	}

	threshold := sr.Threshold(SrSignature)
	if sr.signaturesCollected(threshold) {
		log.Info(fmt.Sprintf("%sStep 2: Subround %s has been finished\n", sr.SyncTimer().FormattedCurrentTime(), sr.Name()))
		sr.SetStatus(SrSignature, spos.SsFinished)
		return true
	}

	return false
}

// signaturesCollected method checks if the signatures received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (sr *subroundSignature) signaturesCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		node := sr.ConsensusGroup()[i]

		//if isBitmapJobDone {
		isSignJobDone, err := sr.JobDone(node, SrSignature)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		// Start (new add)
		if isSignJobDone {
			n++
		}
		// End (new add)
	}

	return n >= threshold
}
