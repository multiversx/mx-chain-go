package spos

import (
	"bytes"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

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
	ts []byte) *ConsensusData {

	return &ConsensusData{
		Data:      dta,
		PubKeys:   pks,
		Signature: sig,
		MsgType:   msg,
		TimeStamp: ts}
}

// Message defines the data needed by spos to comunicate between nodes which are in the validators group
type Message struct {
	P2p *p2p.Messenger
	Cns *Consensus

	Hdr  *block.Header
	Blk  *block.Block
	Blkc *blockchain.BlockChain

	Rounds          int // only for statistic
	RoundsWithBlock int // only for statistic

	ChRcvMsg map[MessageType]chan *ConsensusData
}

// NewMessage creates a new Message object
func NewMessage(p2p *p2p.Messenger, cns *Consensus) *Message {
	msg := Message{P2p: p2p, Cns: cns}

	if msg.P2p != nil {
		(*msg.P2p).SetOnRecvMsg(msg.ReceiveMessage)
	}

	msg.Blkc = &blockchain.BlockChain{}

	msg.Rounds = 0          // only for statistic
	msg.RoundsWithBlock = 0 // only for statistic

	msg.ChRcvMsg = make(map[MessageType]chan *ConsensusData)

	if cns == nil || cns.Validators == nil || len(cns.Validators.ConsensusGroup()) == 0 {
		msg.ChRcvMsg[MtBlockBody] = make(chan *ConsensusData)
		msg.ChRcvMsg[MtBlockHeader] = make(chan *ConsensusData)
		msg.ChRcvMsg[MtCommitmentHash] = make(chan *ConsensusData)
		msg.ChRcvMsg[MtBitmap] = make(chan *ConsensusData)
		msg.ChRcvMsg[MtCommitment] = make(chan *ConsensusData)
		msg.ChRcvMsg[MtSignature] = make(chan *ConsensusData)
	} else {
		msg.ChRcvMsg[MtBlockBody] = make(chan *ConsensusData, len(cns.Validators.ConsensusGroup()))
		msg.ChRcvMsg[MtBlockHeader] = make(chan *ConsensusData, len(cns.Validators.ConsensusGroup()))
		msg.ChRcvMsg[MtCommitmentHash] = make(chan *ConsensusData, len(cns.Validators.ConsensusGroup()))
		msg.ChRcvMsg[MtBitmap] = make(chan *ConsensusData, len(cns.Validators.ConsensusGroup()))
		msg.ChRcvMsg[MtCommitment] = make(chan *ConsensusData, len(cns.Validators.ConsensusGroup()))
		msg.ChRcvMsg[MtSignature] = make(chan *ConsensusData, len(cns.Validators.ConsensusGroup()))
	}

	go msg.CheckChannels()

	return &msg
}

// StartRound method is the function which actually do the job of the StartRound subround (it is used as the handler
// function of the doSubroundJob pointer variable function in Subround struct from spos package)
func (msg *Message) StartRound() bool {
	msg.Blk = nil
	msg.Hdr = nil
	msg.Cns.Data = nil
	msg.Cns.ResetRoundStatus()
	msg.Cns.ResetAgreement()

	leader, err := msg.Cns.GetLeader()

	if err != nil {
		leader = "Unknown"
	}

	if leader == msg.Cns.Self() {
		leader += " (MY TURN)"
	}

	msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())+
		"Step 0: Preparing for this round with leader %s ", leader))

	return true
}

