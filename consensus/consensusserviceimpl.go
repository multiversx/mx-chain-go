package consensus

import (
	"errors"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/sync"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/validators"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	_ "github.com/davecgh/go-spew/spew"
	"strings"
	"time"
)

type ConsensusServiceImpl struct {
	Node  string
	Nodes []string

	Block      block.Block
	BlockChain blockchain.BlockChain

	Round         round.Round
	GenesisTime   time.Time
	RoundDuration time.Duration
	RoundDivision []time.Duration

	SelfRoundState round.RoundState

	DoRun         bool
	DoLog         bool
	Validators    map[string]RoundStateValidation
	PBFTThreshold int

	P2PNode *p2p.Node

	ChRcvMsg chan []byte

	SyncTime    sync.SyncTime
	ClockOffset time.Duration
}

type RoundStateValidation struct {
	Block         bool
	ComitmentHash bool
	Bitmap        bool
	Comitment     bool
	Signature     bool
}

func NewConsensusServiceImpl(p2pNode *p2p.Node, v *validators.Validators, genesisRoundTimeStamp time.Time, roundDuration time.Duration) *ConsensusServiceImpl {
	csi := ConsensusServiceImpl{}
	rs := chronology.GetRounderService()

	csi.DoRun = true
	csi.DoLog = true

	csi.SyncTime = sync.New()
	csi.ChRcvMsg = make(chan []byte)

	csi.P2PNode = p2pNode
	csi.P2PNode.OnMsgRecv = csi.recv

	csi.Node = v.GetSelf()
	csi.Nodes = v.GetConsensusGroup()

	csi.PBFTThreshold = len(csi.Nodes)*2/3 + 1

	csi.Validators = make(map[string]RoundStateValidation)
	csi.ResetValidators()

	csi.ResetBlock()

	csi.BlockChain = blockchain.New(nil)

	csi.GenesisTime = genesisRoundTimeStamp
	csi.RoundDuration = roundDuration
	csi.RoundDivision = chronology.GetRounderService().CreateRoundTimeDivision(roundDuration)

	csi.ClockOffset = csi.SyncTime.GetClockOffset()
	csi.SelfRoundState = round.RS_UNKNOWN

	csi.Round = rs.CreateRoundFromDateTime(csi.GenesisTime, csi.GetCurrentTime(), csi.RoundDuration, csi.RoundDivision)

	return &csi
}

func (c *ConsensusServiceImpl) StartRounds() {

	rs := chronology.GetRounderService()

	c.ClockOffset = c.SyncTime.GetClockOffset()
	c.SelfRoundState = round.RS_START_ROUND

	c.Round = rs.CreateRoundFromDateTime(c.GenesisTime, c.GetCurrentTime(), c.RoundDuration, c.RoundDivision)

	for c.DoRun {

		time.Sleep(5 * time.Millisecond)

		oldRoundIndex := c.Round.GetIndex()
		oldRoundState := c.Round.GetRoundState()

		rs.UpdateRoundFromDateTime(c.GenesisTime, c.GetCurrentTime(), &c.Round)

		if oldRoundIndex != c.Round.GetIndex() {
			c.Log(fmt.Sprintf("\n"+FormatTime(c.GetCurrentTime())+"################################################## ROUND %d BEGINS ##################################################\n", c.Round.GetIndex()))
			c.InitRound()
		}

		if oldRoundState != c.Round.GetRoundState() {
			c.Log(fmt.Sprintf("\n" + FormatTime(c.GetCurrentTime()) + ".................... SUBROUND " + rs.GetRoundStateName(c.Round.GetRoundState()) + " BEGINS ....................\n"))
		}

		timeRoundState := c.Round.GetRoundState()

		c.OptimizeRoundState(timeRoundState)

		switch timeRoundState {
		case round.RS_START_ROUND:
			if c.SelfRoundState == timeRoundState {
				if c.DoStartRound() {
					c.SelfRoundState = round.RS_BLOCK
				}
			}
		case round.RS_BLOCK:
			if c.SelfRoundState == timeRoundState {
				if c.DoBlock() {
					c.SelfRoundState = round.RS_COMITMENT_HASH
				}
			}
		case round.RS_COMITMENT_HASH:
			if c.SelfRoundState == timeRoundState {
				if c.DoComitmentHash() {
					c.SelfRoundState = round.RS_BITMAP
				}
			}
		case round.RS_BITMAP:
			if c.SelfRoundState == timeRoundState {
				if c.DoBitmap() {
					c.SelfRoundState = round.RS_COMITMENT
				}
			}
		case round.RS_COMITMENT:
			if c.SelfRoundState == timeRoundState {
				if c.DoComitment() {
					c.SelfRoundState = round.RS_SIGNATURE
				}
			}
		case round.RS_SIGNATURE:
			if c.SelfRoundState == timeRoundState {
				if c.DoSignature() {
					c.SelfRoundState = round.RS_END_ROUND
				}
			}
		case round.RS_END_ROUND:
			if c.SelfRoundState == timeRoundState {
				if c.DoEndRound() {
					c.SelfRoundState = round.RS_START_ROUND
				}
			}
		default:
		}
	}

	close(c.ChRcvMsg)
}

