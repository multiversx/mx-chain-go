package bn

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

// doBlockJob method actually send the proposed block in the Block subround, when this node is leader
// (it is used as a handler function of the doSubroundJob pointer function declared in Subround struct,
// from spos package)
func (wrk *Worker) doBlockJob() bool {
	if wrk.boot.ShouldSync() { // if node is not synchronized yet, it has to continue the bootstrapping mechanism
		return false
	}

	if !wrk.SPoS.IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	if wrk.isSelfJobDone(SrBlock) {
		return false
	}

	if wrk.isCurrentRoundFinished(SrBlock) {
		return false
	}

	if !wrk.sendBlockBody() ||
		!wrk.sendBlockHeader() {
		return false
	}

	err := wrk.SPoS.SetSelfJobDone(SrBlock, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	wrk.multiSigner.SetMessage(wrk.SPoS.Data)

	return true
}

// sendBlockBody method send the proposed block body in the Block subround
func (wrk *Worker) sendBlockBody() bool {
	haveTime := func() bool {
		if wrk.SPoS.Chr.GetSubround() > chronology.SubroundId(SrBlock) {
			return false
		}

		return true
	}

	blk, err := wrk.blockProcessor.CreateTxBlockBody(
		wrk.shardCoordinator.ShardForCurrentNode(),
		maxTransactionsInBlock,
		wrk.SPoS.Chr.Round().Index(),
		haveTime,
	)

	blkStr, err := wrk.marshalizer.Marshal(blk)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	dta := spos.NewConsensusData(
		nil,
		blkStr,
		[]byte(wrk.SPoS.SelfPubKey()),
		nil,
		int(MtBlockBody),
		wrk.SPoS.Chr.RoundTimeStamp(),
		wrk.SPoS.Chr.Round().Index())

	if !wrk.sendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: Sending block body\n", wrk.SPoS.Chr.GetFormattedTime()))

	wrk.BlockBody = blk

	return true
}

// sendBlockHeader method send the proposed block header in the Block subround
func (wrk *Worker) sendBlockHeader() bool {
	hdr := &block.Header{}

	hdr.Round = uint32(wrk.SPoS.Chr.Round().Index())
	hdr.TimeStamp = wrk.SPoS.Chr.RoundTimeStamp()

	if wrk.BlockChain.CurrentBlockHeader == nil {
		hdr.Nonce = 1
		hdr.PrevHash = wrk.BlockChain.GenesisHeaderHash
	} else {
		hdr.Nonce = wrk.BlockChain.CurrentBlockHeader.Nonce + 1
		hdr.PrevHash = wrk.BlockChain.CurrentBlockHeaderHash
	}

	blkStr, err := wrk.marshalizer.Marshal(wrk.BlockBody)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	hdr.BlockBodyHash = wrk.hasher.Compute(string(blkStr))

	hdrStr, err := wrk.marshalizer.Marshal(hdr)

	hdrHash := wrk.hasher.Compute(string(hdrStr))

	if err != nil {
		log.Error(err.Error())
		return false
	}

	dta := spos.NewConsensusData(
		hdrHash,
		hdrStr,
		[]byte(wrk.SPoS.SelfPubKey()),
		nil,
		int(MtBlockHeader),
		wrk.SPoS.Chr.RoundTimeStamp(),
		wrk.SPoS.Chr.Round().Index())

	if !wrk.sendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: Sending block header with nonce %d and hash %s\n",
		wrk.SPoS.Chr.GetFormattedTime(), hdr.Nonce, toB64(hdrHash)))

	wrk.SPoS.Data = hdrHash
	wrk.Header = hdr

	return true
}

