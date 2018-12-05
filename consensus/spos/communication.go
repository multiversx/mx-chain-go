package spos

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transactionPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"math/big"
)

const shardsCount = 1024            // TODO: current numbers of shards (this should be injected, and this const should be removed later)
const maxTransactionsInBlock = 1000 // TODO: maximum transactions in one block (this should be injected, and this const should be removed later)

// MessageType specifies what type of message was received
type MessageType int

const (
	// MtBlockBody defines ID of a message that has a block body inside
	MtBlockBody MessageType = iota
	// MtBlockHeader defines ID of a message that has a block header inside
	MtBlockHeader
	// MtCommitmentHash defines ID of a message that has a commitment hash inside
	MtCommitmentHash
	// MtBitmap defines ID of a message that has a bitmap inside
	MtBitmap
	// MtCommitment defines ID of a message that has a commitment inside
	MtCommitment
	// MtSignature defines ID of a message that has a Signature inside
	MtSignature
	// MtUnknown defines ID of a message that has an unknown Data inside
	MtUnknown
)

// ConsensusData defines the data needed by spos to comunicate between nodes over network in all subrounds
type ConsensusData struct {
	Data      []byte
	PubKeys   [][]byte
	Signature []byte
	MsgType   MessageType
	TimeStamp []byte
}

// NewConsensusData creates a new ConsensusData object
func NewConsensusData(
	dta []byte,
	pks [][]byte,
	sig []byte,
	msg MessageType,
	tms []byte,
) *ConsensusData {

	return &ConsensusData{
		Data:      dta,
		PubKeys:   pks,
		Signature: sig,
		MsgType:   msg,
		TimeStamp: tms}
}

// Communication defines the data needed by spos to comunicate between nodes which are in the validators group
type Communication struct {
	log bool

	Cns *Consensus

	Hdr  *block.Header
	Blk  *block.Block
	Blkc *blockchain.BlockChain
	TxP  *transactionPool.TransactionPool

	Rounds          int // only for statistic
	RoundsWithBlock int // only for statistic

	ChRcvMsg map[MessageType]chan *ConsensusData

	OnSendMessage func([]byte) // this is a pointer to a function which actually send the message from a node to the
	// network
}

// NewCommunication creates a new Communication object
func NewCommunication(
	log bool,
	cns *Consensus,
	blkc *blockchain.BlockChain,
	txp *transactionPool.TransactionPool,
) *Communication {

	com := Communication{
		log:  log,
		Cns:  cns,
		Blkc: blkc,
		TxP:  txp}

	com.Rounds = 0          // only for statistic
	com.RoundsWithBlock = 0 // only for statistic

	com.ChRcvMsg = make(map[MessageType]chan *ConsensusData)

	nodes := 0

	if cns != nil &&
		cns.Validators != nil {
		nodes = len(cns.Validators.ConsensusGroup())
	}

	com.ChRcvMsg[MtBlockBody] = make(chan *ConsensusData, nodes)
	com.ChRcvMsg[MtBlockHeader] = make(chan *ConsensusData, nodes)
	com.ChRcvMsg[MtCommitmentHash] = make(chan *ConsensusData, nodes)
	com.ChRcvMsg[MtBitmap] = make(chan *ConsensusData, nodes)
	com.ChRcvMsg[MtCommitment] = make(chan *ConsensusData, nodes)
	com.ChRcvMsg[MtSignature] = make(chan *ConsensusData, nodes)

	go com.CheckChannels()

	return &com
}

// StartRound method is the function which actually do the job of the StartRound subround (it is used as the handler
// function of the doSubroundJob pointer variable function in Subround struct from spos package)
func (com *Communication) StartRound() bool {
	com.Blk = nil
	com.Hdr = nil
	com.Cns.Data = nil
	com.Cns.ResetRoundStatus()
	com.Cns.ResetValidation()

	leader, err := com.Cns.GetLeader()

	if err != nil {
		leader = "Unknown"
	}

	if leader == com.Cns.SelfId() {
		leader = fmt.Sprintf(leader + " (MY TURN)")
	}

	com.Log(fmt.Sprintf(com.GetFormatedTime()+"Step 0: Preparing for this round with leader %s ", leader))

	return true
}

