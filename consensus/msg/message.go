package msg

import (
	"fmt"
	"strings"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
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

func (cns *Consensus) SendMessage(subround chronology.Subround) bool {
	for {
		switch subround {

		case chronology.Subround(srBlock):
			if cns.RoundStatus.Block == ssFinished {
				return false
			}

			if cns.Validators.ValidationMap[cns.Validators.Self].Block {
				return false
			}

			if !cns.IsNodeLeaderInCurrentRound(cns.Validators.Self) {
				return false
			}

			return cns.SendBlock()
		case chronology.Subround(srComitmentHash):
			if cns.RoundStatus.ComitmentHash == ssFinished {
				return false
			}

			if cns.Validators.ValidationMap[cns.Validators.Self].ComitmentHash {
				return false
			}

			if cns.RoundStatus.Block != ssFinished {
				subround = chronology.Subround(srBlock)
				continue
			}

			return cns.SendComitmentHash()
		case chronology.Subround(srBitmap):
			if cns.RoundStatus.Bitmap == ssFinished {
				return false
			}

			if cns.Validators.ValidationMap[cns.Validators.Self].Bitmap {
				return false
			}

			if !cns.IsNodeLeaderInCurrentRound(cns.Validators.Self) {
				return false
			}

			if cns.RoundStatus.ComitmentHash != ssFinished {
				subround = chronology.Subround(srComitmentHash)
				continue
			}

			return cns.SendBitmap()
		case chronology.Subround(srComitment):
			if cns.RoundStatus.Comitment == ssFinished {
				return false
			}

			if cns.Validators.ValidationMap[cns.Validators.Self].Comitment {
				return false
			}

			if cns.RoundStatus.Bitmap != ssFinished {
				subround = chronology.Subround(srBitmap)
				continue
			}

			if !cns.IsNodeInBitmapGroup(cns.Validators.Self) {
				return false
			}

			return cns.SendComitment()
		case chronology.Subround(srSignature):
			if cns.RoundStatus.Signature == ssFinished {
				return false
			}

			if cns.Validators.ValidationMap[cns.Validators.Self].Signature {
				return false
			}

			if cns.RoundStatus.Comitment != ssFinished {
				subround = chronology.Subround(srComitment)
				continue
			}

			if !cns.IsNodeInBitmapGroup(cns.Validators.Self) {
				return false
			}

			return cns.SendSignature()
		default:
		}

		break
	}

	return false
}

func (cns *Consensus) SendBlock() bool {
	currentBlock := cns.BlockChain.GetCurrentBlock()

	if currentBlock == nil {
		cns.Block = block.NewBlock(0, cns.chr.GetFormatedCurrentTime(), cns.Self, "", "", cns.GetMessageTypeName(mtBlock))
	} else {
		cns.Block = block.NewBlock(currentBlock.Nonce+1, cns.chr.GetFormatedCurrentTime(), cns.Self, "", currentBlock.Hash, cns.GetMessageTypeName(mtBlock))
	}

	cns.Block.Hash = cns.Block.CalculateHash()

	if !cns.BroadcastBlock(cns.Block) {
		return false
	}

	cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 1: Sending block"))

	rsv := cns.ValidationMap[cns.Self]
	rsv.Block = true
	cns.ValidationMap[cns.Self] = rsv

	return true
}

func (cns *Consensus) SendComitmentHash() bool {
	block := *cns.Block

	block.Signature = cns.Self
	block.MetaData = cns.GetMessageTypeName(mtComitmentHash)

	if !cns.BroadcastBlock(&block) {
		return false
	}

	cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 2: Sending comitment hash"))

	rsv := cns.ValidationMap[cns.Self]
	rsv.ComitmentHash = true
	cns.ValidationMap[cns.Self] = rsv

	return true
}

func (cns *Consensus) SendBitmap() bool {
	block := *cns.Block

	block.Signature = cns.Self
	block.MetaData = cns.GetMessageTypeName(mtBitmap)

	for i := 0; i < len(cns.ConsensusGroup); i++ {
		if cns.ValidationMap[cns.ConsensusGroup[i]].ComitmentHash {
			block.MetaData += "," + cns.ConsensusGroup[i]
		}
	}

	if !cns.BroadcastBlock(&block) {
		return false
	}

	cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 3: Sending bitmap"))

	for i := 0; i < len(cns.ConsensusGroup); i++ {
		if cns.ValidationMap[cns.ConsensusGroup[i]].ComitmentHash {
			rsv := cns.ValidationMap[cns.ConsensusGroup[i]]
			rsv.Bitmap = true
			cns.ValidationMap[cns.ConsensusGroup[i]] = rsv
		}
	}

	return true
}

func (cns *Consensus) SendComitment() bool {
	block := *cns.Block

	block.Signature = cns.Self
	block.MetaData = cns.GetMessageTypeName(mtComitment)

	if !cns.BroadcastBlock(&block) {
		return false
	}

	cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 4: Sending comitment"))

	rsv := cns.ValidationMap[cns.Self]
	rsv.Comitment = true
	cns.ValidationMap[cns.Self] = rsv

	return true
}

func (cns *Consensus) SendSignature() bool {
	block := *cns.Block

	block.Signature = cns.Self
	block.MetaData = cns.GetMessageTypeName(mtSignature)

	if !cns.BroadcastBlock(&block) {
		return false
	}

	cns.Log(fmt.Sprintf(cns.chr.GetFormatedCurrentTime() + "Step 5: Sending signature"))

	rsv := cns.ValidationMap[cns.Self]
	rsv.Signature = true
	cns.ValidationMap[cns.Self] = rsv

	return true
}

func (cns *Consensus) BroadcastBlock(block *block.Block) bool {
	marsh := &mock.MockMarshalizer{}

	message, err := marsh.Marshal(block)

	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return false
	}

	//(*c.P2PNode).BroadcastString(string(message), []string{})
	m := p2p.NewMessage((*cns.P2PNode).ID().Pretty(), message, marsh)
	(*cns.P2PNode).BroadcastMessage(m, []string{})

	return true
}