/*
Check node round state vs. time round state and decide if node state could be changed analyzing his tasks which have been done in current round
*/
func (c *ConsensusServiceImpl) OptimizeRoundState(timeRoundState round.RoundState) {

	//rs := chronology.GetRounderService()

	switch timeRoundState {
	case round.RS_BLOCK:
		if c.SelfRoundState == round.RS_START_ROUND {
			c.SelfRoundState = round.RS_BLOCK
		}
	default:
	}

	if timeRoundState != c.SelfRoundState {
		select {
		case rcvMsg := <-c.ChRcvMsg:
			if c.SelfRoundState != round.RS_ABORDED {
				if c.ConsumeReceivedMessage(&rcvMsg, timeRoundState) {
					//c.Log(fmt.Sprintf("\n" + FormatTime(c.SyncTime.GetCurrentTime())+"Received message in round time transition state: %s -> %s", rs.GetRoundStateName(timeRoundState), rs.GetRoundStateName(c.SelfRoundState)))
				}
			}
		default:
		}
	}
}

func (c *ConsensusServiceImpl) DoStartRound() bool {

	rs := chronology.GetRounderService()

	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(5 * time.Millisecond)

		c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 0: Preparing for this round"))
		return true
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 0: Aborded round %d in subround %s", c.Round.GetIndex(), rs.GetRoundStateName(round.RS_START_ROUND)))
	return false
}

func (c *ConsensusServiceImpl) DoBlock() bool {

	rs := chronology.GetRounderService()

	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(5 * time.Millisecond)

		if c.SendMessage(round.RS_BLOCK) {
			return true
		}

		if rs.GetRoundStateFromDateTime(&c.Round, c.GetCurrentTime()) != round.RS_BLOCK {
			c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 1: Extended the " + rs.GetRoundStateName(round.RS_BLOCK) + " subround"))
			return true // Try to give a chance to this round if the block from leader will arrive later
		}

		select {
		case rcvMsg := <-c.ChRcvMsg:
			if c.ConsumeReceivedMessage(&rcvMsg, round.RS_BLOCK) {
				c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 1: Synchronized block"))
				return true
			}
		default:
		}
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 1: Aborded round %d in subround %s", c.Round.GetIndex(), rs.GetRoundStateName(round.RS_BLOCK)))
	return false
}

func (c *ConsensusServiceImpl) DoComitmentHash() bool {

	rs := chronology.GetRounderService()

	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(5 * time.Millisecond)

		c.SendMessage(round.RS_COMITMENT_HASH)

		if rs.GetRoundStateFromDateTime(&c.Round, c.GetCurrentTime()) != round.RS_COMITMENT_HASH {
			c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 2: Extended the " + rs.GetRoundStateName(round.RS_COMITMENT_HASH) + " subround"))
			return true // Try to give a chance to this round if the necesary comitment hashes will arrive later
		}

		select {
		case rcvMsg := <-c.ChRcvMsg:
			consensusType := CT_PBFT

			if !c.IsNodeLeaderInCurrentRound(c.Node) {
				consensusType = CT_FULL
			}

			if c.ConsumeReceivedMessage(&rcvMsg, round.RS_COMITMENT_HASH) {
				if ok, n := c.CheckConsensus(round.RS_COMITMENT_HASH, consensusType, true); ok {
					if n == len(c.Nodes) {
						c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 2: Received all (%d from %d) comitment hashes", n, len(c.Nodes)))
					} else {
						c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 2: Received %d from %d comitment hashes, which are enough", n, len(c.Nodes)))
					}
					return true
				}
			}
		default:
		}
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 2: Aborded round %d in subround %s", c.Round.GetIndex(), rs.GetRoundStateName(round.RS_COMITMENT_HASH)))
	return false
}