// EndRound method is the function which actually do the job of the EndRound subround (it is used as the handler
// function of the doSubroundJob pointer variable function in Subround struct from spos package)
func (com *Communication) EndRound() bool {
	if !com.Cns.CheckConsensus(SrEndRound) {
		return false
	}

	header, err := marshal.DefMarsh.Marshal(com.Hdr)

	if err != nil {
		com.Log(fmt.Sprintf(com.GetFormatedTime() + err.Error()))
		return false
	}

	body, err := marshal.DefMarsh.Marshal(com.Blk)

	if err != nil {
		com.Log(fmt.Sprintf(com.GetFormatedTime() + err.Error()))
		return false
	}

	com.Blkc.CurrentBlock = com.Hdr
	com.Blkc.LocalHeight = new(big.Int).SetUint64(com.Hdr.Nonce)
	com.Blkc.Put(blockchain.BlockHeaderUnit, com.Hdr.BlockHash, header)
	com.Blkc.Put(blockchain.BlockUnit, com.Hdr.BlockHash, body)

	// TODO: Here the block should be add in the block pool, when its implementation will be finished

	if com.Blk != nil { // remove transactions included in the committed block, from the transaction pool
		for i := 0; i < len(com.Blk.MiniBlocks); i++ {
			for j := 0; j < len(com.Blk.MiniBlocks[i].TxHashes); j++ {
				com.TxP.RemoveTransaction(com.Blk.MiniBlocks[i].TxHashes[j], com.Blk.MiniBlocks[i].DestShardID)
			}
		}
	}

	if com.Cns.IsNodeLeaderInCurrentRound(com.Cns.SelfId()) {
		com.Log(fmt.Sprintf("\n"+com.GetFormatedTime()+
			">>>>>>>>>>>>>>>>>>>> ADDED PROPOSED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n",
			com.Hdr.Nonce))
	} else {
		com.Log(fmt.Sprintf("\n"+com.GetFormatedTime()+
			">>>>>>>>>>>>>>>>>>>> ADDED SYNCHRONIZED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n",
			com.Hdr.Nonce))
	}

	com.Rounds++          // only for statistic
	com.RoundsWithBlock++ // only for statistic

	return true
}

// SendBlock method actually send the proposed block in the Block subround, when this node is leader (it is used as
// a handler function of the doSubroundJob pointer function declared in Subround struct from spos package)
func (com *Communication) SendBlock() bool {
	if com.ShouldSynch() { // if node is not synchronized yet, it has to continue the bootstrapping mechanism
		com.Log(fmt.Sprintf(com.GetFormatedTime()+
			"Canceled round %d in subround %s", com.Cns.Chr.Round().Index(), com.Cns.GetSubroundName(SrBlock)))
		com.Cns.Chr.SetSelfSubround(-1)
		return false
	}

	if com.Cns.Status(SrBlock) == SsFinished || // check if the Block subround is already finished
		com.Cns.GetValidation(com.Cns.SelfId(), SrBlock) || // check if the block has been already sent
		!com.Cns.IsNodeLeaderInCurrentRound(com.Cns.SelfId()) { // check if the leader of this round is another node
		return false
	}

	if !com.SendBlockBody() ||
		!com.SendBlockHeader() {
		return false
	}

	com.Cns.SetValidation(com.Cns.SelfId(), SrBlock, true)
	com.Cns.SetShouldCheckConsensus(true)

	return true
}

// SendBlockBody method send the proposed block body in the Block subround
func (com *Communication) SendBlockBody() bool {
	blk := &block.Block{}

	blk.MiniBlocks = com.CreateMiniBlocks()

	message, err := marshal.DefMarsh.Marshal(blk)

	if err != nil {
		com.Log(fmt.Sprintf(com.GetFormatedTime() + err.Error()))
		return false
	}

	dta := NewConsensusData(
		message,
		nil,
		[]byte(com.Cns.SelfId()),
		MtBlockBody,
		[]byte(com.GetTime()))

	if !com.BroadcastMessage(dta) {
		return false
	}

	com.Log(fmt.Sprintf(com.GetFormatedTime() + "Step 1: Sending block body"))

	com.Blk = blk

	return true
}

