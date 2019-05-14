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

	sendConsensusMessage func(*consensus.Message) bool
}

// NewSubroundBlock creates a subroundBlock object
func NewSubroundBlock(
	subround *subround,
	sendConsensusMessage func(*consensus.Message) bool,
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
	sendConsensusMessage func(*consensus.Message) bool,
) error {
	if subround == nil {
		return spos.ErrNilSubround
	}

	if subround.ConsensusState == nil {
		return spos.ErrNilConsensusState
	}

	if sendConsensusMessage == nil {
		return spos.ErrNilSendConsensusMessageFunction
	}

	err := spos.ValidateConsensusCore(subround.ConsensusCoreHandler)

	return err
}

// doBlockJob method does the job of the block subround
func (sr *subroundBlock) doBlockJob() bool {
	if !sr.IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	if sr.IsSelfJobDone(SrBlock) {
		return false
	}

	if sr.IsCurrentSubroundFinished(SrBlock) {
		return false
	}

	if !sr.sendBlockBody() ||
		!sr.sendBlockHeader() {
		return false
	}

	err := sr.SetSelfJobDone(SrBlock, true)
	if err != nil {
		log.Error(err.Error())
		return false
	}

	return true
}

// sendBlockBody method job the proposed block body in the Block subround
func (sr *subroundBlock) sendBlockBody() bool {
	startTime := time.Time{}
	startTime = sr.RoundTimeStamp
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
		[]byte(sr.SelfPubKey()),
		nil,
		int(MtBlockBody),
		uint64(sr.Rounder().TimeStamp().Unix()),
		sr.Rounder().Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: block body has been sent\n", sr.SyncTimer().FormattedCurrentTime()))

	sr.BlockBody = blockBody

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
		[]byte(sr.SelfPubKey()),
		nil,
		int(MtBlockHeader),
		uint64(sr.Rounder().TimeStamp().Unix()),
		sr.Rounder().Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: block header with nonce %d and hash %s has been sent\n",
		sr.SyncTimer().FormattedCurrentTime(), hdr.GetNonce(), toB64(hdrHash)))

	sr.Data = hdrHash
	sr.Header = hdr

	return true
}

func (sr *subroundBlock) createHeader() (data.HeaderHandler, error) {
	startTime := time.Time{}
	startTime = sr.RoundTimeStamp
	maxTime := time.Duration(sr.EndTime())
	haveTimeInCurrentSubround := func() bool {
		return sr.Rounder().RemainingTime(startTime, maxTime) > 0
	}

	hdr, err := sr.BlockProcessor().CreateBlockHeader(
		sr.BlockBody,
		sr.Rounder().Index(),
		haveTimeInCurrentSubround)
	if err != nil {
		return nil, err
	}

	hdr.SetRound(uint32(sr.Rounder().Index()))
	hdr.SetTimeStamp(uint64(sr.Rounder().TimeStamp().Unix()))

	var prevRandSeed []byte
	if sr.Blockchain().GetCurrentBlockHeader() == nil {
		hdr.SetNonce(1)
		hdr.SetPrevHash(sr.Blockchain().GetGenesisHeaderHash())

		prevRandSeed = sr.Blockchain().GetGenesisHeader().GetRandSeed()
	} else {
		hdr.SetNonce(sr.Blockchain().GetCurrentBlockHeader().GetNonce() + 1)
		hdr.SetPrevHash(sr.Blockchain().GetCurrentBlockHeaderHash())

		prevRandSeed = sr.Blockchain().GetCurrentBlockHeader().GetRandSeed()
	}
	randSeed, err := sr.BlsSingleSigner().Sign(sr.BlsPrivateKey(), prevRandSeed)
	// Cannot propose block if unable to create random seed
	if err != nil {
		return nil, err
	}

	hdr.SetPrevRandSeed(prevRandSeed)
	hdr.SetRandSeed(randSeed)
	log.Info(fmt.Sprintf("random seed for the next round is %s", toB64(randSeed)))

	return hdr, nil
}

// receivedBlockBody method is called when a block body is received through the block body channel.
func (sr *subroundBlock) receivedBlockBody(cnsDta *consensus.Message) bool {
	node := string(cnsDta.PubKey)

	if sr.IsBlockBodyAlreadyReceived() {
		return false
	}

	if !sr.IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.Rounder().Index(), SrBlock) {
		return false
	}

	sr.BlockBody = sr.decodeBlockBody(cnsDta.SubRoundData)

	if sr.BlockBody == nil {
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
func (sr *subroundBlock) receivedBlockHeader(cnsDta *consensus.Message) bool {
	node := string(cnsDta.PubKey)

	if sr.IsConsensusDataSet() {
		return false
	}

	if sr.IsHeaderAlreadyReceived() {
		return false
	}

	if !sr.IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !sr.CanProcessReceivedMessage(cnsDta, sr.Rounder().Index(), SrBlock) {
		return false
	}

	sr.Data = cnsDta.BlockHeaderHash
	sr.Header = sr.decodeBlockHeader(cnsDta.SubRoundData)

	if sr.Header == nil {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: block header with nonce %d and hash %s has been received\n",
		sr.SyncTimer().FormattedCurrentTime(), sr.Header.GetNonce(), toB64(cnsDta.BlockHeaderHash)))

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

func (sr *subroundBlock) processReceivedBlock(cnsDta *consensus.Message) bool {
	if sr.BlockBody == nil ||
		sr.Header == nil {
		return false
	}

	defer func() {
		sr.SetProcessingBlock(false)
	}()

	sr.SetProcessingBlock(true)

	node := string(cnsDta.PubKey)

	startTime := time.Time{}
	startTime = sr.RoundTimeStamp
	maxTime := sr.Rounder().TimeDuration() * processingThresholdPercent / 100
	remainingTimeInCurrentRound := func() time.Duration {
		return sr.Rounder().RemainingTime(startTime, maxTime)
	}

	err := sr.BlockProcessor().ProcessBlock(
		sr.Blockchain(),
		sr.Header,
		sr.BlockBody,
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
			sr.RoundCanceled = true
		}
		return false
	}

	err = sr.SetJobDone(node, SrBlock, true)
	if err != nil {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, %s\n",
			sr.Rounder().Index(), getSubroundName(SrBlock), err.Error()))
		return false
	}

	return true
}

// doBlockConsensusCheck method checks if the consensus in the <BLOCK> subround is achieved
func (sr *subroundBlock) doBlockConsensusCheck() bool {
	if sr.RoundCanceled {
		return false
	}

	if sr.Status(SrBlock) == spos.SsFinished {
		return true
	}

	threshold := sr.Threshold(SrBlock)
	if sr.isBlockReceived(threshold) {
		log.Info(fmt.Sprintf("%sStep 1: subround %s has been finished\n", sr.SyncTimer().FormattedCurrentTime(), sr.Name()))
		sr.SetStatus(SrBlock, spos.SsFinished)
		return true
	}

	return false
}

// isBlockReceived method checks if the block was received from the leader in the current round
func (sr *subroundBlock) isBlockReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.ConsensusGroup()); i++ {
		node := sr.ConsensusGroup()[i]
		isJobDone, err := sr.JobDone(node, SrBlock)

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
