package chronology

import (
	"errors"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/validators"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/statistic"

	"strings"
	"time"
)

const SLEEP_TIME = 5

type ChronologyIn struct {
	GenesisTime time.Time
	P2PNode     *p2p.Messenger
	Block       *block.Block
	BlockChain  *blockchain.BlockChain
	Validators  *validators.Validators
	Consensus   *consensus.Consensus
	Round       *round.Round
	Statistic   *statistic.Chronology
	SyncTime    *ntp.SyncTime
}

type Chronology struct {
	DoRun      bool
	DoLog      bool
	DoSyncMode bool

	*ChronologyIn

	SelfRoundState round.RoundState
	ChRcvMsg       chan []byte
	ClockOffset    time.Duration
}

func New(chronologyIn *ChronologyIn) *Chronology {
	csi := Chronology{}

	csi.DoRun = true

	csi.ChronologyIn = chronologyIn

	csi.ChRcvMsg = make(chan []byte, len(chronologyIn.Validators.ConsensusGroup))

	(*csi.P2PNode).SetOnRecvMsg(csi.recv)

	csi.InitRound()

	return &csi
}

func (c *Chronology) StartRounds() {
	for c.DoRun {

		time.Sleep(SLEEP_TIME * time.Millisecond)

		_, roundState := c.UpdateRound()

		switch roundState {
		case round.RS_START_ROUND:
			if c.SelfRoundState == roundState {
				if c.DoStartRound() {
					c.SelfRoundState = round.RS_BLOCK
				}
			}
		case round.RS_BLOCK:
			if c.SelfRoundState == roundState {
				if c.DoBlock() {
					c.SelfRoundState = round.RS_COMITMENT_HASH
				}
			}
		case round.RS_COMITMENT_HASH:
			if c.SelfRoundState == roundState {
				if c.DoComitmentHash() {
					c.SelfRoundState = round.RS_BITMAP
				}
			}
		case round.RS_BITMAP:
			if c.SelfRoundState == roundState {
				if c.DoBitmap() {
					c.SelfRoundState = round.RS_COMITMENT
				}
			}
		case round.RS_COMITMENT:
			if c.SelfRoundState == roundState {
				if c.DoComitment() {
					c.SelfRoundState = round.RS_SIGNATURE
				}
			}
		case round.RS_SIGNATURE:
			if c.SelfRoundState == roundState {
				if c.DoSignature() {
					c.SelfRoundState = round.RS_END_ROUND
				}
			}
		case round.RS_END_ROUND:
			if c.SelfRoundState == roundState {
				if c.DoEndRound() {
					c.SelfRoundState = round.RS_BEFORE_ROUND
				}
			}
		default:
		}
	}

	close(c.ChRcvMsg)
}

func (c *Chronology) UpdateRound() (int, round.RoundState) {
	oldRoundIndex := c.Round.Index
	oldRoundState := c.Round.State

	c.Round.UpdateRoundFromDateTime(c.GenesisTime, c.GetCurrentTime())

	if oldRoundIndex != c.Round.Index {
		c.Statistic.AddRound() // only for statistic

		leader, err := c.GetLeader()
		if err != nil {
			leader = "Unknown"
		}

		if leader == c.Validators.Self {
			leader += " (MY TURN)"
		}

		c.Log(fmt.Sprintf("\n"+c.GetFormatedCurrentTime()+"############################## ROUND %d BEGINS WITH LEADER  %s  ##############################\n", c.Round.Index, leader))
		c.InitRound()
	}

	if oldRoundState != c.Round.State {
		c.Log(fmt.Sprintf("\n" + c.GetFormatedCurrentTime() + ".................... SUBROUND " + c.Round.GetRoundStateName(c.Round.State) + " BEGINS ....................\n"))
	}

	roundState := c.SelfRoundState

	if c.DoSyncMode {
		roundState = c.Round.State
	}

	c.OptimizeRoundState(roundState)

	return c.Round.Index, roundState
}

