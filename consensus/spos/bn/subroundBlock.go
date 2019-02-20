package bn

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

type subroundBlock struct {
	*subround

	blockChain       *blockchain.BlockChain
	blockProcessor   process.BlockProcessor
	consensusState   *spos.ConsensusState
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	multiSigner      crypto.MultiSigner
	rounder          round.Rounder
	shardCoordinator sharding.ShardCoordinator
	syncTimer        ntp.SyncTimer

	sendConsensusMessage func(*spos.ConsensusData) bool
}

// NewSubroundBlock creates a subroundBlock object
func NewSubroundBlock(
	subround *subround,
	blockChain *blockchain.BlockChain,
	blockProcessor process.BlockProcessor,
	consensusState *spos.ConsensusState,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	multiSigner crypto.MultiSigner,
	rounder round.Rounder,
	shardCoordinator sharding.ShardCoordinator,
	syncTimer ntp.SyncTimer,
	sendConsensusMessage func(*spos.ConsensusData) bool,
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
	blockChain *blockchain.BlockChain,
	blockProcessor process.BlockProcessor,
	consensusState *spos.ConsensusState,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	multiSigner crypto.MultiSigner,
	rounder round.Rounder,
	shardCoordinator sharding.ShardCoordinator,
	syncTimer ntp.SyncTimer,
	sendConsensusMessage func(*spos.ConsensusData) bool,
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

// doBlockJob method actually job the proposed block in the Block subround, when this node is leader
// (it is used as a handler function of the doSubroundJob pointer function declared in subround struct,
// from spos package)
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
	haveTimeInCurrentSubround := func() bool {
		roundStartTime := sr.rounder.TimeStamp()
		currentTime := sr.syncTimer.CurrentTime()
		elapsedTime := currentTime.Sub(roundStartTime)
		haveTime := float64(sr.EndTime()) - float64(elapsedTime)

		return time.Duration(haveTime) > 0
	}

	blk, err := sr.blockProcessor.CreateTxBlockBody(
		sr.shardCoordinator.ShardForCurrentNode(),
		maxTransactionsInBlock,
		sr.rounder.Index(),
		haveTimeInCurrentSubround,
	)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	blkStr, err := sr.marshalizer.Marshal(blk)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	dta := spos.NewConsensusData(
		nil,
		blkStr,
		[]byte(sr.consensusState.SelfPubKey()),
		nil,
		int(MtBlockBody),
		uint64(sr.rounder.TimeStamp().Unix()),
		sr.rounder.Index())

	if !sr.sendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: Sending block body\n", sr.syncTimer.FormattedCurrentTime()))

	sr.consensusState.BlockBody = blk

	return true
}

// sendBlockHeader method job the proposed block header in the Block subround
func (sr *subroundBlock) sendBlockHeader() bool {
	hdr := &block.Header{}

	hdr.Round = uint32(sr.rounder.Index())
	hdr.TimeStamp = uint64(sr.rounder.TimeStamp().Unix())

	if sr.blockChain.CurrentBlockHeader == nil {
		hdr.Nonce = 1
		hdr.PrevHash = sr.blockChain.GenesisHeaderHash
	} else {
		hdr.Nonce = sr.blockChain.CurrentBlockHeader.Nonce + 1
		hdr.PrevHash = sr.blockChain.CurrentBlockHeaderHash
	}

	blkStr, err := sr.marshalizer.Marshal(sr.consensusState.BlockBody)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	hdr.BlockBodyHash = sr.hasher.Compute(string(blkStr))

	hdrStr, err := sr.marshalizer.Marshal(hdr)

	if err != nil {
		log.Error(err.Error())
		return false
	}

	hdrHash := sr.hasher.Compute(string(hdrStr))

	dta := spos.NewConsensusData(
		hdrHash,
		hdrStr,
		[]byte(sr.consensusState.SelfPubKey()),
		nil,
		int(MtBlockHeader),
		uint64(sr.rounder.TimeStamp().Unix()),
		sr.rounder.Index())

	if !sr.sendConsensusMessage(dta) {
		return false
	}

	log.Info(fmt.Sprintf("%sStep 1: Sending block header with nonce %d and hash %s\n",
		sr.syncTimer.FormattedCurrentTime(), hdr.Nonce, toB64(hdrHash)))

	sr.consensusState.Data = hdrHash
	sr.consensusState.Header = hdr

	return true
}

// receivedBlockBody method is called when a block body is received through the block body channel.
func (sr *subroundBlock) receivedBlockBody(cnsDta *spos.ConsensusData) bool {
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

	log.Info(fmt.Sprintf("%sStep 1: Received block body\n", sr.syncTimer.FormattedCurrentTime()))

	blockProcessedWithSuccess := sr.processReceivedBlock(cnsDta)

	return blockProcessedWithSuccess
}

