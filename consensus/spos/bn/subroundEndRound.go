package bn

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus/spos"
)

type subroundEndRound struct {
	*spos.Subround
}

// NewSubroundEndRound creates a subroundEndRound object
func NewSubroundEndRound(
	baseSubround *spos.Subround,
	extend func(subroundId int),
) (*subroundEndRound, error) {
	err := checkNewSubroundEndRoundParams(
		baseSubround,
	)
	if err != nil {
		return nil, err
	}

	srEndRound := subroundEndRound{
		baseSubround,
	}
	srEndRound.Job = srEndRound.doEndRoundJob
	srEndRound.Check = srEndRound.doEndRoundConsensusCheck
	srEndRound.Extend = extend

	return &srEndRound, nil
}

func checkNewSubroundEndRoundParams(
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

// doEndRoundJob method does the job of the subround EndRound
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

	timeBefore := time.Now()
	// Commit the block (commits also the account state)
	err = sr.BlockProcessor().CommitBlock(sr.Blockchain(), sr.Header, sr.BlockBody)
	if err != nil {
		log.Error(err.Error())
		return false
	}
	timeAfter := time.Now()

	log.Info(fmt.Sprintf("time elapsed to commit block: %v sec\n", timeAfter.Sub(timeBefore).Seconds()))

	sr.SetStatus(SrEndRound, spos.SsFinished)

	// broadcast block body and header
	err = sr.BroadcastMessenger().BroadcastBlock(sr.BlockBody, sr.Header)
	if err != nil {
		log.Error(err.Error())
	}

	// broadcast header to metachain
	err = sr.BroadcastMessenger().BroadcastHeader(sr.Header)
	if err != nil {
		log.Error(err.Error())
	}

	sr.BlocksTracker().SetBlockBroadcastRound(sr.Header.GetNonce(), sr.RoundIndex)

	log.Info(fmt.Sprintf("%sStep 6: TxBlockBody and Header has been committed and broadcast\n", sr.SyncTimer().FormattedCurrentTime()))

	err = sr.broadcastMiniBlocksAndTransactions()
	if err != nil {
		log.Error(err.Error())
	}

	actionMsg := "synchronized"
	if sr.IsSelfLeaderInCurrentRound() {
		actionMsg = "proposed"
	}

	msg := fmt.Sprintf("Added %s block with nonce  %d  in blockchain", actionMsg, sr.Header.GetNonce())
	log.Info(log.Headline(msg, sr.SyncTimer().FormattedCurrentTime(), "+"))

	return true
}

func (sr *subroundEndRound) broadcastMiniBlocksAndTransactions() error {
	miniBlocks, transactions, err := sr.BlockProcessor().MarshalizedDataToBroadcast(sr.Header, sr.BlockBody)
	if err != nil {
		return err
	}

	err = sr.BroadcastMessenger().BroadcastMiniBlocks(miniBlocks)
	if err != nil {
		return err
	}

	err = sr.BroadcastMessenger().BroadcastTransactions(transactions)
	if err != nil {
		return err
	}

	return nil
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
		isSigJobDone, err := sr.JobDone(pubKey, SrSignature)
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
		err = sr.MultiSigner().VerifySignatureShare(uint16(i), signature, sr.GetData(), bitmap)
		if err != nil {
			return err
		}
	}

	return nil
}
