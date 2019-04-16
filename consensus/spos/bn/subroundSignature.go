package bn

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

type subroundSignature struct {
	*subround

	sendConsensusMessage func(*spos.ConsensusMessage) bool
}

// NewSubroundSignature creates a subroundSignature object
func NewSubroundSignature(
	subround *subround,
	sendConsensusMessage func(*spos.ConsensusMessage) bool,
	extend func(subroundId int),
) (*subroundSignature, error) {

	err := checkNewSubroundSignatureParams(
		subround,
		sendConsensusMessage,
	)

	if err != nil {
		return nil, err
	}

	srSignature := subroundSignature{
		subround,
		sendConsensusMessage,
	}

	srSignature.job = srSignature.doSignatureJob
	srSignature.check = srSignature.doSignatureConsensusCheck
	srSignature.extend = extend

	return &srSignature, nil
}

func checkNewSubroundSignatureParams(
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

// doSignatureJob method does the job of the signatuure subround
func (sr *subroundSignature) doSignatureJob() bool {
	if !sr.GetConsensusState().IsSelfJobDone(SrBitmap) { // is NOT self in the leader's bitmap?
		return false
	}

	if !sr.GetConsensusState().CanDoSubroundJob(SrSignature) {
		return false
	}

	bitmap := sr.GetConsensusState().GenerateBitmap(SrBitmap)

	err := sr.checkCommitmentsValidity(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	// first compute commitment aggregation
	err = sr.GetMultiSigner().AggregateCommitments(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	sigPart, err := sr.GetMultiSigner().CreateSignatureShare(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	msg := spos.NewConsensusMessage(
		sr.GetConsensusState().Data,
		sigPart,
		[]byte(sr.GetConsensusState().SelfPubKey()),
		nil,
		int(MtSignature),
		uint64(sr.GetRounder().TimeStamp().Unix()),
		sr.GetRounder().Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 5: signature has been sent\n", sr.GetSyncTimer().FormattedCurrentTime()))

	err = sr.GetConsensusState().SetSelfJobDone(SrSignature, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

func (sr *subroundSignature) checkCommitmentsValidity(bitmap []byte) error {
	nbBitsBitmap := len(bitmap) * 8
	consensusGroup := sr.GetConsensusState().ConsensusGroup()
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
		isCommJobDone, err := sr.GetConsensusState().JobDone(pubKey, SrCommitment)

		if err != nil {
			return err
		}

		if !isCommJobDone {
			return spos.ErrNilCommitment
		}

		commitment, err := sr.GetMultiSigner().Commitment(uint16(i))

		if err != nil {
			return err
		}

		computedCommitmentHash := sr.GetHasher().Compute(string(commitment))
		receivedCommitmentHash, err := sr.GetMultiSigner().CommitmentHash(uint16(i))

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
func (sr *subroundSignature) receivedSignature(cnsDta *spos.ConsensusMessage) bool {
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

	if !sr.GetConsensusState().CanProcessReceivedMessage(cnsDta, sr.GetRounder().Index(), SrSignature) {
		return false
	}

	index, err := sr.GetConsensusState().ConsensusGroupIndex(node)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sr.GetMultiSigner().StoreSignatureShare(uint16(index), cnsDta.SubRoundData)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sr.GetConsensusState().SetJobDone(node, SrSignature, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	threshold := sr.GetConsensusState().Threshold(SrSignature)
	if sr.signaturesCollected(threshold) {
		n := sr.GetConsensusState().ComputeSize(SrSignature)
		log.Info(fmt.Sprintf("%sStep 5: received %d from %d signatures\n",
			sr.GetSyncTimer().FormattedCurrentTime(), n, len(sr.GetConsensusState().ConsensusGroup())))
	}

	return true
}

// doSignatureConsensusCheck method checks if the consensus in the <SIGNATURE> subround is achieved
func (sr *subroundSignature) doSignatureConsensusCheck() bool {
	if sr.GetConsensusState().RoundCanceled {
		return false
	}

	if sr.GetConsensusState().Status(SrSignature) == spos.SsFinished {
		return true
	}

	threshold := sr.GetConsensusState().Threshold(SrSignature)
	if sr.signaturesCollected(threshold) {
		log.Info(fmt.Sprintf("%sStep 5: subround %s has been finished\n", sr.GetSyncTimer().FormattedCurrentTime(), sr.Name()))
		sr.GetConsensusState().SetStatus(SrSignature, spos.SsFinished)
		return true
	}

	return false
}

// signaturesCollected method checks if the signatures received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (sr *subroundSignature) signaturesCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.GetConsensusState().ConsensusGroup()); i++ {
		node := sr.GetConsensusState().ConsensusGroup()[i]
		isBitmapJobDone, err := sr.GetConsensusState().JobDone(node, SrBitmap)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isBitmapJobDone {
			isSignJobDone, err := sr.GetConsensusState().JobDone(node, SrSignature)

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