// decodeBlockBody method decodes block body which is marshalized in the received message
func (sr *subroundBlock) decodeBlockBody(dta []byte) *block.TxBlockBody {
	if dta == nil {
		return nil
	}

	var blk block.TxBlockBody

	err := sr.marshalizer.Unmarshal(&blk, dta)

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return &blk
}

// receivedBlockHeader method is called when a block header is received through the block header channel.
// If the block header is valid, than the validatorRoundStates map coresponding to the node which sent it,
// is set on true for the subround Block
func (sr *subroundBlock) receivedBlockHeader(cnsDta *spos.ConsensusData) bool {
	node := string(cnsDta.PubKey)

	if sr.consensusState.IsConsensusDataAlreadySet() {
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

	log.Info(fmt.Sprintf("%sStep 1: Received block header with nonce %d and hash %s\n",
		sr.syncTimer.FormattedCurrentTime(), sr.consensusState.Header.Nonce, toB64(cnsDta.BlockHeaderHash)))

	if !sr.checkIfBlockIsValid(sr.consensusState.Header) {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, INVALID BLOCK\n",
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

func (sr *subroundBlock) processReceivedBlock(cnsDta *spos.ConsensusData) bool {
	if sr.consensusState.BlockBody == nil ||
		sr.consensusState.Header == nil {
		return false
	}

	node := string(cnsDta.PubKey)

	haveTimeInCurrentRound := func() time.Duration {
		roundStartTime := sr.rounder.TimeStamp()
		currentTime := sr.syncTimer.CurrentTime()
		elapsedTime := currentTime.Sub(roundStartTime)
		haveTime := float64(sr.rounder.TimeDuration())*maxBlockProcessingTimePercent - float64(elapsedTime)

		return time.Duration(haveTime)
	}

	err := sr.blockProcessor.ProcessBlock(sr.blockChain, sr.consensusState.Header, sr.consensusState.BlockBody, haveTimeInCurrentRound)

	if err != nil {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, %s\n",
			sr.rounder.Index(), getSubroundName(SrBlock), err.Error()))

		return false
	}

	if haveTimeInCurrentRound() < 0 {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, TIME IS OUT\n",
			cnsDta.RoundIndex, getSubroundName(SrBlock)))

		sr.blockProcessor.RevertAccountState()

		return false
	}

	sr.multiSigner.SetMessage(sr.consensusState.Data)
	err = sr.consensusState.SetJobDone(node, SrBlock, true)

	if err != nil {
		log.Info(fmt.Sprintf("Canceled round %d in subround %s, %s\n",
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
		sr.printBlockCM() // only for printing block consensus messages
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
		isJobDone, err := sr.consensusState.GetJobDone(node, SrBlock)

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

// checkIfBlockIsValid method checks if the received block is valid
func (sr *subroundBlock) checkIfBlockIsValid(receivedHeader *block.Header) bool {
	if sr.blockChain.CurrentBlockHeader == nil {
		if receivedHeader.Nonce == 1 { // first block after genesis
			if bytes.Equal(receivedHeader.PrevHash, sr.blockChain.GenesisHeaderHash) {
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

	if receivedHeader.Nonce < sr.blockChain.CurrentBlockHeader.Nonce+1 {
		log.Info(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d\n",
			sr.blockChain.CurrentBlockHeader.Nonce, receivedHeader.Nonce))

		return false
	}

	if receivedHeader.Nonce == sr.blockChain.CurrentBlockHeader.Nonce+1 {
		prevHeaderHash := sr.getHeaderHash(sr.blockChain.CurrentBlockHeader)

		if bytes.Equal(receivedHeader.PrevHash, prevHeaderHash) {
			return true
		}

		log.Info(fmt.Sprintf("Hash not match: local block hash is %s and node received block with previous hash %s\n",
			toB64(prevHeaderHash), toB64(receivedHeader.PrevHash)))

		return false
	}

	log.Info(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d\n",
		sr.blockChain.CurrentBlockHeader.Nonce, receivedHeader.Nonce))

	return false
}

func (sr *subroundBlock) getHeaderHash(hdr *block.Header) []byte {
	headerMarsh, err := sr.marshalizer.Marshal(hdr)

	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return sr.hasher.Compute(string(headerMarsh))
}

// printBlockCM method prints the (BLOCK) subround consensus messages
func (sr *subroundBlock) printBlockCM() {
	if !sr.consensusState.IsSelfLeaderInCurrentRound() {
		log.Info(fmt.Sprintf("%sStep 1: Synchronized block\n", sr.syncTimer.FormattedCurrentTime()))
	}

	log.Info(fmt.Sprintf("%sStep 1: subround (BLOCK) has been finished\n", sr.syncTimer.FormattedCurrentTime()))
}

// toB64 convert a byte array to a base64 string
func toB64(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}

	return base64.StdEncoding.EncodeToString(buff)
}