//Check node round state vs. time round state and decide if node state could be changed analyzing his tasks which have been done in current round
func (c *Chronology) OptimizeRoundState(roundState round.RoundState) {
	switch roundState {
	case round.RS_BLOCK:
		if c.SelfRoundState == round.RS_START_ROUND {
			c.SelfRoundState = round.RS_BLOCK
		}
	default:
	}

	if roundState < round.RS_START_ROUND || roundState > round.RS_END_ROUND || roundState != c.SelfRoundState {
		select {
		case rcvMsg := <-c.ChRcvMsg:
			if c.SelfRoundState >= round.RS_START_ROUND && c.SelfRoundState <= round.RS_END_ROUND {
				if c.ConsumeReceivedMessage(&rcvMsg, c.Round.State) {
					//c.Log(fmt.Sprintf("\n" + FormatTime(c.SyncTime.GetCurrentTime())+"Received message in time round state %s and self round state %s", rs.GetRoundStateName(c.Round.GetRoundState()), rs.GetRoundStateName(c.SelfRoundState)))
				}
			}
		default:
		}
	}
}

func (c *Chronology) DoStartRound() bool {
	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)

		c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Step 0: Preparing for this round"))
		return true
	}

	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 0: Aborded round %d in subround %s", c.Round.Index, c.Round.GetRoundStateName(round.RS_START_ROUND)))
	return false
}

func (c *Chronology) DoBlock() bool {
	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)

		bActionDone := c.SendMessage(round.RS_BLOCK)

		if bActionDone {
			bActionDone = false
			if ok, _ := c.CheckConsensus(round.RS_BLOCK, round.RS_BLOCK); ok {
				return true
			}
		}

		timeRoundState := c.Round.GetRoundStateFromDateTime(c.GetCurrentTime())

		if timeRoundState > round.RS_BLOCK {
			c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Step 1: Extended the " + c.Round.GetRoundStateName(round.RS_BLOCK) + " subround"))
			c.Round.Subround.Block = round.SS_EXTENDED
			return true // Try to give a chance to this round if the block from leader will arrive later
		}

		select {
		case rcvMsg := <-c.ChRcvMsg:
			if c.ConsumeReceivedMessage(&rcvMsg, timeRoundState) {
				bActionDone = true
			}
		default:
		}

		if bActionDone {
			bActionDone = false
			if ok, _ := c.CheckConsensus(round.RS_BLOCK, round.RS_BLOCK); ok {
				c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Step 1: Synchronized block"))
				return true
			}
		}
	}

	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 1: Aborded round %d in subround %s", c.Round.Index, c.Round.GetRoundStateName(round.RS_BLOCK)))
	return false
}

func (c *Chronology) DoComitmentHash() bool {
	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)

		bActionDone := c.SendMessage(round.RS_COMITMENT_HASH)

		timeRoundState := c.Round.GetRoundStateFromDateTime(c.GetCurrentTime())

		if timeRoundState > round.RS_COMITMENT_HASH {
			if c.GetComitmentHashes() < c.Consensus.Threshold.ComitmentHash {
				c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 2: Extended the "+c.Round.GetRoundStateName(round.RS_COMITMENT_HASH)+" subround. Got only %d from %d commitment hashes which are not enough", c.GetComitmentHashes(), len(c.Validators.ConsensusGroup)))
			} else {
				c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Step 2: Extended the " + c.Round.GetRoundStateName(round.RS_COMITMENT_HASH) + " subround"))
			}
			c.Round.ComitmentHash = round.SS_EXTENDED
			return true // Try to give a chance to this round if the necesary comitment hashes will arrive later
		}

		select {
		case rcvMsg := <-c.ChRcvMsg:
			if c.ConsumeReceivedMessage(&rcvMsg, timeRoundState) {
				bActionDone = true
			}
		default:
		}

		if bActionDone {
			bActionDone = false
			if ok, n := c.CheckConsensus(round.RS_BLOCK, round.RS_COMITMENT_HASH); ok {
				if n == len(c.Validators.ConsensusGroup) {
					c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 2: Received all (%d from %d) comitment hashes", n, len(c.Validators.ConsensusGroup)))
				} else {
					c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 2: Received %d from %d comitment hashes, which are enough", n, len(c.Validators.ConsensusGroup)))
				}
				return true
			}
		}
	}

	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 2: Aborded round %d in subround %s", c.Round.Index, c.Round.GetRoundStateName(round.RS_COMITMENT_HASH)))
	return false
}

