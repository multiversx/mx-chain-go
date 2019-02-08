package bn

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

// doAdvanceJob method is the function which actually does the job of the Advance subround (it is used as the handler
// function of the doSubroundJob pointer variable function in Subround struct, from spos package)
func (wrk *Worker) doAdvanceJob() bool {
	if wrk.SPoS.Status(SrEndRound) == spos.SsFinished {
		return false
	}

	wrk.blockProcessor.RevertAccountState()

	log.Info(fmt.Sprintf("%sStep 7: Creating and broadcasting an empty block\n", wrk.SPoS.Chr.GetFormattedTime()))

	wrk.createEmptyBlock()

	return true
}

// checkAdvanceConsensus method checks if the consensus is achieved in the advance subround.
func (wrk *Worker) checkAdvanceConsensus() bool {
	return true
}

// createEmptyBlock creates, commits and broadcasts an empty block at the end of the round if no block was proposed or
// syncronized in this round
func (wrk *Worker) createEmptyBlock() bool {
	blk := wrk.blockProcessor.CreateEmptyBlockBody(
		wrk.shardCoordinator.ShardForCurrentNode(),
		wrk.SPoS.Chr.Round().Index())

	hdr := &block.Header{}
	hdr.Round = uint32(wrk.SPoS.Chr.Round().Index())
	hdr.TimeStamp = wrk.SPoS.Chr.RoundTimeStamp()

	var prevHeaderHash []byte

	if wrk.BlockChain.CurrentBlockHeader == nil {
		hdr.Nonce = 1
		prevHeaderHash = wrk.BlockChain.GenesisHeaderHash
	} else {
		hdr.Nonce = wrk.BlockChain.CurrentBlockHeader.Nonce + 1
		prevHeaderHash = wrk.BlockChain.CurrentBlockHeaderHash
	}

	hdr.PrevHash = prevHeaderHash
	blkStr, err := wrk.marshalizer.Marshal(blk)

	if err != nil {
		log.Info(err.Error())
		return false
	}

	hdr.BlockBodyHash = wrk.hasher.Compute(string(blkStr))

	cnsGroup := wrk.SPoS.ConsensusGroup()
	cnsGroupSize := len(cnsGroup)

	hdr.PubKeysBitmap = make([]byte, cnsGroupSize/8+1)

	// TODO: decide the signature for the empty block
	headerStr, err := wrk.marshalizer.Marshal(hdr)
	hdrHash := wrk.hasher.Compute(string(headerStr))
	hdr.Signature = hdrHash
	hdr.Commitment = hdrHash

	// Commit the block (commits also the account state)
	err = wrk.blockProcessor.CommitBlock(wrk.BlockChain, hdr, blk)

	if err != nil {
		log.Info(err.Error())
		return false
	}

	// broadcast block body
	err = wrk.broadcastTxBlockBody(blk)

	if err != nil {
		log.Info(err.Error())
	}

	// broadcast header
	err = wrk.broadcastHeader(hdr)

	if err != nil {
		log.Info(err.Error())
	}

	log.Info(fmt.Sprintf("\n%s******************** ADDED EMPTY BLOCK WITH NONCE  %d  IN BLOCKCHAIN ********************\n\n",
		wrk.SPoS.Chr.GetFormattedTime(), hdr.Nonce))

	return true
}