func (c *ConsensusServiceImpl) DoBitmap() bool {

	rs := chronology.GetRounderService()

	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(5 * time.Millisecond)

		if c.SendMessage(round.RS_BITMAP) {
			return true
		}

		if rs.GetRoundStateFromDateTime(&c.Round, c.GetCurrentTime()) != round.RS_BITMAP {
			c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 3: Extended the " + rs.GetRoundStateName(round.RS_BITMAP) + " subround"))
			return true // Try to give a chance to this round if the bitmap from leader will arrive later
		}

		select {
		case rcvMsg := <-c.ChRcvMsg:
			if c.ConsumeReceivedMessage(&rcvMsg, round.RS_BITMAP) {
				if ok, n := c.CheckConsensus(round.RS_BITMAP, CT_PBFT, true); ok {
					addMessage := "BUT I WAS NOT selected in this bitmap"
					if c.IsNodeInBitmapGroup(c.Node) {
						addMessage = "AND I WAS selected in this bitmap"
					}
					c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 3: Received bitmap from leader, matching with my own, and it got %d from %d comitment hashes, which are enough, %s", n, len(c.Nodes), addMessage))
					return true
				}
			}
		default:
		}
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 3: Aborded round %d in subround %s", c.Round.GetIndex(), rs.GetRoundStateName(round.RS_BITMAP)))
	return false
}

func (c *ConsensusServiceImpl) DoComitment() bool {

	rs := chronology.GetRounderService()

	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(5 * time.Millisecond)

		c.SendMessage(round.RS_COMITMENT)

		if rs.GetRoundStateFromDateTime(&c.Round, c.GetCurrentTime()) != round.RS_COMITMENT {
			c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 4: Extended the " + rs.GetRoundStateName(round.RS_COMITMENT) + " subround"))
			return true // Try to give a chance to this round if the necesary comitments will arrive later
		}

		select {
		case rcvMsg := <-c.ChRcvMsg:
			if c.ConsumeReceivedMessage(&rcvMsg, round.RS_COMITMENT) {
				if ok, n := c.CheckConsensus(round.RS_COMITMENT, CT_PBFT, true); ok {
					c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 4: Received %d from %d comitments, which are matching with bitmap and are enough", n, len(c.Nodes)))
					return true
				}
			}
		default:
		}
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 4: Aborded round %d in subround %s", c.Round.GetIndex(), rs.GetRoundStateName(round.RS_COMITMENT)))
	return false
}

func (c *ConsensusServiceImpl) DoSignature() bool {

	rs := chronology.GetRounderService()

	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(5 * time.Millisecond)

		c.SendMessage(round.RS_SIGNATURE)

		if rs.GetRoundStateFromDateTime(&c.Round, c.GetCurrentTime()) != round.RS_SIGNATURE {
			c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 5: Extended the " + rs.GetRoundStateName(round.RS_SIGNATURE) + " subround"))
			return true // Try to give a chance to this round if the necesary signatures will arrive later
		}

		select {
		case rcvMsg := <-c.ChRcvMsg:
			if c.ConsumeReceivedMessage(&rcvMsg, round.RS_SIGNATURE) {
				if ok, n := c.CheckConsensus(round.RS_SIGNATURE, CT_PBFT, true); ok {
					c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 5: Received %d from %d signatures, which are matching with bitmap and are enough", n, len(c.Nodes)))
					return true
				}
			}
		default:
		}
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Step 5: Aborded round %d in subround %s", c.Round.GetIndex(), rs.GetRoundStateName(round.RS_SIGNATURE)))
	return false
}

