package msg

import (
	"fmt"
	"strings"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
)

const sleepTime = 5

// MessageType specifies what kind of message was received
type MessageType int

const (
	mtBlock MessageType = iota
	mtComitmentHash
	mtBitmap
	mtComitment
	mtSignature
	mtUnknown
)

type Message struct {
	p2p  *p2p.Messenger
	cns  *spos.Consensus
	blk  *block.Block
	blkc *blockchain.BlockChain

	ChRcvMsg map[spos.Subround]chan block.Block
}

func NewMessage(p2p *p2p.Messenger, cns *spos.Consensus, blk *block.Block, blkc *blockchain.BlockChain) *Message {
	msg := Message{p2p: p2p, cns: cns, blk: blk, blkc: blkc}

	if msg.p2p != nil {
		(*msg.p2p).SetOnRecvMsg(msg.ReceiveMessage)
	}

	msg.ChRcvMsg = make(map[spos.Subround]chan block.Block)

	if cns == nil || cns.Validators == nil || len(cns.Validators.ConsensusGroup) == 0 {
		msg.ChRcvMsg[spos.SrBlock] = make(chan block.Block)
		msg.ChRcvMsg[spos.SrComitmentHash] = make(chan block.Block)
		msg.ChRcvMsg[spos.SrBitmap] = make(chan block.Block)
		msg.ChRcvMsg[spos.SrComitment] = make(chan block.Block)
		msg.ChRcvMsg[spos.SrSignature] = make(chan block.Block)
	} else {
		msg.ChRcvMsg[spos.SrBlock] = make(chan block.Block, len(cns.Validators.ConsensusGroup))
		msg.ChRcvMsg[spos.SrComitmentHash] = make(chan block.Block, len(cns.Validators.ConsensusGroup))
		msg.ChRcvMsg[spos.SrBitmap] = make(chan block.Block, len(cns.Validators.ConsensusGroup))
		msg.ChRcvMsg[spos.SrComitment] = make(chan block.Block, len(cns.Validators.ConsensusGroup))
		msg.ChRcvMsg[spos.SrSignature] = make(chan block.Block, len(cns.Validators.ConsensusGroup))
	}

	go msg.CheckChannels()

	return &msg
}

func (msg *Message) OnStartRound() {
	msg.blk.ResetBlock()
	msg.cns.ResetRoundStatus()
	msg.cns.ResetValidationMap()
}

func (msg *Message) OnEndRound() {
	msg.blkc.AddBlock(*msg.blk)

	if msg.cns.IsNodeLeaderInCurrentRound(msg.cns.Self) {
		msg.cns.Log(fmt.Sprintf("\n"+msg.cns.Chr.SyncTime().FormatedCurrentTime(msg.cns.Chr.ClockOffset())+">>>>>>>>>>>>>>>>>>>> ADDED PROPOSED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n", msg.blk.Nonce))
	} else {
		msg.cns.Log(fmt.Sprintf("\n"+msg.cns.Chr.SyncTime().FormatedCurrentTime(msg.cns.Chr.ClockOffset())+">>>>>>>>>>>>>>>>>>>> ADDED SYNCHRONIZED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n", msg.blk.Nonce))
	}
}

func (msg *Message) SendBlock() bool {

	if msg.cns.RoundStatus.Block == spos.SsFinished {
		return false
	}

	if msg.cns.ValidationMap[msg.cns.Self].Block {
		return false
	}

	if !msg.cns.IsNodeLeaderInCurrentRound(msg.cns.Self) {
		return false
	}

	currentBlock := msg.blkc.GetCurrentBlock()

	if currentBlock == nil {
		msg.blk = block.NewBlock(0, msg.cns.Chr.SyncTime().FormatedCurrentTime(msg.cns.Chr.ClockOffset()), msg.cns.Self, "", "", msg.GetMessageTypeName(mtBlock))
	} else {
		msg.blk = block.NewBlock(currentBlock.Nonce+1, msg.cns.Chr.SyncTime().FormatedCurrentTime(msg.cns.Chr.ClockOffset()), msg.cns.Self, "", currentBlock.Hash, msg.GetMessageTypeName(mtBlock))
	}

	msg.blk.Hash = msg.blk.CalculateHash()

	if !msg.BroadcastBlock(msg.blk) {
		return false
	}

	msg.cns.Log(fmt.Sprintf(msg.cns.Chr.SyncTime().FormatedCurrentTime(msg.cns.Chr.ClockOffset()) + "Step 1: Sending blk"))

	msg.cns.ValidationMap[msg.cns.Self].Block = true

	return true
}