// EndRound method is the function which actually do the job of the EndRound subround (it is used as the handler
// function of the doSubroundJob pointer variable function in Subround struct from spos package)
func (msg *Message) EndRound() bool {
	if !msg.Cns.CheckConsensus(SrEndRound) {
		return false
	}

	msg.Blkc.CurrentBlock = msg.Hdr

	if msg.Cns.IsNodeLeaderInCurrentRound(msg.Cns.Self()) {
		msg.Cns.Log(fmt.Sprintf("\n"+msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())+
			">>>>>>>>>>>>>>>>>>>> ADDED PROPOSED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n",
			msg.Hdr.Nonce))
	} else {
		msg.Cns.Log(fmt.Sprintf("\n"+msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())+
			">>>>>>>>>>>>>>>>>>>> ADDED SYNCHRONIZED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n",
			msg.Hdr.Nonce))
	}

	msg.Rounds++          // only for statistic
	msg.RoundsWithBlock++ // only for statistic

	return true
}

// SendBlock method is the function which is actually used to send the proposed block in the Block subround, when this
// node is leader (it is used as the handler function of the doSubroundJob pointer variable function in Subround struct
// from spos package)
func (msg *Message) SendBlock() bool {
	// check if the Block subround is already finished
	if msg.Cns.Status(SrBlock) == SsFinished {
		return false
	}

	// check if the block has been already sent
	if msg.Cns.Agreement(msg.Cns.Self(), SrBlock) {
		return false
	}

	// check if the leader of this round is another node
	if !msg.Cns.IsNodeLeaderInCurrentRound(msg.Cns.Self()) {
		return false
	}

	blk := &block.Block{}

	message, err := marshal.DefMarsh.Marshal(blk)

	if err != nil {
		msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
			err.Error()))
		return false
	}

	dta := NewConsensusData(
		message,
		nil,
		[]byte(msg.Cns.Self()),
		MtBlockBody,
		[]byte(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())))

	if !msg.BroadcastMessage(dta) {
		return false
	}

	msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
		"Step 1: Sending block body"))

	msg.Blk = blk

	currentBlock := msg.Blkc.CurrentBlock
	hdr := &block.Header{}

	if currentBlock == nil {
		hdr.Nonce = 1
		hdr.TimeStamp = []byte(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()))
	} else {
		hdr.Nonce = currentBlock.Nonce + 1
		hdr.TimeStamp = []byte(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()))
		hdr.PrevHash = currentBlock.BlockHash
	}

	message, err = marshal.DefMarsh.Marshal(hdr)

	if err != nil {
		msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
			err.Error()))
		return false
	}

	hdr.BlockHash = hashing.DefHash.Compute(string(message))

	message, err = marshal.DefMarsh.Marshal(hdr)

	if err != nil {
		msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
			err.Error()))
		return false
	}

	dta = NewConsensusData(
		message,
		nil,
		[]byte(msg.Cns.Self()),
		MtBlockHeader,
		[]byte(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())))

	if !msg.BroadcastMessage(dta) {
		return false
	}

	msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
		"Step 1: Sending block header"))

	msg.Hdr = hdr
	msg.Cns.Data = &msg.Hdr.BlockHash
	msg.Cns.SetAgreement(msg.Cns.Self(), SrBlock, true)
	msg.Cns.SetShouldCheckConsensus(true)

	return true
}

// SendCommitmentHash method is the function which is actually used to send the commitment hash for the received block
// from the leader in the CommitmentHash subround (it is used as the handler function of the doSubroundJob pointer
// variable function in Subround struct from spos package)
func (msg *Message) SendCommitmentHash() bool {
	// check if the CommitmentHash subround is already finished
	if msg.Cns.Status(SrCommitmentHash) == SsFinished {
		return false
	}

	// check if the Block subround is not finished
	if msg.Cns.Status(SrBlock) != SsFinished {
		return msg.SendBlock()
	}

	// check if the commitment hash has been already sent
	if msg.Cns.Agreement(msg.Cns.Self(), SrCommitmentHash) {
		return false
	}

	if msg.Cns.Data == nil {
		return false
	}

	dta := NewConsensusData(
		*msg.Cns.Data,
		nil,
		[]byte(msg.Cns.Self()),
		MtCommitmentHash,
		[]byte(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())))

	if !msg.BroadcastMessage(dta) {
		return false
	}

	msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
		"Step 2: Sending commitment hash"))

	msg.Cns.SetAgreement(msg.Cns.Self(), SrCommitmentHash, true)
	msg.Cns.SetShouldCheckConsensus(true)

	return true
}