func (c *Chronology) DoBitmap() bool {
	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)

		bActionDone := c.SendMessage(round.RS_BITMAP)

		if bActionDone {
			bActionDone = false
			if ok, _ := c.CheckConsensus(round.RS_BLOCK, round.RS_BITMAP); ok {
				return true
			}
		}

		timeRoundState := c.Round.GetRoundStateFromDateTime(c.GetCurrentTime())

		if timeRoundState > round.RS_BITMAP {
			c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Step 3: Extended the " + c.Round.GetRoundStateName(round.RS_BITMAP) + " subround"))
			c.Round.Bitmap = round.SS_EXTENDED
			return true // Try to give a chance to this round if the bitmap from leader will arrive later
		}

		select {
		case rcvMsg := <-c.ChRcvMsg:
			if c.ConsumeReceivedMessage(&rcvMsg, timeRoundState) {
				bActionDone = true
			}
		default:
		}

		if bActionDone {
			bActionDone = false
			if ok, n := c.CheckConsensus(round.RS_BLOCK, round.RS_BITMAP); ok {
				addMessage := "BUT I WAS NOT selected in this bitmap"
				if c.IsNodeInBitmapGroup(c.Validators.Self) {
					addMessage = "AND I WAS selected in this bitmap"
				}
				c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 3: Received bitmap from leader, matching with my own, and it got %d from %d comitment hashes, which are enough, %s", n, len(c.Validators.ConsensusGroup), addMessage))
				return true
			}
		}
	}

	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 3: Aborded round %d in subround %s", c.Round.Index, c.Round.GetRoundStateName(round.RS_BITMAP)))
	return false
}

func (c *Chronology) DoComitment() bool {
	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)

		bActionDone := c.SendMessage(round.RS_COMITMENT)

		timeRoundState := c.Round.GetRoundStateFromDateTime(c.GetCurrentTime())

		if timeRoundState > round.RS_COMITMENT {
			c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 4: Extended the "+c.Round.GetRoundStateName(round.RS_COMITMENT)+" subround. Got only %d from %d commitments which are not enough", c.GetComitments(), len(c.Validators.ConsensusGroup)))
			c.Round.Comitment = round.SS_EXTENDED
			return true // Try to give a chance to this round if the necesary comitments will arrive later
		}

		select {
		case rcvMsg := <-c.ChRcvMsg:
			if c.ConsumeReceivedMessage(&rcvMsg, timeRoundState) {
				bActionDone = true
			}
		default:
		}

		if bActionDone {
			bActionDone = false
			if ok, n := c.CheckConsensus(round.RS_BLOCK, round.RS_COMITMENT); ok {
				c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 4: Received %d from %d comitments, which are matching with bitmap and are enough", n, len(c.Validators.ConsensusGroup)))
				return true
			}
		}
	}

	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 4: Aborded round %d in subround %s", c.Round.Index, c.Round.GetRoundStateName(round.RS_COMITMENT)))
	return false
}

func (c *Chronology) DoSignature() bool {
	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)

		bActionDone := c.SendMessage(round.RS_SIGNATURE)

		timeRoundState := c.Round.GetRoundStateFromDateTime(c.GetCurrentTime())

		if timeRoundState > round.RS_SIGNATURE {
			c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 5: Extended the "+c.Round.GetRoundStateName(round.RS_SIGNATURE)+" subround. Got only %d from %d sigantures which are not enough", c.GetSignatures(), len(c.Validators.ConsensusGroup)))
			c.Round.Signature = round.SS_EXTENDED
			return true // Try to give a chance to this round if the necesary signatures will arrive later
		}

		select {
		case rcvMsg := <-c.ChRcvMsg:
			if c.ConsumeReceivedMessage(&rcvMsg, timeRoundState) {
				bActionDone = true
			}
		default:
		}

		if bActionDone {
			bActionDone = false
			if ok, n := c.CheckConsensus(round.RS_BLOCK, round.RS_SIGNATURE); ok {
				c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 5: Received %d from %d signatures, which are matching with bitmap and are enough", n, len(c.Validators.ConsensusGroup)))
				return true
			}
		}
	}

	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 5: Aborded round %d in subround %s", c.Round.Index, c.Round.GetRoundStateName(round.RS_SIGNATURE)))
	return false
}

