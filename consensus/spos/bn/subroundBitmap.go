package bn

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
)

type subroundBitmap struct {
	*spos.Subround
}

// NewSubroundBitmap creates a subroundBitmap object
func NewSubroundBitmap(
	baseSubround *spos.Subround,
	extend func(subroundId int),
) (*subroundBitmap, error) {
	err := checkNewSubroundBitmapParams(
		baseSubround,
	)
	if err != nil {
		return nil, err
	}

	srBitmap := subroundBitmap{
		baseSubround,
	}
	srBitmap.Job = srBitmap.doBitmapJob
	srBitmap.Check = srBitmap.doBitmapConsensusCheck
	srBitmap.Extend = extend

	return &srBitmap, nil
}

func checkNewSubroundBitmapParams(
	baseSubround *spos.Subround,
) error {
	if baseSubround == nil {
		return spos.ErrNilSubround
	}

	if baseSubround.ConsensusState == nil {
		return spos.ErrNilConsensusState
	}

	err := spos.ValidateConsensusCore(baseSubround.ConsensusCoreHandler)

	return err
}

// doBitmapJob method does the job of the subround Bitmap
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

	err := sr.BroadcastMessenger().BroadcastConsensusMessage(msg)
	if err != nil {
		log.Debug("BroadcastConsensusMessage",
			"type", "spos/bn",
			"error", err.Error())
		return false
	}

	log.Debug("step 3: bitmap has been sent",
		"type", "spos/bn",
		"time [s]", sr.SyncTimer().FormattedCurrentTime())

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		pubKey := sr.ConsensusGroup()[i]
		isJobCommHashJobDone, err := sr.JobDone(pubKey, SrCommitmentHash)
		if err != nil {
			log.Debug("spos/bn JobDone", "error", err.Error())
			continue
		}

		if isJobCommHashJobDone {
			err = sr.SetJobDone(pubKey, SrBitmap, true)
			if err != nil {
				log.Debug("SetJobDone", "type", "spos/bn", "error", err.Error())
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
		log.Debug("canceled round, too few signers in bitmap",
			"type", "spos/bn",
			"round", sr.Rounder().Index(),
			"subround", getSubroundName(SrBitmap))

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
				log.Debug("SetJobDone", "type", "spos/bn", "error", err.Error())
				return false
			}
		}
	}

	n := sr.ComputeSize(SrBitmap)
	log.Debug("step 3: received commitment hashes",
		"type", "spos/bn",
		"time [s]", sr.SyncTimer().FormattedCurrentTime(),
		"received", n,
		"total", len(sr.ConsensusGroup()))

	if !sr.IsSelfJobDone(SrBitmap) {
		log.Debug("canceled round, not included in the bitmap",
			"type", "spos/bn",
			"round", sr.Rounder().Index(),
			"subround", getSubroundName(SrBitmap))

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

// doBitmapConsensusCheck method checks if the consensus in the subround Bitmap is achieved
func (sr *subroundBitmap) doBitmapConsensusCheck() bool {
	if sr.RoundCanceled {
		return false
	}

	if sr.Status(SrBitmap) == spos.SsFinished {
		return true
	}

	threshold := sr.Threshold(SrBitmap)
	if sr.isBitmapReceived(threshold) {
		log.Debug("step 3: subround "+sr.Name()+" has been finished",
			"type", "spos/bn",
			"time [s]", sr.SyncTimer().FormattedCurrentTime())
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
			log.Debug("JobDone", "type", "spos/bn", "error", err.Error())
			continue
		}

		if isJobDone {
			n++
		}
	}

	return n >= threshold
}
