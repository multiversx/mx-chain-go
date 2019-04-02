package bn

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

type subroundBlock struct {
	*subround

	blockChain       data.ChainHandler
	blockProcessor   process.BlockProcessor
	consensusState   *spos.ConsensusState
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	multiSigner      crypto.MultiSigner
	rounder          consensus.Rounder
	shardCoordinator sharding.Coordinator
	syncTimer        ntp.SyncTimer

	sendConsensusMessage func(*spos.ConsensusMessage) bool
}

// NewSubroundBlock creates a subroundBlock object
func NewSubroundBlock(
	subround *subround,
	blockChain data.ChainHandler,
	blockProcessor process.BlockProcessor,
	consensusState *spos.ConsensusState,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	multiSigner crypto.MultiSigner,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	syncTimer ntp.SyncTimer,
	sendConsensusMessage func(*spos.ConsensusMessage) bool,
	extend func(subroundId int),
) (*subroundBlock, error) {

	err := checkNewSubroundBlockParams(
		subround,
		blockChain,
		blockProcessor,
		consensusState,
		hasher,
		marshalizer,
		multiSigner,
		rounder,
		shardCoordinator,
		syncTimer,
		sendConsensusMessage,
	)

	if err != nil {
		return nil, err
	}

	srBlock := subroundBlock{
		subround,
		blockChain,
		blockProcessor,
		consensusState,
		hasher,
		marshalizer,
		multiSigner,
		rounder,
		shardCoordinator,
		syncTimer,
		sendConsensusMessage,
	}

	srBlock.job = srBlock.doBlockJob
	srBlock.check = srBlock.doBlockConsensusCheck
	srBlock.extend = extend

	return &srBlock, nil
}

func checkNewSubroundBlockParams(
	subround *subround,
	blockChain data.ChainHandler,
	blockProcessor process.BlockProcessor,
	consensusState *spos.ConsensusState,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	multiSigner crypto.MultiSigner,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	syncTimer ntp.SyncTimer,
	sendConsensusMessage func(*spos.ConsensusMessage) bool,
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

	if hasher == nil {
		return spos.ErrNilHasher
	}

	if marshalizer == nil {
		return spos.ErrNilMarshalizer
	}

	if multiSigner == nil {
		return spos.ErrNilMultiSigner
	}

	if rounder == nil {
		return spos.ErrNilRounder
	}

	if shardCoordinator == nil {
		return spos.ErrNilShardCoordinator
	}

	if syncTimer == nil {
		return spos.ErrNilSyncTimer
	}

	if sendConsensusMessage == nil {
		return spos.ErrNilSendConsensusMessageFunction
	}

	return nil
}

// doBlockJob method does the job of the block subround
func (sr *subroundBlock) doBlockJob() bool {
	if !sr.consensusState.IsSelfLeaderInCurrentRound() { // is NOT self leader in this round?
		return false
	}

	if sr.consensusState.IsSelfJobDone(SrBlock) {
		return false
	}

	if sr.consensusState.IsCurrentSubroundFinished(SrBlock) {
		return false
	}

	if !sr.sendBlockBody() ||
		!sr.sendBlockHeader() {
		return false
	}

	err := sr.consensusState.SetSelfJobDone(SrBlock, true)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	sr.multiSigner.SetMessage(sr.consensusState.Data)

	return true
}

