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
)

//TODO: Split in multiple structs, with Single Responsibility

const (
	// SrStartRound defines ID of subround "Start round"
	SrStartRound chronology.SubroundId = iota
	// SrBlock defines ID of subround "block"
	SrBlock
	// SrCommitmentHash defines ID of subround "commitment hash"
	SrCommitmentHash
	// SrBitmap defines ID of subround "bitmap"
	SrBitmap
	// SrCommitment defines ID of subround "commitment"
	SrCommitment
	// SrSignature defines ID of subround "signature"
	SrSignature
	// SrEndRound defines ID of subround "End round"
	SrEndRound
)

//TODO: current numbers of shards (this should be injected, and this const should be removed later)
const shardsCount = 1024

//TODO: maximum transactions in one block (this should be injected, and this const should be removed later)
const maxTransactionsInBlock = 1000

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

// SPOSConsensusWorker defines the data needed by spos to comunicate between nodes which are in the validators group
type SPOSConsensusWorker struct {
	log bool

	Cns *Consensus

	Hdr  *block.Header
	Blk  *block.Block
	Blkc *blockchain.BlockChain
	TxP  *transactionPool.TransactionPool

	Rounds          int // only for statistic
	RoundsWithBlock int // only for statistic

	ChRcvMsg map[MessageType]chan *ConsensusData

	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
	// this is a pointer to a function which actually send the message from a node to the network
	OnSendMessage func([]byte)
}