// SendBitmap method is the function which is actually used to send the bitmap with the commitment hashes received,
// in the Bitmap subround, when this node is leader (it is used as the handler function of the doSubroundJob pointer
// variable function in Subround struct from spos package)
func (msg *Message) SendBitmap() bool {
	// check if the Bitmap subround is already finished
	if msg.Cns.Status(SrBitmap) == SsFinished {
		return false
	}

	// check if the CommitmentHash subround is not finished
	if msg.Cns.Status(SrCommitmentHash) != SsFinished {
		return msg.SendCommitmentHash()
	}

	// check if the bitmap has been already sent
	if msg.Cns.Agreement(msg.Cns.Self(), SrBitmap) {
		return false
	}

	// check if this node is leader in the current round
	if !msg.Cns.IsNodeLeaderInCurrentRound(msg.Cns.Self()) {
		return false
	}

	if msg.Cns.Data == nil {
		return false
	}

	pks := make([][]byte, 0)

	for i := 0; i < len(msg.Cns.ConsensusGroup()); i++ {
		if msg.Cns.Agreement(msg.Cns.ConsensusGroup()[i], SrCommitmentHash) {
			pks = append(pks, []byte(msg.Cns.ConsensusGroup()[i]))
		}
	}

	dta := NewConsensusData(
		*msg.Cns.Data,
		pks,
		[]byte(msg.Cns.Self()),
		MtBitmap,
		[]byte(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())))

	if !msg.BroadcastMessage(dta) {
		return false
	}

	msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
		"Step 3: Sending bitmap"))

	for i := 0; i < len(msg.Cns.ConsensusGroup()); i++ {
		if msg.Cns.Agreement(msg.Cns.ConsensusGroup()[i], SrCommitmentHash) {
			msg.Cns.SetAgreement(msg.Cns.ConsensusGroup()[i], SrBitmap, true)
		}
	}

	msg.Cns.SetShouldCheckConsensus(true)

	return true
}

// SendCommitment method is the function which is actually used to send the commitment for the received block, in the
// Commitment subround (it is used as the handler function of the doSubroundJob pointer variable function in Subround
// struct from spos package)
func (msg *Message) SendCommitment() bool {
	// check if the Commitment subround is already finished
	if msg.Cns.Status(SrCommitment) == SsFinished {
		return false
	}

	// check if the Bitmap subround is not finished
	if msg.Cns.Status(SrBitmap) != SsFinished {
		return msg.SendBitmap()
	}

	// check if the commitment has been already sent
	if msg.Cns.Agreement(msg.Cns.Self(), SrCommitment) {
		return false
	}

	// check if this node is not in the bitmap received from the leader
	if !msg.Cns.IsNodeInBitmapGroup(msg.Cns.Self()) {
		return false
	}

	if msg.Cns.Data == nil {
		return false
	}

	dta := NewConsensusData(
		*msg.Cns.Data,
		nil,
		[]byte(msg.Cns.Self()),
		MtCommitment,
		[]byte(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())))

	if !msg.BroadcastMessage(dta) {
		return false
	}

	msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
		"Step 4: Sending commitment"))

	msg.Cns.SetAgreement(msg.Cns.Self(), SrCommitment, true)
	msg.Cns.SetShouldCheckConsensus(true)

	return true
}

