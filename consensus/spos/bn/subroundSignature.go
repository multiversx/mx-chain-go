package bn

import (
	"bytes"

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
	if !sr.IsSelfJobDone(SrBitmap) { // is NOT self in the leader's bitmap?
		return false
	}

	if !sr.CanDoSubroundJob(SrSignature) {
		return false
	}

	bitmap := sr.GenerateBitmap(SrBitmap)

	err := sr.checkCommitmentsValidity(bitmap)
	if err != nil {
		log.Debug("checkCommitmentsValidity", "type", "spos/bn", "error", err.Error())
		return false
	}

	currentMultiSigner, err := getBnMultiSigner(sr.MultiSigner())
	if err != nil {
		log.Debug("currentMultiSigner", "type", "spos/bn", "error", err.Error())
		return false
	}

	// first compute commitment aggregation
	err = currentMultiSigner.AggregateCommitments(bitmap)
	if err != nil {
		log.Debug("AggregateCommitments", "type", "spos/bn", "error", err.Error())
		return false
	}

	sigPart, err := currentMultiSigner.CreateSignatureShare(sr.GetData(), bitmap)
	if err != nil {
		log.Debug("CreateSignatureShare", "type", "spos/bn", "error", err.Error())
		return false
	}

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
		log.Debug("BroadcastConsensusMessage", "type", "spos/bn", "error", err.Error())
		return false
	}

	log.Debug("step 5: signature has been sent",
		"type", "spos/bn",
		"time [s]", sr.SyncTimer().FormattedCurrentTime())

	err = sr.SetSelfJobDone(SrSignature, true)
	if err != nil {
		log.Debug("SetSelfJobDone",
			"type", "spos/bn",
			"error", err.Error())
		return false
	}

	return true
}

func (sr *subroundSignature) checkCommitmentsValidity(bitmap []byte) error {
	nbBitsBitmap := len(bitmap) * 8
	consensusGroup := sr.ConsensusGroup()
	consensusGroupSize := len(consensusGroup)
	size := consensusGroupSize

	if consensusGroupSize > nbBitsBitmap {
		size = nbBitsBitmap
	}

	currentMultiSigner, err := getBnMultiSigner(sr.MultiSigner())
	if err != nil {
		return err
	}

	for i := 0; i < size; i++ {
		indexRequired := (bitmap[i/8] & (1 << uint16(i%8))) > 0

		if !indexRequired {
			continue
		}

		pubKey := consensusGroup[i]
		isCommJobDone, err := sr.JobDone(pubKey, SrCommitment)
		if err != nil {
			return err
		}

		if !isCommJobDone {
			return spos.ErrNilCommitment
		}

		commitment, err := currentMultiSigner.Commitment(uint16(i))
		if err != nil {
			return err
		}

		computedCommitmentHash := sr.Hasher().Compute(string(commitment))
		receivedCommitmentHash, err := currentMultiSigner.CommitmentHash(uint16(i))

		if err != nil {
			return err
		}
		if !bytes.Equal(computedCommitmentHash, receivedCommitmentHash) {
			return spos.ErrCommitmentHashDoesNotMatch
		}
	}

	return nil
}

// receivedSignature method is called when a Signature is received through the Signature channel.
// If the Signature is valid, than the jobDone map corresponding to the node which sent it,
// is set on true for the subround Signature
func (sr *subroundSignature) receivedSignature(cnsDta *consensus.Message) bool {
	node := string(cnsDta.PubKey)

	if !sr.IsConsensusDataSet() {
		return false
	}

	if !sr.IsConsensusDataEqual(cnsDta.BlockHeaderHash) {
		return false
	}

	if !sr.IsJobDone(node, SrBitmap) { // is NOT this node in the bitmap group?
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.Rounder().Index(), SrSignature) {
		return false
	}

	index, err := sr.ConsensusGroupIndex(node)
	if err != nil {
		log.Debug("ConsensusGroupIndex", "type", "spos/bn", "error", err.Error())
		return false
	}

	currentMultiSigner := sr.MultiSigner()
	err = currentMultiSigner.StoreSignatureShare(uint16(index), cnsDta.SubRoundData)
	if err != nil {
		log.Debug("StoreSignatureShare", "type", "spos/bn", "error", err.Error())
		return false
	}

	err = sr.SetJobDone(node, SrSignature, true)
	if err != nil {
		log.Debug("SetJobDone SrSignature", "type", "spos/bn", "error", err.Error())
		return false
	}

	threshold := sr.Threshold(SrSignature)
	if sr.signaturesCollected(threshold) {
		n := sr.ComputeSize(SrSignature)
		log.Debug("step 5: received signatures",
			"type", "spos/bn",
			"time [s]", sr.SyncTimer().FormattedCurrentTime(),
			"received", n,
			"total", len(sr.ConsensusGroup()))
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
	if sr.signaturesCollected(threshold) {
		log.Debug("step 5: subround has been finished",
			"type", "spos/bn",
			"time [s]", sr.SyncTimer().FormattedCurrentTime(),
			"subround", sr.Name())
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
		isBitmapJobDone, err := sr.JobDone(node, SrBitmap)
		if err != nil {
			log.Debug("SetJobDone SrSignature", "type", "spos/bn", "error", err.Error())
			continue
		}

		if isBitmapJobDone {
			isSignJobDone, err := sr.JobDone(node, SrSignature)
			if err != nil {
				log.Debug("SetJobDone SrSignature", "type", "spos/bn", "error", err.Error())
				continue
			}

			if !isSignJobDone {
				return false
			}
			n++
		}
	}

	return n >= threshold
}
