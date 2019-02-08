package bn

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
)

type subroundCommitment struct {
	*subround

	consensusState *spos.ConsensusState
	multiSigner    crypto.MultiSigner
	rounder        round.Rounder
	syncTimer      ntp.SyncTimer

	sendConsensusMessage func(*spos.ConsensusData) bool
}

// NewSubroundCommitment creates a subroundCommitment object
func NewSubroundCommitment(
	subround *subround,
	consensusState *spos.ConsensusState,
	multiSigner crypto.MultiSigner,
	rounder round.Rounder,
	syncTimer ntp.SyncTimer,
	sendConsensusMessage func(*spos.ConsensusData) bool,
	extend func(subroundId int),
) (*subroundCommitment, error) {

	err := checkNewSubroundCommitmentParams(
		subround,
		consensusState,
		multiSigner,
		rounder,
		syncTimer,
		sendConsensusMessage,
	)

	if err != nil {
		return nil, err
	}

	srCommitment := subroundCommitment{
		subround,
		consensusState,
		multiSigner,
		rounder,
		syncTimer,
		sendConsensusMessage,
	}

	srCommitment.job = srCommitment.doCommitmentJob
	srCommitment.check = srCommitment.doCommitmentConsensusCheck
	srCommitment.extend = extend

	return &srCommitment, nil
}

func checkNewSubroundCommitmentParams(
	subround *subround,
	consensusState *spos.ConsensusState,
	multiSigner crypto.MultiSigner,
	rounder round.Rounder,
	syncTimer ntp.SyncTimer,
	sendConsensusMessage func(*spos.ConsensusData) bool,
) error {
	if subround == nil {
		return spos.ErrNilSubround
	}

	if consensusState == nil {
		return spos.ErrNilConsensusState
	}

	if multiSigner == nil {
		return spos.ErrNilMultiSigner
	}

	if rounder == nil {
		return spos.ErrNilRounder
	}

	if syncTimer == nil {
		return spos.ErrNilSyncTimer
	}

	if sendConsensusMessage == nil {
		return spos.ErrNilSendConsensusMessageFunction
	}

	return nil
}

// doCommitmentJob method is the function which is actually used to job the commitment for the received block,
// in the Commitment subround (it is used as the handler function of the doSubroundJob pointer variable function
// in subround struct, from spos package)
func (sr *subroundCommitment) doCommitmentJob() bool {
	if !sr.consensusState.IsSelfJobDone(SrBitmap) { // is NOT self in the leader's bitmap?
		return false
	}

	if !sr.consensusState.CanDoSubroundJob(SrCommitment) {
		return false
	}

	selfIndex, err := sr.consensusState.IndexSelfConsensusGroup()

	if err != nil {
		log.Error(err.Error())
		return false
	}

	// commitment
	commitment, err := sr.multiSigner.Commitment(uint16(selfIndex))

	if err != nil {
		log.Error(err.Error())
		return false
	}

	dta := spos.NewConsensusData(
		sr.consensusState.Data,
		commitment,
		[]byte(sr.consensusState.SelfPubKey()),
		nil,
		int(MtCommitment),
		uint64(sr.rounder.TimeStamp().Unix()),
		sr.rounder.Index())

	if !sr.sendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 4: Sending commitment\n", sr.syncTimer.FormattedCurrentTime()))

	err = sr.consensusState.SetSelfJobDone(SrCommitment, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

// receivedCommitment method is called when a commitment is received through the commitment channel.
// If the commitment is valid, than the jobDone map coresponding to the node which sent it,
// is set on true for the subround Comitment
func (sr *subroundCommitment) receivedCommitment(cnsDta *spos.ConsensusData) bool {
	node := string(cnsDta.PubKey)

	if sr.consensusState.IsConsensusDataNotSet() {
		return false
	}

	if !sr.consensusState.IsJobDone(node, SrBitmap) { // is NOT this node in the bitmap group?
		return false
	}

	if !sr.consensusState.CanProcessReceivedMessage(cnsDta, sr.rounder.Index(), SrCommitment) {
		return false
	}

	index, err := sr.consensusState.ConsensusGroupIndex(node)

	if err != nil {
		log.Info(err.Error())
		return false
	}

	err = sr.multiSigner.AddCommitment(uint16(index), cnsDta.SubRoundData)

	if err != nil {
		log.Info(err.Error())
		return false
	}

	err = sr.consensusState.SetJobDone(node, SrCommitment, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

// doCommitmentConsensusCheck method checks if the consensus in the <COMMITMENT> subround is achieved
func (sr *subroundCommitment) doCommitmentConsensusCheck() bool {
	if sr.consensusState.RoundCanceled {
		return false
	}

	if sr.consensusState.Status(SrCommitment) == spos.SsFinished {
		return true
	}

	threshold := sr.consensusState.Threshold(SrCommitment)

	if sr.commitmentsCollected(threshold) {
		sr.printCommitmentCM() // only for printing commitment consensus messages
		sr.consensusState.SetStatus(SrCommitment, spos.SsFinished)

		return true
	}

	return false
}

// commitmentsCollected method checks if the commitments received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (sr *subroundCommitment) commitmentsCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.consensusState.ConsensusGroup()); i++ {
		node := sr.consensusState.ConsensusGroup()[i]
		isBitmapJobDone, err := sr.consensusState.GetJobDone(node, SrBitmap)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isBitmapJobDone {
			isCommJobDone, err := sr.consensusState.GetJobDone(node, SrCommitment)

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

// printCommitmentCM method prints the (COMMITMENT) subround consensus messages
func (sr *subroundCommitment) printCommitmentCM() {
	log.Info(fmt.Sprintf("%sStep 4: Received %d from %d commitments, which are matching with bitmap and are enough\n",
		sr.syncTimer.FormattedCurrentTime(), sr.consensusState.ComputeSize(SrCommitment), len(sr.consensusState.ConsensusGroup())))

	log.Info(fmt.Sprintf("%sStep 4: subround (COMMITMENT) has been finished\n", sr.syncTimer.FormattedCurrentTime()))
}
