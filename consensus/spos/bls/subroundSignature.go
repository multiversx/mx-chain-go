package bls

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
)

type subroundSignature struct {
	*spos.Subround
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
		baseSubround,
	}
	srSignature.Job = srSignature.doSignatureJob
	srSignature.Check = srSignature.doSignatureConsensusCheck
	srSignature.Extend = extend

	return &srSignature, nil
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
	if !sr.CanDoSubroundJob(SrSignature) {
		return false
	}

	sigPart, err := sr.MultiSigner().CreateSignatureShare(sr.GetData(), nil)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	if !sr.IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		//TODO: Check if it is possible to send message only to leader with O(1) instead of O(n)
		msg := consensus.NewConsensusMessage(
			sr.GetData(),
			sigPart,
			[]byte(sr.SelfPubKey()),
			nil,
			int(MtSignature),
			uint64(sr.Rounder().TimeStamp().Unix()),
			sr.Rounder().Index())

		err = sr.BroadcastMessenger().BroadcastConsensusMessage(msg)
		if err != nil {
			log.Info(err.Error())
			return false
		}

		log.Info(fmt.Sprintf("%sStep 2: signature has been sent\n", sr.SyncTimer().FormattedCurrentTime()))

		// Validator has finished its job for this round
		sr.RoundCanceled = true
	}

	err = sr.SetSelfJobDone(SrSignature, true)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

// receivedSignature method is called when a signature is received through the signature channel.
// If the signature is valid, than the jobDone map corresponding to the node which sent it,
// is set on true for the subround Signature
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

	// if this node is leader in this round and it already received 2/3 + 1 of signatures
	// it will ignore any others received later
	if sr.IsSelfLeaderInCurrentRound() {
		threshold := sr.Threshold(SrSignature)
		if ok, _ := sr.signaturesCollected(threshold); ok {
			return false
		}
	}

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

	if sr.IsSelfLeaderInCurrentRound() {
		threshold := sr.Threshold(SrSignature)
		if ok, n := sr.signaturesCollected(threshold); ok {
			log.Info(fmt.Sprintf("%sStep 2: received %d from %d signatures\n",
				sr.SyncTimer().FormattedCurrentTime(), n, len(sr.ConsensusGroup())))
		}
	}

	return true
}

// doSignatureConsensusCheck method checks if the consensus in the subround Signature is achieved
func (sr *subroundSignature) doSignatureConsensusCheck() bool {
	if sr.RoundCanceled {
		return false
	}

	if sr.Status(SrSignature) == spos.SsFinished {
		return true
	}

	threshold := sr.Threshold(SrSignature)
	if ok, _ := sr.signaturesCollected(threshold); ok {
		log.Info(fmt.Sprintf("%sStep 2: Subround %s has been finished\n", sr.SyncTimer().FormattedCurrentTime(), sr.Name()))
		sr.SetStatus(SrSignature, spos.SsFinished)
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

		isSignJobDone, err := sr.JobDone(node, SrSignature)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isSignJobDone {
			n++
		}
	}

	return n >= threshold, n
}
