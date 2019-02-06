package bn

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

// doCommitmentHashJob method is the function which is actually used to send the commitment hash for the received
// block from the leader in the CommitmentHash subround (it is used as the handler function of the doSubroundJob
// pointer variable function in Subround struct, from spos package)
func (wrk *Worker) doCommitmentHashJob() bool {
	if !wrk.canDoCommitmentHashJob() {
		return false
	}

	commitmentHash, err := wrk.genCommitmentHash()

	if err != nil {
		log.Error(err.Error())
		return false
	}

	dta := spos.NewConsensusData(
		wrk.SPoS.Data,
		commitmentHash,
		[]byte(wrk.SPoS.SelfPubKey()),
		nil,
		int(MtCommitmentHash),
		wrk.SPoS.Chr.RoundTimeStamp(),
		wrk.SPoS.Chr.Round().Index())

	if !wrk.sendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 2: Sending commitment hash\n", wrk.SPoS.Chr.GetFormattedTime()))

	err = wrk.SPoS.SetSelfJobDone(SrCommitmentHash, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

func (wrk *Worker) canDoCommitmentHashJob() bool {
	isLastRoundUnfinished := wrk.SPoS.Status(SrBlock) != spos.SsFinished

	if isLastRoundUnfinished {
		if !wrk.doBlockJob() {
			return false
		}

		if !wrk.checkBlockConsensus() {
			return false
		}
	}

	isCurrentRoundFinished := wrk.SPoS.Status(SrCommitmentHash) == spos.SsFinished

	if isCurrentRoundFinished {
		return false
	}

	isJobDone, err := wrk.SPoS.GetSelfJobDone(SrCommitmentHash)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if isJobDone { // has been commitment hash already sent?
		return false
	}

	isConsensusDataNotSet := wrk.SPoS.Data == nil

	if isConsensusDataNotSet {
		return false
	}

	return true
}

// receivedCommitmentHash method is called when a commitment hash is received through the commitment hash
// channel. If the commitment hash is valid, than the jobDone map coresponding to the node which sent it,
// is set on true for the subround ComitmentHash
func (wrk *Worker) receivedCommitmentHash(cnsDta *spos.ConsensusData) bool {
	if !wrk.canReceiveCommitmentHash(cnsDta) {
		return false
	}

	// if this node is leader in this round and it already received 2/3 + 1 of commitment hashes
	// it will ignore any others received later
	if wrk.SPoS.IsSelfLeaderInCurrentRound() {
		threshold := wrk.SPoS.Threshold(SrCommitmentHash)
		if wrk.isCommitmentHashReceived(threshold) {
			return false
		}
	}

	node := string(cnsDta.PubKey)

	index, err := wrk.SPoS.ConsensusGroupIndex(node)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = wrk.multiSigner.AddCommitmentHash(uint16(index), cnsDta.SubRoundData)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = wrk.SPoS.RoundConsensus.SetJobDone(node, SrCommitmentHash, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

func (wrk *Worker) canReceiveCommitmentHash(cnsDta *spos.ConsensusData) bool {
	node := string(cnsDta.PubKey)

	isMessageReceivedFromItself := node == wrk.SPoS.SelfPubKey()

	if isMessageReceivedFromItself {
		return false
	}

	isCurrentRoundFinished := wrk.SPoS.Status(SrCommitmentHash) == spos.SsFinished

	if isCurrentRoundFinished {
		return false
	}

	if !wrk.SPoS.IsNodeInConsensusGroup(node) { // isn't node in the consensus group?
		return false
	}

	isConsensusDataNotSet := wrk.SPoS.Data == nil

	if isConsensusDataNotSet {
		return false
	}

	isMessageReceivedForAnotherRound := !bytes.Equal(cnsDta.BlockHeaderHash, wrk.SPoS.Data)

	if isMessageReceivedForAnotherRound {
		return false
	}

	isMessageReceivedTooLate := wrk.SPoS.Chr.GetSubround() > chronology.SubroundId(SrEndRound)

	if isMessageReceivedTooLate {
		return false
	}

	isJobDone, err := wrk.SPoS.RoundConsensus.GetJobDone(node, SrCommitmentHash)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	if isJobDone {
		return false
	}

	return true
}

// checkCommitmentHashConsensus method checks if the consensus in the <COMMITMENT_HASH> subround is achieved
func (wrk *Worker) checkCommitmentHashConsensus() bool {
	wrk.mutCheckConsensus.Lock()
	defer wrk.mutCheckConsensus.Unlock()

	if wrk.SPoS.Chr.IsCancelled() {
		return false
	}

	if wrk.SPoS.Status(SrCommitmentHash) == spos.SsFinished {
		return true
	}

	threshold := wrk.SPoS.Threshold(SrCommitmentHash)

	if !wrk.SPoS.IsSelfLeaderInCurrentRound() {
		threshold = len(wrk.SPoS.ConsensusGroup())
	}

	if wrk.isCommitmentHashReceived(threshold) {
		wrk.printCommitmentHashCM() // only for printing commitment hash consensus messages
		wrk.SPoS.SetStatus(SrCommitmentHash, spos.SsFinished)

		return true
	}

	threshold = wrk.SPoS.Threshold(SrBitmap)

	if wrk.commitmentHashesCollected(threshold) {
		wrk.printCommitmentHashCM() // only for printing commitment hash consensus messages
		wrk.SPoS.SetStatus(SrCommitmentHash, spos.SsFinished)

		return true
	}

	return false
}

// isCommitmentHashReceived method checks if the commitment hashes from the nodes, belonging to the current jobDone
// group, was received in current round
func (wrk *Worker) isCommitmentHashReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		node := wrk.SPoS.ConsensusGroup()[i]
		isJobDone, err := wrk.SPoS.GetJobDone(node, SrCommitmentHash)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isJobDone {
			n++
		}
	}

	return n >= threshold
}

// commitmentHashesCollected method checks if the commitment hashes received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (wrk *Worker) commitmentHashesCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		node := wrk.SPoS.ConsensusGroup()[i]
		isBitmapJobDone, err := wrk.SPoS.GetJobDone(node, SrBitmap)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isBitmapJobDone {
			isCommHashJobDone, err := wrk.SPoS.GetJobDone(node, SrCommitmentHash)

			if err != nil {
				log.Error(err.Error())
				continue
			}

			if !isCommHashJobDone {
				return false
			}
			n++
		}
	}

	return n >= threshold
}