func (c *Chronology) DoEndRound() bool {
	bActionDone := true

	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)

		timeRoundState := c.Round.GetRoundStateFromDateTime(c.GetCurrentTime())

		if timeRoundState > round.RS_END_ROUND {
			c.Log(fmt.Sprintf("\n" + c.GetFormatedCurrentTime() + ">>>>>>>>>>>>>>>>>>>> THIS ROUND NO BLOCK WAS ADDED TO THE BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n"))
			return true
		}

		select {
		case rcvMsg := <-c.ChRcvMsg:
			if c.ConsumeReceivedMessage(&rcvMsg, timeRoundState) {
				bActionDone = true
			}
		default:
		}

		if bActionDone {
			bActionDone = false
			if ok, _ := c.CheckConsensus(round.RS_BLOCK, round.RS_SIGNATURE); ok {
				c.Statistic.AddRoundWithBlock() // only for statistic

				c.BlockChain.AddBlock(*c.Block)

				if c.IsNodeLeaderInCurrentRound(c.Validators.Self) {
					c.Log(fmt.Sprintf("\n"+c.GetFormatedCurrentTime()+">>>>>>>>>>>>>>>>>>>> ADDED PROPOSED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n", c.Block.Nonce))
				} else {
					c.Log(fmt.Sprintf("\n"+c.GetFormatedCurrentTime()+">>>>>>>>>>>>>>>>>>>> ADDED SYNCHRONIZED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n", c.Block.Nonce))
				}

				return true
			}
		}
	}

	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 6: Aborded round %d in subround %s", c.Round.Index, c.Round.GetRoundStateName(round.RS_END_ROUND)))
	return false
}

func (c *Chronology) ConsumeReceivedMessage(rcvMsg *[]byte, timeRoundState round.RoundState) bool {
	msgType, msgData := c.DecodeMessage(rcvMsg)

	switch msgType {
	case MT_BLOCK:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.Signature

		//if timeRoundState > round.RS_BLOCK || c.SelfRoundState > round.RS_BLOCK {
		//	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Received late " + c.GetMessageTypeName(MT_BLOCK) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		//}

		if c.Round.Block != round.SS_FINISHED {
			if c.IsNodeLeaderInCurrentRound(node) {
				if !c.BlockChain.CheckIfBlockIsValid(rcvBlock) {
					c.SelfRoundState = round.RS_ABORDED
					return false
				}

				rsv := c.Validators.ValidationMap[node]
				rsv.Block = true
				c.Validators.ValidationMap[node] = rsv
				*c.Block = *rcvBlock
				return true
			}
		}
	case MT_COMITMENT_HASH:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.Signature

		//if timeRoundState > round.RS_COMITMENT_HASH || c.SelfRoundState > round.RS_COMITMENT_HASH {
		//	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Received late " + c.GetMessageTypeName(MT_COMITMENT_HASH) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		//}

		if c.Round.ComitmentHash != round.SS_FINISHED {
			if c.IsNodeInValidationGroup(node) {
				if !c.Validators.ValidationMap[node].ComitmentHash {
					if rcvBlock.Hash == c.Block.Hash {
						rsv := c.Validators.ValidationMap[node]
						rsv.ComitmentHash = true
						c.Validators.ValidationMap[node] = rsv
						return true
					}
				}
			}
		}
	case MT_BITMAP:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.Signature

		//if timeRoundState > round.RS_BITMAP || c.SelfRoundState > round.RS_BITMAP {
		//	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Received late " + c.GetMessageTypeName(MT_BITMAP) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		//}

		if c.Round.Bitmap != round.SS_FINISHED {
			if c.IsNodeLeaderInCurrentRound(node) {
				if rcvBlock.Hash == c.Block.Hash {
					nodes := strings.Split(rcvBlock.MetaData[len(c.GetMessageTypeName(MT_BITMAP))+1:], ",")
					if len(nodes) < c.Consensus.Threshold.Bitmap {
						c.SelfRoundState = round.RS_ABORDED
						return false
					}

					for i := 0; i < len(nodes); i++ {
						if !c.IsNodeInValidationGroup(nodes[i]) {
							c.SelfRoundState = round.RS_ABORDED
							return false
						}
					}

					for i := 0; i < len(nodes); i++ {
						rsv := c.Validators.ValidationMap[nodes[i]]
						rsv.Bitmap = true
						c.Validators.ValidationMap[nodes[i]] = rsv
					}

					return true
				}
			}
		}
	case MT_COMITMENT:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.Signature

		//if timeRoundState > round.RS_COMITMENT || c.SelfRoundState > round.RS_COMITMENT {
		//	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Received late " + c.GetMessageTypeName(MT_COMITMENT) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		//}

		if c.Round.Comitment != round.SS_FINISHED {
			if c.IsNodeInBitmapGroup(node) {
				if !c.Validators.ValidationMap[node].Comitment {
					if rcvBlock.Hash == c.Block.Hash {
						rsv := c.Validators.ValidationMap[node]
						rsv.Comitment = true
						c.Validators.ValidationMap[node] = rsv
						return true
					}
				}
			}
		}
	case MT_SIGNATURE:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.Signature

		//if timeRoundState > round.RS_SIGNATURE || c.SelfRoundState > round.RS_SIGNATURE {
		//	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Received late " + c.GetMessageTypeName(MT_SIGNATURE) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		//}

		if c.Round.Signature != round.SS_FINISHED {
			if c.IsNodeInBitmapGroup(node) {
				if !c.Validators.ValidationMap[node].Signature {
					if rcvBlock.Hash == c.Block.Hash {
						rsv := c.Validators.ValidationMap[node]
						rsv.Signature = true
						c.Validators.ValidationMap[node] = rsv
						return true
					}
				}
			}
		}
	default:
	}

	return false
}