// SendSignature method is the function which is actually used to send the Signature for the received block, in the
// Signature subround (it is used as the handler function of the doSubroundJob pointer variable function in Subround
// struct from spos package)
func (msg *Message) SendSignature() bool {
	// check if the Signature subround is already finished
	if msg.Cns.Status(SrSignature) == SsFinished {
		return false
	}

	// check if the Commitment subround is not finished
	if msg.Cns.Status(SrCommitment) != SsFinished {
		return msg.SendCommitment()
	}

	// check if the Signature has been already sent
	if msg.Cns.Agreement(msg.Cns.Self(), SrSignature) {
		return false
	}

	// check if this node is not in the bitmap received from the leader
	if !msg.Cns.IsNodeInBitmapGroup(msg.Cns.Self()) {
		return false
	}

	if msg.Cns.Data == nil {
		return false
	}

	dta := NewConsensusData(
		*msg.Cns.Data,
		nil,
		[]byte(msg.Cns.Self()),
		MtSignature,
		[]byte(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())))

	if !msg.BroadcastMessage(dta) {
		return false
	}

	msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
		"Step 5: Sending signature"))

	msg.Cns.SetAgreement(msg.Cns.Self(), SrSignature, true)
	msg.Cns.SetShouldCheckConsensus(true)

	return true
}

// BroadcastMessage method send the message to the nodes which are in the validators group
func (msg *Message) BroadcastMessage(cnsDta *ConsensusData) bool {
	message, err := marshal.DefMarsh.Marshal(cnsDta)

	if err != nil {
		msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
			err.Error()))
		return false
	}

	m := p2p.NewMessage((*msg.P2p).ID().Pretty(), message, marshal.DefMarsh)
	go (*msg.P2p).BroadcastMessage(m, []string{})

	return true
}

// ExtendBlock method put this subround in the extended mode and print some messages
func (msg *Message) ExtendBlock() {
	msg.Cns.SetStatus(SrBlock, SsExtended)
	msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
		"Step 1: Extended the <BLOCK> subround"))
}

// ExtendCommitmentHash method put this subround in the extended mode and print some messages
func (msg *Message) ExtendCommitmentHash() {
	msg.Cns.SetStatus(SrCommitmentHash, SsExtended)
	if msg.Cns.ComputeSize(SrCommitmentHash) < msg.Cns.Threshold(SrCommitmentHash) {
		msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())+
			"Step 2: Extended the <COMMITMENT_HASH> subround. Got only %d from %d commitment hashes which are not enough",
			msg.Cns.ComputeSize(SrCommitmentHash), len(msg.Cns.ConsensusGroup())))
	} else {
		msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
			"Step 2: Extended the <COMMITMENT_HASH> subround"))
	}
}

// ExtendBitmap method put this subround in the extended mode and print some messages
func (msg *Message) ExtendBitmap() {
	msg.Cns.SetStatus(SrBitmap, SsExtended)
	msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
		"Step 3: Extended the <BITMAP> subround"))
}

// ExtendCommitment method put this subround in the extended mode and print some messages
func (msg *Message) ExtendCommitment() {
	msg.Cns.SetStatus(SrCommitment, SsExtended)
	msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())+
		"Step 4: Extended the <COMMITMENT> subround. Got only %d from %d commitments which are not enough",
		msg.Cns.ComputeSize(SrCommitment), len(msg.Cns.ConsensusGroup())))
}

// ExtendSignature method put this subround in the extended mode and print some messages
func (msg *Message) ExtendSignature() {
	msg.Cns.SetStatus(SrSignature, SsExtended)
	msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())+
		"Step 5: Extended the <SIGNATURE> subround. Got only %d from %d sigantures which are not enough",
		msg.Cns.ComputeSize(SrSignature), len(msg.Cns.ConsensusGroup())))
}

// ExtendEndRound method just print some messages as no extend will be permited, because a new round will be start
func (msg *Message) ExtendEndRound() {
	msg.Cns.Log(fmt.Sprintf("\n" + msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
		">>>>>>>>>>>>>>>>>>>> THIS ROUND NO BLOCK WAS ADDED TO THE BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n"))
	msg.Rounds++ // only for statistic
}

