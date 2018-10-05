package chronology

import (
	"errors"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/epoch"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/sync"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/validators"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/statistic"

	"strings"
	"time"
)

type Epocher interface {
	Print(epoch *epoch.Epoch)
}

type Rounder interface {
	CreateRoundFromDateTime(genesisRoundTimeStamp time.Time, timeStamp time.Time, roundTimeDuration time.Duration, roundTimeDivision []time.Duration) round.Round
	UpdateRoundFromDateTime(genesisRoundTimeStamp time.Time, timeStamp time.Time, round *round.Round)
	CreateRoundTimeDivision(time.Duration) []time.Duration
	GetRoundStateFromDateTime(round *round.Round, timeStamp time.Time) round.RoundState
	GetRoundStateName(roundState round.RoundState) string
}

const SLEEP_TIME = 5

type Chronology struct {
	Node  string
	Nodes []string

	Block      block.Block
	BlockChain blockchain.BlockChain

	Round         round.Round
	GenesisTime   time.Time
	RoundDuration time.Duration
	RoundDivision []time.Duration

	Subround Subround

	SelfRoundState round.RoundState

	DoRun              bool
	DoLog              bool
	DoSyncMode         bool
	Validators         map[string]RoundStateValidation
	PBFTThreshold      int
	ConsensusThreshold ConsensusThreshold

	P2PNode *p2p.Messenger

	ChRcvMsg chan []byte

	SyncTime    sync.SyncTime
	ClockOffset time.Duration

	ChronologyStatistic *statistic.ChronologyStatistic
}

type RoundStateValidation struct {
	Block         bool
	ComitmentHash bool
	Bitmap        bool
	Comitment     bool
	Signature     bool
}

func New(p2pNode *p2p.Messenger, v *validators.Validators, chronologyStatistic *statistic.ChronologyStatistic, genesisRoundTimeStamp time.Time, roundDuration time.Duration) *Chronology {
	csi := Chronology{}
	rs := GetRounderService()

	csi.DoRun = true

	csi.SyncTime = sync.New(roundDuration)
	csi.ChRcvMsg = make(chan []byte, len(v.GetConsensusGroup()))
	//csi.ChRcvMsg = make(chan []byte)

	csi.P2PNode = p2pNode
	(*csi.P2PNode).SetOnRecvMsg(csi.recv)

	csi.Node = v.GetSelf()
	csi.Nodes = v.GetConsensusGroup()

	csi.PBFTThreshold = len(csi.Nodes)*2/3 + 1

	csi.Validators = make(map[string]RoundStateValidation)

	csi.BlockChain = blockchain.New(nil)

	csi.InitRound()

	csi.GenesisTime = genesisRoundTimeStamp
	csi.RoundDuration = roundDuration
	csi.RoundDivision = GetRounderService().CreateRoundTimeDivision(roundDuration)
	csi.Round = rs.CreateRoundFromDateTime(csi.GenesisTime, csi.GetCurrentTime(), csi.RoundDuration, csi.RoundDivision)

	csi.ChronologyStatistic = chronologyStatistic

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
	rs := GetRounderService()

	oldRoundIndex := c.Round.GetIndex()
	oldRoundState := c.Round.GetRoundState()

	rs.UpdateRoundFromDateTime(c.GenesisTime, c.GetCurrentTime(), &c.Round)

	if oldRoundIndex != c.Round.GetIndex() {
		if c.DoLog { // only for statistic
			c.ChronologyStatistic.AddRound()
		}

		leader, err := c.GetLeader()
		if err != nil {
			leader = "Unknown"
		}

		if leader == c.Node {
			leader += " (MY TURN)"
		}

		c.Log(fmt.Sprintf("\n"+FormatTime(c.GetCurrentTime())+"############################## ROUND %d BEGINS WITH LEADER  %s  ##############################\n", c.Round.GetIndex(), leader))
		c.InitRound()
	}

	if oldRoundState != c.Round.GetRoundState() {
		c.Log(fmt.Sprintf("\n" + FormatTime(c.GetCurrentTime()) + ".................... SUBROUND " + rs.GetRoundStateName(c.Round.GetRoundState()) + " BEGINS ....................\n"))
	}

	roundState := c.SelfRoundState

	if c.DoSyncMode {
		roundState = c.Round.GetRoundState()
	}

	c.OptimizeRoundState(roundState)

	return c.Round.GetIndex(), roundState
}

