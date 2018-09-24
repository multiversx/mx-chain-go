package consensus

import (
	"errors"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"time"
)

type RoundStateValidation struct {
	ProposeBlock       bool
	ComitmentHash      bool
	Bitmap             bool
	Comitment          bool
	AggregateComitment bool
}

type ConsensusServiceImpl struct {
	Node  string
	Nodes []string

	Block          data.Block
	ReceivedBlocks map[string]data.Block
	BlockChain     data.BlockChain

	Round                 *chronology.Round
	GenesisRoundTimeStamp time.Time
	RoundDuration         time.Duration
	RoundDivision         []time.Duration

	DoRunning bool

	Validators map[string]RoundStateValidation

	P2PNode *p2p.Node

	ChStdOut *chan string
	ChRcvMsg chan string
}

func (c *ConsensusServiceImpl) recv(sender *p2p.Node, peerID string, m *p2p.Message) {

	m.AddHop(sender.P2pNode.ID().Pretty())

	var receivedBlock data.Block

	json := marshal.JsonMarshalizer{}
	err := json.Unmarshal(&receivedBlock, m.Payload)

	if err != nil {
		*c.ChStdOut <- err.Error()
		return
	}

	c.ReceivedBlocks[peerID] = receivedBlock

	*c.ChStdOut <- fmt.Sprintf(FormatTime(time.Now())+"Peer with ID = %s got a message from peer with ID = %s, with %s for block with Nonce = %d and Hash = %s\n", sender.P2pNode.ID().Pretty(), peerID, c.ReceivedBlocks[peerID].MetaData, c.ReceivedBlocks[peerID].Nonce, c.ReceivedBlocks[peerID].Hash)

	c.DecodeMessage(peerID)

	c.ChRcvMsg <- peerID

	sender.BroadcastMessage(m, []string{peerID})
}

func NewConsensusServiceImpl(p2pNodes []*p2p.Node, index int, genesisRoundTimeStamp time.Time, roundDuration time.Duration, ch *chan string) ConsensusServiceImpl {
	csi := ConsensusServiceImpl{}

	csi.DoRunning = true

	csi.ChStdOut = ch
	csi.ChRcvMsg = make(chan string)

	csi.P2PNode = p2pNodes[index]
	csi.P2PNode.OnMsgRecv = csi.recv

	csi.Node = p2pNodes[index].P2pNode.ID().Pretty()
	csi.Nodes = make([]string, len(p2pNodes))

	for i := 0; i < len(p2pNodes); i++ {
		csi.Nodes[i] = p2pNodes[i].P2pNode.ID().Pretty()
	}

	csi.Validators = make(map[string]RoundStateValidation)
	csi.ResetValidators(true)

	csi.Block = data.NewBlock(-1, "", "", "", "", "")

	csi.ReceivedBlocks = make(map[string]data.Block)
	csi.ResetReceivedBlocks(true)

	csi.BlockChain = data.NewBlockChain(nil)

	csi.GenesisRoundTimeStamp = genesisRoundTimeStamp
	csi.RoundDuration = roundDuration
	csi.RoundDivision = chronology.ChronologyServiceImpl{}.CreateRoundTimeDivision(roundDuration)

	return csi
}

func (c *ConsensusServiceImpl) StartRounds() {

	chr := chronology.ChronologyServiceImpl{}

	c.Round = chr.CreateRoundFromDateTime(c.GenesisRoundTimeStamp, time.Now(), c.RoundDuration, c.RoundDivision)

	for c.DoRunning {

		time.Sleep(10 * time.Millisecond)

		oldRoundIndex := c.Round.GetIndex()

		now := time.Now()

		chr.UpdateRoundFromDateTime(c.GenesisRoundTimeStamp, now, c.Round)

		if oldRoundIndex != c.Round.GetIndex() {
			c.ResetValidators(false)
			c.ResetBlock()
			c.ResetReceivedBlocks(false)
		}

		roundState := chr.GetRoundStateFromDateTime(c.Round, now)

		switch roundState {
		case chronology.RS_PROPOSE_BLOCK:
			if c.Round.GetRoundState() == roundState {
				if c.DoProposeBlock() {
					c.Round.SetRoundState(chronology.RS_COMITMENT_HASH)
				}
			}
		case chronology.RS_COMITMENT_HASH:
			if c.Round.GetRoundState() == roundState {
				if c.DoComitmentHash() {
					c.Round.SetRoundState(chronology.RS_BITMAP)
				}
			}
		case chronology.RS_BITMAP:
			if c.Round.GetRoundState() == roundState {
				if c.DoBitmap() {
					c.Round.SetRoundState(chronology.RS_COMITMENT)
				}
			}
		case chronology.RS_COMITMENT:
			if c.Round.GetRoundState() == roundState {
				if c.DoComitment() {
					c.Round.SetRoundState(chronology.RS_AGGREGATE_COMITMENT)
				}
			}
		case chronology.RS_AGGREGATE_COMITMENT:
			if c.Round.GetRoundState() == roundState {
				if c.DoAggregateComitment() {
					c.Round.SetRoundState(chronology.RS_END_ROUND)
				}
			}
		default:
		}
	}

	close(c.ChRcvMsg)
	close(*c.ChStdOut)
}