// receivedBlockBody method is called when a block body is received through the block body channel.
func (wrk *Worker) receivedBlockBody(cnsDta *spos.ConsensusData) bool {
	node := string(cnsDta.PubKey)

	if wrk.isBlockBodyAlreadyReceived() {
		return false
	}

	if !wrk.SPoS.IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !wrk.canReceiveMessage(node, cnsDta.RoundIndex, SrBlock) {
		return false
	}

	wrk.BlockBody = wrk.decodeBlockBody(cnsDta.SubRoundData)

	if wrk.BlockBody == nil {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: Received block body\n", wrk.SPoS.Chr.GetFormattedTime()))

	blockProcessedWithSuccess := wrk.processReceivedBlock(cnsDta)

	return blockProcessedWithSuccess
}

// decodeBlockBody method decodes block body which is marshalized in the received message
func (wrk *Worker) decodeBlockBody(dta []byte) *block.TxBlockBody {
	if dta == nil {
		return nil
	}

	var blk block.TxBlockBody

	err := wrk.marshalizer.Unmarshal(&blk, dta)

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return &blk
}

// receivedBlockHeader method is called when a block header is received through the block header channel.
// If the block header is valid, than the validatorRoundStates map coresponding to the node which sent it,
// is set on true for the subround Block
func (wrk *Worker) receivedBlockHeader(cnsDta *spos.ConsensusData) bool {
	node := string(cnsDta.PubKey)

	if wrk.isConsensusDataAlreadySet() {
		return false
	}

	if wrk.isHeaderAlreadyReceived() {
		return false
	}

	if !wrk.SPoS.IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !wrk.canReceiveMessage(node, cnsDta.RoundIndex, SrBlock) {
		return false
	}

	wrk.SPoS.Data = cnsDta.BlockHeaderHash
	wrk.Header = wrk.decodeBlockHeader(cnsDta.SubRoundData)

	if wrk.Header == nil {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: Received block header with nonce %d and hash %s\n",
		wrk.SPoS.Chr.GetFormattedTime(), wrk.Header.Nonce, toB64(cnsDta.BlockHeaderHash)))

	if !wrk.checkIfBlockIsValid(wrk.Header) {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, INVALID BLOCK\n",
			wrk.SPoS.Chr.Round().Index(), getSubroundName(SrBlock)))

		return false
	}

	blockProcessedWithSuccess := wrk.processReceivedBlock(cnsDta)

	return blockProcessedWithSuccess
}

// decodeBlockHeader method decodes block header which is marshalized in the received message
func (wrk *Worker) decodeBlockHeader(dta []byte) *block.Header {
	if dta == nil {
		return nil
	}

	var hdr block.Header

	err := wrk.marshalizer.Unmarshal(&hdr, dta)

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return &hdr
}

func (wrk *Worker) processReceivedBlock(cnsDta *spos.ConsensusData) bool {
	if wrk.BlockBody == nil ||
		wrk.Header == nil {
		return false
	}

	node := string(cnsDta.PubKey)

	haveTime := func() time.Duration {
		chr := wrk.SPoS.Chr

		roundStartTime := chr.Round().TimeStamp()
		currentTime := chr.SyncTimer().CurrentTime(chr.ClockOffset())
		elapsedTime := currentTime.Sub(roundStartTime)
		haveTime := float64(chr.Round().TimeDuration())*maxBlockProcessingTimePercent - float64(elapsedTime)

		return time.Duration(haveTime)
	}

	err := wrk.blockProcessor.ProcessBlock(wrk.BlockChain, wrk.Header, wrk.BlockBody, haveTime)

	if err != nil {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, %s\n",
			wrk.SPoS.Chr.Round().Index(), getSubroundName(SrBlock), err.Error()))

		return false
	}

	subround := wrk.SPoS.Chr.GetSubround()

	if cnsDta.RoundIndex != wrk.SPoS.Chr.Round().Index() {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, CURRENT ROUND IS %d\n",
			cnsDta.RoundIndex, getSubroundName(SrBlock), wrk.SPoS.Chr.Round().Index()))

		wrk.blockProcessor.RevertAccountState()

		return false
	}

	if subround > chronology.SubroundId(SrEndRound) {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, CURRENT SUBROUND IS %s\n",
			cnsDta.RoundIndex, getSubroundName(SrBlock), getSubroundName(subround)))

		wrk.blockProcessor.RevertAccountState()

		return false
	}

	wrk.multiSigner.SetMessage(wrk.SPoS.Data)
	err = wrk.SPoS.RoundConsensus.SetJobDone(node, SrBlock, true)

	if err != nil {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, %s\n",
			wrk.SPoS.Chr.Round().Index(), getSubroundName(SrBlock), err.Error()))

		return false
	}

	return true
}