// CreateMiniBlocks method will create mini blocks with transactions from mini pool. It will check if the transaction
// is valid and than it will add it in the specific miniblock, depending of the destination shard id. If the transactions
// count or the time needed for this action will exceed the limits given, it will stop the procedure and will return
// the state of its work until than.
func (com *Communication) CreateMiniBlocks() []block.MiniBlock {
	currentSubRound := com.GetSubround()
	mblkc := make([]block.MiniBlock, 0)

	for i, txs := 0, 0; i < shardsCount; i++ {
		txStore := com.TxP.MiniPoolTxStore(uint32(i))

		if txStore == nil {
			continue
		}

		mblk := block.MiniBlock{}
		mblk.DestShardID = uint32(i)
		mblk.TxHashes = make([][]byte, 0)

		for _, txHash := range txStore.Keys() {
			// TODO: Here the transaction execution should be called, when its implementation will be finished, to check
			// TODO: if the selected transaction is ok to be added

			mblk.TxHashes = append(mblk.TxHashes, txHash)
			txs++

			if txs >= maxTransactionsInBlock { // max transactions count in one block was reached
				mblkc = append(mblkc, mblk)
				return mblkc
			}
		}

		if com.GetSubround() > currentSubRound { // time is out
			mblkc = append(mblkc, mblk)
			return mblkc
		}

		mblkc = append(mblkc, mblk)
	}

	return mblkc
}

// GetSubround method returns current subround taking in consideration the current time
func (com *Communication) GetSubround() chronology.SubroundId {
	return com.Cns.Chr.GetSubroundFromDateTime(com.Cns.Chr.SyncTime().CurrentTime(com.Cns.Chr.ClockOffset()))
}

// SendBlockHeader method send the proposed block header in the Block subround
func (com *Communication) SendBlockHeader() bool {
	hdr := &block.Header{}

	if com.Blkc.CurrentBlock == nil {
		hdr.Nonce = 1
		hdr.Round = uint32(com.Cns.Chr.Round().Index())
		hdr.TimeStamp = []byte(com.GetTime())
	} else {
		hdr.Nonce = com.Blkc.CurrentBlock.Nonce + 1
		hdr.Round = uint32(com.Cns.Chr.Round().Index())
		hdr.TimeStamp = []byte(com.GetTime())
		hdr.PrevHash = com.Blkc.CurrentBlock.BlockHash
	}

	message, err := marshal.DefMarsh.Marshal(com.Blk)

	if err != nil {
		com.Log(fmt.Sprintf(com.GetFormatedTime() + err.Error()))
		return false
	}

	hdr.BlockHash = hashing.DefHash.Compute(string(message))

	message, err = marshal.DefMarsh.Marshal(hdr)

	if err != nil {
		com.Log(fmt.Sprintf(com.GetFormatedTime() + err.Error()))
		return false
	}

	dta := NewConsensusData(
		message,
		nil,
		[]byte(com.Cns.SelfId()),
		MtBlockHeader,
		[]byte(com.GetTime()))

	if !com.BroadcastMessage(dta) {
		return false
	}

	com.Log(fmt.Sprintf(com.GetFormatedTime() + "Step 1: Sending block header"))

	com.Hdr = hdr
	com.Cns.Data = &com.Hdr.BlockHash

	return true
}

// SendCommitmentHash method is the function which is actually used to send the commitment hash for the received block
// from the leader in the CommitmentHash subround (it is used as the handler function of the doSubroundJob pointer
// variable function in Subround struct from spos package)
func (com *Communication) SendCommitmentHash() bool {
	// check if the Block subround is not finished
	if com.Cns.Status(SrBlock) != SsFinished {
		return com.SendBlock()
	}

	if com.Cns.Status(SrCommitmentHash) == SsFinished || // check if the CommitmentHash subround is already finished
		com.Cns.GetValidation(com.Cns.SelfId(), SrCommitmentHash) || // check if the commitment hash has been already sent
		com.Cns.Data == nil { // check if this node has a consensus data on which it should send the commitment hash
		return false
	}

	dta := NewConsensusData(
		*com.Cns.Data,
		nil,
		[]byte(com.Cns.SelfId()),
		MtCommitmentHash,
		[]byte(com.GetTime()))

	if !com.BroadcastMessage(dta) {
		return false
	}

	com.Log(fmt.Sprintf(com.GetFormatedTime() + "Step 2: Sending commitment hash"))

	com.Cns.SetValidation(com.Cns.SelfId(), SrCommitmentHash, true)
	com.Cns.SetShouldCheckConsensus(true)

	return true
}

