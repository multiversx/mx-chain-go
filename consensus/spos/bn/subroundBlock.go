package bn

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type subroundBlock struct {
	*subround

	sendConsensusMessage func(*spos.ConsensusMessage) bool
}

// NewSubroundBlock creates a subroundBlock object
func NewSubroundBlock(
	subround *subround,
	sendConsensusMessage func(*spos.ConsensusMessage) bool,
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
	sendConsensusMessage func(*spos.ConsensusMessage) bool,
) error {
	if subround == nil {
		return spos.ErrNilSubround
	}

	if sendConsensusMessage == nil {
		return spos.ErrNilSendConsensusMessageFunction
	}

	return nil
}

// doBlockJob method does the job of the block subround
func (sr *subroundBlock) doBlockJob() bool {
	if !sr.GetConsensusState().IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	if sr.GetConsensusState().IsSelfJobDone(SrBlock) {
		return false
	}

	if sr.GetConsensusState().IsCurrentSubroundFinished(SrBlock) {
		return false
	}

	if !sr.sendBlockBody() ||
		!sr.sendBlockHeader() {
		return false
	}

	err := sr.GetConsensusState().SetSelfJobDone(SrBlock, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	sr.GetMultiSigner().SetMessage(sr.GetConsensusState().Data)

	return true
}

// sendBlockBody method job the proposed block body in the Block subround
func (sr *subroundBlock) sendBlockBody() bool {
	startTime := time.Time{}
	startTime = sr.GetConsensusState().RoundTimeStamp
	maxTime := time.Duration(sr.EndTime())
	haveTimeInCurrentSubround := func() bool {
		return sr.GetRounder().RemainingTime(startTime, maxTime) > 0
	}

	blockBody, err := sr.GetBlockProcessor().CreateBlockBody(
		sr.GetRounder().Index(),
		haveTimeInCurrentSubround,
	)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	blkStr, err := sr.GetMarshalizer().Marshal(blockBody)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	msg := spos.NewConsensusMessage(
		nil,
		blkStr,
		[]byte(sr.GetConsensusState().SelfPubKey()),
		nil,
		int(MtBlockBody),
		uint64(sr.GetRounder().TimeStamp().Unix()),
		sr.GetRounder().Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: block body has been sent\n", sr.GetSyncTimer().FormattedCurrentTime()))

	sr.GetConsensusState().BlockBody = blockBody

	return true
}

// sendBlockHeader method job the proposed block header in the Block subround
func (sr *subroundBlock) sendBlockHeader() bool {
	hdr, err := sr.createHeader()
	if err != nil {
		log.Error(err.Error())
		return false
	}

	hdrStr, err := sr.GetMarshalizer().Marshal(hdr)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	hdrHash := sr.GetHasher().Compute(string(hdrStr))

	msg := spos.NewConsensusMessage(
		hdrHash,
		hdrStr,
		[]byte(sr.GetConsensusState().SelfPubKey()),
		nil,
		int(MtBlockHeader),
		uint64(sr.GetRounder().TimeStamp().Unix()),
		sr.GetRounder().Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: block header with nonce %d and hash %s has been sent\n",
		sr.GetSyncTimer().FormattedCurrentTime(), hdr.GetNonce(), toB64(hdrHash)))

	sr.GetConsensusState().Data = hdrHash
	sr.GetConsensusState().Header = hdr

	return true
}

func (sr *subroundBlock) createHeader() (data.HeaderHandler, error) {
	hdr, err := sr.GetBlockProcessor().CreateBlockHeader(sr.GetConsensusState().BlockBody)

	if err != nil {
		return nil, err
	}

	hdr.SetRound(uint32(sr.GetRounder().Index()))
	hdr.SetTimeStamp(uint64(sr.GetRounder().TimeStamp().Unix()))

	if sr.GetChainHandler().GetCurrentBlockHeader() == nil {
		hdr.SetNonce(1)
		hdr.SetPrevHash(sr.GetChainHandler().GetGenesisHeaderHash())
		// Previous random seed is the signature of the previous block
		hdr.SetPrevRandSeed(sr.GetChainHandler().GetGenesisHeader().GetSignature())
	} else {
		hdr.SetNonce(sr.GetChainHandler().GetCurrentBlockHeader().GetNonce() + 1)
		hdr.SetPrevHash(sr.GetChainHandler().GetCurrentBlockHeaderHash())
		// Previous random seed is the signature of the previous block
		hdr.SetPrevRandSeed(sr.GetChainHandler().GetCurrentBlockHeader().GetSignature())
	}

	// currently for bnSPoS RandSeed field is not used
	hdr.SetRandSeed([]byte{0})

	return hdr, nil
}

// receivedBlockBody method is called when a block body is received through the block body channel.
func (sr *subroundBlock) receivedBlockBody(cnsDta *spos.ConsensusMessage) bool {
	node := string(cnsDta.PubKey)

	if sr.GetConsensusState().IsBlockBodyAlreadyReceived() {
		return false
	}

	if !sr.GetConsensusState().IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !sr.GetConsensusState().CanProcessReceivedMessage(cnsDta, sr.GetRounder().Index(), SrBlock) {
		return false
	}

	sr.GetConsensusState().BlockBody = sr.decodeBlockBody(cnsDta.SubRoundData)

	if sr.GetConsensusState().BlockBody == nil {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: block body has been received\n", sr.GetSyncTimer().FormattedCurrentTime()))

	blockProcessedWithSuccess := sr.processReceivedBlock(cnsDta)

	return blockProcessedWithSuccess
}

// decodeBlockBody method decodes block body which is marshalized in the received message
func (sr *subroundBlock) decodeBlockBody(dta []byte) block.Body {
	if dta == nil {
		return nil
	}

	var blk block.Body

	err := sr.GetMarshalizer().Unmarshal(&blk, dta)

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return blk
}

// receivedBlockHeader method is called when a block header is received through the block header channel.
// If the block header is valid, than the validatorRoundStates map corresponding to the node which sent it,
// is set on true for the subround Block
func (sr *subroundBlock) receivedBlockHeader(cnsDta *spos.ConsensusMessage) bool {
	node := string(cnsDta.PubKey)

	if sr.GetConsensusState().IsConsensusDataSet() {
		return false
	}

	if sr.GetConsensusState().IsHeaderAlreadyReceived() {
		return false
	}

	if !sr.GetConsensusState().IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !sr.GetConsensusState().CanProcessReceivedMessage(cnsDta, sr.GetRounder().Index(), SrBlock) {
		return false
	}

	sr.GetConsensusState().Data = cnsDta.BlockHeaderHash
	sr.GetConsensusState().Header = sr.decodeBlockHeader(cnsDta.SubRoundData)

	if sr.GetConsensusState().Header == nil {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: block header with nonce %d and hash %s has been received\n",
		sr.GetSyncTimer().FormattedCurrentTime(), sr.GetConsensusState().Header.GetNonce(), toB64(cnsDta.BlockHeaderHash)))

	if !sr.GetBlockProcessor().CheckBlockValidity(sr.GetChainHandler(), sr.GetConsensusState().Header, nil) {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, invalid block\n",
			sr.GetRounder().Index(), getSubroundName(SrBlock)))

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

	err := sr.GetMarshalizer().Unmarshal(&hdr, dta)

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return &hdr
}