func (c *Chronology) CheckConsensus(startRoundState round.RoundState, endRoundState round.RoundState) (bool, int) {
	var n int
	var ok bool

	for i := startRoundState; i <= endRoundState; i++ {
		switch i {
		case round.RS_BLOCK:
			if c.Round.Block != round.SS_FINISHED {
				if ok, n = c.IsBlock(c.Consensus.Threshold.Block); ok {
					c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 1: Subround %s has been finished", c.Round.GetRoundStateName(round.RS_BLOCK)))
					c.Round.Block = round.SS_FINISHED
				} else {
					return false, n
				}
			}
		case round.RS_COMITMENT_HASH:
			if c.Round.ComitmentHash != round.SS_FINISHED {
				threshold := c.Consensus.Threshold.ComitmentHash
				if !c.IsNodeLeaderInCurrentRound(c.Validators.Self) {
					threshold = len(c.Validators.ConsensusGroup)
				}
				if ok, n = c.IsComitmentHash(threshold); ok {
					c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 2: Subround %s has been finished", c.Round.GetRoundStateName(round.RS_COMITMENT_HASH)))
					c.Round.ComitmentHash = round.SS_FINISHED
				} else if ok, n = c.IsComitmentHashInBitmap(c.Consensus.Threshold.Bitmap); ok {
					c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 2: Subround %s has been finished", c.Round.GetRoundStateName(round.RS_COMITMENT_HASH)))
					c.Round.ComitmentHash = round.SS_FINISHED
				} else {
					return false, n
				}
			}
		case round.RS_BITMAP:
			if c.Round.Bitmap != round.SS_FINISHED {
				if ok, n = c.IsComitmentHashInBitmap(c.Consensus.Threshold.Bitmap); ok {
					c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 3: Subround %s has been finished", c.Round.GetRoundStateName(round.RS_BITMAP)))
					c.Round.Bitmap = round.SS_FINISHED
				} else {
					return false, n
				}
			}
		case round.RS_COMITMENT:
			if c.Round.Comitment != round.SS_FINISHED {
				if ok, n = c.IsBitmapInComitment(c.Consensus.Threshold.Comitment); ok {
					c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 4: Subround %s has been finished", c.Round.GetRoundStateName(round.RS_COMITMENT)))
					c.Round.Comitment = round.SS_FINISHED
				} else {
					return false, n
				}
			}
		case round.RS_SIGNATURE:
			if c.Round.Signature != round.SS_FINISHED {
				if ok, n = c.IsComitmentInSignature(c.Consensus.Threshold.Signature); ok {
					c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Step 5: Subround %s has been finished", c.Round.GetRoundStateName(round.RS_SIGNATURE)))
					c.Round.Signature = round.SS_FINISHED
				} else {
					return false, n
				}
			}
		default:
			return false, -1
		}
	}

	return true, n
}

