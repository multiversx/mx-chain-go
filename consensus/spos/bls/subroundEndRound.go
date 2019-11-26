package bls

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
)

type subroundEndRound struct {
	*spos.Subround

	appStatusHandler core.AppStatusHandler
}

// SetAppStatusHandler method set appStatusHandler
func (sr *subroundEndRound) SetAppStatusHandler(ash core.AppStatusHandler) error {
	if ash == nil || ash.IsInterfaceNil() {
		return spos.ErrNilAppStatusHandler
	}

	sr.appStatusHandler = ash
	return nil
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
		statusHandler.NewNilStatusHandler(),
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
	if !sr.IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	bitmap := sr.GenerateBitmap(SrSignature)
	err := sr.checkSignaturesValidity(bitmap)
	if err != nil {
		debugError("checkSignaturesValidity", err)
		return false
	}

	// Aggregate sig and add it to the block
	sig, err := sr.MultiSigner().AggregateSigs(bitmap)
	if err != nil {
		debugError("multisigner.AggregateSigs", err)
		return false
	}

	sr.Header.SetPubKeysBitmap(bitmap)
	sr.Header.SetSignature(sig)

	timeBefore := time.Now()
	// Commit the block (commits also the account state)
	err = sr.BlockProcessor().CommitBlock(sr.Blockchain(), sr.Header, sr.BlockBody)
	if err != nil {
		debugError("commit block", err)
		return false
	}
	timeAfter := time.Now()

	log.Debug("commit block", "type", "spos/bls", "time elapsed [s]", timeAfter.Sub(timeBefore).Seconds())

	sr.SetStatus(SrEndRound, spos.SsFinished)

	// broadcast section

	// broadcast block body and header
	err = sr.BroadcastMessenger().BroadcastBlock(sr.BlockBody, sr.Header)
	if err != nil {
		debugError("BroadcastBlock", err)
	}

	// broadcast header to metachain
	err = sr.BroadcastMessenger().BroadcastShardHeader(sr.Header)
	if err != nil {
		debugError("BroadcastShardHeader", err)
	}

	log.Debug("step 3: BlockBody and Header has been committed and broadcast",
		"type", "spos/bls",
		"time [s]", sr.SyncTimer().FormattedCurrentTime())

	err = sr.broadcastMiniBlocksAndTransactions()
	if err != nil {
		debugError("broadcastMiniBlocksAndTransactions", err)
	}

	msg := fmt.Sprintf("Added proposed block with nonce  %d  in blockchain", sr.Header.GetNonce())
	log.Debug(display.Headline(msg, sr.SyncTimer().FormattedCurrentTime(), "+"))

	sr.updateMetricsForLeader()

	return true
}

func (sr *subroundEndRound) updateMetricsForLeader() {
	sr.appStatusHandler.Increment(core.MetricCountAcceptedBlocks)
	sr.appStatusHandler.SetStringValue(core.MetricConsensusRoundState,
		fmt.Sprintf("valid block produced in %f sec", time.Now().Sub(sr.Rounder().TimeStamp()).Seconds()))
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

		err = sr.MultiSigner().VerifySignatureShare(uint16(i), signature, sr.GetData(), bitmap)
		if err != nil {
			return err
		}
	}

	return nil
}