func (cns *Consensus) ReceiveMessage(sender p2p.Messenger, peerID string, m *p2p.Message) {
	//c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Peer with ID = %s got a message from peer with ID = %s which traversed %d peers\n", sender.P2pNode.ID().Pretty(), peerID, len(m.Peers)))
	m.AddHop(sender.ID().Pretty())
	cns.ChRcvMsg <- m.Payload
	//sender.BroadcastMessage(m, m.Peers)
	sender.BroadcastMessage(m, []string{})
}

func (cns *Consensus) ConsumeReceivedMessage(rcvMsg *[]byte, chr *chronology.Chronology) bool {
	msgType, msgData := cns.DecodeMessage(rcvMsg)

	switch msgType {
	case mtBlock:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.Signature

		//if timeRoundState > round.RS_BLOCK || c.SelfRoundState > round.RS_BLOCK {
		//	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Received late " + c.GetMessageTypeName(mtBlock) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		//}

		if cns.RoundStatus.Block != ssFinished {
			if cns.IsNodeLeaderInCurrentRound(node) {
				if !cns.BlockChain.CheckIfBlockIsValid(rcvBlock) {
					chr.SetSelfSubround(chronology.SrCanceled)
					return false
				}

				rsv := cns.Validators.ValidationMap[node]
				rsv.Block = true
				cns.Validators.ValidationMap[node] = rsv
				*cns.Block = *rcvBlock
				return true
			}
		}
	case mtComitmentHash:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.Signature

		//if timeRoundState > round.RS_COMITMENT_HASH || c.SelfRoundState > round.RS_COMITMENT_HASH {
		//	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Received late " + c.GetMessageTypeName(mtComitmentHash) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		//}

		if cns.RoundStatus.ComitmentHash != ssFinished {
			if cns.IsNodeInValidationGroup(node) {
				if !cns.Validators.ValidationMap[node].ComitmentHash {
					if rcvBlock.Hash == cns.Block.Hash {
						rsv := cns.Validators.ValidationMap[node]
						rsv.ComitmentHash = true
						cns.Validators.ValidationMap[node] = rsv
						return true
					}
				}
			}
		}
	case mtBitmap:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.Signature

		//if timeRoundState > round.RS_BITMAP || c.SelfRoundState > round.RS_BITMAP {
		//	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Received late " + c.GetMessageTypeName(mtBitmap) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		//}

		if cns.RoundStatus.Bitmap != ssFinished {
			if cns.IsNodeLeaderInCurrentRound(node) {
				if rcvBlock.Hash == cns.Block.Hash {
					nodes := strings.Split(rcvBlock.MetaData[len(cns.GetMessageTypeName(mtBitmap))+1:], ",")
					if len(nodes) < cns.Threshold.Bitmap {
						chr.SetSelfSubround(chronology.SrCanceled)
						return false
					}

					for i := 0; i < len(nodes); i++ {
						if !cns.IsNodeInValidationGroup(nodes[i]) {
							chr.SetSelfSubround(chronology.SrCanceled)
							return false
						}
					}

					for i := 0; i < len(nodes); i++ {
						rsv := cns.Validators.ValidationMap[nodes[i]]
						rsv.Bitmap = true
						cns.Validators.ValidationMap[nodes[i]] = rsv
					}

					return true
				}
			}
		}
	case mtComitment:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.Signature

		//if timeRoundState > round.RS_COMITMENT || c.SelfRoundState > round.RS_COMITMENT {
		//	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Received late " + c.GetMessageTypeName(mtComitment) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		//}

		if cns.RoundStatus.Comitment != ssFinished {
			if cns.IsNodeInBitmapGroup(node) {
				if !cns.Validators.ValidationMap[node].Comitment {
					if rcvBlock.Hash == cns.Block.Hash {
						rsv := cns.Validators.ValidationMap[node]
						rsv.Comitment = true
						cns.Validators.ValidationMap[node] = rsv
						return true
					}
				}
			}
		}
	case mtSignature:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.Signature

		//if timeRoundState > round.RS_SIGNATURE || c.SelfRoundState > round.RS_SIGNATURE {
		//	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Received late " + c.GetMessageTypeName(mtSignature) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		//}

		if cns.RoundStatus.Signature != ssFinished {
			if cns.IsNodeInBitmapGroup(node) {
				if !cns.Validators.ValidationMap[node].Signature {
					if rcvBlock.Hash == cns.Block.Hash {
						rsv := cns.Validators.ValidationMap[node]
						rsv.Signature = true
						cns.Validators.ValidationMap[node] = rsv
						return true
					}
				}
			}
		}
	default:
	}

	return false
}

