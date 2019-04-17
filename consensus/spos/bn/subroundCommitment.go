package bn

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

type subroundCommitment struct {
	*subround

	sendConsensusMessage func(*consensus.ConsensusMessage) bool
}

// NewSubroundCommitment creates a subroundCommitment object
func NewSubroundCommitment(
	subround *subround,
	sendConsensusMessage func(*consensus.ConsensusMessage) bool,
	extend func(subroundId int),
) (*subroundCommitment, error) {

	err := checkNewSubroundCommitmentParams(
		subround,
		sendConsensusMessage,
	)

	if err != nil {
		return nil, err
	}

	srCommitment := subroundCommitment{
		subround,
		sendConsensusMessage,
	}

	srCommitment.job = srCommitment.doCommitmentJob
	srCommitment.check = srCommitment.doCommitmentConsensusCheck
	srCommitment.extend = extend

	return &srCommitment, nil
}

func checkNewSubroundCommitmentParams(
	subround *subround,
	sendConsensusMessage func(*consensus.ConsensusMessage) bool,
) error {
	if subround == nil {
		return spos.ErrNilSubround
	}

	if sendConsensusMessage == nil {
		return spos.ErrNilSendConsensusMessageFunction
	}

	return nil
}

// doCommitmentJob method does the job of the commitment subround
func (sr *subroundCommitment) doCommitmentJob() bool {
	if !sr.ConsensusState().IsSelfJobDone(SrBitmap) { // is NOT self in the leader's bitmap?
		return false
	}

	if !sr.ConsensusState().CanDoSubroundJob(SrCommitment) {
		return false
	}

	selfIndex, err := sr.ConsensusState().SelfConsensusGroupIndex()

	if err != nil {
		log.Error(err.Error())
		return false
	}

	// commitment
	commitment, err := sr.MultiSigner().Commitment(uint16(selfIndex))

	if err != nil {
		log.Error(err.Error())
		return false
	}

	msg := consensus.NewConsensusMessage(
		sr.ConsensusState().Data,
		commitment,
		[]byte(sr.ConsensusState().SelfPubKey()),
		nil,
		int(MtCommitment),
		uint64(sr.Rounder().TimeStamp().Unix()),
		sr.Rounder().Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 4: commitment has been sent\n", sr.SyncTimer().FormattedCurrentTime()))

	err = sr.ConsensusState().SetSelfJobDone(SrCommitment, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

// receivedCommitment method is called when a commitment is received through the commitment channel.
// If the commitment is valid, than the jobDone map corresponding to the node which sent it,
// is set on true for the subround Commitment
func (sr *subroundCommitment) receivedCommitment(cnsDta *consensus.ConsensusMessage) bool {
	node := string(cnsDta.PubKey)

	if !sr.ConsensusState().IsConsensusDataSet() {
		return false
	}

	if !sr.ConsensusState().IsConsensusDataEqual(cnsDta.BlockHeaderHash) {
		return false
	}

	if !sr.ConsensusState().IsJobDone(node, SrBitmap) { // is NOT this node in the bitmap group?
		return false
	}

	if !sr.ConsensusState().CanProcessReceivedMessage(cnsDta, sr.Rounder().Index(), SrCommitment) {
		return false
	}

	index, err := sr.ConsensusState().ConsensusGroupIndex(node)

	if err != nil {
		log.Info(err.Error())
		return false
	}

	err = sr.MultiSigner().StoreCommitment(uint16(index), cnsDta.SubRoundData)

	if err != nil {
		log.Info(err.Error())
		return false
	}

	err = sr.ConsensusState().SetJobDone(node, SrCommitment, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	threshold := sr.ConsensusState().Threshold(SrCommitment)
	if sr.commitmentsCollected(threshold) {
		n := sr.ConsensusState().ComputeSize(SrCommitment)
		log.Info(fmt.Sprintf("%sStep 4: received %d from %d commitments\n",
			sr.SyncTimer().FormattedCurrentTime(), n, len(sr.ConsensusState().ConsensusGroup())))
	}

	return true
}

// doCommitmentConsensusCheck method checks if the consensus in the <COMMITMENT> subround is achieved
func (sr *subroundCommitment) doCommitmentConsensusCheck() bool {
	if sr.ConsensusState().RoundCanceled {
		return false
	}

	if sr.ConsensusState().Status(SrCommitment) == spos.SsFinished {
		return true
	}

	threshold := sr.ConsensusState().Threshold(SrCommitment)
	if sr.commitmentsCollected(threshold) {
		log.Info(fmt.Sprintf("%sStep 4: subround %s has been finished\n", sr.SyncTimer().FormattedCurrentTime(), sr.Name()))
		sr.ConsensusState().SetStatus(SrCommitment, spos.SsFinished)
		return true
	}

	return false
}

// commitmentsCollected method checks if the commitments received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (sr *subroundCommitment) commitmentsCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.ConsensusState().ConsensusGroup()); i++ {
		node := sr.ConsensusState().ConsensusGroup()[i]
		isBitmapJobDone, err := sr.ConsensusState().JobDone(node, SrBitmap)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isBitmapJobDone {
			isCommJobDone, err := sr.ConsensusState().JobDone(node, SrCommitment)

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