// SendBitmap method is the function which is actually used to send the bitmap with the commitment hashes received,
// in the Bitmap subround, when this node is leader (it is used as the handler function of the doSubroundJob pointer
// variable function in Subround struct from spos package)
func (com *Communication) SendBitmap() bool {
	// check if the CommitmentHash subround is not finished
	if com.Cns.Status(SrCommitmentHash) != SsFinished {
		return com.SendCommitmentHash()
	}

	if com.Cns.Status(SrBitmap) == SsFinished || // check if the Bitmap subround is already finished
		com.Cns.GetValidation(com.Cns.SelfId(), SrBitmap) || // check if the bitmap has been already sent
		!com.Cns.IsNodeLeaderInCurrentRound(com.Cns.SelfId()) || // check if this node is leader in the current round
		com.Cns.Data == nil { // check if this node has a consensus data on which it should send the bitmap
		return false
	}

	pks := make([][]byte, 0)

	for i := 0; i < len(com.Cns.ConsensusGroup()); i++ {
		if com.Cns.GetValidation(com.Cns.ConsensusGroup()[i], SrCommitmentHash) {
			pks = append(pks, []byte(com.Cns.ConsensusGroup()[i]))
		}
	}

	dta := NewConsensusData(
		*com.Cns.Data,
		pks,
		[]byte(com.Cns.SelfId()),
		MtBitmap,
		[]byte(com.GetTime()))

	if !com.BroadcastMessage(dta) {
		return false
	}

	com.Log(fmt.Sprintf(com.GetFormatedTime() + "Step 3: Sending bitmap"))

	for i := 0; i < len(com.Cns.ConsensusGroup()); i++ {
		if com.Cns.GetValidation(com.Cns.ConsensusGroup()[i], SrCommitmentHash) {
			com.Cns.SetValidation(com.Cns.ConsensusGroup()[i], SrBitmap, true)
		}
	}

	com.Cns.SetShouldCheckConsensus(true)

	return true
}

// SendCommitment method is the function which is actually used to send the commitment for the received block, in the
// Commitment subround (it is used as the handler function of the doSubroundJob pointer variable function in Subround
// struct from spos package)
func (com *Communication) SendCommitment() bool {
	// check if the Bitmap subround is not finished
	if com.Cns.Status(SrBitmap) != SsFinished {
		return com.SendBitmap()
	}

	if com.Cns.Status(SrCommitment) == SsFinished || // check if the Commitment subround is already finished
		com.Cns.GetValidation(com.Cns.SelfId(), SrCommitment) || // check if the commitment has been already sent
		!com.Cns.IsNodeInBitmapGroup(com.Cns.SelfId()) || // check if this node is not in the bitmap received from the leader
		com.Cns.Data == nil { // check if this node has a consensus data on which it should send the commitment
		return false
	}

	dta := NewConsensusData(
		*com.Cns.Data,
		nil,
		[]byte(com.Cns.SelfId()),
		MtCommitment,
		[]byte(com.GetTime()))

	if !com.BroadcastMessage(dta) {
		return false
	}

	com.Log(fmt.Sprintf(com.GetFormatedTime() + "Step 4: Sending commitment"))

	com.Cns.SetValidation(com.Cns.SelfId(), SrCommitment, true)
	com.Cns.SetShouldCheckConsensus(true)

	return true
}

// SendSignature method is the function which is actually used to send the Signature for the received block, in the
// Signature subround (it is used as the handler function of the doSubroundJob pointer variable function in Subround
// struct from spos package)
func (com *Communication) SendSignature() bool {
	// check if the Commitment subround is not finished
	if com.Cns.Status(SrCommitment) != SsFinished {
		return com.SendCommitment()
	}

	if com.Cns.Status(SrSignature) == SsFinished || // check if the Signature subround is already finished
		com.Cns.GetValidation(com.Cns.SelfId(), SrSignature) || // check if the signature has been already sent
		!com.Cns.IsNodeInBitmapGroup(com.Cns.SelfId()) || // check if this node is not in the bitmap received from the leader
		com.Cns.Data == nil { // check if this node has a consensus data on which it should send the signature
		return false
	}

	dta := NewConsensusData(
		*com.Cns.Data,
		nil,
		[]byte(com.Cns.SelfId()),
		MtSignature,
		[]byte(com.GetTime()))

	if !com.BroadcastMessage(dta) {
		return false
	}

	com.Log(fmt.Sprintf(com.GetFormatedTime() + "Step 5: Sending signature"))

	com.Cns.SetValidation(com.Cns.SelfId(), SrSignature, true)
	com.Cns.SetShouldCheckConsensus(true)

	return true
}