func (cns *Consensus) DecodeMessage(rcvMsg *[]byte) (MessageType, interface{}) {
	if ok, msgBlock := cns.IsBlockInMessage(rcvMsg); ok {
		//c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Got a message with %s for block with Signature = %s and Nonce = %d and Hash = %s\n", msgBlock.MetaData, msgBlock.Signature, msgBlock.Nonce, msgBlock.Hash))
		if strings.Contains(msgBlock.MetaData, cns.GetMessageTypeName(mtBlock)) {
			return mtBlock, msgBlock
		}

		if strings.Contains(msgBlock.MetaData, cns.GetMessageTypeName(mtComitmentHash)) {
			return mtComitmentHash, msgBlock
		}

		if strings.Contains(msgBlock.MetaData, cns.GetMessageTypeName(mtBitmap)) {
			return mtBitmap, msgBlock
		}

		if strings.Contains(msgBlock.MetaData, cns.GetMessageTypeName(mtComitment)) {
			return mtComitment, msgBlock
		}

		if strings.Contains(msgBlock.MetaData, cns.GetMessageTypeName(mtSignature)) {
			return mtSignature, msgBlock
		}
	}

	return mtUnknown, nil
}

func (c *Consensus) IsBlockInMessage(rcvMsg *[]byte) (bool, *block.Block) {
	var msgBlock block.Block

	json := marshal.JsonMarshalizer{}
	err := json.Unmarshal(&msgBlock, *rcvMsg)

	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return false, nil
	}

	return true, &msgBlock
}

func (c *Consensus) GetMessageTypeName(messageType MessageType) string {
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