// ReceiveMessage is the handler function for the received P2p messages
func (msg *Message) ReceiveMessage(sender p2p.Messenger, peerID string, m *p2p.Message) {
	m.AddHop(sender.ID().Pretty())
	msg.ConsumeReceivedMessage(&m.Payload)
	sender.BroadcastMessage(m, []string{})
}

// ConsumeReceivedMessage method redirects the received message to the channel which should handle it
func (msg *Message) ConsumeReceivedMessage(rcvMsg *[]byte) {
	msgType, msgData := msg.DecodeMessage(rcvMsg)

	if ch, ok := msg.ChRcvMsg[msgType]; ok {
		ch <- msgData
	}
}

// DecodeMessage method decodes the received message
func (msg *Message) DecodeMessage(rcvMsg *[]byte) (MessageType, *ConsensusData) {
	if rcvMsg == nil {
		return MtUnknown, nil
	}

	cnsDta := ConsensusData{}

	err := marshal.DefMarsh.Unmarshal(&cnsDta, *rcvMsg)

	if err != nil {
		msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
			err.Error()))
		return MtUnknown, nil
	}

	return cnsDta.MsgType, &cnsDta
}

// DecodeBlockBody method decodes block body which is marshalized in the received message
func (msg *Message) DecodeBlockBody(dta *[]byte) *block.Block {
	if dta == nil {
		return nil
	}

	var blk block.Block

	err := marshal.DefMarsh.Unmarshal(&blk, *dta)

	if err != nil {
		msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
			err.Error()))
		return nil
	}

	return &blk
}

// DecodeBlockHeader method decodes block header which is marshalized in the received message
func (msg *Message) DecodeBlockHeader(dta *[]byte) *block.Header {
	if dta == nil {
		return nil
	}

	var hdr block.Header

	err := marshal.DefMarsh.Unmarshal(&hdr, *dta)

	if err != nil {
		msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
			err.Error()))
		return nil
	}

	return &hdr
}

// CheckChannels method is used to listen to the channels through which node receives and consumes, during the round,
// different messages from the nodes which are in the validators group
func (msg *Message) CheckChannels() {
	for {
		select {
		case rcvDta := <-msg.ChRcvMsg[MtBlockBody]:
			if msg.ReceivedBlockBody(rcvDta) {
				msg.Cns.SetShouldCheckConsensus(true)
			}
		case rcvDta := <-msg.ChRcvMsg[MtBlockHeader]:
			if msg.ReceivedBlockHeader(rcvDta) {
				msg.Cns.SetShouldCheckConsensus(true)
			}
		case rcvDta := <-msg.ChRcvMsg[MtCommitmentHash]:
			if msg.ReceivedCommitmentHash(rcvDta) {
				msg.Cns.SetShouldCheckConsensus(true)
			}
		case rcvDta := <-msg.ChRcvMsg[MtBitmap]:
			if msg.ReceivedBitmap(rcvDta) {
				msg.Cns.SetShouldCheckConsensus(true)
			}
		case rcvDta := <-msg.ChRcvMsg[MtCommitment]:
			if msg.ReceivedCommitment(rcvDta) {
				msg.Cns.SetShouldCheckConsensus(true)
			}
		case rcvDta := <-msg.ChRcvMsg[MtSignature]:
			if msg.ReceivedSignature(rcvDta) {
				msg.Cns.SetShouldCheckConsensus(true)
			}
		}
	}
}

// ReceivedBlockBody method is called when a block body is received through the block body channel.
func (msg *Message) ReceivedBlockBody(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == msg.Cns.Self() ||
		msg.Blk != nil ||
		!msg.Cns.IsNodeLeaderInCurrentRound(node) {
		return false
	}

	msg.Blk = msg.DecodeBlockBody(&cnsDta.Data)
	msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
		"Step 1: Received block body"))

	return true
}

