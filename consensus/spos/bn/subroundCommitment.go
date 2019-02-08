package bn

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

// doCommitmentJob method is the function which is actually used to send the commitment for the received block,
// in the Commitment subround (it is used as the handler function of the doSubroundJob pointer variable function
// in Subround struct, from spos package)
func (wrk *Worker) doCommitmentJob() bool {
	if wrk.isBitmapSubroundUnfinished() {
		return false
	}

	if !wrk.isSelfInBitmap() { // is NOT self in the leader's bitmap?
		return false
	}

	if !wrk.canDoSubroundJob(SrCommitment) {
		return false
	}

	selfIndex, err := wrk.SPoS.IndexSelfConsensusGroup()

	if err != nil {
		log.Error(err.Error())
		return false
	}

	// commitment
	commitment, err := wrk.multiSigner.Commitment(uint16(selfIndex))

	if err != nil {
		log.Error(err.Error())
		return false
	}

	dta := spos.NewConsensusData(
		wrk.SPoS.Data,
		commitment,
		[]byte(wrk.SPoS.SelfPubKey()),
		nil,
		int(MtCommitment),
		wrk.SPoS.Chr.RoundTimeStamp(),
		wrk.SPoS.Chr.Round().Index())

	if !wrk.sendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 4: Sending commitment\n", wrk.SPoS.Chr.GetFormattedTime()))

	err = wrk.SPoS.SetSelfJobDone(SrCommitment, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

func (wrk *Worker) isBitmapSubroundUnfinished() bool {
	isBitmapSubroundUnfinished := wrk.SPoS.Status(SrBitmap) != spos.SsFinished

	if isBitmapSubroundUnfinished {
		if !wrk.doBitmapJob() {
			return true
		}

		if !wrk.checkBitmapConsensus() {
			return true
		}
	}

	return false
}

// receivedCommitment method is called when a commitment is received through the commitment channel.
// If the commitment is valid, than the jobDone map coresponding to the node which sent it,
// is set on true for the subround Comitment
func (wrk *Worker) receivedCommitment(cnsDta *spos.ConsensusData) bool {
	node := string(cnsDta.PubKey)

	if wrk.isConsensusDataNotSet() {
		return false
	}

	if !wrk.isValidatorInBitmap(node) { // is NOT this node in the bitmap group?
		return false
	}

	if !wrk.canReceiveMessage(node, cnsDta.RoundIndex, SrCommitment) {
		return false
	}

	index, err := wrk.SPoS.ConsensusGroupIndex(node)

	if err != nil {
		log.Info(err.Error())
		return false
	}

	err = wrk.multiSigner.AddCommitment(uint16(index), cnsDta.SubRoundData)

	if err != nil {
		log.Info(err.Error())
		return false
	}

	err = wrk.SPoS.RoundConsensus.SetJobDone(node, SrCommitment, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

// checkCommitmentConsensus method checks if the consensus in the <COMMITMENT> subround is achieved
func (wrk *Worker) checkCommitmentConsensus() bool {
	wrk.mutCheckConsensus.Lock()
	defer wrk.mutCheckConsensus.Unlock()

	if wrk.SPoS.Chr.IsCancelled() {
		return false
	}

	if wrk.SPoS.Status(SrCommitment) == spos.SsFinished {
		return true
	}

	threshold := wrk.SPoS.Threshold(SrCommitment)

	if wrk.commitmentsCollected(threshold) {
		wrk.printCommitmentCM() // only for printing commitment consensus messages
		wrk.SPoS.SetStatus(SrCommitment, spos.SsFinished)

		return true
	}

	return false
}

// commitmentsCollected method checks if the commitments received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (wrk *Worker) commitmentsCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		node := wrk.SPoS.ConsensusGroup()[i]
		isBitmapJobDone, err := wrk.SPoS.GetJobDone(node, SrBitmap)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isBitmapJobDone {
			isCommJobDone, err := wrk.SPoS.GetJobDone(node, SrCommitment)

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

// extendCommitment method put this subround in the extended mode and print some messages
func (wrk *Worker) extendCommitment() {
	wrk.SPoS.SetStatus(SrCommitment, spos.SsExtended)

	log.Info(fmt.Sprintf("%sStep 4: Extended the (COMMITMENT) subround. Got only %d from %d commitments which are not enough\n",
		wrk.SPoS.Chr.GetFormattedTime(), wrk.SPoS.ComputeSize(SrCommitment), len(wrk.SPoS.ConsensusGroup())))
}

// printCommitmentCM method prints the (COMMITMENT) subround consensus messages
func (wrk *Worker) printCommitmentCM() {
	log.Info(fmt.Sprintf("%sStep 4: Received %d from %d commitments, which are matching with bitmap and are enough\n",
		wrk.SPoS.Chr.GetFormattedTime(), wrk.SPoS.ComputeSize(SrCommitment), len(wrk.SPoS.ConsensusGroup())))

	log.Info(fmt.Sprintf("%sStep 4: Subround (COMMITMENT) has been finished\n", wrk.SPoS.Chr.GetFormattedTime()))
}