// BroadcastMessage method send the message to the nodes which are in the validators group
func (com *Communication) BroadcastMessage(cnsDta *ConsensusData) bool {
	message, err := marshal.DefMarsh.Marshal(cnsDta)

	if err != nil {
		com.Log(fmt.Sprintf(com.GetFormatedTime() + err.Error()))
		return false
	}

	if com.OnSendMessage == nil {
		com.Log(fmt.Sprintf(com.GetFormatedTime() + "OnSendMessage call back function is not set"))
		return false
	}

	go com.OnSendMessage(message)

	return true
}

// ExtendBlock method put this subround in the extended mode and print some messages
func (com *Communication) ExtendBlock() {
	com.Cns.SetStatus(SrBlock, SsExtended)
	com.Log(fmt.Sprintf(com.GetFormatedTime() + "Step 1: Extended the <BLOCK> subround"))
}

// ExtendCommitmentHash method put this subround in the extended mode and print some messages
func (com *Communication) ExtendCommitmentHash() {
	com.Cns.SetStatus(SrCommitmentHash, SsExtended)
	if com.Cns.ComputeSize(SrCommitmentHash) < com.Cns.Threshold(SrCommitmentHash) {
		com.Log(fmt.Sprintf(com.GetFormatedTime()+
			"Step 2: Extended the <COMMITMENT_HASH> subround. Got only %d from %d commitment hashes which are not enough",
			com.Cns.ComputeSize(SrCommitmentHash), len(com.Cns.ConsensusGroup())))
	} else {
		com.Log(fmt.Sprintf(com.GetFormatedTime() +
			"Step 2: Extended the <COMMITMENT_HASH> subround"))
	}
}

// ExtendBitmap method put this subround in the extended mode and print some messages
func (com *Communication) ExtendBitmap() {
	com.Cns.SetStatus(SrBitmap, SsExtended)
	com.Log(fmt.Sprintf(com.GetFormatedTime() + "Step 3: Extended the <BITMAP> subround"))
}

// ExtendCommitment method put this subround in the extended mode and print some messages
func (com *Communication) ExtendCommitment() {
	com.Cns.SetStatus(SrCommitment, SsExtended)
	com.Log(fmt.Sprintf(com.GetFormatedTime()+
		"Step 4: Extended the <COMMITMENT> subround. Got only %d from %d commitments which are not enough",
		com.Cns.ComputeSize(SrCommitment), len(com.Cns.ConsensusGroup())))
}

// ExtendSignature method put this subround in the extended mode and print some messages
func (com *Communication) ExtendSignature() {
	com.Cns.SetStatus(SrSignature, SsExtended)
	com.Log(fmt.Sprintf(com.GetFormatedTime()+
		"Step 5: Extended the <SIGNATURE> subround. Got only %d from %d sigantures which are not enough",
		com.Cns.ComputeSize(SrSignature), len(com.Cns.ConsensusGroup())))
}

// ExtendEndRound method just print some messages as no extend will be permited, because a new round will be start
func (com *Communication) ExtendEndRound() {
	com.Log(fmt.Sprintf("\n" + com.GetFormatedTime() +
		">>>>>>>>>>>>>>>>>>>> THIS ROUND NO BLOCK WAS ADDED TO THE BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n"))
	com.Rounds++ // only for statistic
}

// ReceivedMessage method redirects the received message to the channel which should handle it
func (com *Communication) ReceivedMessage(cnsData *ConsensusData) {
	if ch, ok := com.ChRcvMsg[cnsData.MsgType]; ok {
		ch <- cnsData
	}
}