func (msg *Message) SendComitmentHash() bool {

	if msg.cns.RoundStatus.ComitmentHash == spos.SsFinished {
		return false
	}

	if msg.cns.RoundStatus.Block != spos.SsFinished {
		return msg.SendBlock()
	}

	if msg.cns.ValidationMap[msg.cns.Self].ComitmentHash {
		return false
	}

	blk := *msg.blk

	blk.Signature = msg.cns.Self
	blk.MetaData = msg.GetMessageTypeName(mtComitmentHash)

	if !msg.BroadcastBlock(&blk) {
		return false
	}

	msg.cns.Log(fmt.Sprintf(msg.cns.Chr.SyncTime().FormatedCurrentTime(msg.cns.Chr.ClockOffset()) + "Step 2: Sending comitment hash"))

	msg.cns.ValidationMap[msg.cns.Self].ComitmentHash = true

	return true
}

func (msg *Message) SendBitmap() bool {

	if msg.cns.RoundStatus.Bitmap == spos.SsFinished {
		return false
	}

	if msg.cns.RoundStatus.ComitmentHash != spos.SsFinished {
		return msg.SendComitmentHash()
	}

	if msg.cns.ValidationMap[msg.cns.Self].Bitmap {
		return false
	}

	if !msg.cns.IsNodeLeaderInCurrentRound(msg.cns.Self) {
		return false
	}

	blk := *msg.blk

	blk.Signature = msg.cns.Self
	blk.MetaData = msg.GetMessageTypeName(mtBitmap)

	for i := 0; i < len(msg.cns.ConsensusGroup); i++ {
		if msg.cns.ValidationMap[msg.cns.ConsensusGroup[i]].ComitmentHash {
			blk.MetaData += "," + msg.cns.ConsensusGroup[i]
		}
	}

	if !msg.BroadcastBlock(&blk) {
		return false
	}

	msg.cns.Log(fmt.Sprintf(msg.cns.Chr.SyncTime().FormatedCurrentTime(msg.cns.Chr.ClockOffset()) + "Step 3: Sending bitmap"))

	for i := 0; i < len(msg.cns.ConsensusGroup); i++ {
		if msg.cns.ValidationMap[msg.cns.ConsensusGroup[i]].ComitmentHash {
			msg.cns.ValidationMap[msg.cns.ConsensusGroup[i]].Bitmap = true
		}
	}

	return true
}

func (msg *Message) SendComitment() bool {

	if msg.cns.RoundStatus.Comitment == spos.SsFinished {
		return false
	}

	if msg.cns.RoundStatus.Bitmap != spos.SsFinished {
		return msg.SendBitmap()
	}

	if msg.cns.ValidationMap[msg.cns.Self].Comitment {
		return false
	}

	if !msg.cns.IsNodeInBitmapGroup(msg.cns.Self) {
		return false
	}

	blk := *msg.blk

	blk.Signature = msg.cns.Self
	blk.MetaData = msg.GetMessageTypeName(mtComitment)

	if !msg.BroadcastBlock(&blk) {
		return false
	}

	msg.cns.Log(fmt.Sprintf(msg.cns.Chr.SyncTime().FormatedCurrentTime(msg.cns.Chr.ClockOffset()) + "Step 4: Sending comitment"))

	msg.cns.ValidationMap[msg.cns.Self].Comitment = true

	return true
}

func (msg *Message) SendSignature() bool {

	if msg.cns.RoundStatus.Signature == spos.SsFinished {
		return false
	}

	if msg.cns.RoundStatus.Comitment != spos.SsFinished {
		return msg.SendComitment()
	}

	if msg.cns.ValidationMap[msg.cns.Self].Signature {
		return false
	}

	if !msg.cns.IsNodeInBitmapGroup(msg.cns.Self) {
		return false
	}

	blk := *msg.blk

	blk.Signature = msg.cns.Self
	blk.MetaData = msg.GetMessageTypeName(mtSignature)

	if !msg.BroadcastBlock(&blk) {
		return false
	}

	msg.cns.Log(fmt.Sprintf(msg.cns.Chr.SyncTime().FormatedCurrentTime(msg.cns.Chr.ClockOffset()) + "Step 5: Sending signature"))

	msg.cns.ValidationMap[msg.cns.Self].Signature = true

	return true
}