// sendBlockBody method job the proposed block body in the Block subround
func (sr *subroundBlock) sendBlockBody() bool {
	startTime := time.Time{}
	startTime = sr.consensusState.RoundTimeStamp
	maxTime := time.Duration(sr.EndTime())
	haveTimeInCurrentSubround := func() bool {
		return sr.rounder.RemainingTime(startTime, maxTime) > 0
	}

	blockBody, err := sr.blockProcessor.CreateBlockBody(
		sr.rounder.Index(),
		haveTimeInCurrentSubround,
	)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	blkStr, err := sr.marshalizer.Marshal(blockBody)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	msg := spos.NewConsensusMessage(
		nil,
		blkStr,
		[]byte(sr.consensusState.SelfPubKey()),
		nil,
		int(MtBlockBody),
		uint64(sr.rounder.TimeStamp().Unix()),
		sr.rounder.Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: block body has been sent\n", sr.syncTimer.FormattedCurrentTime()))

	sr.consensusState.BlockBody = blockBody

	return true
}

// sendBlockHeader method job the proposed block header in the Block subround
func (sr *subroundBlock) sendBlockHeader() bool {
	hdr, err := sr.createHeader()
	if err != nil {
		log.Error(err.Error())
		return false
	}

	hdrStr, err := sr.marshalizer.Marshal(hdr)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	hdrHash := sr.hasher.Compute(string(hdrStr))

	msg := spos.NewConsensusMessage(
		hdrHash,
		hdrStr,
		[]byte(sr.consensusState.SelfPubKey()),
		nil,
		int(MtBlockHeader),
		uint64(sr.rounder.TimeStamp().Unix()),
		sr.rounder.Index())

	if !sr.sendConsensusMessage(msg) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: block header with nonce %d and hash %s has been sent\n",
		sr.syncTimer.FormattedCurrentTime(), hdr.GetNonce(), toB64(hdrHash)))

	sr.consensusState.Data = hdrHash
	sr.consensusState.Header = hdr

	return true
}

func (sr *subroundBlock) createHeader() (data.HeaderHandler, error) {
	hdr, err := sr.blockProcessor.CreateBlockHeader(sr.consensusState.BlockBody)

	if err != nil {
		return nil, err
	}

	hdr.SetRound(uint32(sr.rounder.Index()))
	hdr.SetTimeStamp(uint64(sr.rounder.TimeStamp().Unix()))

	if sr.blockChain.GetCurrentBlockHeader() == nil {
		hdr.SetNonce(1)
		hdr.SetPrevHash(sr.blockChain.GetGenesisHeaderHash())
		// Previous random seed is the signature of the previous block
		hdr.SetPrevRandSeed(sr.blockChain.GetGenesisHeader().GetSignature())
	} else {
		hdr.SetNonce(sr.blockChain.GetCurrentBlockHeader().GetNonce() + 1)
		hdr.SetPrevHash(sr.blockChain.GetCurrentBlockHeaderHash())
		// Previous random seed is the signature of the previous block
		hdr.SetPrevRandSeed(sr.blockChain.GetCurrentBlockHeader().GetSignature())
	}

	// currently for bnSPoS RandSeed field is not used
	hdr.SetRandSeed([]byte{0})

	return hdr, nil
}

// receivedBlockBody method is called when a block body is received through the block body channel.
func (sr *subroundBlock) receivedBlockBody(cnsDta *spos.ConsensusMessage) bool {
	node := string(cnsDta.PubKey)

	if sr.consensusState.IsBlockBodyAlreadyReceived() {
		return false
	}

	if !sr.consensusState.IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !sr.consensusState.CanProcessReceivedMessage(cnsDta, sr.rounder.Index(), SrBlock) {
		return false
	}

	sr.consensusState.BlockBody = sr.decodeBlockBody(cnsDta.SubRoundData)

	if sr.consensusState.BlockBody == nil {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: block body has been received\n", sr.syncTimer.FormattedCurrentTime()))

	blockProcessedWithSuccess := sr.processReceivedBlock(cnsDta)

	return blockProcessedWithSuccess
}

// decodeBlockBody method decodes block body which is marshalized in the received message
func (sr *subroundBlock) decodeBlockBody(dta []byte) block.Body {
	if dta == nil {
		return nil
	}

	var blk block.Body

	err := sr.marshalizer.Unmarshal(&blk, dta)

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

	if sr.consensusState.IsConsensusDataSet() {
		return false
	}

	if sr.consensusState.IsHeaderAlreadyReceived() {
		return false
	}

	if !sr.consensusState.IsNodeLeaderInCurrentRound(node) { // is NOT this node leader in current round?
		return false
	}

	if !sr.consensusState.CanProcessReceivedMessage(cnsDta, sr.rounder.Index(), SrBlock) {
		return false
	}

	sr.consensusState.Data = cnsDta.BlockHeaderHash
	sr.consensusState.Header = sr.decodeBlockHeader(cnsDta.SubRoundData)

	if sr.consensusState.Header == nil {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: block header with nonce %d and hash %s has been received\n",
		sr.syncTimer.FormattedCurrentTime(), sr.consensusState.Header.GetNonce(), toB64(cnsDta.BlockHeaderHash)))

	if !sr.blockProcessor.CheckBlockValidity(sr.blockChain, sr.consensusState.Header, nil) {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, invalid block\n",
			sr.rounder.Index(), getSubroundName(SrBlock)))

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

	err := sr.marshalizer.Unmarshal(&hdr, dta)

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return &hdr
}

