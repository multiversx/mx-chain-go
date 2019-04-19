package bn

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

type subroundBitmap struct {
	*subround

	sendConsensusMessage func(*consensus.Message) bool
}

// NewSubroundBitmap creates a subroundBitmap object
func NewSubroundBitmap(
	subround *subround,
	sendConsensusMessage func(*consensus.Message) bool,
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

// doBitmapJob method does the job of the bitmap subround
func (sr *subroundBitmap) doBitmapJob() bool {
	if !sr.IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	if !sr.CanDoSubroundJob(SrBitmap) {
		return false
	}

	bitmap := sr.GenerateBitmap(SrCommitmentHash)

	msg := consensus.NewConsensusMessage(
		sr.Data,
		bitmap,
		[]byte(sr.SelfPubKey()),
		nil,
		int(MtBitmap),
		uint64(sr.Rounder().TimeStamp().Unix()),
		sr.Rounder().Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 3: bitmap has been sent\n", sr.SyncTimer().FormattedCurrentTime()))

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		pubKey := sr.ConsensusGroup()[i]
		isJobCommHashJobDone, err := sr.JobDone(pubKey, SrCommitmentHash)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isJobCommHashJobDone {
			err = sr.SetJobDone(pubKey, SrBitmap, true)

			if err != nil {
				log.Error(err.Error())
				return false
			}
		}
	}

	sr.Header.SetPubKeysBitmap(bitmap)

	return true
}

// receivedBitmap method is called when a bitmap is received through the bitmap channel.
// If the bitmap is valid, than the jobDone map corresponding to the node which sent it,
// is set on true for the subround Bitmap
func (sr *subroundBitmap) receivedBitmap(cnsDta *consensus.Message) bool {
	node := string(cnsDta.PubKey)

	if !sr.IsConsensusDataSet() {
		return false
	}

	if !sr.IsConsensusDataEqual(cnsDta.BlockHeaderHash) {
		return false
	}

	if !sr.IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.Rounder().Index(), SrBitmap) {
		return false
	}

	signersBitmap := cnsDta.SubRoundData

	// count signers
	nbSigners := countBitmapFlags(signersBitmap)

	if int(nbSigners) < sr.Threshold(SrBitmap) {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, too few signers in bitmap\n",
			sr.Rounder().Index(), getSubroundName(SrBitmap)))

		return false
	}

	publicKeys := sr.ConsensusGroup()

	for i := 0; i < len(publicKeys); i++ {
		byteNb := i / 8
		bitNb := i % 8
		isNodeSigner := (signersBitmap[byteNb] & (1 << uint8(bitNb))) != 0

		if isNodeSigner {
			err := sr.SetJobDone(publicKeys[i], SrBitmap, true)
			if err != nil {
				log.Error(err.Error())
				return false
			}
		}
	}

	n := sr.ComputeSize(SrBitmap)
	log.Info(fmt.Sprintf("%sStep 3: received bitmap from leader and it got %d from %d commitment hashes\n",
		sr.SyncTimer().FormattedCurrentTime(), n, len(sr.ConsensusGroup())))

	if !sr.IsSelfJobDone(SrBitmap) {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, not included in the bitmap\n",
			sr.Rounder().Index(), getSubroundName(SrBitmap)))

		sr.RoundCanceled = true

		return false
	}

	sr.Header.SetPubKeysBitmap(signersBitmap)

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
	if sr.RoundCanceled {
		return false
	}

	if sr.Status(SrBitmap) == spos.SsFinished {
		return true
	}

	threshold := sr.Threshold(SrBitmap)
	if sr.isBitmapReceived(threshold) {
		log.Info(fmt.Sprintf("%sStep 3: subround %s has been finished\n", sr.SyncTimer().FormattedCurrentTime(), sr.Name()))
		sr.SetStatus(SrBitmap, spos.SsFinished)
		return true
	}

	return false
}

// isBitmapReceived method checks if the bitmap was received from the leader in current round
func (sr *subroundBitmap) isBitmapReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		node := sr.ConsensusGroup()[i]
		isJobDone, err := sr.JobDone(node, SrBitmap)

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