// NewCommunication creates a new SPOSConsensusWorker object
func NewCommunication(
	log bool,
	cns *Consensus,
	blkc *blockchain.BlockChain,
	txp *transactionPool.TransactionPool,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
) *SPOSConsensusWorker {

	com := SPOSConsensusWorker{
		log:         log,
		Cns:         cns,
		Blkc:        blkc,
		TxP:         txp,
		hasher:      hasher,
		marshalizer: marshalizer,
	}

	com.Rounds = 0          // only for statistic
	com.RoundsWithBlock = 0 // only for statistic

	com.ChRcvMsg = make(map[MessageType]chan *ConsensusData)

	nodes := 0

	if cns != nil &&
		cns.RoundConsensus != nil {
		nodes = len(cns.RoundConsensus.ConsensusGroup())
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

// DoStartRoundJob method is the function which actually do the job of the StartRound subround
// (it is used as the handler function of the doSubroundJob pointer variable function in Subround struct,
// from spos package)
func (com *SPOSConsensusWorker) DoStartRoundJob() bool {
	com.Blk = nil
	com.Hdr = nil
	com.Cns.Data = nil
	com.Cns.ResetRoundStatus()
	com.Cns.ResetRoundState()

	leader, err := com.Cns.GetLeader()

	if err != nil {
		leader = "Unknown"
	}

	if leader == com.Cns.SelfId() {
		leader = fmt.Sprintf(leader + " (MY TURN)")
	}

	com.Log(fmt.Sprintf(com.GetFormatedTime()+"Step 0: Preparing for this round with leader %s ",
		leader))

	return true
}

// DoEndRoundJob method is the function which actually do the job of the EndRound subround
// (it is used as the handler function of the doSubroundJob pointer variable function in Subround struct,
// from spos package)
func (com *SPOSConsensusWorker) DoEndRoundJob() bool {
	if !com.Cns.CheckEndRoundConsensus() {
		return false
	}

	header, err := com.marshalizer.Marshal(com.Hdr)

	if err != nil {
		com.Log(fmt.Sprintf(com.GetFormatedTime() + err.Error()))
		return false
	}

	body, err := com.marshalizer.Marshal(com.Blk)

	if err != nil {
		com.Log(fmt.Sprintf(com.GetFormatedTime() + err.Error()))
		return false
	}

	com.Blkc.CurrentBlockHeader = com.Hdr
	com.Blkc.LocalHeight = int64(com.Hdr.Nonce)
	com.Blkc.Put(blockchain.BlockHeaderUnit, com.Hdr.BlockHash, header)
	com.Blkc.Put(blockchain.BlockUnit, com.Hdr.BlockHash, body)

	// TODO: Here the block should be add in the block pool, when its implementation will be finished

	if com.Blk != nil { // remove transactions included in the committed block, from the transaction pool
		for i := 0; i < len(com.Blk.MiniBlocks); i++ {
			for j := 0; j < len(com.Blk.MiniBlocks[i].TxHashes); j++ {
				com.TxP.RemoveTransaction(com.Blk.MiniBlocks[i].TxHashes[j],
					com.Blk.MiniBlocks[i].DestShardID)
			}
		}
	}

	if com.Cns.IsNodeLeaderInCurrentRound(com.Cns.SelfId()) {
		com.Log(fmt.Sprintf("\n"+com.GetFormatedTime()+
			">>>>>>>>>>>>>>>>>>>> ADDED PROPOSED BLOCK WITH NONCE  %d  IN BLOCKCHAIN "+
			"<<<<<<<<<<<<<<<<<<<<\n", com.Hdr.Nonce))
	} else {
		com.Log(fmt.Sprintf("\n"+com.GetFormatedTime()+
			">>>>>>>>>>>>>>>>>>>> ADDED SYNCHRONIZED BLOCK WITH NONCE  %d  IN BLOCKCHAIN "+
			"<<<<<<<<<<<<<<<<<<<<\n", com.Hdr.Nonce))
	}

	com.Rounds++          // only for statistic
	com.RoundsWithBlock++ // only for statistic

	return true
}

// DoBlockJob method actually send the proposed block in the Block subround, when this node is leader
// (it is used as a handler function of the doSubroundJob pointer function declared in Subround struct,
// from spos package)
func (com *SPOSConsensusWorker) DoBlockJob() bool {
	if com.ShouldSynch() { // if node is not synchronized yet, it has to continue the bootstrapping mechanism
		com.Log(fmt.Sprintf(com.GetFormatedTime()+"Canceled round %d in subround %s",
			com.Cns.Chr.Round().Index(), com.Cns.GetSubroundName(SrBlock)))
		com.Cns.Chr.SetSelfSubround(-1)
		return false
	}

	if com.Cns.Status(SrBlock) == SsFinished || // is subround Block already finished?
		com.Cns.GetJobDone(com.Cns.SelfId(), SrBlock) || // has been block already sent?
		!com.Cns.IsNodeLeaderInCurrentRound(com.Cns.SelfId()) { // is another node leader in this round?
		return false
	}

	if !com.SendBlockBody() ||
		!com.SendBlockHeader() {
		return false
	}

	com.Cns.SetJobDone(com.Cns.SelfId(), SrBlock, true)

	return true
}

// SendBlockBody method send the proposed block body in the Block subround
func (com *SPOSConsensusWorker) SendBlockBody() bool {
	blk := &block.Block{}

	blk.MiniBlocks = com.CreateMiniBlocks()

	message, err := com.marshalizer.Marshal(blk)

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

// CreateMiniBlocks method will create mini blocks with transactions from mini pool. It will check if the
// transaction is valid and than it will add it in the specific miniblock, depending of the destination shard id.
// If the transactions count or the time needed for this action will exceed the limits given, it will stop the
// procedure and will return the state of its work until than.
func (com *SPOSConsensusWorker) CreateMiniBlocks() []block.MiniBlock {
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
			// TODO: Here the transaction execution should be called, when its implementation will be
			// TODO: finished, to check if the selected transaction is ok to be added

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
func (com *SPOSConsensusWorker) GetSubround() chronology.SubroundId {
	return com.Cns.Chr.GetSubroundFromDateTime(com.Cns.Chr.SyncTime().CurrentTime(com.Cns.Chr.ClockOffset()))
}

// SendBlockHeader method send the proposed block header in the Block subround
func (com *SPOSConsensusWorker) SendBlockHeader() bool {
	hdr := &block.Header{}

	if com.Blkc.CurrentBlockHeader == nil {
		hdr.Nonce = 1
		hdr.Round = uint32(com.Cns.Chr.Round().Index())
		hdr.TimeStamp = []byte(com.GetTime())
	} else {
		hdr.Nonce = com.Blkc.CurrentBlockHeader.Nonce + 1
		hdr.Round = uint32(com.Cns.Chr.Round().Index())
		hdr.TimeStamp = []byte(com.GetTime())
		hdr.PrevHash = com.Blkc.CurrentBlockHeader.BlockHash
	}

	message, err := com.marshalizer.Marshal(com.Blk)

	if err != nil {
		com.Log(fmt.Sprintf(com.GetFormatedTime() + err.Error()))
		return false
	}

	hdr.BlockHash = com.hasher.Compute(string(message))

	message, err = com.marshalizer.Marshal(hdr)

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

// DoCommitmentHashJob method is the function which is actually used to send the commitment hash for the received
// block from the leader in the CommitmentHash subround (it is used as the handler function of the doSubroundJob
// pointer variable function in Subround struct, from spos package)
func (com *SPOSConsensusWorker) DoCommitmentHashJob() bool {
	if com.Cns.Status(SrBlock) != SsFinished { // is subround Block not finished?
		return com.DoBlockJob()
	}

	if com.Cns.Status(SrCommitmentHash) == SsFinished || // is subround CommitmentHash already finished?
		com.Cns.GetJobDone(com.Cns.SelfId(), SrCommitmentHash) || // is commitment hash already sent?
		com.Cns.Data == nil { // is consensus data not set?
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

	com.Cns.SetJobDone(com.Cns.SelfId(), SrCommitmentHash, true)

	return true
}

// DoBitmapJob method is the function which is actually used to send the bitmap with the commitment hashes
// received, in the Bitmap subround, when this node is leader (it is used as the handler function of the
// doSubroundJob pointer variable function in Subround struct, from spos package)
func (com *SPOSConsensusWorker) DoBitmapJob() bool {
	if com.Cns.Status(SrCommitmentHash) != SsFinished { // is subround CommitmentHash not finished?
		return com.DoCommitmentHashJob()
	}

	if com.Cns.Status(SrBitmap) == SsFinished || // is subround Bitmap already finished?
		com.Cns.GetJobDone(com.Cns.SelfId(), SrBitmap) || // has been bitmap already sent?
		!com.Cns.IsNodeLeaderInCurrentRound(com.Cns.SelfId()) || // is another node leader in this round?
		com.Cns.Data == nil { // is consensus data not set?
		return false
	}

	pks := make([][]byte, 0)

	for i := 0; i < len(com.Cns.ConsensusGroup()); i++ {
		if com.Cns.GetJobDone(com.Cns.ConsensusGroup()[i], SrCommitmentHash) {
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
		if com.Cns.GetJobDone(com.Cns.ConsensusGroup()[i], SrCommitmentHash) {
			com.Cns.SetJobDone(com.Cns.ConsensusGroup()[i], SrBitmap, true)
		}
	}

	return true
}

// DoCommitmentJob method is the function which is actually used to send the commitment for the received block,
// in the Commitment subround (it is used as the handler function of the doSubroundJob pointer variable function
// in Subround struct, from spos package)
func (com *SPOSConsensusWorker) DoCommitmentJob() bool {
	if com.Cns.Status(SrBitmap) != SsFinished { // is subround Bitmap not finished?
		return com.DoBitmapJob()
	}

	if com.Cns.Status(SrCommitment) == SsFinished || // is subround Commitment already finished?
		com.Cns.GetJobDone(com.Cns.SelfId(), SrCommitment) || // has been commitment already sent?
		!com.Cns.IsValidatorInBitmap(com.Cns.SelfId()) || // isn't node in the leader's bitmap?
		com.Cns.Data == nil { // is consensus data not set?
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

	com.Cns.SetJobDone(com.Cns.SelfId(), SrCommitment, true)

	return true
}

// DoSignatureJob method is the function which is actually used to send the Signature for the received block,
// in the Signature subround (it is used as the handler function of the doSubroundJob pointer variable function
// in Subround struct, from spos package)
func (com *SPOSConsensusWorker) DoSignatureJob() bool {
	if com.Cns.Status(SrCommitment) != SsFinished { // is subround Commitment not finished?
		return com.DoCommitmentJob()
	}

	if com.Cns.Status(SrSignature) == SsFinished || // is subround Signature already finished?
		com.Cns.GetJobDone(com.Cns.SelfId(), SrSignature) || // has been signature already sent?
		!com.Cns.IsValidatorInBitmap(com.Cns.SelfId()) || // isn't node in the leader's bitmap?
		com.Cns.Data == nil { // is consensus data not set?
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

	com.Cns.SetJobDone(com.Cns.SelfId(), SrSignature, true)

	return true
}

// BroadcastMessage method send the message to the nodes which are in the validators group
func (com *SPOSConsensusWorker) BroadcastMessage(cnsDta *ConsensusData) bool {
	message, err := com.marshalizer.Marshal(cnsDta)

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
func (com *SPOSConsensusWorker) ExtendBlock() {
	com.Cns.SetStatus(SrBlock, SsExtended)
	com.Log(fmt.Sprintf(com.GetFormatedTime() + "Step 1: Extended the <BLOCK> subround"))
}

// ExtendCommitmentHash method put this subround in the extended mode and print some messages
func (com *SPOSConsensusWorker) ExtendCommitmentHash() {
	com.Cns.SetStatus(SrCommitmentHash, SsExtended)
	if com.Cns.ComputeSize(SrCommitmentHash) < com.Cns.Threshold(SrCommitmentHash) {
		com.Log(fmt.Sprintf(com.GetFormatedTime()+
			"Step 2: Extended the <COMMITMENT_HASH> subround. Got only %d from %d commitment hashes"+
			" which are not enough", com.Cns.ComputeSize(SrCommitmentHash), len(com.Cns.ConsensusGroup())))
	} else {
		com.Log(fmt.Sprintf(com.GetFormatedTime() +
			"Step 2: Extended the <COMMITMENT_HASH> subround"))
	}
}

// ExtendBitmap method put this subround in the extended mode and print some messages
func (com *SPOSConsensusWorker) ExtendBitmap() {
	com.Cns.SetStatus(SrBitmap, SsExtended)
	com.Log(fmt.Sprintf(com.GetFormatedTime() + "Step 3: Extended the <BITMAP> subround"))
}

// ExtendCommitment method put this subround in the extended mode and print some messages
func (com *SPOSConsensusWorker) ExtendCommitment() {
	com.Cns.SetStatus(SrCommitment, SsExtended)
	com.Log(fmt.Sprintf(com.GetFormatedTime()+
		"Step 4: Extended the <COMMITMENT> subround. Got only %d from %d commitments"+
		" which are not enough", com.Cns.ComputeSize(SrCommitment), len(com.Cns.ConsensusGroup())))
}

// ExtendSignature method put this subround in the extended mode and print some messages
func (com *SPOSConsensusWorker) ExtendSignature() {
	com.Cns.SetStatus(SrSignature, SsExtended)
	com.Log(fmt.Sprintf(com.GetFormatedTime()+
		"Step 5: Extended the <SIGNATURE> subround. Got only %d from %d sigantures"+
		" which are not enough", com.Cns.ComputeSize(SrSignature), len(com.Cns.ConsensusGroup())))
}

// ExtendEndRound method just print some messages as no extend will be permited, because a new round
// will be start
func (com *SPOSConsensusWorker) ExtendEndRound() {
	com.Log(fmt.Sprintf("\n" + com.GetFormatedTime() +
		">>>>>>>>>>>>>>>>>>>> THIS ROUND NO BLOCK WAS ADDED TO THE BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n"))
	com.Rounds++ // only for statistic
}

// ReceivedMessage method redirects the received message to the channel which should handle it
func (com *SPOSConsensusWorker) ReceivedMessage(cnsData *ConsensusData) {
	if ch, ok := com.ChRcvMsg[cnsData.MsgType]; ok {
		ch <- cnsData
	}
}

// CheckChannels method is used to listen to the channels through which node receives and consumes,
// during the round, different messages from the nodes which are in the validators group
func (com *SPOSConsensusWorker) CheckChannels() {
	for {
		select {
		case rcvDta := <-com.ChRcvMsg[MtBlockBody]:
			com.ReceivedBlockBody(rcvDta)
		case rcvDta := <-com.ChRcvMsg[MtBlockHeader]:
			com.ReceivedBlockHeader(rcvDta)
		case rcvDta := <-com.ChRcvMsg[MtCommitmentHash]:
			com.ReceivedCommitmentHash(rcvDta)
		case rcvDta := <-com.ChRcvMsg[MtBitmap]:
			com.ReceivedBitmap(rcvDta)
		case rcvDta := <-com.ChRcvMsg[MtCommitment]:
			com.ReceivedCommitment(rcvDta)
		case rcvDta := <-com.ChRcvMsg[MtSignature]:
			com.ReceivedSignature(rcvDta)
		}
	}
}

// ReceivedBlockBody method is called when a block body is received through the block body channel.
func (com *SPOSConsensusWorker) ReceivedBlockBody(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == com.Cns.SelfId() || // is block body received from myself?
		!com.Cns.IsNodeLeaderInCurrentRound(node) || // is another node leader in this round?
		com.Blk != nil { // is block body already received?
		return false
	}

	com.Log(fmt.Sprintf(com.GetFormatedTime() + "Step 1: Received block body"))

	com.Blk = com.DecodeBlockBody(&cnsDta.Data)

	return true
}

// DecodeBlockBody method decodes block body which is marshalized in the received message
func (com *SPOSConsensusWorker) DecodeBlockBody(dta *[]byte) *block.Block {
	if dta == nil {
		return nil
	}

	var blk block.Block

	err := com.marshalizer.Unmarshal(&blk, *dta)

	if err != nil {
		com.Log(fmt.Sprintf(com.GetFormatedTime() + err.Error()))
		return nil
	}

	return &blk
}

// ReceivedBlockHeader method is called when a block header is received through the block header channel.
// If the block header is valid, than the validatorRoundStates map coresponding to the node which sent it,
// is set on true for the subround Block
func (com *SPOSConsensusWorker) ReceivedBlockHeader(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == com.Cns.SelfId() || // is block header received from myself?
		com.Cns.Status(SrBlock) == SsFinished || // is subround Block already finished?
		!com.Cns.IsNodeLeaderInCurrentRound(node) || // is another node leader in this round?
		com.Cns.RoundConsensus.GetJobDone(node, SrBlock) { // is block header already received?
		return false
	}

	hdr := com.DecodeBlockHeader(&cnsDta.Data)

	if !com.CheckIfBlockIsValid(hdr) {
		com.Log(fmt.Sprintf(com.GetFormatedTime()+"Canceled round %d in subround %s",
			com.Cns.Chr.Round().Index(), com.Cns.GetSubroundName(SrBlock)))
		com.Cns.Chr.SetSelfSubround(-1)
		return false
	}

	com.Log(fmt.Sprintf(com.GetFormatedTime() + "Step 1: Received block header"))

	com.Hdr = hdr
	com.Cns.Data = &com.Hdr.BlockHash

	com.Cns.RoundConsensus.SetJobDone(node, SrBlock, true)
	return true
}

// DecodeBlockHeader method decodes block header which is marshalized in the received message
func (com *SPOSConsensusWorker) DecodeBlockHeader(dta *[]byte) *block.Header {
	if dta == nil {
		return nil
	}

	var hdr block.Header

	err := com.marshalizer.Unmarshal(&hdr, *dta)

	if err != nil {
		com.Log(fmt.Sprintf(com.GetFormatedTime() + err.Error()))
		return nil
	}

	return &hdr
}

// ReceivedCommitmentHash method is called when a commitment hash is received through the commitment hash
// channel. If the commitment hash is valid, than the jobDone map coresponding to the node which sent it,
// is set on true for the subround ComitmentHash
func (com *SPOSConsensusWorker) ReceivedCommitmentHash(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == com.Cns.SelfId() || // is commitment hash received from myself?
		com.Cns.Status(SrCommitmentHash) == SsFinished || // is subround CommitmentHash already finished?
		!com.Cns.IsNodeInConsensusGroup(node) || // isn't node in the jobDone group?
		com.Cns.RoundConsensus.GetJobDone(node, SrCommitmentHash) || // is commitment hash already received?
		com.Cns.Data == nil || // is consensus data not set?
		!bytes.Equal(cnsDta.Data, *com.Cns.Data) { // is this the consesnus data of this round?
		return false
	}

	// if this node is leader in this round and already he received 2/3 + 1 of commitment hashes
	// he will ignore any others received later
	if com.Cns.IsNodeLeaderInCurrentRound(com.Cns.SelfId()) {
		if com.Cns.IsCommitmentHashReceived(com.Cns.Threshold(SrCommitmentHash)) {
			return false
		}
	}

	com.Cns.RoundConsensus.SetJobDone(node, SrCommitmentHash, true)
	return true
}

// ReceivedBitmap method is called when a bitmap is received through the bitmap channel.
// If the bitmap is valid, than the jobDone map coresponding to the node which sent it,
// is set on true for the subround Bitmap
func (com *SPOSConsensusWorker) ReceivedBitmap(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == com.Cns.SelfId() || // is bitmap received from myself?
		com.Cns.Status(SrBitmap) == SsFinished || // is subround Bitmap already finished?
		!com.Cns.IsNodeLeaderInCurrentRound(node) || // is another node leader in this round?
		com.Cns.RoundConsensus.GetJobDone(node, SrBitmap) || // is bitmap already received?
		com.Cns.Data == nil || // is consensus data not set?
		!bytes.Equal(cnsDta.Data, *com.Cns.Data) { // is this the consesnus data of this round?
		return false
	}

	nodes := cnsDta.PubKeys

	if len(nodes) < com.Cns.Threshold(SrBitmap) {
		com.Log(fmt.Sprintf(com.GetFormatedTime()+"Canceled round %d in subround %s",
			com.Cns.Chr.Round().Index(), com.Cns.GetSubroundName(SrBitmap)))
		com.Cns.Chr.SetSelfSubround(-1)
		return false
	}

	for i := 0; i < len(nodes); i++ {
		if !com.Cns.IsNodeInConsensusGroup(string(nodes[i])) {
			com.Log(fmt.Sprintf(com.GetFormatedTime()+"Canceled round %d in subround %s",
				com.Cns.Chr.Round().Index(), com.Cns.GetSubroundName(SrBitmap)))
			com.Cns.Chr.SetSelfSubround(-1)
			return false
		}
	}

	for i := 0; i < len(nodes); i++ {
		com.Cns.RoundConsensus.SetJobDone(string(nodes[i]), SrBitmap, true)
	}

	return true
}

// ReceivedCommitment method is called when a commitment is received through the commitment channel.
// If the commitment is valid, than the jobDone map coresponding to the node which sent it,
// is set on true for the subround Comitment
func (com *SPOSConsensusWorker) ReceivedCommitment(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == com.Cns.SelfId() || // is commitment received from myself?
		com.Cns.Status(SrCommitment) == SsFinished || // is subround Commitment already finished?
		!com.Cns.IsValidatorInBitmap(node) || // isn't node in the bitmap group?
		com.Cns.RoundConsensus.GetJobDone(node, SrCommitment) || // is commitment already received?
		com.Cns.Data == nil || // is consensus data not set?
		!bytes.Equal(cnsDta.Data, *com.Cns.Data) { // is this the consesnus data of this round?
		return false
	}

	com.Cns.RoundConsensus.SetJobDone(node, SrCommitment, true)
	return true
}

// ReceivedSignature method is called when a Signature is received through the Signature channel.
// If the Signature is valid, than the jobDone map coresponding to the node which sent it,
// is set on true for the subround Signature
func (com *SPOSConsensusWorker) ReceivedSignature(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == com.Cns.SelfId() || // is signature received from myself?
		com.Cns.Status(SrSignature) == SsFinished || // is subround Signature already finished?
		!com.Cns.IsValidatorInBitmap(node) || // isn't node in the bitmap group?
		com.Cns.RoundConsensus.GetJobDone(node, SrSignature) || // is signature already received?
		com.Cns.Data == nil || // is consensus data not set?
		!bytes.Equal(cnsDta.Data, *com.Cns.Data) { // is this the consesnus data of this round?
		return false
	}

	com.Cns.RoundConsensus.SetJobDone(node, SrSignature, true)
	return true
}

// CheckIfBlockIsValid method checks if the received block is valid
func (com *SPOSConsensusWorker) CheckIfBlockIsValid(receivedHeader *block.Header) bool {
	// TODO: This logic is temporary and it should be refactored after the bootstrap mechanism
	// TODO: will be implemented

	if com.Blkc.CurrentBlockHeader == nil {
		if receivedHeader.Nonce == 1 { // first block after genesis
			if bytes.Equal(receivedHeader.PrevHash, []byte("")) {
				return true
			}

			com.Log(fmt.Sprintf("Hash not match: local block hash is empty and node received block "+
				"with previous hash %s", receivedHeader.PrevHash))
			return false
		}

		// to resolve the situation when a node comes later in the network and it has the
		// bootstrap mechanism not implemented yet (he will accept the block received)
		com.Log(fmt.Sprintf("Nonce not match: local block nonce is 0 and node received block "+
			"with nonce %d", receivedHeader.Nonce))
		com.Log(fmt.Sprintf("\n"+com.GetFormatedTime()+
			">>>>>>>>>>>>>>>>>>>> ACCEPTED BLOCK WITH NONCE %d BECAUSE BOOSTRAP IS NOT "+
			"IMPLEMENTED YET <<<<<<<<<<<<<<<<<<<<\n", receivedHeader.Nonce))
		return true
	}

	if receivedHeader.Nonce < com.Blkc.CurrentBlockHeader.Nonce+1 {
		com.Log(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block "+
			"with nonce %d", com.Blkc.CurrentBlockHeader.Nonce, receivedHeader.Nonce))
		return false
	}

	if receivedHeader.Nonce == com.Blkc.CurrentBlockHeader.Nonce+1 {
		if bytes.Equal(receivedHeader.PrevHash, com.Blkc.CurrentBlockHeader.BlockHash) {
			return true
		}

		com.Log(fmt.Sprintf("Hash not match: local block hash is %s and node received block "+
			"with previous hash %s", com.Blkc.CurrentBlockHeader.BlockHash, receivedHeader.PrevHash))
		return false
	}

	// to resolve the situation when a node misses some Blocks and it has the bootstrap mechanism
	// not implemented yet (he will accept the block received)
	com.Log(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block "+
		"with nonce %d", com.Blkc.CurrentBlockHeader.Nonce, receivedHeader.Nonce))
	com.Log(fmt.Sprintf("\n"+com.GetFormatedTime()+
		">>>>>>>>>>>>>>>>>>>> ACCEPTED BLOCK WITH NONCE %d BECAUSE BOOSTRAP IS NOT "+
		"IMPLEMENTED YET <<<<<<<<<<<<<<<<<<<<\n", receivedHeader.Nonce))
	return true
}

// ShouldSynch method returns the synch state of the node. If it returns 'true', this means that the node
// is not synchronized yet and it has to continue the bootstrapping mechanism, otherwise the node is already
// synched and it can participate to the consensus, if it is in the jobDone group of this round
func (com *SPOSConsensusWorker) ShouldSynch() bool {
	if com.Cns == nil ||
		com.Cns.Chr == nil ||
		com.Cns.Chr.Round() == nil {
		return true
	}

	rnd := com.Cns.Chr.Round()

	if com.Blkc == nil ||
		com.Blkc.CurrentBlockHeader == nil {
		return rnd.Index() > 0
	}

	return com.Blkc.CurrentBlockHeader.Round+1 < uint32(rnd.Index())
}

// GetMessageTypeName method returns the name of the message from a given message ID
func (com *SPOSConsensusWorker) GetMessageTypeName(messageType MessageType) string {
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
func (com *SPOSConsensusWorker) GetFormatedTime() string {
	return com.Cns.Chr.SyncTime().FormatedCurrentTime(com.Cns.Chr.ClockOffset())
}

// GetTime method returns a string containing the current time
func (com *SPOSConsensusWorker) GetTime() string {
	return com.Cns.Chr.SyncTime().CurrentTime(com.Cns.Chr.ClockOffset()).String()
}

// Log method prints info about consensus (if log is true)
func (com *SPOSConsensusWorker) Log(message string) {
	if com.log {
		fmt.Printf(message + "\n")
	}
}

// CheckEndRoundConsensus method checks if the consensus is achieved in each subround from first subround to the given
// subround. If the consensus is achieved in one subround, the subround status is marked as finished
func (cns *Consensus) CheckEndRoundConsensus() bool {
	for i := SrBlock; i <= SrSignature; i++ {
		currentSubRound := cns.Chr.SubroundHandlers()[i]
		if !currentSubRound.Check() {
			return false
		}
	}

	return true
}

// CheckEndRoundConsensus method checks if the consensus is achieved in each subround from first subround to the given
// subround. If the consensus is achieved in one subround, the subround status is marked as finished
func (cns *Consensus) CheckStartRoundConsensus() bool {
	return true
}

// CheckBlockConsensus method checks if the consensus in the <BLOCK> subround is achieved
func (cns *Consensus) CheckBlockConsensus() bool {
	if cns.Status(SrBlock) == SsFinished {
		return true
	}

	if cns.IsBlockReceived(cns.Threshold(SrBlock)) {
		cns.PrintBlockCM() // only for printing block consensus messages
		cns.SetStatus(SrBlock, SsFinished)
		return true
	}

	return false
}

// CheckCommitmentHashConsensus method checks if the consensus in the <COMMITMENT_HASH> subround is achieved
func (cns *Consensus) CheckCommitmentHashConsensus() bool {
	if cns.Status(SrCommitmentHash) == SsFinished {
		return true
	}

	threshold := cns.Threshold(SrCommitmentHash)

	if !cns.IsNodeLeaderInCurrentRound(cns.selfId) {
		threshold = len(cns.consensusGroup)
	}

	if cns.IsCommitmentHashReceived(threshold) {
		cns.PrintCommitmentHashCM() // only for printing commitment hash consensus messages
		cns.SetStatus(SrCommitmentHash, SsFinished)
		return true
	}

	if cns.CommitmentHashesCollected(cns.Threshold(SrBitmap)) {
		cns.PrintCommitmentHashCM() // only for printing commitment hash consensus messages
		cns.SetStatus(SrCommitmentHash, SsFinished)
		return true
	}

	return false
}

// CheckBitmapConsensus method checks if the consensus in the <BITMAP> subround is achieved
func (cns *Consensus) CheckBitmapConsensus() bool {
	if cns.Status(SrBitmap) == SsFinished {
		return true
	}

	if cns.CommitmentHashesCollected(cns.Threshold(SrBitmap)) {
		cns.PrintBitmapCM() // only for printing bitmap consensus messages
		cns.SetStatus(SrBitmap, SsFinished)
		return true
	}

	return false
}

// CheckCommitmentConsensus method checks if the consensus in the <COMMITMENT> subround is achieved
func (cns *Consensus) CheckCommitmentConsensus() bool {
	if cns.Status(SrCommitment) == SsFinished {
		return true
	}

	if cns.CommitmentsCollected(cns.Threshold(SrCommitment)) {
		cns.PrintCommitmentCM() // only for printing commitment consensus messages
		cns.SetStatus(SrCommitment, SsFinished)
		return true
	}

	return false
}

// CheckSignatureConsensus method checks if the consensus in the <SIGNATURE> subround is achieved
func (cns *Consensus) CheckSignatureConsensus() bool {
	if cns.Status(SrSignature) == SsFinished {
		return true
	}

	if cns.SignaturesCollected(cns.Threshold(SrSignature)) {
		cns.PrintSignatureCM() // only for printing signature consensus messages
		cns.SetStatus(SrSignature, SsFinished)
		return true
	}

	return false
}

// GetSubroundName returns the name of each subround from a given subround ID
func (cns *Consensus) GetSubroundName(subroundId chronology.SubroundId) string {
	switch subroundId {
	case SrStartRound:
		return "<START_ROUND>"
	case SrBlock:
		return "<BLOCK>"
	case SrCommitmentHash:
		return "<COMMITMENT_HASH>"
	case SrBitmap:
		return "<BITMAP>"
	case SrCommitment:
		return "<COMMITMENT>"
	case SrSignature:
		return "<SIGNATURE>"
	case SrEndRound:
		return "<END_ROUND>"
	default:
		return "Undifined subround"
	}
}

// PrintBlockCM method prints the <BLOCK> consensus messages
func (cns *Consensus) PrintBlockCM() {
	if !cns.IsNodeLeaderInCurrentRound(cns.selfId) {
		cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) +
			"Step 1: Synchronized block"))
	}
	cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) +
		"Step 1: SubroundId <BLOCK> has been finished"))
}

// PrintCommitmentHashCM method prints the <COMMITMENT_HASH> consensus messages
func (cns *Consensus) PrintCommitmentHashCM() {
	n := cns.ComputeSize(SrCommitmentHash)
	if n == len(cns.consensusGroup) {
		cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset())+
			"Step 2: Received all (%d from %d) commitment hashes", n, len(cns.consensusGroup)))
	} else {
		cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset())+
			"Step 2: Received %d from %d commitment hashes, which are enough", n, len(cns.consensusGroup)))
	}
	cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) +
		"Step 2: SubroundId <COMMITMENT_HASH> has been finished"))
}

