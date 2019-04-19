package bn

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

type subroundSignature struct {
	*subround

	sendConsensusMessage func(*consensus.Message) bool
}

// NewSubroundSignature creates a subroundSignature object
func NewSubroundSignature(
	subround *subround,
	sendConsensusMessage func(*consensus.Message) bool,
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
	sendConsensusMessage func(*consensus.Message) bool,
) error {
	if subround == nil {
		return spos.ErrNilSubround
	}

	if subround.ConsensusState == nil {
		return spos.ErrNilConsensusState
	}

	if sendConsensusMessage == nil {
		return spos.ErrNilSendConsensusMessageFunction
	}

	err := spos.ValidateConsensusCore(subround.ConsensusCoreHandler)

	return err
}

// doSignatureJob method does the job of the signatuure subround
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
		log.Error(err.Error())
		return false
	}

	// first compute commitment aggregation
	err = sr.MultiSigner().AggregateCommitments(bitmap)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	sigPart, err := sr.MultiSigner().CreateSignatureShare(bitmap)
	if err != nil {
		log.Error(err.Error())
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

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 5: signature has been sent\n", sr.SyncTimer().FormattedCurrentTime()))

	err = sr.SetSelfJobDone(SrSignature, true)
	if err != nil {
		log.Error(err.Error())
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

		commitment, err := sr.MultiSigner().Commitment(uint16(i))
		if err != nil {
			return err
		}

		computedCommitmentHash := sr.Hasher().Compute(string(commitment))
		receivedCommitmentHash, err := sr.MultiSigner().CommitmentHash(uint16(i))

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
		log.Error(err.Error())
		return false
	}

	currentMultiSigner := sr.MultiSigner()
	err = currentMultiSigner.StoreSignatureShare(uint16(index), cnsDta.SubRoundData)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sr.SetJobDone(node, SrSignature, true)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	threshold := sr.Threshold(SrSignature)
	if sr.signaturesCollected(threshold) {
		n := sr.ComputeSize(SrSignature)
		log.Info(fmt.Sprintf("%sStep 5: received %d from %d signatures\n",
			sr.SyncTimer().FormattedCurrentTime(), n, len(sr.ConsensusGroup())))
	}

	return true
}

// doSignatureConsensusCheck method checks if the consensus in the <SIGNATURE> subround is achieved
func (sr *subroundSignature) doSignatureConsensusCheck() bool {
	if sr.RoundCanceled {
		return false
	}

	if sr.Status(SrSignature) == spos.SsFinished {
		return true
	}

	threshold := sr.Threshold(SrSignature)
	if sr.signaturesCollected(threshold) {
		log.Info(fmt.Sprintf("%sStep 5: subround %s has been finished\n", sr.SyncTimer().FormattedCurrentTime(), sr.Name()))
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
			log.Error(err.Error())
			continue
		}

		if isBitmapJobDone {
			isSignJobDone, err := sr.JobDone(node, SrSignature)

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