func (sr *subroundBlock) processReceivedBlock(cnsDta *spos.ConsensusMessage) bool {
	if sr.GetConsensusState().BlockBody == nil ||
		sr.GetConsensusState().Header == nil {
		return false
	}

	defer func() {
		sr.GetConsensusState().SetProcessingBlock(false)
	}()

	sr.GetConsensusState().SetProcessingBlock(true)

	node := string(cnsDta.PubKey)

	startTime := time.Time{}
	startTime = sr.GetConsensusState().RoundTimeStamp
	maxTime := sr.GetRounder().TimeDuration() * processingThresholdPercent / 100
	remainingTimeInCurrentRound := func() time.Duration {
		return sr.GetRounder().RemainingTime(startTime, maxTime)
	}

	err := sr.GetBlockProcessor().ProcessBlock(
		sr.GetChainHandler(),
		sr.GetConsensusState().Header,
		sr.GetConsensusState().BlockBody,
		remainingTimeInCurrentRound,
	)

	if cnsDta.RoundIndex < sr.GetRounder().Index() {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, meantime round index has been changed to %d\n",
			cnsDta.RoundIndex, getSubroundName(SrBlock), sr.GetRounder().Index()))
		return false
	}

	if err != nil {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, %s\n",
			sr.GetRounder().Index(), getSubroundName(SrBlock), err.Error()))
		if err == process.ErrTimeIsOut {
			sr.GetConsensusState().RoundCanceled = true
		}
		return false
	}

	sr.GetMultiSigner().SetMessage(sr.GetConsensusState().Data)
	err = sr.GetConsensusState().SetJobDone(node, SrBlock, true)
	if err != nil {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, %s\n",
			sr.GetRounder().Index(), getSubroundName(SrBlock), err.Error()))
		return false
	}

	return true
}

// doBlockConsensusCheck method checks if the consensus in the <BLOCK> subround is achieved
func (sr *subroundBlock) doBlockConsensusCheck() bool {
	if sr.GetConsensusState().RoundCanceled {
		return false
	}

	if sr.GetConsensusState().Status(SrBlock) == spos.SsFinished {
		return true
	}

	threshold := sr.GetConsensusState().Threshold(SrBlock)
	if sr.isBlockReceived(threshold) {
		log.Info(fmt.Sprintf("%sStep 1: subround %s has been finished\n", sr.GetSyncTimer().FormattedCurrentTime(), sr.Name()))
		sr.GetConsensusState().SetStatus(SrBlock, spos.SsFinished)
		return true
	}

	return false
}

// isBlockReceived method checks if the block was received from the leader in the current round
func (sr *subroundBlock) isBlockReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.GetConsensusState().ConsensusGroup()); i++ {
		node := sr.GetConsensusState().ConsensusGroup()[i]
		isJobDone, err := sr.GetConsensusState().JobDone(node, SrBlock)

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