func (c *ConsensusServiceImpl) DoEndRound() bool {

	bcs := data.GetBlockChainerService()
	rs := chronology.GetRounderService()

	for c.SelfRoundState != round.RS_ABORDED {
		time.Sleep(5 * time.Millisecond)

		if ok, _ := c.CheckConsensus(round.RS_SIGNATURE, CT_PBFT, true); ok {

			bcs.AddBlock(&c.BlockChain, c.Block)

			if c.IsNodeLeaderInCurrentRound(c.Node) {
				c.Log(fmt.Sprintf("\n"+FormatTime(c.GetCurrentTime())+">>>>>>>>>>>>>>>>>>>> ADDED PROPOSED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n", c.Block.GetNonce()))
			} else {
				c.Log(fmt.Sprintf("\n"+FormatTime(c.GetCurrentTime())+">>>>>>>>>>>>>>>>>>>> ADDED SYNCHRONIZED BLOCK WITH NONCE  %d  IN BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n", c.Block.GetNonce()))
			}

			return true
		}

		if rs.GetRoundStateFromDateTime(&c.Round, c.GetCurrentTime()) != round.RS_END_ROUND {
			c.Log(fmt.Sprintf("\n" + FormatTime(c.GetCurrentTime()) + ">>>>>>>>>>>>>>>>>>>> THIS ROUND NO BLOCK WAS ADDED TO THE BLOCKCHAIN <<<<<<<<<<<<<<<<<<<<\n"))
			return true
		}

		select {
		case rcvMsg := <-c.ChRcvMsg:
			c.ConsumeReceivedMessage(&rcvMsg, round.RS_END_ROUND)
		default:
		}
	}

	return false
}

func (c *ConsensusServiceImpl) SendMessage(roundState round.RoundState) bool {

	for {
		switch roundState {

		case round.RS_BLOCK:
			if !c.IsNodeLeaderInCurrentRound(c.Node) {
				return false
			}

			if c.Validators[c.Node].Block {
				return false
			}

			return c.SendBlock()
		case round.RS_COMITMENT_HASH:
			if c.Validators[c.Node].ComitmentHash {
				return false
			}

			if !c.IsBlockDirty(1) {
				roundState = round.RS_BLOCK
				continue
			}

			return c.SendComitmentHash()
		case round.RS_BITMAP:
			if !c.IsNodeLeaderInCurrentRound(c.Node) {
				return false
			}

			if c.Validators[c.Node].Bitmap {
				return false
			}

			if !c.IsComitmentHashDirty(c.PBFTThreshold) {
				roundState = round.RS_COMITMENT_HASH
				continue
			}

			return c.SendBitmap()
		case round.RS_COMITMENT:
			if !c.IsNodeInBitmapGroup(c.Node) {
				return false
			}

			if c.Validators[c.Node].Comitment {
				return false
			}

			if ok, _ := c.IsBitmapInComitmentHash(c.PBFTThreshold); !ok {
				roundState = round.RS_BITMAP
				continue
			}

			return c.SendComitment()
		case round.RS_SIGNATURE:
			if !c.IsNodeInBitmapGroup(c.Node) {
				return false
			}

			if c.Validators[c.Node].Signature {
				return false
			}

			if ok, _ := c.IsBitmapInComitment(c.PBFTThreshold); !ok {
				roundState = round.RS_COMITMENT
				continue
			}

			return c.SendSignature()
		default:
		}

		break
	}

	return false
}

func (c *ConsensusServiceImpl) SendBlock() bool {

	bs := data.GetBlockerService()
	bcs := data.GetBlockChainerService()

	currentBlock := bcs.GetCurrentBlock(&c.BlockChain)

	if currentBlock == nil {
		c.Block = block.New(0, c.GetCurrentTime().String(), c.Node, "", "", c.GetMessageTypeName(MT_BLOCK))
	} else {
		c.Block = block.New(currentBlock.GetNonce()+1, c.GetCurrentTime().String(), c.Node, "", currentBlock.GetHash(), c.GetMessageTypeName(MT_BLOCK))
	}

	c.Block.Hash = bs.CalculateHash(&c.Block)

	json := marshal.JsonMarshalizer{}
	message, err := json.Marshal(c.Block)

	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return false
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 1: Sending block"))

	rsv := c.Validators[c.Node]
	rsv.Block = true
	c.Validators[c.Node] = rsv

	c.P2PNode.BroadcastString(string(message), []string{})
	return true
}

