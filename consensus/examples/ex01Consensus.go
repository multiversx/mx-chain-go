package main

import (
	"flag"
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/service"
	"github.com/davecgh/go-spew/spew"
	"strconv"
	"time"
)

var CGS *int
var RT *int
var NID *string

var Node string
var Nodes []string

type RoundStateValidation struct {
	ProposeBlock       bool
	ComitmentHash      bool
	Bitmap             bool
	Comitment          bool
	AggregateComitment bool
}

var Block data.Block
var BlockChain data.BlockChain
var Round *chronology.Round
var GenesisRoundTimeStamp time.Time = time.Date(2018, time.September, 18, 14, 0, 0, 0, time.UTC)
var RoundDuration time.Duration
var RoundDivision []time.Duration
var IsRunning = true

var Validators map[string]RoundStateValidation

func main() {

	// Parse options from the command line
	CGS = flag.Int("cgs", 4, "consensus group size")
	RT = flag.Int("rt", 4, "round time in seconds")
	NID = flag.String("nid", "1", "node id")
	flag.Parse()

	Node = *NID

	if Nodes == nil {
		Nodes = make([]string, *CGS)
	}

	for i := 0; i < *CGS; i++ {
		Nodes[i] = strconv.Itoa(i + 1)
	}

	if Validators == nil {
		Validators = make(map[string]RoundStateValidation)
	}

	ResetValidators()

	Block = data.NewBlock(-1, "", "", "", "", "")
	BlockChain = data.NewBlockChain(nil)

	RoundDuration = time.Duration(time.Duration(*RT) * time.Second)
	RoundDivision = service.GetChronologyService().CreateRoundTimeDivision(RoundDuration)

	ch := make(chan string)

	go StartRounds(ch)

	oldNonce := -1

	for v := range ch {
		spew.Println(v)

		currentBlock := service.GetBlockChainService().GetCurrentBlock(&BlockChain)

		if currentBlock == nil {
			continue
		}

		if currentBlock.GetNonce() > oldNonce {
			oldNonce = currentBlock.GetNonce()
			spew.Println("")
			spew.Dump(Block)
			spew.Println("")
		}
	}

	IsRunning = false
}

func StartRounds(ch chan string) {

	chr := service.GetChronologyService()

	Round = chr.CreateRoundFromDateTime(GenesisRoundTimeStamp, time.Now(), RoundDuration, RoundDivision)

	for IsRunning {

		time.Sleep(10 * time.Millisecond)

		now := time.Now()

		chr.UpdateRoundFromDateTime(GenesisRoundTimeStamp, now, Round)

		roundState := chr.GetRoundStateFromDateTime(Round, now)

		switch roundState {
		case chronology.RS_PROPOSE_BLOCK:
			if Round.GetRoundState() == roundState {
				if DoProposeBlock(ch) {
					Round.SetRoundState(chronology.RS_COMITMENT_HASH)
				}
			}
		case chronology.RS_COMITMENT_HASH:
			if Round.GetRoundState() == roundState {
				if DoComitmentHash(ch) {
					Round.SetRoundState(chronology.RS_BITMAP)
				}
			}
		case chronology.RS_BITMAP:
			if Round.GetRoundState() == roundState {
				if DoBitmap(ch) {
					Round.SetRoundState(chronology.RS_COMITMENT)
				}
			}
		case chronology.RS_COMITMENT:
			if Round.GetRoundState() == roundState {
				if DoComitment(ch) {
					Round.SetRoundState(chronology.RS_AGGREGATE_COMITMENT)
				}
			}
		case chronology.RS_AGGREGATE_COMITMENT:
			if Round.GetRoundState() == roundState {
				if DoAggregateComitment(ch) {
					Round.SetRoundState(chronology.RS_END_ROUND)
				}
			}
		default:
		}
	}

	close(ch)
}

func DoProposeBlock(ch chan string) bool {
	con := service.GetConsensusService()
	leader, err := con.ComputeLeader(Nodes, Round)

	if err != nil {
		ch <- err.Error()
		return false
	}

	if leader == Node {
		ProposeBlock()
		ch <- FormatTime(time.Now()) + "Step 1: It is my turn to propose block..."
	} else {
		ch <- FormatTime(time.Now()) + "Step 1: It is node with ID = " + leader + " turn to propose block..."
	}

	return true
}

func DoComitmentHash(ch chan string) bool {
	ch <- FormatTime(time.Now()) + "Step 2: Send comitment hash..."
	return true
}

func DoBitmap(ch chan string) bool {
	ch <- FormatTime(time.Now()) + "Step 3: Send bitmap..."
	return true
}

func DoComitment(ch chan string) bool {
	ch <- FormatTime(time.Now()) + "Step 4: Send comitment..."
	return true
}

func DoAggregateComitment(ch chan string) bool {

	ch <- FormatTime(time.Now()) + "Step 5: Send aggregate comitment..."

	bcs := service.GetBlockChainService()

	currentBlock := bcs.GetCurrentBlock(&BlockChain)

	if currentBlock == nil {
		if Block.GetNonce() >= 0 {
			ch <- "\nAdded genesis block in blockchain..."
			bcs.AddBlock(&BlockChain, Block)
		}
	} else {
		if currentBlock.GetNonce() < Block.GetNonce() {
			ch <- "\nAdded block in blockchain..."
			bcs.AddBlock(&BlockChain, Block)
		}
	}

	ch <- ""

	return true
}

func ProposeBlock() {
	currentBlock := service.GetBlockChainService().GetCurrentBlock(&BlockChain)

	if currentBlock == nil {
		Block = data.NewBlock(0, time.Now().String(), Node, "", "", "PROPOSE_BLOCK")
	} else {

		Block = data.NewBlock(currentBlock.GetNonce()+1, time.Now().String(), Node, "", currentBlock.GetHash(), "PROPOSE_BLOCK")
	}

	hash := service.GetBlockService().CalculateHash(&Block)
	Block.SetHash(hash)
}

func ResetValidators() {
	for i := 0; i < len(Nodes); i++ {
		Validators[Nodes[i]] = RoundStateValidation{false, false, false, false, false}
	}
}

func FormatTime(time time.Time) string {

	str := fmt.Sprintf("%.4d-%.2d-%.2d %.2d:%.2d:%.2d.%.9d ", time.Year(), time.Month(), time.Day(), time.Hour(), time.Minute(), time.Second(), time.Nanosecond())
	return str
}