func (msg *Message) BroadcastBlock(block *block.Block) bool {
	marsh := &mock.MockMarshalizer{}

	message, err := marsh.Marshal(block)

	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return false
	}

	//(*c.P2PNode).BroadcastString(string(message), []string{})
	m := p2p.NewMessage((*msg.p2p).ID().Pretty(), message, marsh)
	(*msg.p2p).BroadcastMessage(m, []string{})

	return true
}

func (msg *Message) ReceiveMessage(sender p2p.Messenger, peerID string, m *p2p.Message) {
	//c.Log(fmt.Sprintf(c.FormatedCurrentTime()+"Peer with ID = %s got a message from peer with ID = %s which traversed %d peers\n", sender.P2pNode.ID().Pretty(), peerID, len(m.Peers)))
	m.AddHop(sender.ID().Pretty())
	//msg.cns.ChRcvMsg <- m.Payload
	msg.ConsumeReceivedMessage(m.Payload)
	//sender.BroadcastMessage(m, m.Peers)
	sender.BroadcastMessage(m, []string{})
}

func (msg *Message) ConsumeReceivedMessage(rcvMsg []byte) {
	msgType, msgData := msg.DecodeMessage(&rcvMsg)

	switch msgType {
	case mtBlock:
		msg.ChRcvMsg[spos.SrBlock] <- msgData.(block.Block)
	case mtComitmentHash:
		msg.ChRcvMsg[spos.SrComitmentHash] <- msgData.(block.Block)
	case mtBitmap:
		msg.ChRcvMsg[spos.SrBitmap] <- msgData.(block.Block)
	case mtComitment:
		msg.ChRcvMsg[spos.SrComitment] <- msgData.(block.Block)
	case mtSignature:
		msg.ChRcvMsg[spos.SrSignature] <- msgData.(block.Block)
	default:
	}
}

func (msg *Message) CheckChannels() {
	for {
		time.Sleep(sleepTime * time.Millisecond)
		select {
		case rcvBlock := <-msg.ChRcvMsg[spos.SrBlock]:
			if msg.ReceivedBlock(&rcvBlock) {
				msg.cns.SetReceivedMessage(true)
			}
		case rcvBlock := <-msg.ChRcvMsg[spos.SrComitmentHash]:
			if msg.ReceivedComitmentHash(&rcvBlock) {
				msg.cns.SetReceivedMessage(true)
			}
		case rcvBlock := <-msg.ChRcvMsg[spos.SrBitmap]:
			if msg.ReceivedBitmap(&rcvBlock) {
				msg.cns.SetReceivedMessage(true)
			}
		case rcvBlock := <-msg.ChRcvMsg[spos.SrComitment]:
			if msg.ReceivedComitment(&rcvBlock) {
				msg.cns.SetReceivedMessage(true)
			}
		case rcvBlock := <-msg.ChRcvMsg[spos.SrSignature]:
			if msg.ReceivedSignature(&rcvBlock) {
				msg.cns.SetReceivedMessage(true)
			}
		default:
		}
	}
}

// ReceivedBlock method is called when a block is received through the block channel. If this block is valid, than the validation map coresponding
// to the node which sent the block, will be set on true for the subround Block
func (msg *Message) ReceivedBlock(rcvBlock *block.Block) bool {
	node := rcvBlock.Signature

	if msg.cns.RoundStatus.Block != spos.SsFinished {
		if msg.cns.IsNodeLeaderInCurrentRound(node) {
			if !msg.blkc.CheckIfBlockIsValid(rcvBlock) {
				msg.cns.Chr.SetSelfSubround(chronology.SrCanceled)
				return false
			}

			msg.cns.Validators.ValidationMap[node].Block = true
			*msg.blk = *rcvBlock
			return true
		}
	}

	return false
}

func (msg *Message) ReceivedComitmentHash(rcvBlock *block.Block) bool {
	node := rcvBlock.Signature

	if msg.cns.RoundStatus.ComitmentHash != spos.SsFinished {
		if msg.cns.IsNodeInValidationGroup(node) {
			if !msg.cns.Validators.ValidationMap[node].ComitmentHash {
				if rcvBlock.Hash == msg.blk.Hash {
					msg.cns.Validators.ValidationMap[node].ComitmentHash = true
					return true
				}
			}
		}
	}

	return false
}

