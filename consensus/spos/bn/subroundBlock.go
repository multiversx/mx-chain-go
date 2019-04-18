package bn

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type subroundBlock struct {
	*subround

	sendConsensusMessage func(*consensus.ConsensusMessage) bool
}

// NewSubroundBlock creates a subroundBlock object
func NewSubroundBlock(
	subround *subround,
	sendConsensusMessage func(*consensus.ConsensusMessage) bool,
	extend func(subroundId int),
) (*subroundBlock, error) {

	err := checkNewSubroundBlockParams(
		subround,
		sendConsensusMessage,
	)

	if err != nil {
		return nil, err
	}

	srBlock := subroundBlock{
		subround,
		sendConsensusMessage,
	}

	srBlock.job = srBlock.doBlockJob
	srBlock.check = srBlock.doBlockConsensusCheck
	srBlock.extend = extend

	return &srBlock, nil
}

func checkNewSubroundBlockParams(
	subround *subround,
	sendConsensusMessage func(*consensus.ConsensusMessage) bool,
) error {
	if subround == nil {
		return spos.ErrNilSubround
	}

	if sendConsensusMessage == nil {
		return spos.ErrNilSendConsensusMessageFunction
	}

	containerValidator := spos.ConsensusContainerValidator{}
	err := containerValidator.ValidateConsensusDataContainer(subround.consensusDataContainer)

	return err
}

// doBlockJob method does the job of the block subround
func (sr *subroundBlock) doBlockJob() bool {
	if !sr.ConsensusState().IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	if sr.ConsensusState().IsSelfJobDone(SrBlock) {
		return false
	}

	if sr.ConsensusState().IsCurrentSubroundFinished(SrBlock) {
		return false
	}

	if !sr.sendBlockBody() ||
		!sr.sendBlockHeader() {
		return false
	}

	err := sr.ConsensusState().SetSelfJobDone(SrBlock, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	sr.MultiSigner().SetMessage(sr.ConsensusState().GetData())

	return true
}

// sendBlockBody method job the proposed block body in the Block subround
func (sr *subroundBlock) sendBlockBody() bool {
	startTime := time.Time{}
	startTime = sr.ConsensusState().RoundTimeStamp
	maxTime := time.Duration(sr.EndTime())
	haveTimeInCurrentSubround := func() bool {
		return sr.Rounder().RemainingTime(startTime, maxTime) > 0
	}

	blockBody, err := sr.BlockProcessor().CreateBlockBody(
		sr.Rounder().Index(),
		haveTimeInCurrentSubround,
	)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	blkStr, err := sr.Marshalizer().Marshal(blockBody)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	msg := consensus.NewConsensusMessage(
		nil,
		blkStr,
		[]byte(sr.ConsensusState().SelfPubKey()),
		nil,
		int(MtBlockBody),
		uint64(sr.Rounder().TimeStamp().Unix()),
		sr.Rounder().Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: block body has been sent\n", sr.SyncTimer().FormattedCurrentTime()))

	sr.ConsensusState().BlockBody = blockBody

	return true
}

// sendBlockHeader method job the proposed block header in the Block subround
func (sr *subroundBlock) sendBlockHeader() bool {
	hdr, err := sr.createHeader()
	if err != nil {
		log.Error(err.Error())
		return false
	}

	hdrStr, err := sr.Marshalizer().Marshal(hdr)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	hdrHash := sr.Hasher().Compute(string(hdrStr))

	msg := consensus.NewConsensusMessage(
		hdrHash,
		hdrStr,
		[]byte(sr.ConsensusState().SelfPubKey()),
		nil,
		int(MtBlockHeader),
		uint64(sr.Rounder().TimeStamp().Unix()),
		sr.Rounder().Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: block header with nonce %d and hash %s has been sent\n",
		sr.SyncTimer().FormattedCurrentTime(), hdr.GetNonce(), toB64(hdrHash)))

	sr.ConsensusState().Data = hdrHash
	sr.ConsensusState().Header = hdr

	return true
}

func (sr *subroundBlock) createHeader() (data.HeaderHandler, error) {
	hdr, err := sr.BlockProcessor().CreateBlockHeader(sr.ConsensusState().BlockBody)

	if err != nil {
		return nil, err
	}

	hdr.SetRound(uint32(sr.Rounder().Index()))
	hdr.SetTimeStamp(uint64(sr.Rounder().TimeStamp().Unix()))

	if sr.Blockchain().GetCurrentBlockHeader() == nil {
		hdr.SetNonce(1)
		hdr.SetPrevHash(sr.Blockchain().GetGenesisHeaderHash())
		// Previous random seed is the signature of the previous block
		hdr.SetPrevRandSeed(sr.Blockchain().GetGenesisHeader().GetSignature())
	} else {
		hdr.SetNonce(sr.Blockchain().GetCurrentBlockHeader().GetNonce() + 1)
		hdr.SetPrevHash(sr.Blockchain().GetCurrentBlockHeaderHash())
		// Previous random seed is the signature of the previous block
		hdr.SetPrevRandSeed(sr.Blockchain().GetCurrentBlockHeader().GetSignature())
	}

	// currently for bnSPoS RandSeed field is not used
	hdr.SetRandSeed([]byte{0})

	return hdr, nil
}

// receivedBlockBody method is called when a block body is received through the block body channel.
func (sr *subroundBlock) receivedBlockBody(cnsDta *consensus.ConsensusMessage) bool {
	node := string(cnsDta.PubKey)

	if sr.ConsensusState().IsBlockBodyAlreadyReceived() {
		return false
	}

	if !sr.ConsensusState().IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !sr.ConsensusState().CanProcessReceivedMessage(cnsDta, sr.Rounder().Index(), SrBlock) {
		return false
	}

	sr.ConsensusState().BlockBody = sr.decodeBlockBody(cnsDta.SubRoundData)

	if sr.ConsensusState().BlockBody == nil {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: block body has been received\n", sr.SyncTimer().FormattedCurrentTime()))

	blockProcessedWithSuccess := sr.processReceivedBlock(cnsDta)

	return blockProcessedWithSuccess
}

// decodeBlockBody method decodes block body which is marshalized in the received message
func (sr *subroundBlock) decodeBlockBody(dta []byte) block.Body {
	if dta == nil {
		return nil
	}

	var blk block.Body

	err := sr.Marshalizer().Unmarshal(&blk, dta)

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return blk
}

// receivedBlockHeader method is called when a block header is received through the block header channel.
// If the block header is valid, than the validatorRoundStates map corresponding to the node which sent it,
// is set on true for the subround Block
func (sr *subroundBlock) receivedBlockHeader(cnsDta *consensus.ConsensusMessage) bool {
	node := string(cnsDta.PubKey)

	if sr.ConsensusState().IsConsensusDataSet() {
		return false
	}

	if sr.ConsensusState().IsHeaderAlreadyReceived() {
		return false
	}

	if !sr.ConsensusState().IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !sr.ConsensusState().CanProcessReceivedMessage(cnsDta, sr.Rounder().Index(), SrBlock) {
		return false
	}

	sr.ConsensusState().Data = cnsDta.BlockHeaderHash
	sr.ConsensusState().Header = sr.decodeBlockHeader(cnsDta.SubRoundData)

	if sr.ConsensusState().Header == nil {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: block header with nonce %d and hash %s has been received\n",
		sr.SyncTimer().FormattedCurrentTime(), sr.ConsensusState().Header.GetNonce(), toB64(cnsDta.BlockHeaderHash)))

	if !sr.BlockProcessor().CheckBlockValidity(sr.Blockchain(), sr.ConsensusState().Header, nil) {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, invalid block\n",
			sr.Rounder().Index(), getSubroundName(SrBlock)))

		return false
	}

	blockProcessedWithSuccess := sr.processReceivedBlock(cnsDta)

	return blockProcessedWithSuccess
}

