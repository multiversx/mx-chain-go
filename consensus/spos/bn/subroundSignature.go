package bn

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

// doSignatureJob method is the function which is actually used to send the Signature for the received block,
// in the Signature subround (it is used as the handler function of the doSubroundJob pointer variable function
// in Subround struct, from spos package)
func (wrk *Worker) doSignatureJob() bool {
	if wrk.isCommitmentSubroundUnfinished() {
		return false
	}

	if !wrk.isSelfInBitmap() { // is NOT self in the leader's bitmap?
		return false
	}

	if !wrk.canDoSubroundJob(SrSignature) {
		return false
	}

	bitmap := wrk.genBitmap(SrBitmap)

	err := wrk.checkCommitmentsValidity(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	// first compute commitment aggregation
	aggComm, err := wrk.multiSigner.AggregateCommitments(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	sigPart, err := wrk.multiSigner.CreateSignatureShare(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	dta := spos.NewConsensusData(
		wrk.SPoS.Data,
		sigPart,
		[]byte(wrk.SPoS.SelfPubKey()),
		nil,
		int(MtSignature),
		wrk.SPoS.Chr.RoundTimeStamp(),
		wrk.SPoS.Chr.Round().Index())

	if !wrk.sendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 5: Sending signature\n", wrk.SPoS.Chr.GetFormattedTime()))

	selfIndex, err := wrk.SPoS.IndexSelfConsensusGroup()

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = wrk.multiSigner.AddSignatureShare(uint16(selfIndex), sigPart)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = wrk.SPoS.SetSelfJobDone(SrSignature, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	wrk.Header.Commitment = aggComm

	return true
}

func (wrk *Worker) isCommitmentSubroundUnfinished() bool {
	isCommitmentSubroundUnfinished := wrk.SPoS.Status(SrCommitment) != spos.SsFinished

	if isCommitmentSubroundUnfinished {
		if !wrk.doCommitmentJob() {
			return true
		}

		if !wrk.checkCommitmentConsensus() {
			return true
		}
	}

	return false
}

func (wrk *Worker) checkCommitmentsValidity(bitmap []byte) error {
	nbBitsBitmap := len(bitmap) * 8
	consensusGroup := wrk.SPoS.ConsensusGroup()
	consensusGroupSize := len(consensusGroup)
	size := consensusGroupSize

	if consensusGroupSize > nbBitsBitmap {
		size = nbBitsBitmap
	}

	for i := 0; i < size; i++ {
		indexRequired := (bitmap[i/8] & (1 << uint16(i%8))) > 0

		if !indexRequired {
			continue
		}

		pubKey := consensusGroup[i]
		isCommJobDone, err := wrk.SPoS.GetJobDone(pubKey, SrCommitment)

		if err != nil {
			return err
		}

		if !isCommJobDone {
			return spos.ErrNilCommitment
		}

		commitment, err := wrk.multiSigner.Commitment(uint16(i))

		if err != nil {
			return err
		}

		computedCommitmentHash := wrk.hasher.Compute(string(commitment))
		receivedCommitmentHash, err := wrk.multiSigner.CommitmentHash(uint16(i))

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
// If the Signature is valid, than the jobDone map coresponding to the node which sent it,
// is set on true for the subround Signature
func (wrk *Worker) receivedSignature(cnsDta *spos.ConsensusData) bool {
	node := string(cnsDta.PubKey)

	if wrk.isConsensusDataNotSet() {
		return false
	}

	if !wrk.isValidatorInBitmap(node) { // is NOT this node in the bitmap group?
		return false
	}

	if !wrk.canReceiveMessage(node, cnsDta.RoundIndex, SrSignature) {
		return false
	}

	index, err := wrk.SPoS.ConsensusGroupIndex(node)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = wrk.multiSigner.AddSignatureShare(uint16(index), cnsDta.SubRoundData)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = wrk.SPoS.RoundConsensus.SetJobDone(node, SrSignature, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

// checkSignatureConsensus method checks if the consensus in the <SIGNATURE> subround is achieved
func (wrk *Worker) checkSignatureConsensus() bool {
	wrk.mutCheckConsensus.Lock()
	defer wrk.mutCheckConsensus.Unlock()

	if wrk.SPoS.Chr.IsCancelled() {
		return false
	}

	if wrk.SPoS.Status(SrSignature) == spos.SsFinished {
		return true
	}

	threshold := wrk.SPoS.Threshold(SrSignature)

	if wrk.signaturesCollected(threshold) {
		wrk.printSignatureCM() // only for printing signature consensus messages
		wrk.SPoS.SetStatus(SrSignature, spos.SsFinished)

		return true
	}

	return false
}

// signaturesCollected method checks if the signatures received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (wrk *Worker) signaturesCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		node := wrk.SPoS.ConsensusGroup()[i]
		isBitmapJobDone, err := wrk.SPoS.GetJobDone(node, SrBitmap)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isBitmapJobDone {
			isSignJobDone, err := wrk.SPoS.GetJobDone(node, SrSignature)

			if err != nil {
				log.Error(err.Error())
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

// extendSignature method put this subround in the extended mode and print some messages
func (wrk *Worker) extendSignature() {
	wrk.SPoS.SetStatus(SrSignature, spos.SsExtended)

	log.Info(fmt.Sprintf("%sStep 5: Extended the (SIGNATURE) subround. Got only %d from %d signatures which are not enough\n",
		wrk.SPoS.Chr.GetFormattedTime(), wrk.SPoS.ComputeSize(SrSignature), len(wrk.SPoS.ConsensusGroup())))
}

// printSignatureCM method prints the (SIGNATURE) subround consensus messages
func (wrk *Worker) printSignatureCM() {
	log.Info(fmt.Sprintf("%sStep 5: Received %d from %d signatures, which are matching with bitmap and are enough\n",
		wrk.SPoS.Chr.GetFormattedTime(), wrk.SPoS.ComputeSize(SrSignature), len(wrk.SPoS.ConsensusGroup())))

	log.Info(fmt.Sprintf("%sStep 5: Subround (SIGNATURE) has been finished\n", wrk.SPoS.Chr.GetFormattedTime()))
}