func (c *ConsensusServiceImpl) SendComitmentHash() bool {

	if c.Block.Nonce == -1 {
		return false
	}

	block := c.Block

	block.Signature = c.Node
	block.MetaData = c.GetMessageTypeName(MT_COMITMENT_HASH)

	json := marshal.JsonMarshalizer{}
	message, err := json.Marshal(block)

	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return false
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 2: Sending comitment hash"))

	rsv := c.Validators[c.Node]
	rsv.ComitmentHash = true
	c.Validators[c.Node] = rsv

	c.P2PNode.BroadcastString(string(message), []string{})
	return true
}

func (c *ConsensusServiceImpl) SendBitmap() bool {

	if c.Block.Nonce == -1 {
		return false
	}

	block := c.Block

	block.Signature = c.Node
	block.MetaData = c.GetMessageTypeName(MT_BITMAP)

	for i := 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].ComitmentHash {
			block.MetaData += "," + c.Nodes[i]
		}
	}

	json := marshal.JsonMarshalizer{}
	message, err := json.Marshal(block)

	if err != nil {
		fmt.Printf(err.Error() + "\n")
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

	c.P2PNode.BroadcastString(string(message), []string{})
	return true
}

func (c *ConsensusServiceImpl) SendComitment() bool {

	if c.Block.Nonce == -1 {
		return false
	}

	block := c.Block

	block.Signature = c.Node
	block.MetaData = c.GetMessageTypeName(MT_COMITMENT)

	json := marshal.JsonMarshalizer{}
	message, err := json.Marshal(block)

	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return false
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 4: Sending comitment"))

	rsv := c.Validators[c.Node]
	rsv.Comitment = true
	c.Validators[c.Node] = rsv

	c.P2PNode.BroadcastString(string(message), []string{})
	return true
}

func (c *ConsensusServiceImpl) SendSignature() bool {

	if c.Block.Nonce == -1 {
		return false
	}

	block := c.Block

	block.Signature = c.Node
	block.MetaData = c.GetMessageTypeName(MT_SIGNATURE)

	json := marshal.JsonMarshalizer{}
	message, err := json.Marshal(block)

	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return false
	}

	c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Step 5: Sending signature"))

	rsv := c.Validators[c.Node]
	rsv.Signature = true
	c.Validators[c.Node] = rsv

	c.P2PNode.BroadcastString(string(message), []string{})
	return true
}

func (c *ConsensusServiceImpl) CheckIfBlockIsValid(receivedBlock *block.Block) bool {

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

func (c *ConsensusServiceImpl) ComputeLeader(nodes []string, round *round.Round) (string, error) {

	if round == nil {
		return "", errors.New("Round is null")
	}

	if nodes == nil {
		return "", errors.New("List of nodes is null")
	}

	if len(nodes) == 0 {
		return "", errors.New("List of nodes is empty")
	}

	index := round.GetIndex() % int64(len(nodes))
	return nodes[index], nil
}

func (c *ConsensusServiceImpl) IsNodeLeader(node string, nodes []string, round *round.Round) (bool, error) {

	v, err := c.ComputeLeader(nodes, round)

	if err != nil {
		fmt.Println(err)
		return false, err
	}

	return v == node, nil
}

func (c *ConsensusServiceImpl) IsNodeLeaderInCurrentRound(node string) bool {

	leader, err := c.GetLeader()

	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return false
	}

	return leader == node
}

func (c *ConsensusServiceImpl) IsNodeInValidationGroup(node string) bool {

	for i := 0; i < len(c.Nodes); i++ {
		if c.Nodes[i] == node {
			return true
		}
	}

	return false
}

func (c *ConsensusServiceImpl) IsNodeInBitmapGroup(node string) bool {
	return c.Validators[node].Bitmap
}