// decodeBlockHeader method decodes block header which is marshalized in the received message
func (sr *subroundBlock) decodeBlockHeader(dta []byte) *block.Header {
	if dta == nil {
		return nil
	}

	var hdr block.Header

	err := sr.Marshalizer().Unmarshal(&hdr, dta)

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return &hdr
}

func (sr *subroundBlock) processReceivedBlock(cnsDta *consensus.ConsensusMessage) bool {
	if sr.ConsensusState().BlockBody == nil ||
		sr.ConsensusState().Header == nil {
		return false
	}

	defer func() {
		sr.ConsensusState().SetProcessingBlock(false)
	}()

	sr.ConsensusState().SetProcessingBlock(true)

	node := string(cnsDta.PubKey)

	startTime := time.Time{}
	startTime = sr.ConsensusState().RoundTimeStamp
	maxTime := sr.Rounder().TimeDuration() * processingThresholdPercent / 100
	remainingTimeInCurrentRound := func() time.Duration {
		return sr.Rounder().RemainingTime(startTime, maxTime)
	}

	err := sr.BlockProcessor().ProcessBlock(
		sr.Blockchain(),
		sr.ConsensusState().Header,
		sr.ConsensusState().BlockBody,
		remainingTimeInCurrentRound,
	)

	if cnsDta.RoundIndex < sr.Rounder().Index() {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, meantime round index has been changed to %d\n",
			cnsDta.RoundIndex, getSubroundName(SrBlock), sr.Rounder().Index()))
		return false
	}

	if err != nil {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, %s\n",
			sr.Rounder().Index(), getSubroundName(SrBlock), err.Error()))
		if err == process.ErrTimeIsOut {
			sr.ConsensusState().RoundCanceled = true
		}
		return false
	}

	sr.MultiSigner().SetMessage(sr.ConsensusState().Data)
	err = sr.ConsensusState().SetJobDone(node, SrBlock, true)
	if err != nil {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, %s\n",
			sr.Rounder().Index(), getSubroundName(SrBlock), err.Error()))
		return false
	}

	return true
}

// doBlockConsensusCheck method checks if the consensus in the <BLOCK> subround is achieved
func (sr *subroundBlock) doBlockConsensusCheck() bool {
	if sr.ConsensusState().RoundCanceled {
		return false
	}

	if sr.ConsensusState().Status(SrBlock) == spos.SsFinished {
		return true
	}

	threshold := sr.ConsensusState().Threshold(SrBlock)
	if sr.isBlockReceived(threshold) {
		log.Info(fmt.Sprintf("%sStep 1: subround %s has been finished\n", sr.SyncTimer().FormattedCurrentTime(), sr.Name()))
		sr.ConsensusState().SetStatus(SrBlock, spos.SsFinished)
		return true
	}

	return false
}

// isBlockReceived method checks if the block was received from the leader in the current round
func (sr *subroundBlock) isBlockReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.ConsensusState().ConsensusGroup()); i++ {
		node := sr.ConsensusState().ConsensusGroup()[i]
		isJobDone, err := sr.ConsensusState().JobDone(node, SrBlock)

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

// toB64 convert a byte array to a base64 string
func toB64(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}

	return base64.StdEncoding.EncodeToString(buff)
}