// ReceivedBlockHeader method is called when a block header is received through the block header channel. If the block
// header is valid, than the agreement map coresponding to the node which sent it, is set on true for the subround Block
func (msg *Message) ReceivedBlockHeader(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == msg.Cns.Self() ||
		msg.Cns.Status(SrBlock) == SsFinished ||
		!msg.Cns.IsNodeLeaderInCurrentRound(node) {
		return false
	}

	hdr := msg.DecodeBlockHeader(&cnsDta.Data)

	if !msg.CheckIfBlockIsValid(hdr) {
		msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())+
			"Canceled round %d in subround %s",
			msg.Cns.Chr.Round().Index(), msg.Cns.GetSubroundName(SrBlock)))
		msg.Cns.Chr.SetSelfSubround(-1)
		return false
	}

	msg.Hdr = hdr
	msg.Cns.Data = &msg.Hdr.BlockHash
	msg.Cns.Validators.SetAgreement(node, SrBlock, true)
	msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset()) +
		"Step 1: Received block header"))

	return true
}

// ReceivedCommitmentHash method is called when a commitment hash is received through the commitment hash channel.
// If the commitment hash is valid, than the validation map coresponding to the node which sent it, is set on
// true for the subround ComitmentHash
func (msg *Message) ReceivedCommitmentHash(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == msg.Cns.Self() ||
		msg.Cns.Status(SrCommitmentHash) == SsFinished ||
		!msg.Cns.IsNodeInValidationGroup(node) ||
		msg.Cns.Validators.Agreement(node, SrCommitmentHash) ||
		msg.Cns.Data == nil ||
		!bytes.Equal(cnsDta.Data, *msg.Cns.Data) {
		return false
	}

	// if this node is leader in this round and already he received 2/3 + 1 of commitment hashes he will refuse any
	// others received later
	if msg.Cns.IsNodeLeaderInCurrentRound(msg.Cns.Self()) {
		if msg.Cns.IsCommitmentHashReceived(msg.Cns.Threshold(SrCommitmentHash)) {
			return false
		}
	}

	msg.Cns.Validators.SetAgreement(node, SrCommitmentHash, true)
	return true
}

// ReceivedBitmap method is called when a bitmap is received through the bitmap channel. If the bitmap is valid, than
// the validation map coresponding to the node which sent it, is set on true for the subround Bitmap
func (msg *Message) ReceivedBitmap(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == msg.Cns.Self() ||
		msg.Cns.Status(SrBitmap) == SsFinished ||
		!msg.Cns.IsNodeLeaderInCurrentRound(node) ||
		msg.Cns.Data == nil ||
		!bytes.Equal(cnsDta.Data, *msg.Cns.Data) {
		return false
	}

	nodes := cnsDta.PubKeys

	if len(nodes) < msg.Cns.Threshold(SrBitmap) {
		msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())+
			"Canceled round %d in subround %s",
			msg.Cns.Chr.Round().Index(), msg.Cns.GetSubroundName(SrBitmap)))
		msg.Cns.Chr.SetSelfSubround(-1)
		return false
	}

	for i := 0; i < len(nodes); i++ {
		if !msg.Cns.IsNodeInValidationGroup(string(nodes[i])) {
			msg.Cns.Log(fmt.Sprintf(msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())+
				"Canceled round %d in subround %s",
				msg.Cns.Chr.Round().Index(), msg.Cns.GetSubroundName(SrBitmap)))
			msg.Cns.Chr.SetSelfSubround(-1)
			return false
		}
	}

	for i := 0; i < len(nodes); i++ {
		msg.Cns.Validators.SetAgreement(string(nodes[i]), SrBitmap, true)
	}

	return true
}