func (c *Chronology) SendMessage(roundState round.RoundState) bool {
	for {
		switch roundState {

		case round.RS_BLOCK:
			if c.Round.Block == round.SS_FINISHED {
				return false
			}

			if c.Validators.ValidationMap[c.Validators.Self].Block {
				return false
			}

			if !c.IsNodeLeaderInCurrentRound(c.Validators.Self) {
				return false
			}

			return c.SendBlock()
		case round.RS_COMITMENT_HASH:
			if c.Round.ComitmentHash == round.SS_FINISHED {
				return false
			}

			if c.Validators.ValidationMap[c.Validators.Self].ComitmentHash {
				return false
			}

			if c.Round.Block != round.SS_FINISHED {
				roundState = round.RS_BLOCK
				continue
			}

			return c.SendComitmentHash()
		case round.RS_BITMAP:
			if c.Round.Bitmap == round.SS_FINISHED {
				return false
			}

			if c.Validators.ValidationMap[c.Validators.Self].Bitmap {
				return false
			}

			if !c.IsNodeLeaderInCurrentRound(c.Validators.Self) {
				return false
			}

			if c.Round.ComitmentHash != round.SS_FINISHED {
				roundState = round.RS_COMITMENT_HASH
				continue
			}

			return c.SendBitmap()
		case round.RS_COMITMENT:
			if c.Round.Comitment == round.SS_FINISHED {
				return false
			}

			if c.Validators.ValidationMap[c.Validators.Self].Comitment {
				return false
			}

			if c.Round.Bitmap != round.SS_FINISHED {
				roundState = round.RS_BITMAP
				continue
			}

			if !c.IsNodeInBitmapGroup(c.Validators.Self) {
				return false
			}

			return c.SendComitment()
		case round.RS_SIGNATURE:
			if c.Round.Signature == round.SS_FINISHED {
				return false
			}

			if c.Validators.ValidationMap[c.Validators.Self].Signature {
				return false
			}

			if c.Round.Comitment != round.SS_FINISHED {
				roundState = round.RS_COMITMENT
				continue
			}

			if !c.IsNodeInBitmapGroup(c.Validators.Self) {
				return false
			}

			return c.SendSignature()
		default:
		}

		break
	}

	return false
}

func (c *Chronology) SendBlock() bool {
	currentBlock := c.BlockChain.GetCurrentBlock()

	if currentBlock == nil {
		*c.Block = block.New(0, c.GetCurrentTime().String(), c.Validators.Self, "", "", c.GetMessageTypeName(MT_BLOCK))
	} else {
		*c.Block = block.New(currentBlock.Nonce+1, c.GetCurrentTime().String(), c.Validators.Self, "", currentBlock.Hash, c.GetMessageTypeName(MT_BLOCK))
	}

	c.Block.Hash = c.Block.CalculateHash()

	if !c.BroadcastBlock(c.Block) {
		return false
	}

	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Step 1: Sending block"))

	rsv := c.Validators.ValidationMap[c.Validators.Self]
	rsv.Block = true
	c.Validators.ValidationMap[c.Validators.Self] = rsv

	return true
}

func (c *Chronology) SendComitmentHash() bool {
	block := *c.Block

	block.Signature = c.Validators.Self
	block.MetaData = c.GetMessageTypeName(MT_COMITMENT_HASH)

	if !c.BroadcastBlock(&block) {
		return false
	}

	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Step 2: Sending comitment hash"))

	rsv := c.Validators.ValidationMap[c.Validators.Self]
	rsv.ComitmentHash = true
	c.Validators.ValidationMap[c.Validators.Self] = rsv

	return true
}

func (c *Chronology) SendBitmap() bool {
	block := *c.Block

	block.Signature = c.Validators.Self
	block.MetaData = c.GetMessageTypeName(MT_BITMAP)

	for i := 0; i < len(c.Validators.ConsensusGroup); i++ {
		if c.Validators.ValidationMap[c.Validators.ConsensusGroup[i]].ComitmentHash {
			block.MetaData += "," + c.Validators.ConsensusGroup[i]
		}
	}

	if !c.BroadcastBlock(&block) {
		return false
	}

	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Step 3: Sending bitmap"))

	for i := 0; i < len(c.Validators.ConsensusGroup); i++ {
		if c.Validators.ValidationMap[c.Validators.ConsensusGroup[i]].ComitmentHash {
			rsv := c.Validators.ValidationMap[c.Validators.ConsensusGroup[i]]
			rsv.Bitmap = true
			c.Validators.ValidationMap[c.Validators.ConsensusGroup[i]] = rsv
		}
	}

	return true
}

func (c *Chronology) SendComitment() bool {
	block := *c.Block

	block.Signature = c.Validators.Self
	block.MetaData = c.GetMessageTypeName(MT_COMITMENT)

	if !c.BroadcastBlock(&block) {
		return false
	}

	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Step 4: Sending comitment"))

	rsv := c.Validators.ValidationMap[c.Validators.Self]
	rsv.Comitment = true
	c.Validators.ValidationMap[c.Validators.Self] = rsv

	return true
}