func (sr *subroundBlock) processReceivedBlock(cnsDta *spos.ConsensusMessage) bool {
	if sr.consensusState.BlockBody == nil ||
		sr.consensusState.Header == nil {
		return false
	}

	defer func() {
		sr.consensusState.SetProcessingBlock(false)
	}()

	sr.consensusState.SetProcessingBlock(true)

	node := string(cnsDta.PubKey)

	startTime := time.Time{}
	startTime = sr.consensusState.RoundTimeStamp
	maxTime := sr.rounder.TimeDuration() * processingThresholdPercent / 100
	remainingTimeInCurrentRound := func() time.Duration {
		return sr.rounder.RemainingTime(startTime, maxTime)
	}

	err := sr.blockProcessor.ProcessBlock(
		sr.blockChain,
		sr.consensusState.Header,
		sr.consensusState.BlockBody,
		remainingTimeInCurrentRound,
	)

	if cnsDta.RoundIndex < sr.rounder.Index() {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, meantime round index has been changed to %d\n",
			cnsDta.RoundIndex, getSubroundName(SrBlock), sr.rounder.Index()))
		return false
	}

	if err != nil {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, %s\n",
			sr.rounder.Index(), getSubroundName(SrBlock), err.Error()))
		if err == process.ErrTimeIsOut {
			sr.consensusState.RoundCanceled = true
		}
		return false
	}

	sr.multiSigner.SetMessage(sr.consensusState.Data)
	err = sr.consensusState.SetJobDone(node, SrBlock, true)
	if err != nil {
		log.Info(fmt.Sprintf("canceled round %d in subround %s, %s\n",
			sr.rounder.Index(), getSubroundName(SrBlock), err.Error()))
		return false
	}

	return true
}

// doBlockConsensusCheck method checks if the consensus in the <BLOCK> subround is achieved
func (sr *subroundBlock) doBlockConsensusCheck() bool {
	if sr.consensusState.RoundCanceled {
		return false
	}

	if sr.consensusState.Status(SrBlock) == spos.SsFinished {
		return true
	}

	threshold := sr.consensusState.Threshold(SrBlock)
	if sr.isBlockReceived(threshold) {
		log.Info(fmt.Sprintf("%sStep 1: subround %s has been finished\n", sr.syncTimer.FormattedCurrentTime(), sr.Name()))
		sr.consensusState.SetStatus(SrBlock, spos.SsFinished)
		return true
	}

	return false
}

// isBlockReceived method checks if the block was received from the leader in the current round
func (sr *subroundBlock) isBlockReceived(threshold int) bool {
	n := 0

	for i := 0; i < len(sr.consensusState.ConsensusGroup()); i++ {
		node := sr.consensusState.ConsensusGroup()[i]
		isJobDone, err := sr.consensusState.JobDone(node, SrBlock)

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
