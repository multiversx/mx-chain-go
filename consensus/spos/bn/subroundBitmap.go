package bn

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

type subroundBitmap struct {
	*subround

	sendConsensusMessage func(*consensus.ConsensusMessage) bool
}

// NewSubroundBitmap creates a subroundBitmap object
func NewSubroundBitmap(
	subround *subround,
	sendConsensusMessage func(*consensus.ConsensusMessage) bool,
	extend func(subroundId int),
) (*subroundBitmap, error) {

	err := checkNewSubroundBitmapParams(
		subround,
		sendConsensusMessage,
	)

	if err != nil {
		return nil, err
	}

	srBitmap := subroundBitmap{
		subround,
		sendConsensusMessage,
	}

	srBitmap.job = srBitmap.doBitmapJob
	srBitmap.check = srBitmap.doBitmapConsensusCheck
	srBitmap.extend = extend

	return &srBitmap, nil
}

func checkNewSubroundBitmapParams(
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

// doBitmapJob method does the job of the bitmap subround
func (sr *subroundBitmap) doBitmapJob() bool {
	if !sr.ConsensusState().IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	if !sr.ConsensusState().CanDoSubroundJob(SrBitmap) {
		return false
	}

	bitmap := sr.ConsensusState().GenerateBitmap(SrCommitmentHash)

	msg := consensus.NewConsensusMessage(
		sr.ConsensusState().Data,
		bitmap,
		[]byte(sr.ConsensusState().SelfPubKey()),
		nil,
		int(MtBitmap),
		uint64(sr.Rounder().TimeStamp().Unix()),
		sr.Rounder().Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 3: bitmap has been sent\n", sr.SyncTimer().FormattedCurrentTime()))

	for i := 0; i < len(sr.ConsensusState().ConsensusGroup()); i++ {
		pubKey := sr.ConsensusState().ConsensusGroup()[i]
		isJobCommHashJobDone, err := sr.ConsensusState().JobDone(pubKey, SrCommitmentHash)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isJobCommHashJobDone {
			err = sr.ConsensusState().SetJobDone(pubKey, SrBitmap, true)

			if err != nil {
				log.Error(err.Error())
				return false
			}
		}
	}

	sr.ConsensusState().Header.SetPubKeysBitmap(bitmap)

	return true
}

// receivedBitmap method is called when a bitmap is received through the bitmap channel.
// If the bitmap is valid, than the jobDone map corresponding to the node which sent it,
// is set on true for the subround Bitmap
func (sr *subroundBitmap) receivedBitmap(cnsDta *consensus.ConsensusMessage) bool {
	node := string(cnsDta.PubKey)

	if !sr.ConsensusState().IsConsensusDataSet() {
		return false
	}

	if !sr.ConsensusState().IsConsensusDataEqual(cnsDta.BlockHeaderHash) {
		return false
	}

	if !sr.ConsensusState().IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !sr.ConsensusState().CanProcessReceivedMessage(cnsDta, sr.Rounder().Index(), SrBitmap) {
		return false
	}

	signersBitmap := cnsDta.SubRoundData

	// count signers
	nbSigners := countBitmapFlags(signersBitmap)

	if int(nbSigners) < sr.ConsensusState().Threshold(SrBitmap) {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, too few signers in bitmap\n",
			sr.Rounder().Index(), getSubroundName(SrBitmap)))

		return false
	}

	publicKeys := sr.ConsensusState().ConsensusGroup()

	for i := 0; i < len(publicKeys); i++ {
		byteNb := i / 8
		bitNb := i % 8
		isNodeSigner := (signersBitmap[byteNb] & (1 << uint8(bitNb))) != 0

		if isNodeSigner {
			err := sr.ConsensusState().SetJobDone(publicKeys[i], SrBitmap, true)

			if err != nil {
				log.Error(err.Error())
				return false
			}
		}
	}

	n := sr.ConsensusState().ComputeSize(SrBitmap)
	log.Info(fmt.Sprintf("%sStep 3: received bitmap from leader and it got %d from %d commitment hashes\n",
		sr.SyncTimer().FormattedCurrentTime(), n, len(sr.ConsensusState().ConsensusGroup())))

	if !sr.ConsensusState().IsSelfJobDone(SrBitmap) {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, not included in the bitmap\n",
			sr.Rounder().Index(), getSubroundName(SrBitmap)))

		sr.ConsensusState().RoundCanceled = true

		return false
	}

	sr.ConsensusState().Header.SetPubKeysBitmap(signersBitmap)

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

// doBitmapConsensusCheck method checks if the consensus in the <BITMAP> subround is achieved
func (sr *subroundBitmap) doBitmapConsensusCheck() bool {
	if sr.ConsensusState().RoundCanceled {
		return false
	}

	if sr.ConsensusState().Status(SrBitmap) == spos.SsFinished {
		return true
	}

	threshold := sr.ConsensusState().Threshold(SrBitmap)
	if sr.isBitmapReceived(threshold) {
		log.Info(fmt.Sprintf("%sStep 3: subround %s has been finished\n", sr.SyncTimer().FormattedCurrentTime(), sr.Name()))
		sr.ConsensusState().SetStatus(SrBitmap, spos.SsFinished)
		return true
	}

	return false
}

// isBitmapReceived method checks if the bitmap was received from the leader in current round
func (sr *subroundBitmap) isBitmapReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.ConsensusState().ConsensusGroup()); i++ {
		node := sr.ConsensusState().ConsensusGroup()[i]
		isJobDone, err := sr.ConsensusState().JobDone(node, SrBitmap)

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