func (c *ConsensusServiceImpl) IsBlockDirty(threshold int) bool {
	for n, i := 0, 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].Block {
			n++
			if n >= threshold {
				return true
			}
		}
	}

	return false
}

func (c *ConsensusServiceImpl) IsComitmentHashDirty(threshold int) bool {
	for n, i := 0, 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].ComitmentHash {
			n++
			if n >= threshold {
				return true
			}
		}
	}

	return false
}

func (c *ConsensusServiceImpl) IsBitmapDirty(threshold int) bool {
	for n, i := 0, 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].Bitmap {
			n++
			if n >= threshold {
				return true
			}
		}
	}

	return false
}

func (c *ConsensusServiceImpl) IsComitmentDirty(threshold int) bool {
	for n, i := 0, 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].Comitment {
			n++
			if n >= threshold {
				return true
			}
		}
	}

	return false
}

func (c *ConsensusServiceImpl) IsSignatureDirty(threshold int) bool {
	for n, i := 0, 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].Signature {
			n++
			if n >= threshold {
				return true
			}
		}
	}

	return false
}

func (c *ConsensusServiceImpl) IsBlock(threshold int) (bool, int) {

	n := 0

	for i := 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].Block {
			n++
		}
	}

	return n >= threshold, n
}

func (c *ConsensusServiceImpl) IsComitmentHash(threshold int) (bool, int) {

	n := 0

	for i := 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].ComitmentHash {
			n++
		}
	}

	return n >= threshold, n
}

func (c *ConsensusServiceImpl) IsBitmapInComitmentHash(threshold int) (bool, int) {

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

func (c *ConsensusServiceImpl) IsBitmapInComitment(threshold int) (bool, int) {

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

func (c *ConsensusServiceImpl) IsBitmapInSignature(threshold int) (bool, int) {

	n := 0

	for i := 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].Bitmap {
			if !c.Validators[c.Nodes[i]].Signature {
				return false, n
			}
			n++
		}
	}

	return n >= threshold, n
}

func (c *ConsensusServiceImpl) GetLeader() (string, error) {

	if c.Round.GetIndex() == -1 {
		return "", errors.New("Round is not set")
	}

	if c.Nodes == nil {
		return "", errors.New("List of nodes is null")
	}

	if len(c.Nodes) == 0 {
		return "", errors.New("List of nodes is empty")
	}

	index := c.Round.GetIndex() % int64(len(c.Nodes))
	return c.Nodes[index], nil
}

func (c *ConsensusServiceImpl) CheckConsensus(roundState round.RoundState, consensusType ConsensusType, checkPreviousSubrounds bool) (bool, int) {

	var threshold, n int
	var ok bool
	var startRoundState round.RoundState

	if checkPreviousSubrounds {
		startRoundState = round.RS_BLOCK
	} else {
		startRoundState = roundState
	}

	switch consensusType {
	case CT_NONE:
		threshold = 0
	case CT_PBFT:
		threshold = c.PBFTThreshold
	case CT_FULL:
		threshold = len(c.Nodes)
	}

	for i := startRoundState; i <= roundState; i++ {
		switch i {
		case round.RS_BLOCK:
			ok, n = c.IsBlock(1)
		case round.RS_COMITMENT_HASH:
			ok, n = c.IsComitmentHash(threshold)
		case round.RS_BITMAP:
			ok, n = c.IsBitmapInComitmentHash(threshold)
		case round.RS_COMITMENT:
			ok, n = c.IsBitmapInComitment(threshold)
		case round.RS_SIGNATURE:
			ok, n = c.IsBitmapInSignature(threshold)
		default:
		}

		if !ok {
			return false, n
		}
	}

	return true, n
}

func (c *ConsensusServiceImpl) InitRound() {
	c.ClockOffset = c.SyncTime.GetClockOffset()
	c.SelfRoundState = round.RS_START_ROUND
	c.ResetValidators()
	c.ResetBlock()
}

func (c *ConsensusServiceImpl) ResetValidators() {
	for i := 0; i < len(c.Nodes); i++ {
		c.Validators[c.Nodes[i]] = RoundStateValidation{false, false, false, false, false}
	}
}