// CheckChannels method is used to listen to the channels through which node receives and consumes, during the round,
// different messages from the nodes which are in the validators group
func (com *Communication) CheckChannels() {
	for {
		select {
		case rcvDta := <-com.ChRcvMsg[MtBlockBody]:
			if com.ReceivedBlockBody(rcvDta) {
				com.Cns.SetShouldCheckConsensus(true)
			}
		case rcvDta := <-com.ChRcvMsg[MtBlockHeader]:
			if com.ReceivedBlockHeader(rcvDta) {
				com.Cns.SetShouldCheckConsensus(true)
			}
		case rcvDta := <-com.ChRcvMsg[MtCommitmentHash]:
			if com.ReceivedCommitmentHash(rcvDta) {
				com.Cns.SetShouldCheckConsensus(true)
			}
		case rcvDta := <-com.ChRcvMsg[MtBitmap]:
			if com.ReceivedBitmap(rcvDta) {
				com.Cns.SetShouldCheckConsensus(true)
			}
		case rcvDta := <-com.ChRcvMsg[MtCommitment]:
			if com.ReceivedCommitment(rcvDta) {
				com.Cns.SetShouldCheckConsensus(true)
			}
		case rcvDta := <-com.ChRcvMsg[MtSignature]:
			if com.ReceivedSignature(rcvDta) {
				com.Cns.SetShouldCheckConsensus(true)
			}
		}
	}
}

// ReceivedBlockBody method is called when a block body is received through the block body channel.
func (com *Communication) ReceivedBlockBody(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == com.Cns.SelfId() || // check if the block body received is from myself
		!com.Cns.IsNodeLeaderInCurrentRound(node) || // check if the node who sent the block body is not leader in the current round
		com.Blk != nil { // check if the block body was already received
		return false
	}

	com.Log(fmt.Sprintf(com.GetFormatedTime() + "Step 1: Received block body"))

	com.Blk = com.DecodeBlockBody(&cnsDta.Data)

	return true
}

// DecodeBlockBody method decodes block body which is marshalized in the received message
func (com *Communication) DecodeBlockBody(dta *[]byte) *block.Block {
	if dta == nil {
		return nil
	}

	var blk block.Block

	err := marshal.DefMarsh.Unmarshal(&blk, *dta)

	if err != nil {
		com.Log(fmt.Sprintf(com.GetFormatedTime() + err.Error()))
		return nil
	}

	return &blk
}

// ReceivedBlockHeader method is called when a block header is received through the block header channel. If the block
// header is valid, than the agreement map coresponding to the node which sent it, is set on true for the subround Block
func (com *Communication) ReceivedBlockHeader(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == com.Cns.SelfId() || // check if the block header received is from myself
		com.Cns.Status(SrBlock) == SsFinished || // check if the Block subround is already finished
		!com.Cns.IsNodeLeaderInCurrentRound(node) || // check if the node who sent the block header is not leader in the current round
		com.Cns.Validators.GetValidation(node, SrBlock) { // check if the node already sent the block headcer
		return false
	}

	hdr := com.DecodeBlockHeader(&cnsDta.Data)

	if !com.CheckIfBlockIsValid(hdr) {
		com.Log(fmt.Sprintf(com.GetFormatedTime()+
			"Canceled round %d in subround %s", com.Cns.Chr.Round().Index(), com.Cns.GetSubroundName(SrBlock)))
		com.Cns.Chr.SetSelfSubround(-1)
		return false
	}

	com.Log(fmt.Sprintf(com.GetFormatedTime() + "Step 1: Received block header"))

	com.Hdr = hdr
	com.Cns.Data = &com.Hdr.BlockHash

	com.Cns.Validators.SetValidation(node, SrBlock, true)
	return true
}

// DecodeBlockHeader method decodes block header which is marshalized in the received message
func (com *Communication) DecodeBlockHeader(dta *[]byte) *block.Header {
	if dta == nil {
		return nil
	}

	var hdr block.Header

	err := marshal.DefMarsh.Unmarshal(&hdr, *dta)

	if err != nil {
		com.Log(fmt.Sprintf(com.GetFormatedTime() + err.Error()))
		return nil
	}

	return &hdr
}

// ReceivedCommitmentHash method is called when a commitment hash is received through the commitment hash channel.
// If the commitment hash is valid, than the validation map coresponding to the node which sent it, is set on
// true for the subround ComitmentHash
func (com *Communication) ReceivedCommitmentHash(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == com.Cns.SelfId() || // check if the commitment hash received is from myself
		com.Cns.Status(SrCommitmentHash) == SsFinished || // check if the CommitmentHash subround is already finished
		!com.Cns.IsNodeInValidationGroup(node) || // check if the node is not in the validation group
		com.Cns.Validators.GetValidation(node, SrCommitmentHash) || // check if the node already sent the commitment hash
		com.Cns.Data == nil || // check if the consensus data is not set
		!bytes.Equal(cnsDta.Data, *com.Cns.Data) { // check if the data on which the consensus should be done is not the same with the received one
		return false
	}

	// if this node is leader in this round and already he received 2/3 + 1 of commitment hashes he will ignore any
	// others received later
	if com.Cns.IsNodeLeaderInCurrentRound(com.Cns.SelfId()) {
		if com.Cns.IsCommitmentHashReceived(com.Cns.Threshold(SrCommitmentHash)) {
			return false
		}
	}

	com.Cns.Validators.SetValidation(node, SrCommitmentHash, true)
	return true
}