func (c *Chronology) SendSignature() bool {
	block := *c.Block

	block.Signature = c.Validators.Self
	block.MetaData = c.GetMessageTypeName(MT_SIGNATURE)

	if !c.BroadcastBlock(&block) {
		return false
	}

	c.Log(fmt.Sprintf(c.GetFormatedCurrentTime() + "Step 5: Sending signature"))

	rsv := c.Validators.ValidationMap[c.Validators.Self]
	rsv.Signature = true
	c.Validators.ValidationMap[c.Validators.Self] = rsv

	return true
}

func (c *Chronology) BroadcastBlock(block *block.Block) bool {
	marsh := &mock.MockMarshalizer{}

	message, err := marsh.Marshal(block)

	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return false
	}

	//(*c.P2PNode).BroadcastString(string(message), []string{})
	m := p2p.NewMessage((*c.P2PNode).ID().Pretty(), message, marsh)
	(*c.P2PNode).BroadcastMessage(m, []string{})

	return true
}

func (c *Chronology) ComputeLeader(nodes []string, round *round.Round) (string, error) {
	if round == nil {
		return "", errors.New("Round is null")
	}

	if nodes == nil {
		return "", errors.New("List of nodes is null")
	}

	if len(nodes) == 0 {
		return "", errors.New("List of nodes is empty")
	}

	index := round.Index % len(nodes)
	return nodes[index], nil
}

func (c *Chronology) IsNodeLeader(node string, nodes []string, round *round.Round) (bool, error) {
	v, err := c.ComputeLeader(nodes, round)

	if err != nil {
		fmt.Println(err)
		return false, err
	}

	return v == node, nil
}

func (c *Chronology) IsNodeLeaderInCurrentRound(node string) bool {
	leader, err := c.GetLeader()

	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return false
	}

	return leader == node
}

func (c *Chronology) IsNodeInValidationGroup(node string) bool {
	for i := 0; i < len(c.Validators.ConsensusGroup); i++ {
		if c.Validators.ConsensusGroup[i] == node {
			return true
		}
	}

	return false
}

func (c *Chronology) IsNodeInBitmapGroup(node string) bool {
	return c.Validators.ValidationMap[node].Bitmap
}

func (c *Chronology) IsBlock(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(c.Validators.ConsensusGroup); i++ {
		if c.Validators.ValidationMap[c.Validators.ConsensusGroup[i]].Block {
			n++
		}
	}

	return n >= threshold, n
}

func (c *Chronology) IsComitmentHash(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(c.Validators.ConsensusGroup); i++ {
		if c.Validators.ValidationMap[c.Validators.ConsensusGroup[i]].ComitmentHash {
			n++
		}
	}

	return n >= threshold, n
}

func (c *Chronology) IsComitmentHashInBitmap(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(c.Validators.ConsensusGroup); i++ {
		if c.Validators.ValidationMap[c.Validators.ConsensusGroup[i]].Bitmap {
			if !c.Validators.ValidationMap[c.Validators.ConsensusGroup[i]].ComitmentHash {
				return false, n
			}
			n++
		}
	}

	return n >= threshold, n
}

func (c *Chronology) IsBitmapInComitment(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(c.Validators.ConsensusGroup); i++ {
		if c.Validators.ValidationMap[c.Validators.ConsensusGroup[i]].Bitmap {
			if !c.Validators.ValidationMap[c.Validators.ConsensusGroup[i]].Comitment {
				return false, n
			}
			n++
		}
	}

	return n >= threshold, n
}

func (c *Chronology) IsComitmentInSignature(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(c.Validators.ConsensusGroup); i++ {
		if c.Validators.ValidationMap[c.Validators.ConsensusGroup[i]].Comitment {
			if !c.Validators.ValidationMap[c.Validators.ConsensusGroup[i]].Signature {
				return false, n
			}
			n++
		}
	}

	return n >= threshold, n
}

func (c *Chronology) GetComitmentHashes() int {
	n := 0

	for i := 0; i < len(c.Validators.ConsensusGroup); i++ {
		if c.Validators.ValidationMap[c.Validators.ConsensusGroup[i]].ComitmentHash {
			n++
		}
	}

	return n
}

func (c *Chronology) GetComitments() int {
	n := 0

	for i := 0; i < len(c.Validators.ConsensusGroup); i++ {
		if c.Validators.ValidationMap[c.Validators.ConsensusGroup[i]].Comitment {
			n++
		}
	}

	return n
}

func (c *Chronology) GetSignatures() int {
	n := 0

	for i := 0; i < len(c.Validators.ConsensusGroup); i++ {
		if c.Validators.ValidationMap[c.Validators.ConsensusGroup[i]].Signature {
			n++
		}
	}

	return n
}