//Check node round state vs. time round state and decide if node state could be changed analyzing his tasks which have been done in current round
func (c *Chronology) OptimizeRoundState(roundState round.RoundState) {
	//rs := chronology.GetRounderService()

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
				if c.ConsumeReceivedMessage(&rcvMsg, c.Round.GetRoundState()) {
					//c.Log(fmt.Sprintf("\n" + FormatTime(c.SyncTime.GetCurrentTime())+"Received message in time round state %s and self round state %s", rs.GetRoundStateName(c.Round.GetRoundState()), rs.GetRoundStateName(c.SelfRoundState)))
				}
			}
		default:
		}
	}
}

func (c *Chronology) DoStartRound() bool {
	rs := GetRounderService()

	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)

		c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 0: Preparing for this round"))
		return true
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 0: Aborded round %d in subround %s", c.Round.GetIndex(), rs.GetRoundStateName(round.RS_START_ROUND)))
	return false
}

func (c *Chronology) DoBlock() bool {
	rs := GetRounderService()

	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)

		bActionDone := c.SendMessage(round.RS_BLOCK)

		if bActionDone {
			bActionDone = false
			if ok, _ := c.CheckConsensus(round.RS_BLOCK, round.RS_BLOCK); ok {
				return true
			}
		}

		timeRoundState := rs.GetRoundStateFromDateTime(&c.Round, c.GetCurrentTime())

		if timeRoundState > round.RS_BLOCK {
			c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 1: Extended the " + rs.GetRoundStateName(round.RS_BLOCK) + " subround"))
			c.Subround.Block = SS_EXTENDED
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
				c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 1: Synchronized block"))
				return true
			}
		}
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 1: Aborded round %d in subround %s", c.Round.GetIndex(), rs.GetRoundStateName(round.RS_BLOCK)))
	return false
}

func (c *Chronology) DoComitmentHash() bool {
	rs := GetRounderService()

	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)

		bActionDone := c.SendMessage(round.RS_COMITMENT_HASH)

		timeRoundState := rs.GetRoundStateFromDateTime(&c.Round, c.GetCurrentTime())

		if timeRoundState > round.RS_COMITMENT_HASH {
			if c.GetComitmentHashes() < c.PBFTThreshold {
				c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 2: Extended the "+rs.GetRoundStateName(round.RS_COMITMENT_HASH)+" subround. Got only %d from %d commitment hashes which are not enough", c.GetComitmentHashes(), len(c.Nodes)))
			} else {
				c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 2: Extended the " + rs.GetRoundStateName(round.RS_COMITMENT_HASH) + " subround"))
			}
			c.Subround.ComitmentHash = SS_EXTENDED
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
				if n == len(c.Nodes) {
					c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 2: Received all (%d from %d) comitment hashes", n, len(c.Nodes)))
				} else {
					c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 2: Received %d from %d comitment hashes, which are enough", n, len(c.Nodes)))
				}
				return true
			}
		}
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 2: Aborded round %d in subround %s", c.Round.GetIndex(), rs.GetRoundStateName(round.RS_COMITMENT_HASH)))
	return false
}

func (c *Chronology) DoBitmap() bool {
	rs := GetRounderService()

	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)

		bActionDone := c.SendMessage(round.RS_BITMAP)

		if bActionDone {
			bActionDone = false
			if ok, _ := c.CheckConsensus(round.RS_BLOCK, round.RS_BITMAP); ok {
				return true
			}
		}

		timeRoundState := rs.GetRoundStateFromDateTime(&c.Round, c.GetCurrentTime())

		if timeRoundState > round.RS_BITMAP {
			c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 3: Extended the " + rs.GetRoundStateName(round.RS_BITMAP) + " subround"))
			c.Subround.Bitmap = SS_EXTENDED
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
				if c.IsNodeInBitmapGroup(c.Node) {
					addMessage = "AND I WAS selected in this bitmap"
				}
				c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 3: Received bitmap from leader, matching with my own, and it got %d from %d comitment hashes, which are enough, %s", n, len(c.Nodes), addMessage))
				return true
			}
		}
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 3: Aborded round %d in subround %s", c.Round.GetIndex(), rs.GetRoundStateName(round.RS_BITMAP)))
	return false
}

