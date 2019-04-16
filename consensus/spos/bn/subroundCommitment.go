package bn

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

type subroundCommitment struct {
	*subround

	sendConsensusMessage func(*spos.ConsensusMessage) bool
}

// NewSubroundCommitment creates a subroundCommitment object
func NewSubroundCommitment(
	subround *subround,
	sendConsensusMessage func(*spos.ConsensusMessage) bool,
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
	sendConsensusMessage func(*spos.ConsensusMessage) bool,
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
	if !sr.GetConsensusState().IsSelfJobDone(SrBitmap) { // is NOT self in the leader's bitmap?
		return false
	}

	if !sr.GetConsensusState().CanDoSubroundJob(SrCommitment) {
		return false
	}

	selfIndex, err := sr.GetConsensusState().SelfConsensusGroupIndex()

	if err != nil {
		log.Error(err.Error())
		return false
	}

	// commitment
	commitment, err := sr.GetMultiSigner().Commitment(uint16(selfIndex))

	if err != nil {
		log.Error(err.Error())
		return false
	}

	msg := spos.NewConsensusMessage(
		sr.GetConsensusState().Data,
		commitment,
		[]byte(sr.GetConsensusState().SelfPubKey()),
		nil,
		int(MtCommitment),
		uint64(sr.GetRounder().TimeStamp().Unix()),
		sr.GetRounder().Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 4: commitment has been sent\n", sr.GetSyncTimer().FormattedCurrentTime()))

	err = sr.GetConsensusState().SetSelfJobDone(SrCommitment, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

// receivedCommitment method is called when a commitment is received through the commitment channel.
// If the commitment is valid, than the jobDone map corresponding to the node which sent it,
// is set on true for the subround Commitment
func (sr *subroundCommitment) receivedCommitment(cnsDta *spos.ConsensusMessage) bool {
	node := string(cnsDta.PubKey)

	if !sr.GetConsensusState().IsConsensusDataSet() {
		return false
	}

	if !sr.GetConsensusState().IsConsensusDataEqual(cnsDta.BlockHeaderHash) {
		return false
	}

	if !sr.GetConsensusState().IsJobDone(node, SrBitmap) { // is NOT this node in the bitmap group?
		return false
	}

	if !sr.GetConsensusState().CanProcessReceivedMessage(cnsDta, sr.GetRounder().Index(), SrCommitment) {
		return false
	}

	index, err := sr.GetConsensusState().ConsensusGroupIndex(node)

	if err != nil {
		log.Info(err.Error())
		return false
	}

	err = sr.GetMultiSigner().StoreCommitment(uint16(index), cnsDta.SubRoundData)

	if err != nil {
		log.Info(err.Error())
		return false
	}

	err = sr.GetConsensusState().SetJobDone(node, SrCommitment, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	threshold := sr.GetConsensusState().Threshold(SrCommitment)
	if sr.commitmentsCollected(threshold) {
		n := sr.GetConsensusState().ComputeSize(SrCommitment)
		log.Info(fmt.Sprintf("%sStep 4: received %d from %d commitments\n",
			sr.GetSyncTimer().FormattedCurrentTime(), n, len(sr.GetConsensusState().ConsensusGroup())))
	}

	return true
}

// doCommitmentConsensusCheck method checks if the consensus in the <COMMITMENT> subround is achieved
func (sr *subroundCommitment) doCommitmentConsensusCheck() bool {
	if sr.GetConsensusState().RoundCanceled {
		return false
	}

	if sr.GetConsensusState().Status(SrCommitment) == spos.SsFinished {
		return true
	}

	threshold := sr.GetConsensusState().Threshold(SrCommitment)
	if sr.commitmentsCollected(threshold) {
		log.Info(fmt.Sprintf("%sStep 4: subround %s has been finished\n", sr.GetSyncTimer().FormattedCurrentTime(), sr.Name()))
		sr.GetConsensusState().SetStatus(SrCommitment, spos.SsFinished)
		return true
	}

	return false
}

// commitmentsCollected method checks if the commitments received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (sr *subroundCommitment) commitmentsCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.GetConsensusState().ConsensusGroup()); i++ {
		node := sr.GetConsensusState().ConsensusGroup()[i]
		isBitmapJobDone, err := sr.GetConsensusState().JobDone(node, SrBitmap)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isBitmapJobDone {
			isCommJobDone, err := sr.GetConsensusState().JobDone(node, SrCommitment)

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
