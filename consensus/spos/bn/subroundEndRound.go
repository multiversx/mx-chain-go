package bn

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

type subroundEndRound struct {
	*subround

	broadcastBlock func(data.BodyHandler, data.HeaderHandler) error
}

// NewSubroundEndRound creates a subroundEndRound object
func NewSubroundEndRound(
	subround *subround,
	broadcastBlock func(data.BodyHandler, data.HeaderHandler) error,
	extend func(subroundId int),
) (*subroundEndRound, error) {
	err := checkNewSubroundEndRoundParams(
		subround,
		broadcastBlock,
	)
	if err != nil {
		return nil, err
	}

	srEndRound := subroundEndRound{
		subround,
		broadcastBlock,
	}
	srEndRound.job = srEndRound.doEndRoundJob
	srEndRound.check = srEndRound.doEndRoundConsensusCheck
	srEndRound.extend = extend

	return &srEndRound, nil
}

func checkNewSubroundEndRoundParams(
	subround *subround,
	broadcastBlock func(data.BodyHandler, data.HeaderHandler) error,
) error {
	if subround == nil {
		return spos.ErrNilSubround
	}

	if subround.ConsensusState == nil {
		return spos.ErrNilConsensusState
	}

	if broadcastBlock == nil {
		return spos.ErrNilBroadcastBlockFunction
	}

	err := spos.ValidateConsensusCore(subround.ConsensusCoreHandler)

	return err
}

// doEndRoundJob method does the job of the end round subround
func (sr *subroundEndRound) doEndRoundJob() bool {
	bitmap := sr.GenerateBitmap(SrBitmap)
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

	sr.Header.SetSignature(sig)

	// Commit the block (commits also the account state)
	err = sr.BlockProcessor().CommitBlock(sr.Blockchain(), sr.ConsensusState.Header, sr.ConsensusState.BlockBody)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	sr.SetStatus(SrEndRound, spos.SsFinished)

	// broadcast block body and header
	err = sr.broadcastBlock(sr.ConsensusState.BlockBody, sr.ConsensusState.Header)
	if err != nil {
		log.Error(err.Error())
	}

	log.Info(fmt.Sprintf("%sStep 6: TxBlockBody and Header has been commited and broadcasted \n", sr.SyncTimer().FormattedCurrentTime()))

	actionMsg := "synchronized"
	if sr.IsSelfLeaderInCurrentRound() {
		actionMsg = "proposed"
	}

	msg := fmt.Sprintf("Added %s block with nonce  %d  in blockchain", actionMsg, sr.Header.GetNonce())
	log.Info(log.Headline(msg, sr.SyncTimer().FormattedCurrentTime(), "+"))

	return true
}

// doEndRoundConsensusCheck method checks if the consensus is achieved in each subround from first subround to the given
// subround. If the consensus is achieved in one subround, the subround status is marked as finished
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

		// verify partial signature
		err = sr.MultiSigner().VerifySignatureShare(uint16(i), signature, bitmap)
		if err != nil {
			return err
		}
	}

	return nil
}