func (c *Chronology) DoComitment() bool {
	rs := GetRounderService()

	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)

		bActionDone := c.SendMessage(round.RS_COMITMENT)

		timeRoundState := rs.GetRoundStateFromDateTime(&c.Round, c.GetCurrentTime())

		if timeRoundState > round.RS_COMITMENT {
			c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 4: Extended the "+rs.GetRoundStateName(round.RS_COMITMENT)+" subround. Got only %d from %d commitments which are not enough", c.GetComitments(), len(c.Nodes)))
			c.Subround.Comitment = SS_EXTENDED
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
				c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 4: Received %d from %d comitments, which are matching with bitmap and are enough", n, len(c.Nodes)))
				return true
			}
		}
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 4: Aborded round %d in subround %s", c.Round.GetIndex(), rs.GetRoundStateName(round.RS_COMITMENT)))
	return false
}

func (c *Chronology) DoSignature() bool {
	rs := GetRounderService()

	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)

		bActionDone := c.SendMessage(round.RS_SIGNATURE)

		timeRoundState := rs.GetRoundStateFromDateTime(&c.Round, c.GetCurrentTime())

		if timeRoundState > round.RS_SIGNATURE {
			c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 5: Extended the "+rs.GetRoundStateName(round.RS_SIGNATURE)+" subround. Got only %d from %d sigantures which are not enough", c.GetSignatures(), len(c.Nodes)))
			c.Subround.Signature = SS_EXTENDED
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
				c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 5: Received %d from %d signatures, which are matching with bitmap and are enough", n, len(c.Nodes)))
				return true
			}
		}
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 5: Aborded round %d in subround %s", c.Round.GetIndex(), rs.GetRoundStateName(round.RS_SIGNATURE)))
	return false
}

func (c *Chronology) DoEndRound() bool {
	bcs := data.GetBlockChainerService()
	rs := GetRounderService()

	bActionDone := true

	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(SLEEP_TIME * time.Millisecond)

		timeRoundState := rs.GetRoundStateFromDateTime(&c.Round, c.GetCurrentTime())

		if timeRoundState > round.RS_END_ROUND {
			c.Log(fmt.Sprintf("\n" + FormatTime(c.GetCurrentTime()) + ">>>>>>>>>>>>>>>>>>>> THIS ROUND NO BLOCK WAS ADDED TO THE BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n"))
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
				if c.DoLog {
					c.ChronologyStatistic.AddRoundWithBlock()
				} // only for statistic

				bcs.AddBlock(&c.BlockChain, c.Block)

				if c.IsNodeLeaderInCurrentRound(c.Node) {
					c.Log(fmt.Sprintf("\n"+FormatTime(c.GetCurrentTime())+">>>>>>>>>>>>>>>>>>>> ADDED PROPOSED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n", c.Block.GetNonce()))
				} else {
					c.Log(fmt.Sprintf("\n"+FormatTime(c.GetCurrentTime())+">>>>>>>>>>>>>>>>>>>> ADDED SYNCHRONIZED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n", c.Block.GetNonce()))
				}

				return true
			}
		}
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 6: Aborded round %d in subround %s", c.Round.GetIndex(), rs.GetRoundStateName(round.RS_END_ROUND)))
	return false
}

