package bn

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

type subroundBitmap struct {
	*subround

	sendConsensusMessage func(*spos.ConsensusMessage) bool
}

// NewSubroundBitmap creates a subroundBitmap object
func NewSubroundBitmap(
	subround *subround,
	sendConsensusMessage func(*spos.ConsensusMessage) bool,
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

// doBitmapJob method does the job of the bitmap subround
func (sr *subroundBitmap) doBitmapJob() bool {
	if !sr.GetConsensusState().IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	if !sr.GetConsensusState().CanDoSubroundJob(SrBitmap) {
		return false
	}

	bitmap := sr.GetConsensusState().GenerateBitmap(SrCommitmentHash)

	msg := spos.NewConsensusMessage(
		sr.GetConsensusState().Data,
		bitmap,
		[]byte(sr.GetConsensusState().SelfPubKey()),
		nil,
		int(MtBitmap),
		uint64(sr.GetRounder().TimeStamp().Unix()),
		sr.GetRounder().Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 3: bitmap has been sent\n", sr.GetSyncTimer().FormattedCurrentTime()))

	for i := 0; i < len(sr.GetConsensusState().ConsensusGroup()); i++ {
		pubKey := sr.GetConsensusState().ConsensusGroup()[i]
		isJobCommHashJobDone, err := sr.GetConsensusState().JobDone(pubKey, SrCommitmentHash)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isJobCommHashJobDone {
			err = sr.GetConsensusState().SetJobDone(pubKey, SrBitmap, true)

			if err != nil {
				log.Error(err.Error())
				return false
			}
		}
	}

	sr.GetConsensusState().Header.SetPubKeysBitmap(bitmap)

	return true
}

// receivedBitmap method is called when a bitmap is received through the bitmap channel.
// If the bitmap is valid, than the jobDone map corresponding to the node which sent it,
// is set on true for the subround Bitmap
func (sr *subroundBitmap) receivedBitmap(cnsDta *spos.ConsensusMessage) bool {
	node := string(cnsDta.PubKey)

	if !sr.GetConsensusState().IsConsensusDataSet() {
		return false
	}

	if !sr.GetConsensusState().IsConsensusDataEqual(cnsDta.BlockHeaderHash) {
		return false
	}

	if !sr.GetConsensusState().IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !sr.GetConsensusState().CanProcessReceivedMessage(cnsDta, sr.GetRounder().Index(), SrBitmap) {
		return false
	}

	signersBitmap := cnsDta.SubRoundData

	// count signers
	nbSigners := countBitmapFlags(signersBitmap)

	if int(nbSigners) < sr.GetConsensusState().Threshold(SrBitmap) {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, too few signers in bitmap\n",
			sr.GetRounder().Index(), getSubroundName(SrBitmap)))

		return false
	}

	publicKeys := sr.GetConsensusState().ConsensusGroup()

	for i := 0; i < len(publicKeys); i++ {
		byteNb := i / 8
		bitNb := i % 8
		isNodeSigner := (signersBitmap[byteNb] & (1 << uint8(bitNb))) != 0

		if isNodeSigner {
			err := sr.GetConsensusState().SetJobDone(publicKeys[i], SrBitmap, true)

			if err != nil {
				log.Error(err.Error())
				return false
			}
		}
	}

	n := sr.GetConsensusState().ComputeSize(SrBitmap)
	log.Info(fmt.Sprintf("%sStep 3: received bitmap from leader and it got %d from %d commitment hashes\n",
		sr.GetSyncTimer().FormattedCurrentTime(), n, len(sr.GetConsensusState().ConsensusGroup())))

	if !sr.GetConsensusState().IsSelfJobDone(SrBitmap) {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, not included in the bitmap\n",
			sr.GetRounder().Index(), getSubroundName(SrBitmap)))

		sr.GetConsensusState().RoundCanceled = true

		return false
	}

	sr.GetConsensusState().Header.SetPubKeysBitmap(signersBitmap)

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
	if sr.GetConsensusState().RoundCanceled {
		return false
	}

	if sr.GetConsensusState().Status(SrBitmap) == spos.SsFinished {
		return true
	}

	threshold := sr.GetConsensusState().Threshold(SrBitmap)
	if sr.isBitmapReceived(threshold) {
		log.Info(fmt.Sprintf("%sStep 3: subround %s has been finished\n", sr.GetSyncTimer().FormattedCurrentTime(), sr.Name()))
		sr.GetConsensusState().SetStatus(SrBitmap, spos.SsFinished)
		return true
	}

	return false
}

// isBitmapReceived method checks if the bitmap was received from the leader in current round
func (sr *subroundBitmap) isBitmapReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.GetConsensusState().ConsensusGroup()); i++ {
		node := sr.GetConsensusState().ConsensusGroup()[i]
		isJobDone, err := sr.GetConsensusState().JobDone(node, SrBitmap)

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