// ReceivedBitmap method is called when a bitmap is received through the bitmap channel. If the bitmap is valid, than
// the validation map coresponding to the node which sent it, is set on true for the subround Bitmap
func (com *Communication) ReceivedBitmap(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == com.Cns.SelfId() || // check if the bitmap received is from myself
		com.Cns.Status(SrBitmap) == SsFinished || // check if the Bitmap subround is already finished
		!com.Cns.IsNodeLeaderInCurrentRound(node) || // check if the node who sent the bitmap is not leader in the current round
		com.Cns.Validators.GetValidation(node, SrBitmap) || // check if the node already sent the bitmap
		com.Cns.Data == nil || // check if the consensus data is not set
		!bytes.Equal(cnsDta.Data, *com.Cns.Data) { // check if the data on which the consensus should be done is not the same with the received one
		return false
	}

	nodes := cnsDta.PubKeys

	if len(nodes) < com.Cns.Threshold(SrBitmap) {
		com.Log(fmt.Sprintf(com.GetFormatedTime()+
			"Canceled round %d in subround %s", com.Cns.Chr.Round().Index(), com.Cns.GetSubroundName(SrBitmap)))
		com.Cns.Chr.SetSelfSubround(-1)
		return false
	}

	for i := 0; i < len(nodes); i++ {
		if !com.Cns.IsNodeInValidationGroup(string(nodes[i])) {
			com.Log(fmt.Sprintf(com.GetFormatedTime()+
				"Canceled round %d in subround %s", com.Cns.Chr.Round().Index(), com.Cns.GetSubroundName(SrBitmap)))
			com.Cns.Chr.SetSelfSubround(-1)
			return false
		}
	}

	for i := 0; i < len(nodes); i++ {
		com.Cns.Validators.SetValidation(string(nodes[i]), SrBitmap, true)
	}

	return true
}

// ReceivedCommitment method is called when a commitment is received through the commitment channel. If the commitment
// is valid, than the validation map coresponding to the node which sent it, is set on true for the subround Comitment
func (com *Communication) ReceivedCommitment(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == com.Cns.SelfId() || // check if the commitment received is from myself
		com.Cns.Status(SrCommitment) == SsFinished || // check if the Commitment subround is already finished
		!com.Cns.IsNodeInBitmapGroup(node) || // check if the node is not in the bitmap group
		com.Cns.Validators.GetValidation(node, SrCommitment) || // check if the node already sent the commitment
		com.Cns.Data == nil || // check if the consensus data is not set
		!bytes.Equal(cnsDta.Data, *com.Cns.Data) { // check if the data on which the consensus should be done is not the same with the received one
		return false
	}

	com.Cns.Validators.SetValidation(node, SrCommitment, true)
	return true
}

// ReceivedSignature method is called when a Signature is received through the Signature channel. If the Signature
// is valid, than the validation map coresponding to the node which sent it, is set on true for the subround Signature
func (com *Communication) ReceivedSignature(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == com.Cns.SelfId() || // check if the signature received is from myself
		com.Cns.Status(SrSignature) == SsFinished || // check if the Signature subround is already finished
		!com.Cns.IsNodeInBitmapGroup(node) || // check if the node is not in the bitmap group
		com.Cns.Validators.GetValidation(node, SrSignature) || // check if the node already sent the signature
		com.Cns.Data == nil || // check if the consensus data is not set
		!bytes.Equal(cnsDta.Data, *com.Cns.Data) { // check if the data on which the consensus should be done is not the same with the received one
		return false
	}

	com.Cns.Validators.SetValidation(node, SrSignature, true)
	return true
}