func (c *Chronology) ConsumeReceivedMessage(rcvMsg *[]byte, timeRoundState round.RoundState) bool {
	//rs := GetRounderService()

	msgType, msgData := c.DecodeMessage(rcvMsg)

	switch msgType {
	case MT_BLOCK:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.GetSignature()

		//if timeRoundState > round.RS_BLOCK || c.SelfRoundState > round.RS_BLOCK {
		//	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Received late " + c.GetMessageTypeName(MT_BLOCK) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		//}

		if c.Subround.Block != SS_FINISHED {
			if c.IsNodeLeaderInCurrentRound(node) {
				if !c.CheckIfBlockIsValid(rcvBlock) {
					c.SelfRoundState = round.RS_ABORDED
					return false
				}

				rsv := c.Validators[node]
				rsv.Block = true
				c.Validators[node] = rsv
				c.Block = *rcvBlock
				return true
			}
		}
	case MT_COMITMENT_HASH:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.GetSignature()

		//if timeRoundState > round.RS_COMITMENT_HASH || c.SelfRoundState > round.RS_COMITMENT_HASH {
		//	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Received late " + c.GetMessageTypeName(MT_COMITMENT_HASH) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		//}

		if c.Subround.ComitmentHash != SS_FINISHED {
			if c.IsNodeInValidationGroup(node) {
				if !c.Validators[node].ComitmentHash {
					if rcvBlock.GetHash() == c.Block.GetHash() {
						rsv := c.Validators[node]
						rsv.ComitmentHash = true
						c.Validators[node] = rsv
						return true
					}
				}
			}
		}
	case MT_BITMAP:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.GetSignature()

		//if timeRoundState > round.RS_BITMAP || c.SelfRoundState > round.RS_BITMAP {
		//	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Received late " + c.GetMessageTypeName(MT_BITMAP) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		//}

		if c.Subround.Bitmap != SS_FINISHED {
			if c.IsNodeLeaderInCurrentRound(node) {
				if rcvBlock.GetHash() == c.Block.GetHash() {
					nodes := strings.Split(rcvBlock.GetMetaData()[len(c.GetMessageTypeName(MT_BITMAP))+1:], ",")
					if len(nodes) < c.ConsensusThreshold.Bitmap {
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
						rsv := c.Validators[nodes[i]]
						rsv.Bitmap = true
						c.Validators[nodes[i]] = rsv
					}

					return true
				}
			}
		}
	case MT_COMITMENT:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.GetSignature()

		//if timeRoundState > round.RS_COMITMENT || c.SelfRoundState > round.RS_COMITMENT {
		//	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Received late " + c.GetMessageTypeName(MT_COMITMENT) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		//}

		if c.Subround.Comitment != SS_FINISHED {
			if c.IsNodeInBitmapGroup(node) {
				if !c.Validators[node].Comitment {
					if rcvBlock.GetHash() == c.Block.GetHash() {
						rsv := c.Validators[node]
						rsv.Comitment = true
						c.Validators[node] = rsv
						return true
					}
				}
			}
		}
	case MT_SIGNATURE:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.GetSignature()

		//if timeRoundState > round.RS_SIGNATURE || c.SelfRoundState > round.RS_SIGNATURE {
		//	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Received late " + c.GetMessageTypeName(MT_SIGNATURE) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		//}

		if c.Subround.Signature != SS_FINISHED {
			if c.IsNodeInBitmapGroup(node) {
				if !c.Validators[node].Signature {
					if rcvBlock.GetHash() == c.Block.GetHash() {
						rsv := c.Validators[node]
						rsv.Signature = true
						c.Validators[node] = rsv
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
	rs := GetRounderService()

	var n int
	var ok bool

	for i := startRoundState; i <= endRoundState; i++ {
		switch i {
		case round.RS_BLOCK:
			if c.Subround.Block != SS_FINISHED {
				if ok, n = c.IsBlock(c.ConsensusThreshold.Block); ok {
					c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 1: Subround %s has been finished", rs.GetRoundStateName(round.RS_BLOCK)))
					c.Subround.Block = SS_FINISHED
				} else {
					return false, n
				}
			}
		case round.RS_COMITMENT_HASH:
			if c.Subround.ComitmentHash != SS_FINISHED {
				if ok, n = c.IsComitmentHash(c.ConsensusThreshold.ComitmentHash); ok {
					c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 2: Subround %s has been finished", rs.GetRoundStateName(round.RS_COMITMENT_HASH)))
					c.Subround.ComitmentHash = SS_FINISHED
				} else if ok, n = c.IsComitmentHashInBitmap(c.ConsensusThreshold.Bitmap); ok {
					c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 2: Subround %s has been finished", rs.GetRoundStateName(round.RS_COMITMENT_HASH)))
					c.Subround.ComitmentHash = SS_FINISHED
				} else {
					return false, n
				}
			}
		case round.RS_BITMAP:
			if c.Subround.Bitmap != SS_FINISHED {
				if ok, n = c.IsComitmentHashInBitmap(c.ConsensusThreshold.Bitmap); ok {
					c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 3: Subround %s has been finished", rs.GetRoundStateName(round.RS_BITMAP)))
					c.Subround.Bitmap = SS_FINISHED
				} else {
					return false, n
				}
			}
		case round.RS_COMITMENT:
			if c.Subround.Comitment != SS_FINISHED {
				if ok, n = c.IsBitmapInComitment(c.ConsensusThreshold.Comitment); ok {
					c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 4: Subround %s has been finished", rs.GetRoundStateName(round.RS_COMITMENT)))
					c.Subround.Comitment = SS_FINISHED
				} else {
					return false, n
				}
			}
		case round.RS_SIGNATURE:
			if c.Subround.Signature != SS_FINISHED {
				if ok, n = c.IsComitmentInSignature(c.ConsensusThreshold.Signature); ok {
					c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 5: Subround %s has been finished", rs.GetRoundStateName(round.RS_SIGNATURE)))
					c.Subround.Signature = SS_FINISHED
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
			if c.Subround.Block == SS_FINISHED {
				return false
			}

			if c.Validators[c.Node].Block {
				return false
			}

			if !c.IsNodeLeaderInCurrentRound(c.Node) {
				return false
			}

			return c.SendBlock()
		case round.RS_COMITMENT_HASH:
			if c.Subround.ComitmentHash == SS_FINISHED {
				return false
			}

			if c.Validators[c.Node].ComitmentHash {
				return false
			}

			if c.Subround.Block != SS_FINISHED {
				roundState = round.RS_BLOCK
				continue
			}

			return c.SendComitmentHash()
		case round.RS_BITMAP:
			if c.Subround.Bitmap == SS_FINISHED {
				return false
			}

			if c.Validators[c.Node].Bitmap {
				return false
			}

			if !c.IsNodeLeaderInCurrentRound(c.Node) {
				return false
			}

			if c.Subround.ComitmentHash != SS_FINISHED {
				roundState = round.RS_COMITMENT_HASH
				continue
			}

			return c.SendBitmap()
		case round.RS_COMITMENT:
			if c.Subround.Comitment == SS_FINISHED {
				return false
			}

			if c.Validators[c.Node].Comitment {
				return false
			}

			if c.Subround.Bitmap != SS_FINISHED {
				roundState = round.RS_BITMAP
				continue
			}

			if !c.IsNodeInBitmapGroup(c.Node) {
				return false
			}

			return c.SendComitment()
		case round.RS_SIGNATURE:
			if c.Subround.Signature == SS_FINISHED {
				return false
			}

			if c.Validators[c.Node].Signature {
				return false
			}

			if c.Subround.Comitment != SS_FINISHED {
				roundState = round.RS_COMITMENT
				continue
			}

			if !c.IsNodeInBitmapGroup(c.Node) {
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
	bs := data.GetBlockerService()
	bcs := data.GetBlockChainerService()

	currentBlock := bcs.GetCurrentBlock(&c.BlockChain)

	if currentBlock == nil {
		c.Block = block.New(0, c.GetCurrentTime().String(), c.Node, "", "", c.GetMessageTypeName(MT_BLOCK))
	} else {
		c.Block = block.New(currentBlock.GetNonce()+1, c.GetCurrentTime().String(), c.Node, "", currentBlock.GetHash(), c.GetMessageTypeName(MT_BLOCK))
	}

	c.Block.Hash = bs.CalculateHash(&c.Block)

	if !c.BroadcastBlock(&c.Block) {
		return false
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 1: Sending block"))

	rsv := c.Validators[c.Node]
	rsv.Block = true
	c.Validators[c.Node] = rsv

	return true
}

func (c *Chronology) SendComitmentHash() bool {
	block := c.Block

	block.Signature = c.Node
	block.MetaData = c.GetMessageTypeName(MT_COMITMENT_HASH)

	if !c.BroadcastBlock(&block) {
		return false
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 2: Sending comitment hash"))

	rsv := c.Validators[c.Node]
	rsv.ComitmentHash = true
	c.Validators[c.Node] = rsv

	return true
}

func (c *Chronology) SendBitmap() bool {
	block := c.Block

	block.Signature = c.Node
	block.MetaData = c.GetMessageTypeName(MT_BITMAP)

	for i := 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].ComitmentHash {
			block.MetaData += "," + c.Nodes[i]
		}
	}

	if !c.BroadcastBlock(&block) {
		return false
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 3: Sending bitmap"))

	for i := 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].ComitmentHash {
			rsv := c.Validators[c.Nodes[i]]
			rsv.Bitmap = true
			c.Validators[c.Nodes[i]] = rsv
		}
	}

	return true
}

func (c *Chronology) SendComitment() bool {
	block := c.Block

	block.Signature = c.Node
	block.MetaData = c.GetMessageTypeName(MT_COMITMENT)

	if !c.BroadcastBlock(&block) {
		return false
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 4: Sending comitment"))

	rsv := c.Validators[c.Node]
	rsv.Comitment = true
	c.Validators[c.Node] = rsv

	return true
}

func (c *Chronology) SendSignature() bool {
	block := c.Block

	block.Signature = c.Node
	block.MetaData = c.GetMessageTypeName(MT_SIGNATURE)

	if !c.BroadcastBlock(&block) {
		return false
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 5: Sending signature"))

	rsv := c.Validators[c.Node]
	rsv.Signature = true
	c.Validators[c.Node] = rsv

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

func (c *Chronology) CheckIfBlockIsValid(receivedBlock *block.Block) bool {
	bcs := data.GetBlockChainerService()

	currentBlock := bcs.GetCurrentBlock(&c.BlockChain)

	if currentBlock == nil {
		if receivedBlock.GetNonce() == 0 {
			if receivedBlock.PrevHash != "" {
				c.Log(fmt.Sprintf("Hash not match: local block hash is %s and node received block with previous hash %s", currentBlock.GetHash(), receivedBlock.GetHash()))
				return false
			}
		} else if receivedBlock.GetNonce() > 0 { // to resolve the situation when a node comes later in the network and it have not implemented the bootstrap mechanism (he will accept the first block received)
			c.Log(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d", -1, receivedBlock.GetNonce()))
			c.Log(fmt.Sprintf("\n"+FormatTime(c.GetCurrentTime())+">>>>>>>>>>>>>>>>>>>> ACCEPTED BLOCK WITH NONCE %d BECAUSE BOOSTRAP IS NOT IMPLEMENTED YET <<<<<<<<<<<<<<<<<<<<\n", receivedBlock.GetNonce()))
		}

		return true
	}

	if receivedBlock.GetNonce() < currentBlock.GetNonce()+1 {
		c.Log(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d", currentBlock.GetNonce(), receivedBlock.GetNonce()))
		return false

	} else if receivedBlock.GetNonce() == currentBlock.GetNonce()+1 {
		if receivedBlock.GetPrevHash() != currentBlock.GetHash() {
			c.Log(fmt.Sprintf("Hash not match: local block hash is %s and node received block with previous hash %s", currentBlock.GetHash(), receivedBlock.GetHash()))
			return false
		}
	} else if receivedBlock.GetNonce() > currentBlock.GetNonce()+1 { // to resolve the situation when a node misses some blocks and it have not implemented the bootstrap mechanism (he will accept the next block received)
		c.Log(fmt.Sprintf("Nonce not match: local block nonce is %d and node received block with nonce %d", currentBlock.GetNonce(), receivedBlock.GetNonce()))
		c.Log(fmt.Sprintf("\n"+FormatTime(c.GetCurrentTime())+">>>>>>>>>>>>>>>>>>>> ACCEPTED BLOCK WITH NONCE %d BECAUSE BOOSTRAP IS NOT IMPLEMENTED YET <<<<<<<<<<<<<<<<<<<<\n", receivedBlock.GetNonce()))
	}

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

	index := round.GetIndex() % len(nodes)
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
	for i := 0; i < len(c.Nodes); i++ {
		if c.Nodes[i] == node {
			return true
		}
	}

	return false
}

func (c *Chronology) IsNodeInBitmapGroup(node string) bool {
	return c.Validators[node].Bitmap
}

func (c *Chronology) IsBlock(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].Block {
			n++
		}
	}

	return n >= threshold, n
}

func (c *Chronology) IsComitmentHash(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].ComitmentHash {
			n++
		}
	}

	return n >= threshold, n
}

func (c *Chronology) IsComitmentHashInBitmap(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].Bitmap {
			if !c.Validators[c.Nodes[i]].ComitmentHash {
				return false, n
			}
			n++
		}
	}

	return n >= threshold, n
}

func (c *Chronology) IsBitmapInComitment(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].Bitmap {
			if !c.Validators[c.Nodes[i]].Comitment {
				return false, n
			}
			n++
		}
	}

	return n >= threshold, n
}

func (c *Chronology) IsComitmentInSignature(threshold int) (bool, int) {
	n := 0

	for i := 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].Comitment {
			if !c.Validators[c.Nodes[i]].Signature {
				return false, n
			}
			n++
		}
	}

	return n >= threshold, n
}

func (c *Chronology) GetComitmentHashes() int {
	n := 0

	for i := 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].ComitmentHash {
			n++
		}
	}

	return n
}

func (c *Chronology) GetComitments() int {
	n := 0

	for i := 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].Comitment {
			n++
		}
	}

	return n
}