// extendCommitmentHash method put this subround in the extended mode and print some messages
func (wrk *Worker) extendCommitmentHash() {
	wrk.SPoS.SetStatus(SrCommitmentHash, spos.SsExtended)

	if wrk.SPoS.ComputeSize(SrCommitmentHash) < wrk.SPoS.Threshold(SrCommitmentHash) {
		log.Info(fmt.Sprintf("%sStep 2: Extended the (COMMITMENT_HASH) subround. Got only %d from %d commitment hashes which are not enough\n",
			wrk.SPoS.Chr.GetFormattedTime(), wrk.SPoS.ComputeSize(SrCommitmentHash), len(wrk.SPoS.ConsensusGroup())))
	} else {
		log.Info(fmt.Sprintf("%sStep 2: Extended the (COMMITMENT_HASH) subround\n", wrk.SPoS.Chr.GetFormattedTime()))
	}
}

func (wrk *Worker) genCommitmentHash() ([]byte, error) {
	commitmentSecret, commitment, err := wrk.multiSigner.CreateCommitment()

	if err != nil {
		return nil, err
	}

	selfIndex, err := wrk.SPoS.IndexSelfConsensusGroup()

	if err != nil {
		return nil, err
	}

	err = wrk.multiSigner.AddCommitment(uint16(selfIndex), commitment)

	if err != nil {
		return nil, err
	}

	err = wrk.multiSigner.SetCommitmentSecret(commitmentSecret)

	if err != nil {
		return nil, err
	}

	commitmentHash := wrk.hasher.Compute(string(commitment))

	err = wrk.multiSigner.AddCommitmentHash(uint16(selfIndex), commitmentHash)

	if err != nil {
		return nil, err
	}

	return commitmentHash, nil
}

// printCommitmentHashCM method prints the (COMMITMENT_HASH) subround consensus messages
func (wrk *Worker) printCommitmentHashCM() {
	n := wrk.SPoS.ComputeSize(SrCommitmentHash)

	if n == len(wrk.SPoS.ConsensusGroup()) {
		log.Info(fmt.Sprintf("%sStep 2: Received all (%d from %d) commitment hashes\n",
			wrk.SPoS.Chr.GetFormattedTime(), n, len(wrk.SPoS.ConsensusGroup())))
	} else {
		log.Info(fmt.Sprintf("%sStep 2: Received %d from %d commitment hashes, which are enough\n",
			wrk.SPoS.Chr.GetFormattedTime(), n, len(wrk.SPoS.ConsensusGroup())))
	}

	log.Info(fmt.Sprintf("%sStep 2: Subround (COMMITMENT_HASH) has been finished\n", wrk.SPoS.Chr.GetFormattedTime()))
}
