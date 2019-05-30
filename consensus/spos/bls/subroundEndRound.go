package bls

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

type subroundEndRound struct {
	*spos.Subround

	broadcastBlock func(data.BodyHandler, data.HeaderHandler) error
}

// NewSubroundEndRound creates a subroundEndRound object
func NewSubroundEndRound(
	baseSubround *spos.Subround,
	broadcastBlock func(data.BodyHandler, data.HeaderHandler) error,
	extend func(subroundId int),
) (*subroundEndRound, error) {
	err := checkNewSubroundEndRoundParams(
		baseSubround,
		broadcastBlock,
	)
	if err != nil {
		return nil, err
	}

	srEndRound := subroundEndRound{
		baseSubround,
		broadcastBlock,
	}
	srEndRound.Job = srEndRound.doEndRoundJob
	srEndRound.Check = srEndRound.doEndRoundConsensusCheck
	srEndRound.Extend = extend

	return &srEndRound, nil
}

func checkNewSubroundEndRoundParams(
	baseSubround *spos.Subround,
	broadcastBlock func(data.BodyHandler, data.HeaderHandler) error,
) error {
	if baseSubround == nil {
		return spos.ErrNilSubround
	}
	if baseSubround.ConsensusState == nil {
		return spos.ErrNilConsensusState
	}
	if broadcastBlock == nil {
		return spos.ErrNilBroadcastBlockFunction
	}

	err := spos.ValidateConsensusCore(baseSubround.ConsensusCoreHandler)

	return err
}

// doEndRoundJob method does the job of the subround EndRound
func (sr *subroundEndRound) doEndRoundJob() bool {
	if !sr.IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	bitmap := sr.GenerateBitmap(SrSignature)
	err := sr.checkSignaturesValidity(bitmap)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	// Aggregate sig and add it to the block
	sig, err := sr.MultiSigner().AggregateSigs(bitmap)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	sr.Header.SetPubKeysBitmap(bitmap)
	sr.Header.SetSignature(sig)

	timeBefore := time.Now()
	// Commit the block (commits also the account state)
	err = sr.BlockProcessor().CommitBlock(sr.Blockchain(), sr.ConsensusState.Header, sr.ConsensusState.BlockBody)
	if err != nil {
		log.Error(err.Error())
		return false
	}
	timeAfter := time.Now()

	log.Info(fmt.Sprintf("time elapsed to commit block: %v sec\n", timeAfter.Sub(timeBefore).Seconds()))

	sr.SetStatus(SrEndRound, spos.SsFinished)

	// broadcast block body and header
	err = sr.broadcastBlock(sr.ConsensusState.BlockBody, sr.ConsensusState.Header)
	if err != nil {
		log.Error(err.Error())
	}

	log.Info(fmt.Sprintf("%sStep 3: BlockBody and Header has been committed and broadcast\n", sr.SyncTimer().FormattedCurrentTime()))

	msg := fmt.Sprintf("Added proposed block with nonce  %d  in blockchain", sr.Header.GetNonce())
	log.Info(log.Headline(msg, sr.SyncTimer().FormattedCurrentTime(), "+"))

	return true
}

// doEndRoundConsensusCheck method checks if the consensus is achieved
func (sr *subroundEndRound) doEndRoundConsensusCheck() bool {
	if sr.RoundCanceled {
		return false
	}

	if sr.Status(SrEndRound) == spos.SsFinished {
		return true
	}

	return false
}

func (sr *subroundEndRound) checkSignaturesValidity(bitmap []byte) error {
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
		isSigJobDone, err := sr.ConsensusState.JobDone(pubKey, SrSignature)
		if err != nil {
			return err
		}

		if !isSigJobDone {
			return spos.ErrNilSignature
		}

		signature, err := sr.MultiSigner().SignatureShare(uint16(i))
		if err != nil {
			return err
		}

		err = sr.MultiSigner().VerifySignatureShare(uint16(i), signature, sr.GetData(), bitmap)
		if err != nil {
			return err
		}
	}

	return nil
}