func (c *Chronology) GetSignatures() int {
	n := 0

	for i := 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].Signature {
			n++
		}
	}

	return n
}

func (c *Chronology) GetLeader() (string, error) {
	if c.Round.GetIndex() == -1 {
		return "", errors.New("Round is not set")
	}

	if c.Nodes == nil {
		return "", errors.New("List of nodes is null")
	}

	if len(c.Nodes) == 0 {
		return "", errors.New("List of nodes is empty")
	}

	index := c.Round.GetIndex() % len(c.Nodes)
	return c.Nodes[index], nil
}

func (c *Chronology) InitRound() {
	c.ClockOffset = c.SyncTime.GetClockOffset()
	c.SelfRoundState = round.RS_START_ROUND
	c.ResetValidators()
	c.ResetBlock()
	c.ResetSubround()
	c.ResetConsensusThreshold()
}

func (c *Chronology) ResetValidators() {
	for i := 0; i < len(c.Nodes); i++ {
		c.Validators[c.Nodes[i]] = RoundStateValidation{false, false, false, false, false}
	}
}

func (c *Chronology) ResetBlock() {
	c.Block = block.New(-1, "", "", "", "", "")
}

func (c *Chronology) ResetSubround() {
	c.Subround = Subround{SS_NOTFINISHED, SS_NOTFINISHED, SS_NOTFINISHED, SS_NOTFINISHED, SS_NOTFINISHED}
}