// ReceivedCommitment method is called when a commitment is received through the commitment channel. If the commitment
// is valid, than the validation map coresponding to the node which sent it, is set on true for the subround Comitment
func (msg *Message) ReceivedCommitment(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == msg.Cns.Self() ||
		msg.Cns.Status(SrCommitment) == SsFinished ||
		!msg.Cns.IsNodeInBitmapGroup(node) ||
		msg.Cns.Validators.Agreement(node, SrCommitment) ||
		msg.Cns.Data == nil ||
		!bytes.Equal(cnsDta.Data, *msg.Cns.Data) {
		return false
	}

	msg.Cns.Validators.SetAgreement(node, SrCommitment, true)
	return true
}

// ReceivedSignature method is called when a Signature is received through the Signature channel. If the Signature
// is valid, than the validation map coresponding to the node which sent it, is set on true for the subround Signature
func (msg *Message) ReceivedSignature(cnsDta *ConsensusData) bool {
	node := string(cnsDta.Signature)

	if node == msg.Cns.Self() ||
		msg.Cns.Status(SrSignature) == SsFinished ||
		!msg.Cns.IsNodeInBitmapGroup(node) ||
		msg.Cns.Validators.Agreement(node, SrSignature) ||
		msg.Cns.Data == nil ||
		!bytes.Equal(cnsDta.Data, *msg.Cns.Data) {
		return false
	}

	msg.Cns.Validators.SetAgreement(node, SrSignature, true)
	return true
}

// CheckIfBlockIsValid method checks if the received block is valid
func (msg *Message) CheckIfBlockIsValid(receivedHeader *block.Header) bool {
	currentBlock := msg.Blkc.CurrentBlock

	// This logic is temporary and it should be replaced after the bootstrap mechanism will be implemented

	if currentBlock == nil {
		if receivedHeader.Nonce == 1 {
			if bytes.Equal(receivedHeader.PrevHash, []byte("")) {
				return true
			}
			msg.Cns.Log(fmt.Sprintf("Hash not match: local block hash is nil and node received block "+
				"with previous hash %s", receivedHeader.PrevHash))
			return false
		}
		// to resolve the situation when a node comes later in the network and it have not implemented the bootstrap
		// mechanism (he will accept the first block received)
		msg.Cns.Log(fmt.Sprintf("Nonce not match: local block nonce is 0 and node received block "+
			"with nonce %d", receivedHeader.Nonce))
		msg.Cns.Log(fmt.Sprintf("\n"+msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())+
			">>>>>>>>>>>>>>>>>>>> ACCEPTED BLOCK WITH NONCE %d BECAUSE BOOSTRAP IS NOT IMPLEMENTED YET <<<<<<<<<<<<<<<<<<<<\n",
			receivedHeader.Nonce))

		return true
	}

	if receivedHeader.Nonce < currentBlock.Nonce+1 {
		msg.Cns.Log(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d",
			currentBlock.Nonce, receivedHeader.Nonce))
		return false
	}

	if receivedHeader.Nonce == currentBlock.Nonce+1 {
		if bytes.Equal(receivedHeader.PrevHash, currentBlock.BlockHash) {
			return true
		}
		msg.Cns.Log(fmt.Sprintf("Hash not match: local block hash is %s and node received block with "+
			"previous hash %s", currentBlock.BlockHash, receivedHeader.PrevHash))
		return false
	}

	// to resolve the situation when a node misses some Blocks and it have not implemented the bootstrap
	// mechanism (he will accept the next block received)
	msg.Cns.Log(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d",
		currentBlock.Nonce, receivedHeader.Nonce))
	msg.Cns.Log(fmt.Sprintf("\n"+msg.Cns.Chr.SyncTime().FormatedCurrentTime(msg.Cns.Chr.ClockOffset())+
		">>>>>>>>>>>>>>>>>>>> ACCEPTED BLOCK WITH NONCE %d BECAUSE BOOSTRAP IS NOT IMPLEMENTED YET <<<<<<<<<<<<<<<<<<<<\n",
		receivedHeader.Nonce))

	return true
}

// GetMessageTypeName returns the name of the message from a given message ID
func (msg *Message) GetMessageTypeName(messageType MessageType) string {
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
