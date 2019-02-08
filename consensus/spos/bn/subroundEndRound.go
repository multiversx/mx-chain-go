package bn

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
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
	rounder        round.Rounder
	syncTimer      ntp.SyncTimer

	broadcastTxBlockBody func(*block.TxBlockBody) error
	broadcastHeader      func(*block.Header) error
}

// NewSubroundEndRound creates a subroundEndRound object
func NewSubroundEndRound(
	subround *subround,
	blockChain *blockchain.BlockChain,
	blockProcessor process.BlockProcessor,
	consensusState *spos.ConsensusState,
	multiSigner crypto.MultiSigner,
	rounder round.Rounder,
	syncTimer ntp.SyncTimer,
	broadcastTxBlockBody func(*block.TxBlockBody) error,
	broadcastHeader func(*block.Header) error,
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
		broadcastTxBlockBody,
		broadcastHeader,
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
		broadcastTxBlockBody,
		broadcastHeader,
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
	rounder round.Rounder,
	syncTimer ntp.SyncTimer,
	broadcastTxBlockBody func(*block.TxBlockBody) error,
	broadcastHeader func(*block.Header) error,
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

	if broadcastTxBlockBody == nil {
		return spos.ErrNilBroadcastTxBlockBodyFunction
	}

	if broadcastHeader == nil {
		return spos.ErrNilBroadcastHeaderFunction
	}

	return nil
}

// doEndRoundJob method is the function which actually does the job of the EndRound subround
// (it is used as the handler function of the doSubroundJob pointer variable function in subround struct,
// from spos package)
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

	err = sr.blockProcessor.RemoveBlockTxsFromPool(sr.consensusState.BlockBody)

	if err != nil {
		log.Error(err.Error())
	}

	// broadcast block body
	err = sr.broadcastTxBlockBody(sr.consensusState.BlockBody)

	if err != nil {
		log.Error(err.Error())
	}

	// broadcast header
	err = sr.broadcastHeader(sr.consensusState.Header)

	if err != nil {
		log.Error(err.Error())
	}

	log.Info(fmt.Sprintf("%sStep 6: Commiting and broadcasting TxBlockBody and Header\n", sr.syncTimer.FormattedCurrentTime()))

	if sr.consensusState.IsSelfLeaderInCurrentRound() {
		log.Info(fmt.Sprintf("\n%s++++++++++++++++++++ ADDED PROPOSED BLOCK WITH NONCE  %d  IN BLOCKCHAIN ++++++++++++++++++++\n\n",
			sr.syncTimer.FormattedCurrentTime(), sr.consensusState.Header.Nonce))
	} else {
		log.Info(fmt.Sprintf("\n%sxxxxxxxxxxxxxxxxxxxx ADDED SYNCHRONIZED BLOCK WITH NONCE  %d  IN BLOCKCHAIN xxxxxxxxxxxxxxxxxxxx\n\n",
			sr.syncTimer.FormattedCurrentTime(), sr.consensusState.Header.Nonce))
	}

	return true
}

// doEndRoundConsensusCheck method checks if the consensus is achieved in each subround from first subround to the given
// subround. If the consensus is achieved in one subround, the subround status is marked as finished
func (sr *subroundEndRound) doEndRoundConsensusCheck() bool {
	return true
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
		isSigJobDone, err := sr.consensusState.GetJobDone(pubKey, SrSignature)

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
