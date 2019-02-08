package bn

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
)

// doEndRoundJob method is the function which actually does the job of the EndRound subround
// (it is used as the handler function of the doSubroundJob pointer variable function in Subround struct,
// from spos package)
func (wrk *Worker) doEndRoundJob() bool {
	if !wrk.checkEndRoundConsensus() {
		return false
	}

	bitmap := wrk.genBitmap(SrBitmap)

	err := wrk.checkSignaturesValidity(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	// Aggregate sig and add it to the block
	sig, err := wrk.multiSigner.AggregateSigs(bitmap)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	wrk.Header.Signature = sig

	// Commit the block (commits also the account state)
	err = wrk.blockProcessor.CommitBlock(wrk.BlockChain, wrk.Header, wrk.BlockBody)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	wrk.SPoS.SetStatus(SrEndRound, spos.SsFinished)

	err = wrk.blockProcessor.RemoveBlockTxsFromPool(wrk.BlockBody)

	if err != nil {
		log.Error(err.Error())
	}

	// broadcast block body
	err = wrk.broadcastTxBlockBody(wrk.BlockBody)

	if err != nil {
		log.Error(err.Error())
	}

	// broadcast header
	err = wrk.broadcastHeader(wrk.Header)

	if err != nil {
		log.Error(err.Error())
	}

	log.Info(fmt.Sprintf("%sStep 6: Commiting and broadcasting TxBlockBody and Header\n", wrk.SPoS.Chr.GetFormattedTime()))

	if wrk.SPoS.IsSelfLeaderInCurrentRound() {
		log.Info(fmt.Sprintf("\n%s++++++++++++++++++++ ADDED PROPOSED BLOCK WITH NONCE  %d  IN BLOCKCHAIN ++++++++++++++++++++\n\n",
			wrk.SPoS.Chr.GetFormattedTime(), wrk.Header.Nonce))
	} else {
		log.Info(fmt.Sprintf("\n%sxxxxxxxxxxxxxxxxxxxx ADDED SYNCHRONIZED BLOCK WITH NONCE  %d  IN BLOCKCHAIN xxxxxxxxxxxxxxxxxxxx\n\n",
			wrk.SPoS.Chr.GetFormattedTime(), wrk.Header.Nonce))
	}

	return true
}

func (wrk *Worker) genBitmap(subround chronology.SubroundId) []byte {
	// generate bitmap according to set commitment hashes
	sizeConsensus := len(wrk.SPoS.ConsensusGroup())

	bitmap := make([]byte, sizeConsensus/8+1)

	for i := 0; i < sizeConsensus; i++ {
		pubKey := wrk.SPoS.ConsensusGroup()[i]
		isJobDone, err := wrk.SPoS.GetJobDone(pubKey, subround)

		if err != nil {
			log.Error(err.Error())
			continue
		}

		if isJobDone {
			bitmap[i/8] |= 1 << (uint16(i) % 8)
		}
	}

	return bitmap
}

// checkEndRoundConsensus method checks if the consensus is achieved in each subround from first subround to the given
// subround. If the consensus is achieved in one subround, the subround status is marked as finished
func (wrk *Worker) checkEndRoundConsensus() bool {
	for i := SrBlock; i <= SrSignature; i++ {
		currentSubRound := wrk.SPoS.Chr.SubroundHandlers()[i]
		if !currentSubRound.Check() {
			return false
		}
	}

	return true
}

// extendEndRound method just print some messages as no extend will be permited, because a new round will be start
func (wrk *Worker) extendEndRound() {
	wrk.SPoS.SetStatus(SrEndRound, spos.SsExtended)

	log.Info(fmt.Sprintf("%sStep 6: Extended the (END_ROUND) subround\n", wrk.SPoS.Chr.GetFormattedTime()))
}

func (wrk *Worker) checkSignaturesValidity(bitmap []byte) error {
	nbBitsBitmap := len(bitmap) * 8
	consensusGroup := wrk.SPoS.ConsensusGroup()
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
		isSigJobDone, err := wrk.SPoS.GetJobDone(pubKey, SrSignature)

		if err != nil {
			return err
		}

		if !isSigJobDone {
			return spos.ErrNilSignature
		}

		signature, err := wrk.multiSigner.SignatureShare(uint16(i))

		if err != nil {
			return err
		}

		// verify partial signature
		err = wrk.multiSigner.VerifySignatureShare(uint16(i), signature, bitmap)

		if err != nil {
			return err
		}
	}

	return nil
}