func (c *Chronology) GetLeader() (string, error) {
	if c.Round.Index == -1 {
		return "", errors.New("Round is not set")
	}

	if c.Validators.ConsensusGroup == nil {
		return "", errors.New("List of Validators.ConsensusGroup is null")
	}

	if len(c.Validators.ConsensusGroup) == 0 {
		return "", errors.New("List of nodes is empty")
	}

	index := c.Round.Index % len(c.Validators.ConsensusGroup)
	return c.Validators.ConsensusGroup[index], nil
}

func (c *Chronology) InitRound() {
	c.ClockOffset = c.SyncTime.GetClockOffset()
	c.SelfRoundState = round.RS_START_ROUND
	c.Validators.ResetValidationMap()
	c.Block.ResetBlock()
	c.Round.ResetSubround()
}

// A MessageType specifies what kind of message was received
type MessageType int

const (
	MT_BLOCK MessageType = iota
	MT_COMITMENT_HASH
	MT_BITMAP
	MT_COMITMENT
	MT_SIGNATURE
	MT_UNKNOWN
)

func (c *Chronology) GetMessageTypeName(messageType MessageType) string {
	switch messageType {
	case MT_BLOCK:
		return ("<BLOCK>")
	case MT_COMITMENT_HASH:
		return ("<COMITMENT_HASH>")
	case MT_BITMAP:
		return ("<BITMAP>")
	case MT_COMITMENT:
		return ("<COMITMENT>")
	case MT_SIGNATURE:
		return ("<SIGNATURE>")
	case MT_UNKNOWN:
		return ("<UNKNOWN>")
	default:
		return ("Undifined message type")
	}
}

func (c *Chronology) recv(sender p2p.Messenger, peerID string, m *p2p.Message) {
	//c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Peer with ID = %s got a message from peer with ID = %s which traversed %d peers\n", sender.P2pNode.ID().Pretty(), peerID, len(m.Peers)))
	m.AddHop(sender.ID().Pretty())
	c.ChRcvMsg <- m.Payload
	//sender.BroadcastMessage(m, m.Peers)
	sender.BroadcastMessage(m, []string{})
}

func (c *Chronology) DecodeMessage(rcvMsg *[]byte) (MessageType, interface{}) {
	if ok, msgBlock := c.IsBlockInMessage(rcvMsg); ok {
		//c.Log(fmt.Sprintf(c.GetFormatedCurrentTime()+"Got a message with %s for block with Signature = %s and Nonce = %d and Hash = %s\n", msgBlock.MetaData, msgBlock.Signature, msgBlock.Nonce, msgBlock.Hash))
		if strings.Contains(msgBlock.MetaData, c.GetMessageTypeName(MT_BLOCK)) {
			return MT_BLOCK, msgBlock
		}

		if strings.Contains(msgBlock.MetaData, c.GetMessageTypeName(MT_COMITMENT_HASH)) {
			return MT_COMITMENT_HASH, msgBlock
		}

		if strings.Contains(msgBlock.MetaData, c.GetMessageTypeName(MT_BITMAP)) {
			return MT_BITMAP, msgBlock
		}

		if strings.Contains(msgBlock.MetaData, c.GetMessageTypeName(MT_COMITMENT)) {
			return MT_COMITMENT, msgBlock
		}

		if strings.Contains(msgBlock.MetaData, c.GetMessageTypeName(MT_SIGNATURE)) {
			return MT_SIGNATURE, msgBlock
		}
	}

	return MT_UNKNOWN, nil
}

func (c *Chronology) IsBlockInMessage(rcvMsg *[]byte) (bool, *block.Block) {
	var msgBlock block.Block

	json := marshal.JsonMarshalizer{}
	err := json.Unmarshal(&msgBlock, *rcvMsg)

	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return false, nil
	}

	return true, &msgBlock
}

func (c *Chronology) FormatTime(time time.Time) string {
	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ", time.Year(), time.Month(), time.Day(), time.Hour(), time.Minute(), time.Second(), time.Nanosecond())
	return str
}

func (c *Chronology) GetCurrentTime() time.Time {
	return time.Now().Add(c.ClockOffset)
}

func (c *Chronology) GetFormatedCurrentTime() string {
	return c.FormatTime(c.GetCurrentTime())
}

func (c *Chronology) Log(message string) {
	if c.DoLog {
		fmt.Printf(message + "\n")
	}
}