func (c *ConsensusServiceImpl) DoProposeBlock() bool {

	chr := chronology.ChronologyServiceImpl{}

	leader, err := c.GetLeader()

	if err != nil {
		*c.ChStdOut <- err.Error()
		return false
	}

	if leader == c.Node {
		*c.ChStdOut <- FormatTime(time.Now()) + "Step 1: It is my turn to propose block..."
		return c.ProposeBlock()
	} else {
		*c.ChStdOut <- FormatTime(time.Now()) + "Step 1: It is node with ID = " + leader + " turn to propose block..."
		for {
			time.Sleep(10 * time.Millisecond)

			if chr.GetRoundStateFromDateTime(c.Round, time.Now()) != chronology.RS_PROPOSE_BLOCK {
				return false
			}

			select {
			case v := <-c.ChRcvMsg:
				if v == leader {
					if c.Validators[leader].ProposeBlock {
						receivedBlock := c.ReceivedBlocks[leader]
						if c.CheckIfBlockIsValid(&receivedBlock) {
							c.Block = receivedBlock
							return true
						}
					}
				}

			default:
			}
		}
	}
}

func (c *ConsensusServiceImpl) DoComitmentHash() bool {

	*c.ChStdOut <- FormatTime(time.Now()) + "Step 2: Send comitment hash..."
	return true
}

func (c *ConsensusServiceImpl) DoBitmap() bool {

	*c.ChStdOut <- FormatTime(time.Now()) + "Step 3: Send bitmap..."
	return true
}

func (c *ConsensusServiceImpl) DoComitment() bool {

	if c.Block.Nonce != -1 {

		c.Block.MetaData = "COMITMENT"
		c.Block.Hash = data.BlockServiceImpl{}.CalculateHash(&c.Block)

		json := marshal.JsonMarshalizer{}
		message, err := json.Marshal(c.Block)

		if err != nil {
			*c.ChStdOut <- "Error marshalized block"
			return false
		}

		*c.ChStdOut <- FormatTime(time.Now()) + "Step 4: Send comitment..."

		rsv := c.Validators[c.Node]
		rsv.Comitment = true
		c.Validators[c.Node] = rsv

		c.P2PNode.BroadcastString(string(message), []string{})
		return true
	}

	return false
}

func (c *ConsensusServiceImpl) DoAggregateComitment() bool {

	if c.CheckComitmentConsensus() {

		*c.ChStdOut <- FormatTime(time.Now()) + "Step 5: Send aggregate comitment..."
		data.BlockChainServiceImpl{}.AddBlock(&c.BlockChain, c.Block)

		if c.IsSelfLeaderInCurrentRound() {
			*c.ChStdOut <- "\n" + FormatTime(time.Now()) + "ADDED PROPOSED BLOCK IN BLOCKCHAIN\n"
		} else {
			*c.ChStdOut <- "\n" + FormatTime(time.Now()) + "ADDED SYNCHRONIZED BLOCK IN BLOCKCHAIN\n"
		}
	}

	return true
}

