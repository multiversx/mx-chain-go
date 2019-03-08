package bn

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type subroundBitmap struct {
	*subround

	blockProcessor process.BlockProcessor
	consensusState *spos.ConsensusState
	rounder        consensus.Rounder
	syncTimer      ntp.SyncTimer

	sendConsensusMessage func(*spos.ConsensusMessage) bool
}

// NewSubroundBitmap creates a subroundBitmap object
func NewSubroundBitmap(
	subround *subround,
	blockProcessor process.BlockProcessor,
	consensusState *spos.ConsensusState,
	rounder consensus.Rounder,
	syncTimer ntp.SyncTimer,
	sendConsensusMessage func(*spos.ConsensusMessage) bool,
	extend func(subroundId int),
) (*subroundBitmap, error) {

	err := checkNewSubroundBitmapParams(
		subround,
		blockProcessor,
		consensusState,
		rounder,
		syncTimer,
		sendConsensusMessage,
	)

	if err != nil {
		return nil, err
	}

	srBitmap := subroundBitmap{
		subround,
		blockProcessor,
		consensusState,
		rounder,
		syncTimer,
		sendConsensusMessage,
	}

	srBitmap.job = srBitmap.doBitmapJob
	srBitmap.check = srBitmap.doBitmapConsensusCheck
	srBitmap.extend = extend

	return &srBitmap, nil
}

func checkNewSubroundBitmapParams(
	subround *subround,
	blockProcessor process.BlockProcessor,
	consensusState *spos.ConsensusState,
	rounder consensus.Rounder,
	syncTimer ntp.SyncTimer,
	sendConsensusMessage func(*spos.ConsensusMessage) bool,
) error {
	if subround == nil {
		return spos.ErrNilSubround
	}

	if blockProcessor == nil {
		return spos.ErrNilBlockProcessor
	}

	if consensusState == nil {
		return spos.ErrNilConsensusState
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

// doBitmapJob method does the job of the bitmap subround
func (sr *subroundBitmap) doBitmapJob() bool {
	if !sr.consensusState.IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	if !sr.consensusState.CanDoSubroundJob(SrBitmap) {
		return false
	}

	bitmap := sr.consensusState.GenerateBitmap(SrCommitmentHash)

	msg := spos.NewConsensusMessage(
		sr.consensusState.Data,
		bitmap,
		[]byte(sr.consensusState.SelfPubKey()),
		nil,
		int(MtBitmap),
		uint64(sr.rounder.TimeStamp().Unix()),
		sr.rounder.Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 3: bitmap has been sent\n", sr.syncTimer.FormattedCurrentTime()))

	for i := 0; i < len(sr.consensusState.ConsensusGroup()); i++ {
		pubKey := sr.consensusState.ConsensusGroup()[i]
		isJobCommHashJobDone, err := sr.consensusState.JobDone(pubKey, SrCommitmentHash)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isJobCommHashJobDone {
			err = sr.consensusState.SetJobDone(pubKey, SrBitmap, true)

			if err != nil {
				log.Error(err.Error())
				return false
			}
		}
	}

	sr.consensusState.Header.SetPubKeysBitmap(bitmap)

	return true
}

// receivedBitmap method is called when a bitmap is received through the bitmap channel.
// If the bitmap is valid, than the jobDone map corresponding to the node which sent it,
// is set on true for the subround Bitmap
func (sr *subroundBitmap) receivedBitmap(cnsDta *spos.ConsensusMessage) bool {
	node := string(cnsDta.PubKey)

	if !sr.consensusState.IsConsensusDataSet() {
		return false
	}

	if !sr.consensusState.IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !sr.consensusState.CanProcessReceivedMessage(cnsDta, sr.rounder.Index(), SrBitmap) {
		return false
	}

	signersBitmap := cnsDta.SubRoundData

	// count signers
	nbSigners := countBitmapFlags(signersBitmap)

	if int(nbSigners) < sr.consensusState.Threshold(SrBitmap) {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, too few signers in bitmap\n",
			sr.rounder.Index(), getSubroundName(SrBitmap)))

		return false
	}

	publicKeys := sr.consensusState.ConsensusGroup()

	for i := 0; i < len(publicKeys); i++ {
		byteNb := i / 8
		bitNb := i % 8
		isNodeSigner := (signersBitmap[byteNb] & (1 << uint8(bitNb))) != 0

		if isNodeSigner {
			err := sr.consensusState.SetJobDone(publicKeys[i], SrBitmap, true)

			if err != nil {
				log.Error(err.Error())
				return false
			}
		}
	}

	if !sr.consensusState.IsSelfJobDone(SrBitmap) {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, not included in the bitmap\n",
			sr.rounder.Index(), getSubroundName(SrBitmap)))

		sr.consensusState.RoundCanceled = true

		sr.blockProcessor.RevertAccountState()

		return false
	}

	sr.consensusState.Header.SetPubKeysBitmap(signersBitmap)

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
	if sr.consensusState.RoundCanceled {
		return false
	}

	if sr.consensusState.Status(SrBitmap) == spos.SsFinished {
		return true
	}

	threshold := sr.consensusState.Threshold(SrBitmap)

	if sr.isBitmapReceived(threshold) {
		if !sr.consensusState.IsSelfLeaderInCurrentRound() {
			msg := fmt.Sprintf("%sStep 3: received bitmap from leader, matching with my own, and it got %d from %d commitment hashes, which are enough",
				sr.syncTimer.FormattedCurrentTime(), sr.consensusState.ComputeSize(SrBitmap), len(sr.consensusState.ConsensusGroup()))

			if sr.consensusState.IsSelfJobDone(SrBitmap) {
				msg = fmt.Sprintf("%s, and i was selected in this bitmap\n", msg)
			} else {
				msg = fmt.Sprintf("%s, and i was not selected in this bitmap\n", msg)
			}

			log.Info(msg)
		}

		log.Info(fmt.Sprintf("%sStep 3: subround %s has been finished\n", sr.syncTimer.FormattedCurrentTime(), sr.Name()))

		sr.consensusState.SetStatus(SrBitmap, spos.SsFinished)

		return true
	}

	return false
}

// isBitmapReceived method checks if the bitmap was received from the leader in current round
func (sr *subroundBitmap) isBitmapReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.consensusState.ConsensusGroup()); i++ {
		node := sr.consensusState.ConsensusGroup()[i]
		isJobDone, err := sr.consensusState.JobDone(node, SrBitmap)

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