func (c *ConsensusServiceImpl) ResetBlock() {
	c.Block = block.New(-1, "", "", "", "", "")
}

func (c *ConsensusServiceImpl) Log(message string) {
	if c.DoLog {
		fmt.Printf(message + "\n")
	}
}

func FormatTime(time time.Time) string {

	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ", time.Year(), time.Month(), time.Day(), time.Hour(), time.Minute(), time.Second(), time.Nanosecond())
	return str
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

func (c *ConsensusServiceImpl) GetMessageTypeName(messageType MessageType) string {

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

func (c *ConsensusServiceImpl) recv(sender *p2p.Node, peerID string, m *p2p.Message) {
	//c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime())+"Peer with ID = %s got a message from peer with ID = %s which traversed %d peers\n", sender.P2pNode.ID().Pretty(), peerID, len(m.Peers)))
	m.AddHop(sender.P2pNode.ID().Pretty())
	c.ChRcvMsg <- m.Payload
	//sender.BroadcastMessage(m, m.Peers)
	sender.BroadcastMessage(m, []string{})
}

func (c *ConsensusServiceImpl) ConsumeReceivedMessage(rcvMsg *[]byte, timeRoundState round.RoundState) bool {

	//	rs := chronology.GetRounderService()

	msgType, msgData := c.DecodeMessage(rcvMsg)

	switch msgType {
	case MT_BLOCK:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.GetSignature()

		if timeRoundState > round.RS_BLOCK || c.SelfRoundState > round.RS_BLOCK {
			//			c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Received late " + c.GetMessageTypeName(MT_BLOCK) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		}

		if c.IsNodeLeaderInCurrentRound(node) {
			if !c.Validators[node].Block {
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

		if c.IsNodeLeaderInCurrentRound(c.Node) {
			if c.IsComitmentHashDirty(c.PBFTThreshold) {
				//				c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Avoid getting more comitment hashes, they are already enough"))
				return false
			}
		}

		if timeRoundState > round.RS_COMITMENT_HASH || c.SelfRoundState > round.RS_COMITMENT_HASH {
			//			c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Received late " + c.GetMessageTypeName(MT_COMITMENT_HASH) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		}

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
	case MT_BITMAP:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.GetSignature()

		if timeRoundState > round.RS_BITMAP || c.SelfRoundState > round.RS_BITMAP {
			//			c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Received late " + c.GetMessageTypeName(MT_BITMAP) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		}

		if c.IsNodeLeaderInCurrentRound(node) {
			if !c.IsBitmapDirty(c.PBFTThreshold) {
				nodes := strings.Split(rcvBlock.GetMetaData()[len(c.GetMessageTypeName(MT_BITMAP))+1:], ",")
				if len(nodes) < c.PBFTThreshold {
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
	case MT_COMITMENT:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.GetSignature()

		if timeRoundState > round.RS_COMITMENT || c.SelfRoundState > round.RS_COMITMENT {
			//			c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Received late " + c.GetMessageTypeName(MT_COMITMENT) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		}

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
	case MT_SIGNATURE:
		rcvBlock := msgData.(*block.Block)
		node := rcvBlock.GetSignature()

		if timeRoundState > round.RS_SIGNATURE || c.SelfRoundState > round.RS_SIGNATURE {
			//			c.Log(fmt.Sprintf(FormatTime(c.GetCurrentTime()) + "Received late " + c.GetMessageTypeName(MT_SIGNATURE) + " in time round state " + rs.GetRoundStateName(timeRoundState) + " and self round state " + rs.GetRoundStateName(c.SelfRoundState)))
		}

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
	default:
	}

	return false
}

func (c *ConsensusServiceImpl) DecodeMessage(rcvMsg *[]byte) (MessageType, interface{}) {

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

func (c *ConsensusServiceImpl) IsBlockInMessage(rcvMsg *[]byte) (bool, *block.Block) {

	var msgBlock block.Block

	json := marshal.JsonMarshalizer{}
	err := json.Unmarshal(&msgBlock, *rcvMsg)

	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return false, nil
	}

	return true, &msgBlock
}

func (c *ConsensusServiceImpl) GetCurrentTime() time.Time {
	return time.Now().Add(c.ClockOffset)
}