func (c *ConsensusServiceImpl) ProposeBlock() bool {

	currentBlock := data.BlockChainServiceImpl{}.GetCurrentBlock(&c.BlockChain)

	if currentBlock == nil {
		c.Block = data.NewBlock(0, time.Now().String(), c.Node, "", "", "PROPOSE_BLOCK")
	} else {
		c.Block = data.NewBlock(currentBlock.GetNonce()+1, time.Now().String(), c.Node, "", currentBlock.GetHash(), "PROPOSE_BLOCK")
	}

	c.Block.Hash = data.BlockServiceImpl{}.CalculateHash(&c.Block)

	json := marshal.JsonMarshalizer{}
	message, err := json.Marshal(c.Block)

	if err != nil {
		*c.ChStdOut <- "Error marshalized block"
		return false
	}

	c.P2PNode.BroadcastString(string(message), []string{})
	return true
}

func (c *ConsensusServiceImpl) ResetValidators(resetAll bool) bool {

	if resetAll {
		for i := 0; i < len(c.Nodes); i++ {
			c.Validators[c.Nodes[i]] = RoundStateValidation{false, false, false, false, false}
		}
	} else {
		leader, err := c.GetLeader()

		if err != nil {
			*c.ChStdOut <- err.Error()
			return false
		}

		c.Validators[leader] = RoundStateValidation{false, false, false, false, false}
	}

	return true
}

func (c *ConsensusServiceImpl) ResetBlock() {

	c.Block = data.NewBlock(-1, "", "", "", "", "")
}

func (c *ConsensusServiceImpl) ResetReceivedBlocks(resetAll bool) bool {

	if resetAll {
		for i := 0; i < len(c.Nodes); i++ {
			c.ReceivedBlocks[c.Nodes[i]] = data.NewBlock(-1, "", "", "", "", "")
		}
	} else {
		leader, err := c.GetLeader()

		if err != nil {
			*c.ChStdOut <- err.Error()
			return false
		}

		c.ReceivedBlocks[leader] = data.NewBlock(-1, "", "", "", "", "")
	}

	return true
}

func FormatTime(time time.Time) string {

	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ", time.Year(), time.Month(), time.Day(), time.Hour(), time.Minute(), time.Second(), time.Nanosecond())
	return str
}

func (c *ConsensusServiceImpl) DecodeMessage(peerID string) {

	switch c.ReceivedBlocks[peerID].MetaData {
	case "PROPOSE_BLOCK":
		rsv := c.Validators[peerID]
		rsv.ProposeBlock = true
		c.Validators[peerID] = rsv
	case "COMITMENT":
		rsv := c.Validators[peerID]
		rsv.Comitment = true
		c.Validators[peerID] = rsv
	default:
	}
}

func (c *ConsensusServiceImpl) CheckIfBlockIsValid(receivedBlock *data.Block) bool {

	currentBlock := data.BlockChainServiceImpl{}.GetCurrentBlock(&c.BlockChain)

	if currentBlock == nil {
		if receivedBlock.Nonce != 0 {
			*c.ChStdOut <- "Nonce not match"
			return false
		}

		if receivedBlock.PrevHash != "" {
			*c.ChStdOut <- "Hash not match"
			return false
		}

	} else {
		if receivedBlock.Nonce != currentBlock.Nonce+1 {
			*c.ChStdOut <- "Nonce not match"
			return false
		}

		if receivedBlock.PrevHash != currentBlock.Hash {
			*c.ChStdOut <- "Hash not match"
			return false
		}
	}

	return true
}

func (c *ConsensusServiceImpl) ComputeLeader(nodes []string, round *chronology.Round) (string, error) {

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

func (c *ConsensusServiceImpl) IsNodeLeader(node string, nodes []string, round *chronology.Round) (bool, error) {

	v, err := c.ComputeLeader(nodes, round)

	if err != nil {
		fmt.Println(err)
		return false, err
	}

	return v == node, nil
}

func (c *ConsensusServiceImpl) IsSelfLeaderInCurrentRound() bool {

	leader, err := c.GetLeader()

	if err != nil {
		*c.ChStdOut <- err.Error()
		return false
	}

	return leader == c.Node
}

func (c *ConsensusServiceImpl) GetLeader() (string, error) {

	if c.Round == nil {
		return "", errors.New("Round is null")
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

func (c *ConsensusServiceImpl) CheckComitmentConsensus() bool {

	n := 0

	for i := 0; i < len(c.Nodes); i++ {
		if c.Validators[c.Nodes[i]].Comitment {
			n++
		}
	}

	pbft := len(c.Nodes)*2/3 + 1

	return n >= pbft
}