func (msg *Message) ReceivedBitmap(rcvBlock *block.Block) bool {
	node := rcvBlock.Signature

	if msg.cns.RoundStatus.Bitmap != spos.SsFinished {
		if msg.cns.IsNodeLeaderInCurrentRound(node) {
			if rcvBlock.Hash == msg.blk.Hash {
				nodes := strings.Split(rcvBlock.MetaData[len(msg.GetMessageTypeName(mtBitmap))+1:], ",")
				if len(nodes) < msg.cns.Threshold.Bitmap {
					msg.cns.Chr.SetSelfSubround(chronology.SrCanceled)
					return false
				}

				for i := 0; i < len(nodes); i++ {
					if !msg.cns.IsNodeInValidationGroup(nodes[i]) {
						msg.cns.Chr.SetSelfSubround(chronology.SrCanceled)
						return false
					}
				}

				for i := 0; i < len(nodes); i++ {
					msg.cns.Validators.ValidationMap[nodes[i]].Bitmap = true
				}

				return true
			}
		}
	}

	return false
}

func (msg *Message) ReceivedComitment(rcvBlock *block.Block) bool {
	node := rcvBlock.Signature

	if msg.cns.RoundStatus.Comitment != spos.SsFinished {
		if msg.cns.IsNodeInBitmapGroup(node) {
			if !msg.cns.Validators.ValidationMap[node].Comitment {
				if rcvBlock.Hash == msg.blk.Hash {
					msg.cns.Validators.ValidationMap[node].Comitment = true
					return true
				}
			}
		}
	}

	return false
}

func (msg *Message) ReceivedSignature(rcvBlock *block.Block) bool {
	node := rcvBlock.Signature

	if msg.cns.RoundStatus.Signature != spos.SsFinished {
		if msg.cns.IsNodeInBitmapGroup(node) {
			if !msg.cns.Validators.ValidationMap[node].Signature {
				if rcvBlock.Hash == msg.blk.Hash {
					msg.cns.Validators.ValidationMap[node].Signature = true
					return true
				}
			}
		}
	}

	return false
}

func (msg *Message) DecodeMessage(rcvMsg *[]byte) (MessageType, interface{}) {
	if ok, msgBlock := msg.IsBlockInMessage(rcvMsg); ok {
		//c.Log(fmt.Sprintf(c.FormatedCurrentTime()+"Got a message with %s for blk with Signature = %s and Nonce = %d and Hash = %s\n", msgBlock.MetaData, msgBlock.Signature, msgBlock.Nonce, msgBlock.Hash))
		if strings.Contains(msgBlock.MetaData, msg.GetMessageTypeName(mtBlock)) {
			return mtBlock, msgBlock
		}

		if strings.Contains(msgBlock.MetaData, msg.GetMessageTypeName(mtComitmentHash)) {
			return mtComitmentHash, msgBlock
		}

		if strings.Contains(msgBlock.MetaData, msg.GetMessageTypeName(mtBitmap)) {
			return mtBitmap, msgBlock
		}

		if strings.Contains(msgBlock.MetaData, msg.GetMessageTypeName(mtComitment)) {
			return mtComitment, msgBlock
		}

		if strings.Contains(msgBlock.MetaData, msg.GetMessageTypeName(mtSignature)) {
			return mtSignature, msgBlock
		}
	}

	return mtUnknown, nil
}

func (msg *Message) IsBlockInMessage(rcvMsg *[]byte) (bool, *block.Block) {
	var msgBlock block.Block

	json := marshal.JsonMarshalizer{}
	err := json.Unmarshal(&msgBlock, *rcvMsg)

	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return false, nil
	}

	return true, &msgBlock
}

func (msg *Message) GetMessageTypeName(messageType MessageType) string {
	switch messageType {
	case mtBlock:
		return ("<BLOCK>")
	case mtComitmentHash:
		return ("<COMITMENT_HASH>")
	case mtBitmap:
		return ("<BITMAP>")
	case mtComitment:
		return ("<COMITMENT>")
	case mtSignature:
		return ("<SIGNATURE>")
	case mtUnknown:
		return ("<UNKNOWN>")
	default:
		return ("Undifined message type")
	}
}
