package chronology

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/epoch"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/round"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/davecgh/go-spew/spew"

	"testing"
	"time"
)

// ############################## Test Epoch subpackage ##############################

func TestEpochBehaviour(t *testing.T) {
	e := epoch.New(1, time.Now())
	spew.Dump(e)
}

// ############################## Test Round subpackage ##############################

func TestRoundBehaviour(t *testing.T) {
	genesisRoundTimeStamp := time.Date(2018, time.September, 18, 14, 0, 0, 0, time.UTC)

	duration := time.Duration(4 * time.Second)
	division := createRoundTimeDivision(duration)
	subround := round.Subround{round.SS_NOTFINISHED, round.SS_NOTFINISHED, round.SS_NOTFINISHED, round.SS_NOTFINISHED, round.SS_NOTFINISHED}

	r := round.NewRoundFromDateTime(genesisRoundTimeStamp, time.Now(), duration, division, subround)

	roundState := r.GetRoundStateFromDateTime(time.Now())

	if roundState < round.RS_START_ROUND || roundState > round.RS_END_ROUND {
		t.Fatal("Wrong round state")
	}

	spew.Dump(r)
	spew.Dump(roundState)
	spew.Dump(r.GetRoundStateName(roundState))
}

func TestRoundState(t *testing.T) {

	if round.RS_START_ROUND != 1 || round.RS_BLOCK != 2 || round.RS_END_ROUND != 7 {
		t.Fatal("Wrong values in round state enum")
	}
}

// ############################## Test Chronology ##############################

func TestComputeLeader(t *testing.T) {

	var chr Chronology

	genesisRoundTimeStamp := time.Date(2018, time.September, 18, 14, 0, 0, 0, time.UTC)

	duration := time.Duration(4 * time.Second)
	division := createRoundTimeDivision(duration)
	subround := round.Subround{round.SS_NOTFINISHED, round.SS_NOTFINISHED, round.SS_NOTFINISHED, round.SS_NOTFINISHED, round.SS_NOTFINISHED}

	round := round.NewRoundFromDateTime(genesisRoundTimeStamp, time.Now(), duration, division, subround)

	nodes := []string{"1", "2", "3"}

	node, err := chr.ComputeLeader(nodes, &round)

	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(node)
}

func TestNodeLeader(t *testing.T) {

	var chr Chronology

	genesisRoundTimeStamp := time.Date(2018, time.September, 18, 14, 0, 0, 0, time.UTC)

	duration := time.Duration(4 * time.Second)
	division := createRoundTimeDivision(duration)
	subround := round.Subround{round.SS_NOTFINISHED, round.SS_NOTFINISHED, round.SS_NOTFINISHED, round.SS_NOTFINISHED, round.SS_NOTFINISHED}

	round := round.NewRoundFromDateTime(genesisRoundTimeStamp, time.Now(), duration, division, subround)

	node := "3"
	nodes := []string{"1", "2", "3"}

	bIsLeader, err := chr.IsNodeLeader(node, nodes, &round)

	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(bIsLeader)
}

// other necessary functions for testing

func createRoundTimeDivision(duration time.Duration) []time.Duration {

	var d []time.Duration

	for i := round.RS_START_ROUND; i <= round.RS_END_ROUND; i++ {
		switch i {
		case round.RS_START_ROUND:
			d = append(d, time.Duration(5*duration/100))
		case round.RS_BLOCK:
			d = append(d, time.Duration(25*duration/100))
		case round.RS_COMITMENT_HASH:
			d = append(d, time.Duration(40*duration/100))
		case round.RS_BITMAP:
			d = append(d, time.Duration(55*duration/100))
		case round.RS_COMITMENT:
			d = append(d, time.Duration(70*duration/100))
		case round.RS_SIGNATURE:
			d = append(d, time.Duration(85*duration/100))
		case round.RS_END_ROUND:
			d = append(d, time.Duration(100*duration/100))
		}
	}

	return d
}

func TestA(t *testing.T) {
	c := New(&ChronologyIn{})

	c.OnNeedToBroadcastBlock = func(block *block.Block) bool {
		//put test logic here.

		return false
	}

	c.ReceivedBlock(nil)

	c.doBlockOneTime()

}
