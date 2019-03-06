package bn

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type subroundEndRound struct {
	*subround

	blockChain     *blockchain.BlockChain
	blockProcessor process.BlockProcessor
	consensusState *spos.ConsensusState
	multiSigner    crypto.MultiSigner
	rounder        consensus.Rounder
	syncTimer      ntp.SyncTimer

	broadcastBlock func(block.Body, *block.Header) error
}

// NewSubroundEndRound creates a subroundEndRound object
func NewSubroundEndRound(
	subround *subround,
	blockChain *blockchain.BlockChain,
	blockProcessor process.BlockProcessor,
	consensusState *spos.ConsensusState,
	multiSigner crypto.MultiSigner,
	rounder consensus.Rounder,
	syncTimer ntp.SyncTimer,
	broadcastBlock func(block.Body, *block.Header) error,
	extend func(subroundId int),
) (*subroundEndRound, error) {

	err := checkNewSubroundEndRoundParams(
		subround,
		blockChain,
		blockProcessor,
		consensusState,
		multiSigner,
		rounder,
		syncTimer,
		broadcastBlock,
	)

	if err != nil {
		return nil, err
	}

	srEndRound := subroundEndRound{
		subround,
		blockChain,
		blockProcessor,
		consensusState,
		multiSigner,
		rounder,
		syncTimer,
		broadcastBlock,
	}

	srEndRound.job = srEndRound.doEndRoundJob
	srEndRound.check = srEndRound.doEndRoundConsensusCheck
	srEndRound.extend = extend

	return &srEndRound, nil
}

func checkNewSubroundEndRoundParams(
	subround *subround,
	blockChain *blockchain.BlockChain,
	blockProcessor process.BlockProcessor,
	consensusState *spos.ConsensusState,
	multiSigner crypto.MultiSigner,
	rounder consensus.Rounder,
	syncTimer ntp.SyncTimer,
	broadcastBlock func(block.Body, *block.Header) error,
) error {
	if subround == nil {
		return spos.ErrNilSubround
	}

	if blockChain == nil {
		return spos.ErrNilBlockChain
	}

	if blockProcessor == nil {
		return spos.ErrNilBlockProcessor
	}

	if consensusState == nil {
		return spos.ErrNilConsensusState
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

	if broadcastBlock == nil {
		return spos.ErrNilBroadcastBlockFunction
	}

	return nil
}

// doEndRoundJob method does the job of the end round subround
func (sr *subroundEndRound) doEndRoundJob() bool {
	bitmap := sr.consensusState.GenerateBitmap(SrBitmap)

	err := sr.checkSignaturesValidity(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	// Aggregate sig and add it to the block
	sig, err := sr.multiSigner.AggregateSigs(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	sr.consensusState.Header.Signature = sig

	// Commit the block (commits also the account state)
	err = sr.blockProcessor.CommitBlock(sr.blockChain, sr.consensusState.Header, sr.consensusState.BlockBody)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	sr.consensusState.SetStatus(SrEndRound, spos.SsFinished)

	err = sr.blockProcessor.RemoveBlockInfoFromPool(sr.consensusState.BlockBody)

	if err != nil {
		log.Error(err.Error())
	}

	// broadcast block body and header
	err = sr.broadcastBlock(sr.consensusState.BlockBody, sr.consensusState.Header)

	if err != nil {
		log.Error(err.Error())
	}

	log.Info(fmt.Sprintf("%sStep 6: TxBlockBody and Header has been commited and broadcasted \n", sr.syncTimer.FormattedCurrentTime()))

	actionMsg := "synchronized"
	if sr.consensusState.IsSelfLeaderInCurrentRound() {
		actionMsg = "proposed"
	}

	msg := fmt.Sprintf("Added %s block with nonce  %d  in blockchain", actionMsg, sr.consensusState.Header.Nonce)
	log.Info(log.Headline(msg, sr.syncTimer.FormattedCurrentTime(), "+"))

	return true
}

// doEndRoundConsensusCheck method checks if the consensus is achieved in each subround from first subround to the given
// subround. If the consensus is achieved in one subround, the subround status is marked as finished
func (sr *subroundEndRound) doEndRoundConsensusCheck() bool {
	if sr.consensusState.RoundCanceled {
		return false
	}

	if sr.consensusState.Status(SrEndRound) == spos.SsFinished {
		return true
	}

	return false
}

func (sr *subroundEndRound) checkSignaturesValidity(bitmap []byte) error {
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
		isSigJobDone, err := sr.consensusState.JobDone(pubKey, SrSignature)

		if err != nil {
			return err
		}

		if !isSigJobDone {
			return spos.ErrNilSignature
		}

		signature, err := sr.multiSigner.SignatureShare(uint16(i))

		if err != nil {
			return err
		}

		// verify partial signature
		err = sr.multiSigner.VerifySignatureShare(uint16(i), signature, bitmap)

		if err != nil {
			return err
		}
	}

	return nil
}
