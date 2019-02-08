package bn

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

// doBitmapJob method is the function which is actually used to send the bitmap with the commitment hashes
// received, in the Bitmap subround, when this node is leader (it is used as the handler function of the
// doSubroundJob pointer variable function in Subround struct, from spos package)
func (wrk *Worker) doBitmapJob() bool {
	if wrk.isCommitmentHashSubroundUnfinished() {
		return false
	}

	if !wrk.SPoS.IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	if !wrk.canDoSubroundJob(SrBitmap) {
		return false
	}

	bitmap := wrk.genBitmap(SrCommitmentHash)

	dta := spos.NewConsensusData(
		wrk.SPoS.Data,
		bitmap,
		[]byte(wrk.SPoS.SelfPubKey()),
		nil,
		int(MtBitmap),
		wrk.SPoS.Chr.RoundTimeStamp(),
		wrk.SPoS.Chr.Round().Index())

	if !wrk.sendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 3: Sending bitmap\n", wrk.SPoS.Chr.GetFormattedTime()))

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		pubKey := wrk.SPoS.ConsensusGroup()[i]
		isJobCommHashJobDone, err := wrk.SPoS.GetJobDone(pubKey, SrCommitmentHash)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isJobCommHashJobDone {
			err = wrk.SPoS.SetJobDone(pubKey, SrBitmap, true)

			if err != nil {
				log.Error(err.Error())
				return false
			}
		}
	}

	wrk.Header.PubKeysBitmap = bitmap

	return true
}

func (wrk *Worker) isCommitmentHashSubroundUnfinished() bool {
	isCommitmentHashSubroundUnfinished := wrk.SPoS.Status(SrCommitmentHash) != spos.SsFinished

	if isCommitmentHashSubroundUnfinished {
		if !wrk.doCommitmentHashJob() {
			return true
		}

		if !wrk.checkCommitmentHashConsensus() {
			return true
		}
	}

	return false
}

// receivedBitmap method is called when a bitmap is received through the bitmap channel.
// If the bitmap is valid, than the jobDone map coresponding to the node which sent it,
// is set on true for the subround Bitmap
func (wrk *Worker) receivedBitmap(cnsDta *spos.ConsensusData) bool {
	node := string(cnsDta.PubKey)

	if wrk.isConsensusDataNotSet() {
		return false
	}

	if !wrk.SPoS.IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !wrk.canReceiveMessage(node, cnsDta.RoundIndex, SrBitmap) {
		return false
	}

	signersBitmap := cnsDta.SubRoundData

	// count signers
	nbSigners := countBitmapFlags(signersBitmap)

	if int(nbSigners) < wrk.SPoS.Threshold(SrBitmap) {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, TOO FEW SIGNERS IN BITMAP\n",
			wrk.SPoS.Chr.Round().Index(), getSubroundName(SrBitmap)))

		return false
	}

	publicKeys := wrk.SPoS.ConsensusGroup()

	for i := 0; i < len(publicKeys); i++ {
		byteNb := i / 8
		bitNb := i % 8
		isNodeSigner := (signersBitmap[byteNb] & (1 << uint8(bitNb))) != 0

		if isNodeSigner {
			err := wrk.SPoS.RoundConsensus.SetJobDone(publicKeys[i], SrBitmap, true)

			if err != nil {
				log.Error(err.Error())
				return false
			}
		}
	}

	if !wrk.isSelfInBitmap() {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, NOT INCLUDED IN THE BITMAP\n",
			wrk.SPoS.Chr.Round().Index(), getSubroundName(SrBitmap)))

		wrk.SPoS.Chr.SetSelfSubround(-1)

		wrk.blockProcessor.RevertAccountState()

		return false
	}

	wrk.Header.PubKeysBitmap = signersBitmap

	return true
}

func countBitmapFlags(bitmap []byte) uint16 {
	nbBytes := len(bitmap)
	flags := 0
	for i := 0; i < nbBytes; i++ {
		for j := 0; j < 8; j++ {
			if bitmap[i]&(1<<uint8(j)) != 0 {
				flags++
			}
		}
	}
	return uint16(flags)
}

// isValidatorInBitmap method checks if the node is part of the bitmap received from leader
func (wrk *Worker) isValidatorInBitmap(validator string) bool {
	isJobDone, err := wrk.SPoS.GetJobDone(validator, SrBitmap)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	return isJobDone
}

// isSelfInBitmap method checks if the current node is part of the bitmap received from leader
func (wrk *Worker) isSelfInBitmap() bool {
	return wrk.isValidatorInBitmap(wrk.SPoS.SelfPubKey())
}

// checkBitmapConsensus method checks if the consensus in the <BITMAP> subround is achieved
func (wrk *Worker) checkBitmapConsensus() bool {
	wrk.mutCheckConsensus.Lock()
	defer wrk.mutCheckConsensus.Unlock()

	if wrk.SPoS.Chr.IsCancelled() {
		return false
	}

	if wrk.SPoS.Status(SrBitmap) == spos.SsFinished {
		return true
	}

	threshold := wrk.SPoS.Threshold(SrBitmap)

	//	if wrk.commitmentHashesCollected(threshold) {
	if wrk.isBitmapReceived(threshold) {
		wrk.printBitmapCM() // only for printing bitmap consensus messages
		wrk.SPoS.SetStatus(SrBitmap, spos.SsFinished)

		return true
	}

	return false
}

// isBitmapReceived method checks if the bitmap was received from the leader in current round
func (wrk *Worker) isBitmapReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		node := wrk.SPoS.ConsensusGroup()[i]
		isJobDone, err := wrk.SPoS.GetJobDone(node, SrBitmap)

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

// extendBitmap method put this subround in the extended mode and print some messages
func (wrk *Worker) extendBitmap() {
	wrk.SPoS.SetStatus(SrBitmap, spos.SsExtended)

	log.Info(fmt.Sprintf("%sStep 3: Extended the (BITMAP) subround\n", wrk.SPoS.Chr.GetFormattedTime()))
}

// printBitmapCM method prints the (BITMAP) subround consensus messages
func (wrk *Worker) printBitmapCM() {
	if !wrk.SPoS.IsSelfLeaderInCurrentRound() {
		msg := fmt.Sprintf("%sStep 3: Received bitmap from leader, matching with my own, and it got %d from %d commitment hashes, which are enough",
			wrk.SPoS.Chr.GetFormattedTime(), wrk.SPoS.ComputeSize(SrBitmap), len(wrk.SPoS.ConsensusGroup()))

		if wrk.isSelfInBitmap() {
			msg = fmt.Sprintf("%s, AND I WAS selected in this bitmap\n", msg)
		} else {
			msg = fmt.Sprintf("%s, BUT I WAS NOT selected in this bitmap\n", msg)
		}

		log.Info(msg)
	}

	log.Info(fmt.Sprintf("%sStep 3: Subround (BITMAP) has been finished\n", wrk.SPoS.Chr.GetFormattedTime()))
}