func (c *Chronology) ResetConsensusThreshold() {
	if c.IsNodeLeaderInCurrentRound(c.Node) {
		c.ConsensusThreshold = ConsensusThreshold{1, c.PBFTThreshold, c.PBFTThreshold, c.PBFTThreshold, c.PBFTThreshold}
	} else {
		c.ConsensusThreshold = ConsensusThreshold{1, len(c.Nodes), c.PBFTThreshold, c.PBFTThreshold, c.PBFTThreshold}
	}
}

func (c *Chronology) Log(message string) {
	if c.DoLog {
		fmt.Printf(message + "\n")
	}
}

func FormatTime(time time.Time) string {
	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ", time.Year(), time.Month(), time.Day(), time.Hour(), time.Minute(), time.Second(), time.Nanosecond())
	return str
}

type ConsensusThreshold struct {
	Block         int
	ComitmentHash int
	Bitmap        int
	Comitment     int
	Signature     int
}

// A SubroundState specifies what kind of state could have a subround
type SubroundState int

const (
	SS_NOTFINISHED SubroundState = iota
	SS_EXTENDED
	SS_FINISHED
)

type Subround struct {
	Block         SubroundState
	ComitmentHash SubroundState
	Bitmap        SubroundState
	Comitment     SubroundState
	Signature     SubroundState
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
	//c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Peer with ID = %s got a message from peer with ID = %s which traversed %d peers\n", sender.P2pNode.ID().Pretty(), peerID, len(m.Peers)))
	m.AddHop(sender.ID().Pretty())
	c.ChRcvMsg <- m.Payload
	//sender.BroadcastMessage(m, m.Peers)
	sender.BroadcastMessage(m, []string{})
}

