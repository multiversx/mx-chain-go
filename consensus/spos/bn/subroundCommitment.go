package bn

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
)

type subroundCommitment struct {
	*spos.Subround
}

// NewSubroundCommitment creates a subroundCommitment object
func NewSubroundCommitment(
	baseSubround *spos.Subround,
	extend func(subroundId int),
) (*subroundCommitment, error) {

	err := checkNewSubroundCommitmentParams(
		baseSubround,
	)

	if err != nil {
		return nil, err
	}

	srCommitment := subroundCommitment{
		baseSubround,
	}

	srCommitment.Job = srCommitment.doCommitmentJob
	srCommitment.Check = srCommitment.doCommitmentConsensusCheck
	srCommitment.Extend = extend

	return &srCommitment, nil
}

func checkNewSubroundCommitmentParams(
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

// doCommitmentJob method does the job of the subround Commitment
func (sr *subroundCommitment) doCommitmentJob() bool {
	if !sr.IsSelfJobDone(SrBitmap) { // is NOT self in the leader's bitmap?
		return false
	}

	if !sr.CanDoSubroundJob(SrCommitment) {
		return false
	}

	selfIndex, err := sr.SelfConsensusGroupIndex()
	if err != nil {
		log.Error(err.Error())
		return false
	}

	multiSig, err := getBnMultiSigner(sr.MultiSigner())
	if err != nil {
		log.Error(err.Error())
		return false
	}

	// commitment
	commitment, err := multiSig.Commitment(uint16(selfIndex))
	if err != nil {
		log.Error(err.Error())
		return false
	}

	msg := consensus.NewConsensusMessage(
		sr.Data,
		commitment,
		[]byte(sr.SelfPubKey()),
		nil,
		int(MtCommitment),
		uint64(sr.Rounder().TimeStamp().Unix()),
		sr.Rounder().Index())

	err = sr.BroadcastMessenger().BroadcastConsensusMessage(msg)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	log.Info(fmt.Sprintf("%sStep 4: commitment has been sent\n", sr.SyncTimer().FormattedCurrentTime()))

	err = sr.SetSelfJobDone(SrCommitment, true)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

// receivedCommitment method is called when a commitment is received through the commitment channel.
// If the commitment is valid, than the jobDone map corresponding to the node which sent it,
// is set on true for the subround Commitment
func (sr *subroundCommitment) receivedCommitment(cnsDta *consensus.Message) bool {
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

	if !sr.CanProcessReceivedMessage(cnsDta, sr.Rounder().Index(), SrCommitment) {
		return false
	}

	index, err := sr.ConsensusGroupIndex(node)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	currentMultiSigner, err := getBnMultiSigner(sr.MultiSigner())
	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = currentMultiSigner.StoreCommitment(uint16(index), cnsDta.SubRoundData)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sr.SetJobDone(node, SrCommitment, true)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	threshold := sr.Threshold(SrCommitment)
	if sr.commitmentsCollected(threshold) {
		n := sr.ComputeSize(SrCommitment)
		log.Info(fmt.Sprintf("%sStep 4: received %d from %d commitments\n",
			sr.SyncTimer().FormattedCurrentTime(), n, len(sr.ConsensusGroup())))
	}

	return true
}

// doCommitmentConsensusCheck method checks if the consensus in the subround Commitment is achieved
func (sr *subroundCommitment) doCommitmentConsensusCheck() bool {
	if sr.RoundCanceled {
		return false
	}

	if sr.Status(SrCommitment) == spos.SsFinished {
		return true
	}

	threshold := sr.Threshold(SrCommitment)
	if sr.commitmentsCollected(threshold) {
		log.Info(fmt.Sprintf("%sStep 4: subround %s has been finished\n", sr.SyncTimer().FormattedCurrentTime(), sr.Name()))
		sr.SetStatus(SrCommitment, spos.SsFinished)
		return true
	}

	return false
}

// commitmentsCollected method checks if the commitments received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (sr *subroundCommitment) commitmentsCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		node := sr.ConsensusGroup()[i]
		isBitmapJobDone, err := sr.JobDone(node, SrBitmap)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isBitmapJobDone {
			isCommJobDone, err := sr.JobDone(node, SrCommitment)

			if err != nil {
				log.Error(err.Error())
				continue
			}

			if !isCommJobDone {
				return false
			}
			n++
		}
	}

	return n >= threshold
}