// PrintBitmapCM method prints the <BITMAP> consensus messages
func (cns *Consensus) PrintBitmapCM() {
	if !cns.IsNodeLeaderInCurrentRound(cns.selfId) {
		msg := fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset())+
			"Step 3: Received bitmap from leader, matching with my own, and it got %d from %d commitment hashes, which are enough",
			cns.ComputeSize(SrBitmap), len(cns.consensusGroup))

		if cns.IsValidatorInBitmap(cns.selfId) {
			msg = fmt.Sprintf(msg+"%s", ", AND I WAS selected in this bitmap")
		} else {
			msg = fmt.Sprintf(msg+"%s", ", BUT I WAS NOT selected in this bitmap")
		}

		cns.Log(msg)
	}
	cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) +
		"Step 3: SubroundId <BITMAP> has been finished"))
}

// PrintCommitmentCM method prints the <COMMITMENT> consensus messages
func (cns *Consensus) PrintCommitmentCM() {
	cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset())+
		"Step 4: Received %d from %d commitments, which are matching with bitmap and are enough",
		cns.ComputeSize(SrCommitment), len(cns.consensusGroup)))
	cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) +
		"Step 4: SubroundId <COMMITMENT> has been finished"))
}

// PrintSignatureCM method prints the <SIGNATURE> consensus messages
func (cns *Consensus) PrintSignatureCM() {
	cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset())+
		"Step 5: Received %d from %d signatures, which are matching with bitmap and are enough",
		cns.ComputeSize(SrSignature), len(cns.consensusGroup)))
	cns.Log(fmt.Sprintf(cns.Chr.SyncTime().FormatedCurrentTime(cns.Chr.ClockOffset()) +
		"Step 5: SubroundId <SIGNATURE> has been finished"))
}
