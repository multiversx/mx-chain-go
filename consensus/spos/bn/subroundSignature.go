package bn

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
)

type subroundSignature struct {
	*subround

	consensusState *spos.ConsensusState
	hasher         hashing.Hasher
	multiSigner    crypto.MultiSigner
	rounder        round.Rounder
	syncTimer      ntp.SyncTimer

	sendConsensusMessage func(*spos.ConsensusData) bool
}

// NewSubroundSignature creates a subroundSignature object
func NewSubroundSignature(
	subround *subround,
	consensusState *spos.ConsensusState,
	hasher hashing.Hasher,
	multiSigner crypto.MultiSigner,
	rounder round.Rounder,
	syncTimer ntp.SyncTimer,
	sendConsensusMessage func(*spos.ConsensusData) bool,
	extend func(subroundId int),
) (*subroundSignature, error) {

	err := checkNewSubroundSignatureParams(
		subround,
		consensusState,
		hasher,
		multiSigner,
		rounder,
		syncTimer,
		sendConsensusMessage,
	)

	if err != nil {
		return nil, err
	}

	srSignature := subroundSignature{
		subround,
		consensusState,
		hasher,
		multiSigner,
		rounder,
		syncTimer,
		sendConsensusMessage,
	}

	srSignature.job = srSignature.doSignatureJob
	srSignature.check = srSignature.doSignatureConsensusCheck
	srSignature.extend = extend

	return &srSignature, nil
}

func checkNewSubroundSignatureParams(
	subround *subround,
	consensusState *spos.ConsensusState,
	hasher hashing.Hasher,
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

	if hasher == nil {
		return spos.ErrNilHasher
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

// doSignatureJob method is the function which is actually used to job the Signature for the received block,
// in the Signature subround (it is used as the handler function of the doSubroundJob pointer variable function
// in subround struct, from spos package)
func (sr *subroundSignature) doSignatureJob() bool {
	if !sr.consensusState.IsSelfJobDone(SrBitmap) { // is NOT self in the leader's bitmap?
		return false
	}

	if !sr.consensusState.CanDoSubroundJob(SrSignature) {
		return false
	}

	bitmap := sr.consensusState.GenerateBitmap(SrBitmap)

	err := sr.checkCommitmentsValidity(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	// first compute commitment aggregation
	aggComm, err := sr.multiSigner.AggregateCommitments(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	sigPart, err := sr.multiSigner.CreateSignatureShare(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	dta := spos.NewConsensusData(
		sr.consensusState.Data,
		sigPart,
		[]byte(sr.consensusState.SelfPubKey()),
		nil,
		int(MtSignature),
		uint64(sr.rounder.TimeStamp().Unix()),
		sr.rounder.Index())

	if !sr.sendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 5: Sending signature\n", sr.syncTimer.FormattedCurrentTime()))

	selfIndex, err := sr.consensusState.IndexSelfConsensusGroup()

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sr.multiSigner.AddSignatureShare(uint16(selfIndex), sigPart)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sr.consensusState.SetSelfJobDone(SrSignature, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	sr.consensusState.Header.Commitment = aggComm

	return true
}

func (sr *subroundSignature) checkCommitmentsValidity(bitmap []byte) error {
	nbBitsBitmap := len(bitmap) * 8
	consensusGroup := sr.consensusState.ConsensusGroup()
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
		isCommJobDone, err := sr.consensusState.GetJobDone(pubKey, SrCommitment)

		if err != nil {
			return err
		}

		if !isCommJobDone {
			return spos.ErrNilCommitment
		}

		commitment, err := sr.multiSigner.Commitment(uint16(i))

		if err != nil {
			return err
		}

		computedCommitmentHash := sr.hasher.Compute(string(commitment))
		receivedCommitmentHash, err := sr.multiSigner.CommitmentHash(uint16(i))

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
func (sr *subroundSignature) receivedSignature(cnsDta *spos.ConsensusData) bool {
	node := string(cnsDta.PubKey)

	if sr.consensusState.IsConsensusDataNotSet() {
		return false
	}

	if !sr.consensusState.IsJobDone(node, SrBitmap) { // is NOT this node in the bitmap group?
		return false
	}

	if !sr.consensusState.CanProcessReceivedMessage(cnsDta, sr.rounder.Index(), SrSignature) {
		return false
	}

	index, err := sr.consensusState.ConsensusGroupIndex(node)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sr.multiSigner.AddSignatureShare(uint16(index), cnsDta.SubRoundData)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	err = sr.consensusState.SetJobDone(node, SrSignature, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

// doSignatureConsensusCheck method checks if the consensus in the <SIGNATURE> subround is achieved
func (sr *subroundSignature) doSignatureConsensusCheck() bool {
	if sr.consensusState.RoundCanceled {
		return false
	}

	if sr.consensusState.Status(SrSignature) == spos.SsFinished {
		return true
	}

	threshold := sr.consensusState.Threshold(SrSignature)

	if sr.signaturesCollected(threshold) {
		sr.printSignatureCM() // only for printing signature consensus messages
		sr.consensusState.SetStatus(SrSignature, spos.SsFinished)

		return true
	}

	return false
}

// signaturesCollected method checks if the signatures received from the nodes, belonging to the current
// jobDone group, are covering the bitmap received from the leader in the current round
func (sr *subroundSignature) signaturesCollected(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.consensusState.ConsensusGroup()); i++ {
		node := sr.consensusState.ConsensusGroup()[i]
		isBitmapJobDone, err := sr.consensusState.GetJobDone(node, SrBitmap)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isBitmapJobDone {
			isSignJobDone, err := sr.consensusState.GetJobDone(node, SrSignature)

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

// printSignatureCM method prints the (SIGNATURE) subround consensus messages
func (sr *subroundSignature) printSignatureCM() {
	log.Info(fmt.Sprintf("%sStep 5: Received %d from %d signatures, which are matching with bitmap and are enough\n",
		sr.syncTimer.FormattedCurrentTime(), sr.consensusState.ComputeSize(SrSignature), len(sr.consensusState.ConsensusGroup())))

	log.Info(fmt.Sprintf("%sStep 5: subround (SIGNATURE) has been finished\n", sr.syncTimer.FormattedCurrentTime()))
}