// CheckIfBlockIsValid method checks if the received block is valid
func (com *Communication) CheckIfBlockIsValid(receivedHeader *block.Header) bool {
	// TODO: This logic is temporary and it should be refactored after the bootstrap mechanism will be implemented

	if com.Blkc.CurrentBlock == nil {
		if receivedHeader.Nonce == 1 { // first block after genesis
			if bytes.Equal(receivedHeader.PrevHash, []byte("")) {
				return true
			}

			com.Log(fmt.Sprintf("Hash not match: local block hash is empty and node received block with previous "+
				"hash %s", receivedHeader.PrevHash))
			return false
		}

		// to resolve the situation when a node comes later in the network and it has the bootstrap mechanism not
		// implemented yet (he will accept the block received)
		com.Log(fmt.Sprintf("Nonce not match: local block nonce is 0 and node received block with nonce %d",
			receivedHeader.Nonce))
		com.Log(fmt.Sprintf("\n"+com.GetFormatedTime()+
			">>>>>>>>>>>>>>>>>>>> ACCEPTED BLOCK WITH NONCE %d BECAUSE BOOSTRAP IS NOT IMPLEMENTED YET <<<<<<<<<<<<<<<<<<<<\n",
			receivedHeader.Nonce))
		return true
	}

	if receivedHeader.Nonce < com.Blkc.CurrentBlock.Nonce+1 {
		com.Log(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d",
			com.Blkc.CurrentBlock.Nonce, receivedHeader.Nonce))
		return false
	}

	if receivedHeader.Nonce == com.Blkc.CurrentBlock.Nonce+1 {
		if bytes.Equal(receivedHeader.PrevHash, com.Blkc.CurrentBlock.BlockHash) {
			return true
		}

		com.Log(fmt.Sprintf("Hash not match: local block hash is %s and node received block with previous hash %s",
			com.Blkc.CurrentBlock.BlockHash, receivedHeader.PrevHash))
		return false
	}

	// to resolve the situation when a node misses some Blocks and it has the bootstrap mechanism not implemented yet
	// (he will accept the block received)
	com.Log(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d",
		com.Blkc.CurrentBlock.Nonce, receivedHeader.Nonce))
	com.Log(fmt.Sprintf("\n"+com.GetFormatedTime()+
		">>>>>>>>>>>>>>>>>>>> ACCEPTED BLOCK WITH NONCE %d BECAUSE BOOSTRAP IS NOT IMPLEMENTED YET <<<<<<<<<<<<<<<<<<<<\n",
		receivedHeader.Nonce))
	return true
}

// ShouldSynch method returns the synch state of the node. If it returns 'true', this means that the node is not
// synchronized yet and it has to continue the bootstrapping mechanism, otherwise the node is already synched and it can
// participate to the consensus, if it is in the validation group of this round
func (com *Communication) ShouldSynch() bool {
	if com.Cns == nil ||
		com.Cns.Chr == nil ||
		com.Cns.Chr.Round() == nil {
		return true
	}

	rnd := com.Cns.Chr.Round()

	if com.Blkc == nil ||
		com.Blkc.CurrentBlock == nil {
		return rnd.Index() > 0
	}

	return com.Blkc.CurrentBlock.Round+1 < uint32(rnd.Index())
}

// GetMessageTypeName method returns the name of the message from a given message ID
func (com *Communication) GetMessageTypeName(messageType MessageType) string {
	switch messageType {
	case MtBlockBody:
		return "<BLOCK_BODY>"
	case MtBlockHeader:
		return "<BLOCK_HEADER>"
	case MtCommitmentHash:
		return "<COMMITMENT_HASH>"
	case MtBitmap:
		return "<BITMAP>"
	case MtCommitment:
		return "<COMMITMENT>"
	case MtSignature:
		return "<SIGNATURE>"
	case MtUnknown:
		return "<UNKNOWN>"
	default:
		return "Undifined message type"
	}
}

// GetFormatedTime method returns a string containing the formated current time
func (com *Communication) GetFormatedTime() string {
	return com.Cns.Chr.SyncTime().FormatedCurrentTime(com.Cns.Chr.ClockOffset())
}

// GetTime method returns a string containing the current time
func (com *Communication) GetTime() string {
	return com.Cns.Chr.SyncTime().CurrentTime(com.Cns.Chr.ClockOffset()).String()
}

// Log method prints info about consensus (if log is true)
func (com *Communication) Log(message string) {
	if com.log {
		fmt.Printf(message + "\n")
	}
}
