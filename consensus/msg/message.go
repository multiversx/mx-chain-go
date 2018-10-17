package msg

import (
	"fmt"
	"strings"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
)

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
}

func NewMessage(p2p *p2p.Messenger, cns *spos.Consensus, blk *block.Block, blkc *blockchain.BlockChain) *Message {
	msg := Message{p2p: p2p, cns: cns, blk: blk, blkc: blkc}

	if msg.p2p != nil {
		(*msg.p2p).SetOnRecvMsg(msg.ReceiveMessage)
	}

	return &msg
}

func (msg *Message) SendMessage(subround chronology.Subround) bool {
	for {
		switch subround {

		case chronology.Subround(spos.SrBlock):
			if msg.cns.RoundStatus.Block == spos.SsFinished {
				return false
			}

			if msg.cns.Validators.ValidationMap[msg.cns.Validators.Self].Block {
				return false
			}

			if !msg.cns.IsNodeLeaderInCurrentRound(msg.cns.Validators.Self) {
				return false
			}

			return msg.SendBlock()
		case chronology.Subround(spos.SrComitmentHash):
			if msg.cns.RoundStatus.ComitmentHash == spos.SsFinished {
				return false
			}

			if msg.cns.Validators.ValidationMap[msg.cns.Validators.Self].ComitmentHash {
				return false
			}

			if msg.cns.RoundStatus.Block != spos.SsFinished {
				subround = chronology.Subround(spos.SrBlock)
				continue
			}

			return msg.SendComitmentHash()
		case chronology.Subround(spos.SrBitmap):
			if msg.cns.RoundStatus.Bitmap == spos.SsFinished {
				return false
			}

			if msg.cns.Validators.ValidationMap[msg.cns.Validators.Self].Bitmap {
				return false
			}

			if !msg.cns.IsNodeLeaderInCurrentRound(msg.cns.Validators.Self) {
				return false
			}

			if msg.cns.RoundStatus.ComitmentHash != spos.SsFinished {
				subround = chronology.Subround(spos.SrComitmentHash)
				continue
			}

			return msg.SendBitmap()
		case chronology.Subround(spos.SrComitment):
			if msg.cns.RoundStatus.Comitment == spos.SsFinished {
				return false
			}

			if msg.cns.Validators.ValidationMap[msg.cns.Validators.Self].Comitment {
				return false
			}

			if msg.cns.RoundStatus.Bitmap != spos.SsFinished {
				subround = chronology.Subround(spos.SrBitmap)
				continue
			}

			if !msg.cns.IsNodeInBitmapGroup(msg.cns.Validators.Self) {
				return false
			}

			return msg.SendComitment()
		case chronology.Subround(spos.SrSignature):
			if msg.cns.RoundStatus.Signature == spos.SsFinished {
				return false
			}

			if msg.cns.Validators.ValidationMap[msg.cns.Validators.Self].Signature {
				return false
			}

			if msg.cns.RoundStatus.Comitment != spos.SsFinished {
				subround = chronology.Subround(spos.SrComitment)
				continue
			}

			if !msg.cns.IsNodeInBitmapGroup(msg.cns.Validators.Self) {
				return false
			}

			return msg.SendSignature()
		default:
		}

		break
	}

	return false
}

func (msg *Message) SendBlock() bool {
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
	msg.cns.ChRcvMsg <- m.Payload
	//sender.BroadcastMessage(m, m.Peers)
	sender.BroadcastMessage(m, []string{})
}

func (msg *Message) ConsumeReceivedMessage(rcvMsg *[]byte, chr *chronology.Chronology) bool {
	msgType, msgData := msg.DecodeMessage(rcvMsg)

	switch msgType {
	case mtBlock:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.Signature

		if msg.cns.RoundStatus.Block != spos.SsFinished {
			if msg.cns.IsNodeLeaderInCurrentRound(node) {
				if !msg.blkc.CheckIfBlockIsValid(rcvBlock) {
					chr.SetSelfSubround(chronology.SrCanceled)
					return false
				}

				msg.cns.Validators.ValidationMap[node].Block = true
				*msg.blk = *rcvBlock
				return true
			}
		}
	case mtComitmentHash:
		rcvBlock := msgData.(*block.Block)
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
	case mtBitmap:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.Signature

		if msg.cns.RoundStatus.Bitmap != spos.SsFinished {
			if msg.cns.IsNodeLeaderInCurrentRound(node) {
				if rcvBlock.Hash == msg.blk.Hash {
					nodes := strings.Split(rcvBlock.MetaData[len(msg.GetMessageTypeName(mtBitmap))+1:], ",")
					if len(nodes) < msg.cns.Threshold.Bitmap {
						chr.SetSelfSubround(chronology.SrCanceled)
						return false
					}

					for i := 0; i < len(nodes); i++ {
						if !msg.cns.IsNodeInValidationGroup(nodes[i]) {
							chr.SetSelfSubround(chronology.SrCanceled)
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
	case mtComitment:
		rcvBlock := msgData.(*block.Block)
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
	case mtSignature:
		rcvBlock := msgData.(*block.Block)
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
	default:
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