// checkBlockConsensus method checks if the consensus in the <BLOCK> subround is achieved
func (wrk *Worker) checkBlockConsensus() bool {
	wrk.mutCheckConsensus.Lock()
	defer wrk.mutCheckConsensus.Unlock()

	if wrk.SPoS.Chr.IsCancelled() {
		return false
	}

	if wrk.SPoS.Status(SrBlock) == spos.SsFinished {
		return true
	}

	threshold := wrk.SPoS.Threshold(SrBlock)

	if wrk.isBlockReceived(threshold) {
		wrk.printBlockCM() // only for printing block consensus messages
		wrk.SPoS.SetStatus(SrBlock, spos.SsFinished)

		return true
	}

	return false
}

// isBlockReceived method checks if the block was received from the leader in the current round
func (wrk *Worker) isBlockReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(wrk.SPoS.ConsensusGroup()); i++ {
		node := wrk.SPoS.ConsensusGroup()[i]
		isJobDone, err := wrk.SPoS.GetJobDone(node, SrBlock)

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

// extendBlock method put this subround in the extended mode and print some messages
func (wrk *Worker) extendBlock() {
	if wrk.boot.ShouldSync() {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, NOT SYNCRONIZED YET\n",
			wrk.SPoS.Chr.Round().Index(), getSubroundName(SrBlock)))

		wrk.SPoS.Chr.SetSelfSubround(-1)

		return
	}

	wrk.SPoS.SetStatus(SrBlock, spos.SsExtended)

	log.Info(fmt.Sprintf("%sStep 1: Extended the (BLOCK) subround\n", wrk.SPoS.Chr.GetFormattedTime()))
}

// checkIfBlockIsValid method checks if the received block is valid
func (wrk *Worker) checkIfBlockIsValid(receivedHeader *block.Header) bool {
	if wrk.BlockChain.CurrentBlockHeader == nil {
		if receivedHeader.Nonce == 1 { // first block after genesis
			if bytes.Equal(receivedHeader.PrevHash, wrk.BlockChain.GenesisHeaderHash) {
				return true
			}

			log.Info(fmt.Sprintf("Hash not match: local block hash is empty and node received block with previous hash %s\n",
				toB64(receivedHeader.PrevHash)))

			return false
		}

		log.Info(fmt.Sprintf("Nonce not match: local block nonce is 0 and node received block with nonce %d\n",
			receivedHeader.Nonce))

		return false
	}

	if receivedHeader.Nonce < wrk.BlockChain.CurrentBlockHeader.Nonce+1 {
		log.Info(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d\n",
			wrk.BlockChain.CurrentBlockHeader.Nonce, receivedHeader.Nonce))

		return false
	}

	if receivedHeader.Nonce == wrk.BlockChain.CurrentBlockHeader.Nonce+1 {
		prevHeaderHash := wrk.getHeaderHash(wrk.BlockChain.CurrentBlockHeader)

		if bytes.Equal(receivedHeader.PrevHash, prevHeaderHash) {
			return true
		}

		log.Info(fmt.Sprintf("Hash not match: local block hash is %s and node received block with previous hash %s\n",
			toB64(prevHeaderHash), toB64(receivedHeader.PrevHash)))

		return false
	}

	log.Info(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d\n",
		wrk.BlockChain.CurrentBlockHeader.Nonce, receivedHeader.Nonce))

	return false
}

func (wrk *Worker) getHeaderHash(hdr *block.Header) []byte {
	headerMarsh, err := wrk.marshalizer.Marshal(hdr)

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return wrk.hasher.Compute(string(headerMarsh))
}

// printBlockCM method prints the (BLOCK) subround consensus messages
func (wrk *Worker) printBlockCM() {
	if !wrk.SPoS.IsSelfLeaderInCurrentRound() {
		log.Info(fmt.Sprintf("%sStep 1: Synchronized block\n", wrk.SPoS.Chr.GetFormattedTime()))
	}

	log.Info(fmt.Sprintf("%sStep 1: Subround (BLOCK) has been finished\n", wrk.SPoS.Chr.GetFormattedTime()))
}

// toB64 convert a byte array to a base64 string
func toB64(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}

	return base64.StdEncoding.EncodeToString(buff)
}