func (c *Chronology) DecodeMessage(rcvMsg *[]byte) (MessageType, interface{}) {
	if ok, msgBlock := c.IsBlockInMessage(rcvMsg); ok {
		//c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Got a message with %s for block with Signature = %s and Nonce = %d and Hash = %s\n", msgBlock.MetaData, msgBlock.Signature, msgBlock.Nonce, msgBlock.Hash))
		if strings.Contains(msgBlock.GetMetaData(), c.GetMessageTypeName(MT_BLOCK)) {
			return MT_BLOCK, msgBlock
		}

		if strings.Contains(msgBlock.GetMetaData(), c.GetMessageTypeName(MT_COMITMENT_HASH)) {
			return MT_COMITMENT_HASH, msgBlock
		}

		if strings.Contains(msgBlock.GetMetaData(), c.GetMessageTypeName(MT_BITMAP)) {
			return MT_BITMAP, msgBlock
		}

		if strings.Contains(msgBlock.GetMetaData(), c.GetMessageTypeName(MT_COMITMENT)) {
			return MT_COMITMENT, msgBlock
		}

		if strings.Contains(msgBlock.GetMetaData(), c.GetMessageTypeName(MT_SIGNATURE)) {
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

func (c *Chronology) GetCurrentTime() time.Time {
	return time.Now().Add(c.ClockOffset)
}
